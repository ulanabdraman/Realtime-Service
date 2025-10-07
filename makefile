hello:
	@echo GLOT Makefile. Hello!

build_linux:
	@echo === Building for Linux === && \
	set GOOS=linux && \
	set GOARCH=amd64 && \
	set CGO_ENABLED=0 && \
	go build -ldflags="-s -w" ./cmd/realtime

build_windows:
	@echo === Building for Windows === && \
	set GOOS=windows && \
	set GOARCH=amd64 && \
	set CGO_ENABLED=0 && \
	go build -ldflags="-s -w" -o glot_windows.exe ./cmd/realtime

run:
	go run ./cmd/realtime
