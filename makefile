build:
	@go build -o bin/fs

run: build
	@./bin/fs

test:
	@echo "Running tests..."
	@go test ./tests/...

tidy:
	@echo "Cleaning..."
	@go mod tidy

.PHONY: build run clean test