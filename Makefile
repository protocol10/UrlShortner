.PHONY: run build test lint

# Run the application
run:
	go run main.go

# Build the application
build:
	go build -o bin/urlshortener main.go

# Run tests
test:
	go test ./...

# Format and tidy the code
fmt:
	go fmt ./...
	go mod tidy

# Run the golangci-lint linter
lint:
	golangci-lint run
