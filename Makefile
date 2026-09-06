.PHONY: build build-all test clean

VERSION ?= 0.1.0
LDFLAGS := -ldflags "-X github.com/pocker/pocker/cmd.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o pocker .

build-all: build-linux build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/pocker-linux-amd64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/pocker-windows-amd64.exe .

test:
	go test ./...

clean:
	rm -rf pocker dist
