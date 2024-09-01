package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"msg"`
}

// Helper function to log messages in JSON format
func LogJSON(level, msg string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal log entry: %v", err)
		return
	}
	fmt.Println(string(data))
}
