BINARY := promptarmor
PKG := ./cmd/promptarmor

.PHONY: build test lint clean

build:
	go build -o $(BINARY) $(PKG)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
