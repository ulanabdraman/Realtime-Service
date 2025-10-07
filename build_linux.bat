@echo off
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
echo === Building for Linux...
go build -ldflags="-s -w" ./cmd/realtime
