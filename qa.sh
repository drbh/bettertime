 go build -ldflags="-s -w" -o bin/main main.go && \
 go run macapp.go -assets=bin -bin=main -icon=images/logo.png -name=BetterTime -o=target