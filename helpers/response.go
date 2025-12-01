package helpers

import (
	"encoding/json"
	"fmt"

	"github.com/InspectorGadget/aws-compose-service/structs"
)

// send is the single point that emits JSONL out to Docker Compose.
func send(t, m string) {
	response := structs.Response{
		Type:    t,
		Message: m,
	}

	b, _ := json.Marshal(response)
	fmt.Println(string(b))
}

// Info emits an informational message.
func Info(format string, args ...any) {
	send("info", fmt.Sprintf(format, args...))
}

// Debug emits a debug message (you can filter or down-level this in callers).
func Debug(format string, args ...any) {
	send("debug", fmt.Sprintf(format, args...))
}

// Error emits an error message (but does not exit).
func Error(format string, args ...any) {
	send("error", fmt.Sprintf(format, args...))
}

// Setenv tells Docker Compose to set an environment variable.
func Setenv(key, value string) {
	send("setenv", fmt.Sprintf("%s=%s", key, value))
}
