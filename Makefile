# Variables
BINARY_NAME=mock-openai
MAIN_PATH=cmd/server/main.go

.PHONY: all run tidy test clean build help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: Run the server locally
run:
	go run $(MAIN_PATH)

## tidy: Format code and clean up dependencies
tidy:
	go fmt ./...
	go mod tidy

## test: Run all tests with the race detector enabled
test:
	go test -v -race ./...

## build: Compile the binary for the current OS
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

## clean: Remove compiled binaries
clean:
	rm -f $(BINARY_NAME)