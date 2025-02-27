# CLAUDE.md - KrakenD Protocol Buffer to JSON Converter

## Build & Run
```
go build -buildmode=plugin -o krakend-pb-to-json.so .
go run main.go  # For local development
```

## Test
```
go test ./...                 # Run all tests
go test ./pkg/proto -v        # Test proto package with verbose output
go test -run TestProtoDecoder # Run specific test
```

## Lint & Format
```
go fmt ./...                  # Format code
golangci-lint run             # Comprehensive linting
```

## Code Style
- **Imports**: Standard library first, then third-party, grouped with blank lines
- **Naming**: CamelCase for exported functions/types, camelCase for private ones
- **Comments**: Use doc comments above functions & types
- **Error Handling**: Always check errors with `if err != nil` pattern
- **Types**: Use specific types (not `interface{}`) where possible
- **Formatting**: Follow Go standard formatting (gofmt)
- **PR Review**: Ensure imported proto package is properly referenced

This project converts Protocol Buffer data to JSON for use with KrakenD API Gateway.