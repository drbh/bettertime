package main

import (
	"encoding/json"
	"fmt"
	"github.com/getlantern/systray"
	"io/ioutil"
	"os"
	"os/user"
	"time"

	"strings"

	"path/filepath"
)

type MsgCreateUserJson struct {
	Loc string `json:"loc"`
	Abr string `json:"abr"`
}

type TimeZoneNamingPair struct {
	a, b interface{}
}

const (
	// TIME_FORMAT = "15:04:05"
	TIME_FORMAT = "15:04"
)

var GLOBAL_LIST = []TimeZoneNamingPair{}

// we find where the binary is - then concat the location to the config
var dir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
var newDirec = strings.Split(dir, "/")
var final = strings.Join(newDirec[:len(newDirec)-2], "/") + "/Contents/Resources/config.json"
var CONFIG_FILE = final

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func checkUsersConfig() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}
	better_time_config := usr.HomeDir +
		"/Library/Preferences/better-time-config.json"

	if fileExists(better_time_config) {
		fmt.Println("Example file exists")
	} else {
		d1 := []byte(`[
{
    "loc": "America/New_York",
    "abr": "NYC"
},
{
    "loc": "America/Phoenix",
    "abr": "PHX"
},
{
    "loc": "America/Mexico_City",
    "abr": "MCX"
}
]
    `)
		err := ioutil.WriteFile(better_time_config, d1, 0644)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Example file does not exist (or is a directory)")
	}

	CONFIG_FILE = better_time_config

}

func watchFile(filePath string) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func reloadConfig() {

	file, err := ioutil.ReadFile(CONFIG_FILE)

	if err != nil {
		fmt.Println(err)
		// kill if config is missing
		os.Exit(0)
		// return
	}

	data := []MsgCreateUserJson{}
	err = json.Unmarshal([]byte(file), &data)

	if err != nil {
		fmt.Println(err)
		return
	}

	listTimeZones := []TimeZoneNamingPair{}
	for _, timeonTopBar := range data {
		loc, _ := time.LoadLocation(timeonTopBar.Loc)
		abr := timeonTopBar.Abr
		p := TimeZoneNamingPair{loc, abr}
		listTimeZones = append(listTimeZones, p)
	}
	GLOBAL_LIST = listTimeZones
	fmt.Println("updated - config")
	return
}

func onReady() {
	mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuitOrig.ClickedCh
		os.Exit(1)
	}()

	ticker := time.NewTicker(300 * time.Millisecond)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				ps := ""
				for _, tup := range GLOBAL_LIST {
					loc := tup.a.(*time.Location)
					tm := time.Now().In(loc).Format(TIME_FORMAT)
					_ps := fmt.Sprintf(
						"%-5s%-12s",
						tup.b, tm,
					)
					ps = ps + _ps
				}

				topbartext := ps
				systray.SetTitle(
					topbartext,
				)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func watcF() {
	defer func() {
		go watcF()
	}()
	err := watchFile(CONFIG_FILE)
	if err != nil {
		fmt.Println(err)
	}
	reloadConfig()
}

func main() {
	checkUsersConfig()

	fmt.Println(CONFIG_FILE)

	go watcF()
	reloadConfig()
	// bug - this never fires and app ends on L71
	onExit := func() {
		os.Exit(1)
	}
	systray.Run(onReady, onExit)
}
