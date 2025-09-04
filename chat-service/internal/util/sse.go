package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// SendSSEEvent sends a Server-Sent Event to the client.
// It marshals the provided data to JSON and formats it according to the SSE specification.
//
// Parameters:
//   w: The http.ResponseWriter to write the event to.
//   event: The name of the event (e.g., "message", "session_id", "complete").
//   data: The data payload for the event, which will be marshaled to JSON.
func SendSSEEvent(w http.ResponseWriter, event string, data interface{}) {
	// Marshal the data payload to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		// If marshaling fails, log the error and send a generic error event instead.
		log.Printf("Error marshaling SSE data for event '%s': %v", event, err)
		// Fallback to sending an error event, ensuring the client gets some notification.
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Failed to marshal SSE data\"}\n\n")
		return
	}

	// Write the event in the Server-Sent Events format:
	// "event: <event_name>\n"
	// "data: <json_payload>\n\n" (two newlines at the end to signify end of event)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
}

// SendSSEError sends a specific error event to the client.
// This is a convenience function that wraps SendSSEEvent for common error reporting.
//
// Parameters:
//   w: The http.ResponseWriter to write the error event to.
//   errMsg: The error message string to be sent to the client.
func SendSSEError(w http.ResponseWriter, errMsg string) {
	// Create a map to hold the error message, which will be marshaled to JSON.
	errorData := map[string]string{"message": errMsg}
	// Call SendSSEEvent with the "error" event name.
	SendSSEEvent(w, "error", errorData)
}