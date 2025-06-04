// Enhanced utils package with additional functionality
package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// EnvVar function is for reading .env file
func EnvVar(key string, defaultVal string) string {
	godotenv.Load(".env")
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultVal
	}
	return value
}

// EnvVarInt gets environment variable as integer with default
func EnvVarInt(key string, defaultValue int) int {
	if value := EnvVar(key, ""); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// EnvVarBool gets environment variable as boolean with default
func EnvVarBool(key string, defaultValue bool) bool {
	if value := EnvVar(key, ""); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// EnvVarDuration gets environment variable as duration with default
func EnvVarDuration(key string, defaultValue time.Duration) time.Duration {
	if value := EnvVar(key, ""); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// CreateCookieJar creates a new cookie jar
func CreateCookieJar() *cookiejar.Jar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	return jar
}

// CreateHTTPClient creates a new HTTP client with timeout and cookie jar
func CreateHTTPClient() *http.Client {
	jar := CreateCookieJar()
	return &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
}

// CreatePostRequest creates a new HTTP POST request with the given content type and body
func CreatePostRequest(url, contentType string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "High-Seas/2.0")

	return req, nil
}

// ValidateConfig validates required configuration
func ValidateConfig() error {
	required := []string{"JACKETT_API_KEY", "JACKETT_IP", "JACKETT_PORT", "DELUGE_USER", "DELUGE_PASSWORD", "DELUGE_IP", "DELUGE_PORT"}

	for _, key := range required {
		if EnvVar(key, "") == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}
