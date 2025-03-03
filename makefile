# export PATH=$PATH:/bin
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
VERSION := 0.1.0

build:
	@go build -ldflags="-X main.Version=$(VERSION) -X 'main.BuildTime=$(BUILD_TIME)'" -o bin/cool

run: build
	@./bin/cool

test:
	@echo "Running tests..."
	@go test ./tests/...

tidy:
	@echo "Cleaning..."
	@go mod tidy

clean:
	@echo "Removing binary..."
	@rm -f bin/cool

.PHONY: build run clean test tidy