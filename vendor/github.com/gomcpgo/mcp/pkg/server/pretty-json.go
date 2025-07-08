package server

import (
	"bytes"
	"encoding/json"
	"log"
)

// PrettyJSON takes any value and returns a formatted JSON string representation.
// If the input cannot be marshaled to JSON, it returns an error.
func PrettyJSON(v interface{}) string {
	// First marshal the object to JSON
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		log.Printf("failed to marshal to JSON: %v", err)
		return ""
	}

	// Create a buffer for pretty printing
	var prettyJSON bytes.Buffer

	// Use json.Indent to format the JSON with standard indentation
	err = json.Indent(&prettyJSON, jsonBytes, "", "    ")
	if err != nil {
		log.Printf("failed to indent JSON: %v", err)
		return ""
	}

	return prettyJSON.String()
}
