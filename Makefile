build:
	go build -o hapax8asm main.go
install: build
	mv hapax8asm $GOPATH/bin
