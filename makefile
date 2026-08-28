# export PATH=$PATH:/bin
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
VERSION := 0.1.0

build:
	@go build -ldflags="-X github.com/rushikeshg25/coolDb/cmd.Version=$(VERSION) -X 'github.com/rushikeshg25/coolDb/cmd.BuildTime=$(BUILD_TIME)'" -o bin/cool

run: build
	@./bin/cool server

shell: build
	@./bin/cool shell

demo:
	@./scripts/demo.sh

test:
	@echo "Running tests..."
	@go test ./...

tidy:
	@echo "Cleaning..."
	@go mod tidy

clean:
	@echo "Removing binary..."
	@rm -f bin/cool

.PHONY: build run shell demo clean test tidy
