.PHONY: build test lint clean

BINARY = bubble

build:
	go build -o $(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
