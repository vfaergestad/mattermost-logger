package utils

import (
	"fmt"
	"os"
)

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// EnsureDir ensures that a directory exists, and creates it if it does not.
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return err
	}
	return nil
}

// ConstructWebSocketURL constructs the WebSocket URL based on domain and TLS usage
func ConstructWebSocketURL(domain string, port int, useTLS bool) (string, error) {
	var scheme string
	if useTLS {
		scheme = "wss"
	} else {
		scheme = "ws"
	}

	// Construct the full WebSocket URL with port
	wsURL := fmt.Sprintf("%s://%s:%d/api/v4/websocket", scheme, domain, port)
	return wsURL, nil
}

// GetString safely extracts a string from a map
func GetString(data interface{}, key string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if val, exists := m[key].(string); exists {
			return val
		}
	}
	return ""
}
