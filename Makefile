VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test clean

build:
	go build $(LDFLAGS) -o conform .

test:
	go test ./...

clean:
	rm -f conform conform.exe
