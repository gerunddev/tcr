.PHONY: clean test build install

BINARY_NAME=tcr

build:
	go build -o $(BINARY_NAME) .

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	go clean

install:
	go install .
