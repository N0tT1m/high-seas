package utils

import (
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
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

	// Add any common headers here
	req.Header.Set("User-Agent", "High-Seas/1.0")

	return req, nil
}
