// src/config/config.go

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/kirsle/configdir"
)

// Config holds application configuration
type Config struct {
	// UI settings
	UITheme       string `json:"ui_theme"`        // "light", "dark", or "system"
	UIAccentColor string `json:"ui_accent_color"` // Primary color used in the UI
	UIFontSize    int    `json:"ui_font_size"`    // Base font size

	// Player settings
	DefaultVolume       float64 `json:"default_volume"`        // Default volume (0.0-1.0)
	EnableHardwareAccel bool    `json:"enable_hardware_accel"` // Use hardware acceleration when available
	SubtitleFontSize    int     `json:"subtitle_font_size"`    // Subtitle font size

	// Plex settings
	PlexServerURL        string `json:"plex_server_url"`        // URL of the Plex server
	PlexClientIdentifier string `json:"plex_client_identifier"` // Client identifier for Plex API
	PlexUsername         string `json:"plex_username"`          // Plex username (email)
	PlexToken            string `json:"plex_token"`             // Plex authentication token

	// Sync settings
	SyncServerURL string        `json:"sync_server_url"` // URL of the sync server
	SyncInterval  time.Duration `json:"sync_interval"`   // How often to sync playback position (in seconds)
	SyncEnabled   bool          `json:"sync_enabled"`    // Whether sync is enabled

	// Cache settings
	CacheEnabled bool   `json:"cache_enabled"`  // Whether to cache media metadata
	CacheDir     string `json:"cache_dir"`      // Directory to store cache files
	CacheMaxSize int64  `json:"cache_max_size"` // Maximum cache size in MB
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	// Get platform-appropriate cache directory
	cacheDir := configdir.LocalCache("PlexDesktopClient", "cache")
	_ = os.MkdirAll(cacheDir, 0755) // Create cache dir if it doesn't exist

	return &Config{
		UITheme:       "system",
		UIAccentColor: "#E5A00D", // Plex yellow
		UIFontSize:    14,

		DefaultVolume:       0.75,
		EnableHardwareAccel: true,
		SubtitleFontSize:    20,

		PlexServerURL:        "",
		PlexClientIdentifier: "PlexDesktopClient",
		PlexUsername:         "",
		PlexToken:            "",

		SyncServerURL: "",
		SyncInterval:  time.Duration(10) * time.Second,
		SyncEnabled:   true,

		CacheEnabled: true,
		CacheDir:     cacheDir,
		CacheMaxSize: 1024, // 1 GB
	}
}

// LoadConfig loads configuration from disk
func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		config := DefaultConfig()
		if err := config.Save(); err != nil {
			return nil, err
		}
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save writes the config to disk
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	configDir := configdir.LocalConfig("PlexDesktopClient")
	return filepath.Join(configDir, "config.json")
}
