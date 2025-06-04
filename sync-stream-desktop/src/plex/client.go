// src/plex/client.go

package plex

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	plexTVSignInURL    = "https://plex.tv/users/sign_in.json"
	plexTVResourcesURL = "https://plex.tv/api/resources"

	defaultClientProduct = "Plex Desktop Client"
	defaultClientVersion = "1.0.0"
	defaultDeviceName    = "Go Plex Desktop"
)

// Config holds Plex client configuration
type Config struct {
	ServerURL        string
	ClientIdentifier string
	Username         string
	Token            string
}

// Client is a Plex API client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Plex client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Login authenticates with Plex and returns a token
func (c *Client) Login(username, password string) (string, error) {
	// Create request
	req, err := http.NewRequest("POST", plexTVSignInURL, nil)
	if err != nil {
		return "", err
	}

	// Add headers
	req.Header.Add("X-Plex-Client-Identifier", c.config.ClientIdentifier)
	req.Header.Add("X-Plex-Product", defaultClientProduct)
	req.Header.Add("X-Plex-Version", defaultClientVersion)
	req.Header.Add("X-Plex-Device-Name", defaultDeviceName)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Add auth
	req.SetBasicAuth(username, password)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("login failed: %s", resp.Status)
	}

	// Parse response
	var loginResp struct {
		User struct {
			AuthToken string `json:"authToken"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	// Update client config with token
	c.config.Token = loginResp.User.AuthToken

	return loginResp.User.AuthToken, nil
}

// ValidateToken checks if the token is valid
func (c *Client) ValidateToken() error {
	if c.config.Token == "" {
		return fmt.Errorf("no token available")
	}

	// Create request to get account info
	req, err := http.NewRequest("GET", "https://plex.tv/api/v2/user", nil)
	if err != nil {
		return err
	}

	// Add headers
	c.addPlexHeaders(req)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed: %s", resp.Status)
	}

	return nil
}

// GetLibraries returns the available libraries
func (c *Client) GetLibraries() ([]*Library, error) {
	// Create request
	endpoint := fmt.Sprintf("%s/library/sections", c.config.ServerURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	c.addPlexHeaders(req)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get libraries: %s", resp.Status)
	}

	// Parse response
	var mediaContainer struct {
		Directories []*Library `xml:"Directory" json:"Directory"`
	}

	// Try to parse as XML first
	if err := xml.NewDecoder(resp.Body).Decode(&mediaContainer); err != nil {
		// If XML parsing fails, try JSON
		resp.Body.Close()

		// Make the request again for JSON
		req, err = http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		// Add headers for JSON response
		c.addPlexHeaders(req)
		req.Header.Add("Accept", "application/json")

		// Send request
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Parse JSON response
		var jsonResp struct {
			MediaContainer struct {
				Directories []*Library `json:"Directory"`
			} `json:"MediaContainer"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
			return nil, err
		}

		mediaContainer.Directories = jsonResp.MediaContainer.Directories
	}

	return mediaContainer.Directories, nil
}

// GetLibraryItems returns items in a library
func (c *Client) GetLibraryItems(libraryKey string) ([]*MediaItem, error) {
	// Create request
	endpoint := fmt.Sprintf("%s/library/sections/%s/all", c.config.ServerURL, libraryKey)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	c.addPlexHeaders(req)
	req.Header.Add("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get library items: %s", resp.Status)
	}

	// Parse response
	var jsonResp struct {
		MediaContainer struct {
			Metadata []*MediaItem `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, err
	}

	return jsonResp.MediaContainer.Metadata, nil
}

// GetContinueWatching returns items that are in progress
func (c *Client) GetContinueWatching() ([]*MediaItem, error) {
	// Create request
	endpoint := fmt.Sprintf("%s/library/onDeck", c.config.ServerURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	c.addPlexHeaders(req)
	req.Header.Add("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get continue watching: %s", resp.Status)
	}

	// Parse response
	var jsonResp struct {
		MediaContainer struct {
			Metadata []*MediaItem `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, err
	}

	return jsonResp.MediaContainer.Metadata, nil
}

