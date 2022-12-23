build-linux-32:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags "-s -w" -o dist/aws-cf-invalidate-linux-386 src/*.go
build-linux-64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o dist/aws-cf-invalidate-linux src/*.go
build-windows-32:
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags "-s -w" -o dist/aws-cf-invalidate-386.exe src/*.go
build-windows-64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o dist/aws-cf-invalidate.exe src/*.go
build-macos-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o dist/aws-cf-invalidate-mac-arm64 src/*.go

build: build-linux-32 build-linux-64 build-windows-32 build-windows-64 build-macos-arm64

optimize:
	upx --best --lzma dist/*

build-and-install: build-linux-64
	upx --best --lzma dist/aws-cf-invalidate-linux
	sudo cp dist/aws-cf-invalidate-linux /usr/local/bin/aws-cf-invalidate

run:
	go run ./src/*.go

