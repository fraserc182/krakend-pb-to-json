# krakend-pb-to-json

A KrakenD plugin that converts Protocol Buffer data to JSON.

## Installation

### Prerequisites

- KrakenD gateway installed
- Go 1.18+ installed
- Protocol buffer compiler (protoc) installed

### Build Plugin

```bash
# Install the Go protocol buffer compiler plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Generate proto code (if not already generated)
protoc --go_out=. --go_opt=paths=source_relative pkg/proto/gtfs-realtime.proto
protoc --go_out=. --go_opt=paths=source_relative pkg/proto/gtfs.proto

# Build the plugin
go build -buildmode=plugin -o krakend-pb-to-json.so .
```

## Usage

1. Copy the `krakend-pb-to-json.so` file to your KrakenD plugins directory.
2. Configure your KrakenD service to use the plugin:

```json
{
  "version": 3,
  "endpoints": [
    {
      "endpoint": "/api/realtime",
      "method": "GET",
      "backend": [
        {
          "url_pattern": "/gtfs-rt",
          "host": ["http://your-backend-service"],
          "extra_config": {
            "plugin/http-client": {
              "name": ["krakend-pb-to-json.so"],
              "plugin": "proto"
            }
          }
        }
      ]
    }
  ]
}
```

## Development

For local testing and development:

```bash
go run main.go
```

### How it works

This plugin registers a custom decoder that:

1. Reads Protocol Buffer data from the response
2. Unmarshals it into a GTFS FeedMessage structure
3. Converts the structure to JSON
4. Returns the JSON data for KrakenD to process