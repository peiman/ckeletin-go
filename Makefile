.PHONY: build test lint clean

BINARY_NAME=ckeletin-go

build:
	go build -o ${BINARY_NAME} main.go

test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

lint:
	golangci-lint run

clean:
	go clean
	rm -f ${BINARY_NAME}

run:
	./${BINARY_NAME}

install:
	go install

.DEFAULT_GOAL := build
