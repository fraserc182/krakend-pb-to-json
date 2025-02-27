package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/luraproject/lura/v2/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	
	// Import local proto package as pbproto to avoid name conflict
	pbproto "github.com/fraserclark/krakend-pb-to-json/pkg/proto"
)

// Symbol exported by the plugin to comply with KrakenD plugin system
var HandlerRegisterer = registerer("proto")

type registerer string

// Plugin registration function that KrakenD calls to load plugin
func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(interface{}, io.ReadCloser) (io.ReadCloser, error),
)) {
	f(string(r), r.registerProtoDecoder)
	fmt.Fprintf(os.Stderr, "Proto decoder registered as '%s'\n", r)
}

// Utility min function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// The actual plugin handler that wraps our protobuf decoder
func (r registerer) registerProtoDecoder(
	cfg interface{},
	resp io.ReadCloser,
) (io.ReadCloser, error) {
	// Log plugin invocation
	fmt.Fprintf(os.Stderr, "[DEBUG] Proto decoder plugin invoked with config: %+v\n", cfg)
	
	// Create a debug log file in the current working directory
	homeDir, _ := os.UserHomeDir()
	logFilePath := homeDir + "/krakend-proto-debug.log"
	fmt.Fprintf(os.Stderr, "[DEBUG] Writing logs to %s\n", logFilePath)
	
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer logFile.Close()
		fmt.Fprintf(logFile, "--- New request ---\n")
		fmt.Fprintf(logFile, "Config: %+v\n", cfg)
	} else {
		fmt.Fprintf(os.Stderr, "[ERROR] Cannot create log file: %s\n", err.Error())
	}
	
	// Read all the response data
	data, err := io.ReadAll(resp)
	if err != nil {
		errMsg := fmt.Sprintf("[ERROR] Reading protobuf data: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, errMsg)
		if logFile != nil {
			fmt.Fprintf(logFile, "%s\n", errMsg)
		}
		return nil, err
	}
	
	// Log data size and first few bytes for debugging
	debugMsg := fmt.Sprintf("[DEBUG] Received %d bytes of protobuf data\n", len(data))
	fmt.Fprintf(os.Stderr, debugMsg)
	if logFile != nil {
		fmt.Fprintf(logFile, "%s\n", debugMsg)
		if len(data) > 0 {
			fmt.Fprintf(logFile, "First 30 bytes: % x\n", data[:min(30, len(data))])
		}
	}
	
	// Handle empty response
	if len(data) == 0 {
		warnMsg := "[WARN] Empty protobuf response\n"
		fmt.Fprintf(os.Stderr, warnMsg)
		if logFile != nil {
			fmt.Fprintf(logFile, "%s\n", warnMsg)
		}
		return io.NopCloser(strings.NewReader("{}")), nil
	}
	
	// Try to create an empty response directly for debugging
	if logFile != nil {
		fmt.Fprintf(logFile, "First attempting to return a simple JSON response for testing\n")
	}
	
	// Debug: Return a static JSON response to test if the issue is with protobuf parsing
	if true {  // Change to true to test static response
		jsonResponse := `{"test": "This is a test response from the proto decoder plugin"}`
		return io.NopCloser(strings.NewReader(jsonResponse)), nil
	}
	
	// Check if data starts with the protobuf magic number (not always present)
	if len(data) > 4 && logFile != nil {
		fmt.Fprintf(logFile, "Initial bytes check: %v\n", data[:4])
	}
	
	// Create a new GTFS-realtime FeedMessage
	message := &pbproto.FeedMessage{}
	
	// Unmarshal the protobuf data
	if err := proto.Unmarshal(data, message); err != nil {
		errMsg := fmt.Sprintf("[ERROR] Unmarshaling protobuf: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, errMsg)
		if logFile != nil {
			fmt.Fprintf(logFile, "%s\n", errMsg)
			// Dump more raw data for debugging
			if len(data) > 50 {
				fmt.Fprintf(logFile, "First 50 bytes: % x\n", data[:50])
			} else {
				fmt.Fprintf(logFile, "All bytes: % x\n", data)
			}
		}
		
		// Return a friendly error response as JSON instead of failing
		errorResponse := fmt.Sprintf(`{"error": "Failed to parse protobuf data", "details": "%s"}`, err.Error())
		return io.NopCloser(strings.NewReader(errorResponse)), nil
	}
	
	// Configure JSON marshaling options
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	
	// Convert protobuf to JSON
	jsonData, err := marshaler.Marshal(message)
	if err != nil {
		errMsg := fmt.Sprintf("[ERROR] Marshaling to JSON: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, errMsg)
		if logFile != nil {
			fmt.Fprintf(logFile, "%s\n", errMsg)
		}
		// Return a friendly error response
		errorResponse := fmt.Sprintf(`{"error": "Failed to convert protobuf to JSON", "details": "%s"}`, err.Error())
		return io.NopCloser(strings.NewReader(errorResponse)), nil
	}
	
	successMsg := "[DEBUG] Successfully decoded protobuf to JSON\n"
	fmt.Fprintf(os.Stderr, successMsg)
	if logFile != nil {
		fmt.Fprintf(logFile, "%s\n", successMsg)
		if len(jsonData) > 100 {
			fmt.Fprintf(logFile, "First 100 chars of JSON: %s\n", jsonData[:100])
		} else if len(jsonData) > 0 {
			fmt.Fprintf(logFile, "JSON output: %s\n", jsonData)
		}
	}
	
	// Return the JSON data as a ReadCloser
	return io.NopCloser(bytes.NewReader(jsonData)), nil
}

// Legacy function kept for compatibility - now we're using the proper plugin approach
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
    
    // Test function to verify plugin functionality
    fmt.Println("\nRunning self-test...")
    
    // Create a test protobuf message
    homeDir, _ := os.UserHomeDir()
    logFile, _ := os.OpenFile(homeDir+"/krakend-plugin-test.log", os.O_CREATE|os.O_WRONLY, 0644)
    if logFile != nil {
        defer logFile.Close()
        fmt.Fprintf(logFile, "Plugin self-test started\n")
        
        // Create a very simple example JSON response
        testJSON := `{"test": "This is a test from the plugin's main function"}`
        
        // Create a reader with this JSON
        jsonReader := strings.NewReader(testJSON)
        
        // Test if our plugin can handle this
        fmt.Fprintf(logFile, "Creating plugin handler\n")
        handler := HandlerRegisterer  // Already a registerer type
        
        // Create a mock ReadCloser
        mockResponse := io.NopCloser(jsonReader)
        
        // Try to process it
        fmt.Fprintf(logFile, "Calling plugin handler\n")
        result, err := handler.registerProtoDecoder(nil, mockResponse)
        
        if err != nil {
            fmt.Fprintf(logFile, "ERROR: %s\n", err)
        } else if result != nil {
            // Read the result
            resultBytes, _ := io.ReadAll(result)
            fmt.Fprintf(logFile, "SUCCESS: Got result: %s\n", string(resultBytes))
        } else {
            fmt.Fprintf(logFile, "ERROR: Nil result without error\n")
        }
    }
}