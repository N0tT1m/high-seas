package utils

import (
	"log"
	"net/http/cookiejar"
	"os"

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

func CreateCookieJar() *cookiejar.Jar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}

	return jar
}