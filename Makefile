VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build install-local test clean

build:
	go build $(LDFLAGS) -o conform .

install-local:
	go build $(LDFLAGS) -o conform.exe .
	cp conform.exe ~/.local/bin/conform.exe

test:
	go test ./...

clean:
	rm -f conform conform.exe
