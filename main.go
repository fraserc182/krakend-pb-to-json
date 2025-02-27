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
        return err
    }

    // Create a new GTFS-realtime FeedMessage
    message := &pbproto.FeedMessage{}

    // Unmarshal the protobuf data
    if err := proto.Unmarshal(data, message); err != nil {
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
        return fmt.Errorf("failed to marshal to JSON: %v", err)
    }

    // Parse JSON into map[string]interface{}
    if err := json.Unmarshal(jsonData, v); err != nil {
        return fmt.Errorf("failed to parse JSON: %v", err)
    }

    return nil
}


func main() {}