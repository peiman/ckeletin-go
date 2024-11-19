.PHONY: build test lint clean format vuln check install run test-race

BINARY_NAME=ckeletin-go
FILE_MODE_EXEC=0755
FILE_MODE_CONFIG=0600

build:
	go build -o ${BINARY_NAME} main.go

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

test-race:
	go test -v -race ./...

lint:
	golangci-lint run

format:
	gofumpt -l -w .

# Security scan using govulncheck
vuln:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Run all quality checks
check: format lint vuln test test-race

clean:
	go clean
	rm -f ${BINARY_NAME}

run:
	./${BINARY_NAME}

install:
	go install

.DEFAULT_GOAL := build