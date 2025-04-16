// src/sync/client.go

package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Config holds configuration for the sync client
type Config struct {
	ServerURL    string
	SyncInterval time.Duration
	Enabled      bool
}

// MediaSession represents a viewing session
type MediaSession struct {
	MediaKey    string            `json:"mediaKey"`
	Position    int               `json:"position"` // Position in milliseconds
	Duration    int               `json:"duration"` // Duration in milliseconds
	State       string            `json:"state"`    // playing, paused, stopped
	ClientID    string            `json:"clientId"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	LastUpdated time.Time         `json:"lastUpdated"`
}

// WebSocketMessage defines the structure of messages sent over websockets
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client is a sync client for the Plex Sync Server
type Client struct {
	config         *Config
	conn           *websocket.Conn
	httpClient     *http.Client
	clientID       string
	token          string
	connMutex      sync.Mutex
	isConnected    bool
	done           chan struct{}
	sessionHandler func(*MediaSession)
}

// NewClient creates a new sync client
func NewClient(config *Config) *Client {
	return &Client{
		config:      config,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		clientID:    fmt.Sprintf("desktop-%d", time.Now().UnixNano()),
		isConnected: false,
		done:        make(chan struct{}),
	}
}

// Start connects to the sync server and starts listening for events
func (c *Client) Start(token string, sessionHandler func(*MediaSession)) error {
	if !c.config.Enabled || c.config.ServerURL == "" {
		return fmt.Errorf("sync is disabled or server URL is not set")
	}

	c.token = token
	c.sessionHandler = sessionHandler

	// Connect to WebSocket server
	go c.connectAndListen()

	return nil
}

// Stop stops the sync client
func (c *Client) Stop() {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.isConnected && c.conn != nil {
		// Send close frame
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		close(c.done)
		c.isConnected = false
	}
}

// ReportPosition reports the current playback position to the sync server
func (c *Client) ReportPosition(session *MediaSession) error {
	if !c.config.Enabled || c.config.ServerURL == "" {
		return nil // Silently ignore if sync is disabled
	}

	session.ClientID = c.clientID
	session.LastUpdated = time.Now()

	// If connected via WebSocket, send through that
	if c.isConnected && c.conn != nil {
		c.connMutex.Lock()
		defer c.connMutex.Unlock()

		msg := WebSocketMessage{
			Type:    "update_position",
			Payload: session,
		}

		return c.conn.WriteJSON(msg)
	}

	// Otherwise, use HTTP API
	return c.updatePositionHttp(session)
}

// UpdateConfig updates the client configuration
func (c *Client) UpdateConfig(config *Config) {
	// Check if we need to reconnect based on config changes
	reconnect := c.config.ServerURL != config.ServerURL ||
		c.config.Enabled != config.Enabled ||
		(config.Enabled && !c.isConnected)

	// Update config
	c.config = config

	// Reconnect if necessary
	if reconnect && config.Enabled {
		c.Stop()
		go c.connectAndListen()
	} else if !config.Enabled && c.isConnected {
		c.Stop()
	}
}

// IsConnected returns true if connected to the sync server
func (c *Client) IsConnected() bool {
	return c.isConnected
}

// Private methods

// connectAndListen connects to the WebSocket server and listens for messages
func (c *Client) connectAndListen() {
	// Don't connect if not enabled
	if !c.config.Enabled || c.config.ServerURL == "" {
		return
	}

	// Set up connection
	wsURL := fmt.Sprintf("ws%s/ws?clientId=%s",
		c.config.ServerURL[4:], // Replace http with ws
		c.clientID)

	// Add token to URL
	if c.token != "" {
		wsURL += "&token=" + c.token
	}

	// Connect to WebSocket server
	headers := http.Header{}
	headers.Add("X-Plex-Token", c.token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		log.Printf("Failed to connect to sync server: %v", err)
		// Try to reconnect after delay
		time.Sleep(5 * time.Second)
		go c.connectAndListen()
		return
	}

	// Store connection
	c.connMutex.Lock()
	c.conn = conn
	c.isConnected = true
	c.connMutex.Unlock()

	// Create new done channel
	c.done = make(chan struct{})

	// Set up pinger to keep connection alive
	go c.pinger()

	// Listen for messages
	for {
		select {
		case <-c.done:
			return
		default:
			var msg WebSocketMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				c.connMutex.Lock()
				c.isConnected = false
				c.connMutex.Unlock()

				// Try to reconnect after delay
				time.Sleep(5 * time.Second)
				go c.connectAndListen()
				return
			}

			// Handle message
			c.handleMessage(msg)
		}
	}
}

// pinger sends ping messages to keep the connection alive
func (c *Client) pinger() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.connMutex.Lock()
			if c.conn != nil {
				err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Printf("Error sending ping: %v", err)
				}
			}
			c.connMutex.Unlock()
		case <-c.done:
			return
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (c *Client) handleMessage(msg WebSocketMessage) {
	switch msg.Type {
	case "position_update", "play_event", "pause_event", "stop_event":
		// Convert payload to MediaSession
		payloadBytes, err := json.Marshal(msg.Payload)
		if err != nil {
			log.Printf("Error marshaling payload: %v", err)
			return
		}

		var session MediaSession
		if err := json.Unmarshal(payloadBytes, &session); err != nil {
			log.Printf("Error unmarshaling session: %v", err)
			return
		}

		// Only process events from other clients
		if session.ClientID != c.clientID && c.sessionHandler != nil {
			c.sessionHandler(&session)
		}

	case "sessions":
		// Convert payload to array of MediaSession
		payloadBytes, err := json.Marshal(msg.Payload)
		if err != nil {
			log.Printf("Error marshaling payload: %v", err)
			return
		}

		var sessions []MediaSession
		if err := json.Unmarshal(payloadBytes, &sessions); err != nil {
			log.Printf("Error unmarshaling sessions: %v", err)
			return
		}

		// Process each session
		for _, session := range sessions {
			if session.ClientID != c.clientID && c.sessionHandler != nil {
				c.sessionHandler(&session)
			}
		}
	}
}

// updatePositionHttp updates position using HTTP API
func (c *Client) updatePositionHttp(session *MediaSession) error {
	url := fmt.Sprintf("%s/api/media/%s/position", c.config.ServerURL, session.MediaKey)

	// Create request body
	body, err := json.Marshal(map[string]interface{}{
		"position": session.Position,
		"duration": session.Duration,
		"state":    session.State,
		"clientId": c.clientID,
	})
	if err != nil {
		return err
	}

	// Create request
	// With these lines:
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}

	// Add headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Plex-Token", c.token)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to update position: %s", resp.Status)
	}

	return nil
}
