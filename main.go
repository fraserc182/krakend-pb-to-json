package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/luraproject/lura/v2/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	
	// Import local proto package as pbproto to avoid name conflict
	pbproto "github.com/fraserclark/krakend-pb-to-json/pkg/proto"
)

// Plugin registration
func init() {
    // Register our custom decoder factory under the name "proto"
    encoding.GetRegister().Register("proto", func(bool) func(io.Reader, *map[string]interface{}) error {
        return protobufDecoder
    })
}

// Adapted decoder function that complies with encoding.Decoder signature
func protobufDecoder(r io.Reader, v *map[string]interface{}) error {
    // Read the protobuf data
    data, err := io.ReadAll(r)
    if err != nil {
        fmt.Printf("ERROR: Failed to read data: %v\n", err)
        return err
    }

    // Log data size for debugging
    fmt.Printf("DEBUG: Received %d bytes of protobuf data\n", len(data))
    
    // Handle empty response
    if len(data) == 0 {
        fmt.Printf("WARN: Empty response received\n")
        *v = make(map[string]interface{})
        return nil
    }

    // Create a new GTFS-realtime FeedMessage
    message := &pbproto.FeedMessage{}

    // Unmarshal the protobuf data
    if err := proto.Unmarshal(data, message); err != nil {
        fmt.Printf("ERROR: Failed to unmarshal protobuf: %v\n", err)
        // Try to dump some raw data for debugging
        if len(data) > 20 {
            fmt.Printf("DEBUG: First 20 bytes: %v\n", data[:20])
        } else {
            fmt.Printf("DEBUG: Data: %v\n", data)
        }
        return fmt.Errorf("failed to unmarshal protobuf: %v", err)
    }

    // Configure JSON marshaling options
    marshaler := protojson.MarshalOptions{
        UseProtoNames:   true,
        EmitUnpopulated: true,
    }

    // Convert protobuf to JSON
    jsonData, err := marshaler.Marshal(message)
    if err != nil {
        fmt.Printf("ERROR: Failed to marshal to JSON: %v\n", err)
        return fmt.Errorf("failed to marshal to JSON: %v", err)
    }

    // Parse JSON into map[string]interface{}
    if err := json.Unmarshal(jsonData, v); err != nil {
        fmt.Printf("ERROR: Failed to parse JSON: %v\n", err)
        return fmt.Errorf("failed to parse JSON: %v", err)
    }

    fmt.Printf("DEBUG: Successfully decoded protobuf to JSON\n")
    return nil
}


func main() {
    // Simple testing function for local development
    fmt.Println("KrakenD Protocol Buffer to JSON plugin loaded")
    fmt.Println("This is a plugin and is not meant to be run directly.")
    fmt.Println("Build with: go build -buildmode=plugin -o krakend-pb-to-json.so .")
}