// GetRecentlyAdded returns recently added items
func (c *Client) GetRecentlyAdded() ([]*MediaItem, error) {
	// Create request
	endpoint := fmt.Sprintf("%s/library/recentlyAdded", c.config.ServerURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	c.addPlexHeaders(req)
	req.Header.Add("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get recently added: %s", resp.Status)
	}

	// Parse response
	var jsonResp struct {
		MediaContainer struct {
			Metadata []*MediaItem `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, err
	}

	return jsonResp.MediaContainer.Metadata, nil
}

// Search searches across all libraries
func (c *Client) Search(query string) ([]*MediaItem, error) {
	// Create request
	endpoint := fmt.Sprintf("%s/search?query=%s", c.config.ServerURL, url.QueryEscape(query))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	c.addPlexHeaders(req)
	req.Header.Add("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	// Parse response
	var jsonResp struct {
		MediaContainer struct {
			Metadata []*MediaItem `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, err
	}

	return jsonResp.MediaContainer.Metadata, nil
}

// GetMediaInfo gets detailed info about a media item
func (c *Client) GetMediaInfo(key string) (*MediaItem, error) {
	// Create request
	endpoint := fmt.Sprintf("%s%s", c.config.ServerURL, key)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	c.addPlexHeaders(req)
	req.Header.Add("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get media info: %s", resp.Status)
	}

	// Parse response
	var jsonResp struct {
		MediaContainer struct {
			Metadata []*MediaItem `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, err
	}

	if len(jsonResp.MediaContainer.Metadata) == 0 {
		return nil, fmt.Errorf("no metadata found for key %s", key)
	}

	return jsonResp.MediaContainer.Metadata[0], nil
}

// GetStreamURL returns a URL for streaming the media
func (c *Client) GetStreamURL(key string) (string, error) {
	// First get media info to determine if we need transcoding
	mediaInfo, err := c.GetMediaInfo(key)
	if err != nil {
		return "", err
	}

	// Check if media can be direct played
	canDirectPlay := false
	if len(mediaInfo.Media) > 0 && len(mediaInfo.Media[0].Part) > 0 {
		// Simple check - could be expanded based on codec, etc.
		canDirectPlay = true
	}

	var streamURL string
	if canDirectPlay {
		// Direct play
		streamURL = fmt.Sprintf("%s%s?X-Plex-Token=%s",
			c.config.ServerURL,
			mediaInfo.Media[0].Part[0].Key,
			c.config.Token)
	} else {
		// Transcode - you can customize quality settings here
		streamURL = fmt.Sprintf("%s/video/:/transcode/universal/start?path=%s&X-Plex-Token=%s",
			c.config.ServerURL,
			url.QueryEscape(key),
			c.config.Token)
	}

	return streamURL, nil
}

// UpdatePlaybackProgress updates the playback progress on the server
func (c *Client) UpdatePlaybackProgress(key string, positionMs int, state string) error {
	// Determine the state parameter based on player state
	stateParam := "stopped"
	if state == "playing" {
		stateParam = "playing"
	} else if state == "paused" {
		stateParam = "paused"
	}

	// Create request
	endpoint := fmt.Sprintf("%s/:/progress?key=%s&time=%d&state=%s",
		c.config.ServerURL,
		strings.TrimPrefix(key, "/library/metadata/"), // Strip prefix if present
		positionMs,
		stateParam)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	// Add headers
	c.addPlexHeaders(req)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update progress: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetServers returns available Plex servers
func (c *Client) GetServers() ([]Server, error) {
	// Create request
	req, err := http.NewRequest("GET", plexTVResourcesURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	c.addPlexHeaders(req)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get servers: %s", resp.Status)
	}

	// Parse XML response
	var result struct {
		Devices []struct {
			Name        string `xml:"name,attr"`
			Product     string `xml:"product,attr"`
			ProductName string `xml:"productName,attr"`
			Provides    string `xml:"provides,attr"`
			Connections []struct {
				Protocol string `xml:"protocol,attr"`
				Address  string `xml:"address,attr"`
				Port     string `xml:"port,attr"`
				Uri      string `xml:"uri,attr"`
			} `xml:"Connection"`
		} `xml:"Device"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to our server model
	var servers []Server
	for _, device := range result.Devices {
		// Only include Plex Media Server devices
		if !strings.Contains(device.Provides, "server") {
			continue
		}

		// Extract connections
		for _, conn := range device.Connections {
			servers = append(servers, Server{
				Name:    device.Name,
				URL:     conn.Uri,
				Address: conn.Address,
				Port:    conn.Port,
			})
		}
	}

	return servers, nil
}

// Helper to add Plex headers to requests
func (c *Client) addPlexHeaders(req *http.Request) {
	req.Header.Add("X-Plex-Client-Identifier", c.config.ClientIdentifier)
	req.Header.Add("X-Plex-Product", defaultClientProduct)
	req.Header.Add("X-Plex-Version", defaultClientVersion)
	req.Header.Add("X-Plex-Device-Name", defaultDeviceName)

	if c.config.Token != "" {
		req.Header.Add("X-Plex-Token", c.config.Token)
	}
}
