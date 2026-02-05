package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// LogEntry represents a single audit event
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	User      string `json:"user,omitempty"`
	Action    string `json:"action"`
	Resource  string `json:"resource,omitempty"`
	Status    string `json:"status"`
	IP        string `json:"ip"`
}

// Logger handles thread-safe logging to a file
type Logger struct {
	mu       sync.Mutex
	filename string
}

var instance *Logger
var once sync.Once

// Init initializes the global logger
func Init(filename string) {
	once.Do(func() {
		instance = &Logger{
			filename: filename,
		}
	})
}

// Log records an event
func Log(user, action, resource, status, ip string) {
	if instance == nil {
		fmt.Println("[AUDIT WARNING] Logger not initialized")
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		User:      user,
		Action:    action,
		Resource:  resource,
		Status:    status,
		IP:        ip,
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	// Append to file
	f, err := os.OpenFile(instance.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to write audit log: %v\n", err)
		return
	}
	defer f.Close()

	data, _ := json.Marshal(entry)
	f.Write(data)
	f.WriteString("\n")
}

// ReadLogs returns the last N logs
func ReadLogs(limit int) ([]LogEntry, error) {
	if instance == nil {
		return nil, fmt.Errorf("logger not initialized")
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	// This is a naive implementation; for production, reading large files backwards is better.
	content, err := os.ReadFile(instance.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, nil
		}
		return nil, err
	}

	var logs []LogEntry
	lines := splitLines(string(content))

	// Parse lines (reverse order for latest first)
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(lines[i]), &entry); err == nil {
			logs = append(logs, entry)
		}
		if len(logs) >= limit {
			break
		}
	}

	return logs, nil
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
