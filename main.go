package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/luraproject/lura/v2/proxy"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Plugin registration
func init() {
    proxy.RegisterDecoder("proto", ProtoDecoder)  // Changed from encoding.Register
}

// Decoder function that handles protobuf to JSON conversion
func ProtoDecoder(r io.Reader, c *proxy.EntityDecoder) *proxy.Response {
    // Read the protobuf data
    data, err := io.ReadAll(r)
    if err != nil {
        return &proxy.Response{
            IsComplete: true,
            Data:       map[string]interface{}{"error": err.Error()},
        }
    }

    // Create a new GTFS-realtime FeedMessage
    message := &pb.FeedMessage{}

    // Unmarshal the protobuf data
    if err := proto.Unmarshal(data, message); err != nil {
        return &proxy.Response{
            IsComplete: true,
            Data:       map[string]interface{}{"error": fmt.Sprintf("failed to unmarshal protobuf: %v", err)},
        }
    }

    // Configure JSON marshaling options
    marshaler := protojson.MarshalOptions{
        UseProtoNames:   true,
        EmitUnpopulated: true,
    }

    // Convert protobuf to JSON
    jsonData, err := marshaler.Marshal(message)
    if err != nil {
        return &proxy.Response{
            IsComplete: true,
            Data:       map[string]interface{}{"error": fmt.Sprintf("failed to marshal to JSON: %v", err)},
        }
    }

    // Parse JSON into map[string]interface{}
    var result map[string]interface{}
    if err := json.Unmarshal(jsonData, &result); err != nil {
        return &proxy.Response{
            IsComplete: true,
            Data:       map[string]interface{}{"error": fmt.Sprintf("failed to parse JSON: %v", err)},
        }
    }

    return &proxy.Response{
        IsComplete: true,
        Data:       result,
    }
}

func main() {}