build:
	go build -o c8asm main.go
install: build
	mv c8asm $(GOPATH)/bin

test: build
	./c8asm -i test.8asm -o test.bin
	xxd test.bin
