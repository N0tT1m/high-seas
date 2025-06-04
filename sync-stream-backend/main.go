// main.go
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Configuration structure
type Config struct {
	Port           string
	PlexURL        string
	ClientID       string
	AllowedOrigins []string
	SessionTimeout time.Duration
	CertFile       string
	KeyFile        string
	UseHTTPS       bool
	LogLevel       string
}

// User represents an authenticated Plex user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	PlexToken string    `json:"plexToken"`
	LastSeen  time.Time `json:"lastSeen"`
}

// MediaSession represents a viewing session
type MediaSession struct {
	MediaKey    string            `json:"mediaKey"`
	Position    int               `json:"position"` // Position in milliseconds
	Duration    int               `json:"duration"` // Duration in milliseconds
	State       string            `json:"state"`    // playing, paused, stopped
	UserID      string            `json:"userId"`
	ClientID    string            `json:"clientId"`
	Metadata    map[string]string `json:"metadata"`
	LastUpdated time.Time         `json:"lastUpdated"`
}

// WebSocketMessage defines the structure of messages sent over websockets
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Registration request
type RegistrationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Global state
var (
	config        Config
	users         = make(map[string]*User)
	usersMutex    sync.RWMutex
	sessions      = make(map[string]*MediaSession)
	sessionsMutex sync.RWMutex
	// Map of userID -> list of websocket connections
	userConnections      = make(map[string][]*websocket.Conn)
	userConnectionsMutex sync.RWMutex
	// In main.go - Initialize the WebSocket upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			logger.Printf("WebSocket Origin check: %s", r.Header.Get("Origin"))
			return true // Allow all origins for development
		},
	}
	logger *log.Logger
	db     *sql.DB
)

func main() {
	// Setup logging
	logger = log.New(os.Stdout, "[PLEX-SYNC] ", log.LstdFlags)

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		logger.Println("Warning: .env file not found")
	}

	// Initialize configuration
	config = Config{
		Port:           getEnv("PORT", "8080"),
		PlexURL:        getEnv("PLEX_URL", "192.168.1.78:32400"),
		ClientID:       getEnv("PLEX_CLIENT_ID", "ryzen-win"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:4200"), ","),
		SessionTimeout: 24 * time.Hour,
		CertFile:       getEnv("CERT_FILE", ""),
		KeyFile:        getEnv("KEY_FILE", ""),
		UseHTTPS:       getEnvBool("USE_HTTPS", false),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	// Initialize database
	err = initDB()
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer closeDB()

	// In main.go - Make the upgrader more permissive during development:

	upgrader.CheckOrigin = func(r *http.Request) bool {
		// For development, accept all origins
		if config.LogLevel == "debug" {
			return true
		}

		origin := r.Header.Get("Origin")

		// Accept localhost origins during development
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			return true
		}

		// Check against allowed origins
		for _, allowed := range config.AllowedOrigins {
			if origin == allowed {
				return true
			}
		}

		return false
	}

	// Setup routes
	r := mux.NewRouter()

	// Set up global middleware for CORS before any route handling
	r.Use(corsMiddleware)

	// Add CORS OPTIONS handling for all endpoints
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API endpoints
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/auth/login", handleLogin).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/register", handleRegister).Methods("POST", "OPTIONS")
	api.HandleFunc("/servers", withAuth(handleGetServers)).Methods("GET", "OPTIONS")
	api.HandleFunc("/libraries", withAuth(handleGetLibraries)).Methods("GET", "OPTIONS")
	api.HandleFunc("/library/{libraryKey}/items", withAuth(handleGetLibraryItems)).Methods("GET", "OPTIONS")
	api.HandleFunc("/media/{mediaKey}", withAuth(handleGetMediaInfo)).Methods("GET", "OPTIONS")
	api.HandleFunc("/media/{mediaKey}/position", withAuth(handleUpdatePosition)).Methods("POST", "OPTIONS")
	api.HandleFunc("/media/{mediaKey}/position", withAuth(handleGetPosition)).Methods("GET", "OPTIONS")
	// api.HandleFunc("/media/{mediaKey}/stream", withAuth(handleStreamRequest)).Methods("GET", "OPTIONS")
	api.HandleFunc("/continue-watching", withAuth(handleContinueWatching)).Methods("GET", "OPTIONS")
	api.HandleFunc("/recently-added", withAuth(handleRecentlyAdded)).Methods("GET", "OPTIONS")
	// Handle Plex-style media paths
	api.HandleFunc("/media/library/metadata/{id}", withAuth(handleGetMediaInfoByPlexPath)).Methods("GET", "OPTIONS")
	// Add these routes to your API endpoint definitions in main.go
	// Add these after the line with api.HandleFunc("/media/library/metadata/{id}", withAuth(handleGetMediaInfoByPlexPath))
	api.HandleFunc("/media/library/metadata/{id}/position", withAuth(handleGetPlexStylePosition)).Methods("GET", "OPTIONS")
	api.HandleFunc("/media/library/metadata/{id}/position", withAuth(handleUpdatePlexStylePosition)).Methods("POST", "OPTIONS")

	// 1. First, add this route to your API routes section (find where other routes are registered)
	api.HandleFunc("/media/library/metadata/{id}/stream", withAuth(handleStreamRequest)).Methods("GET", "OPTIONS")

	// Add this to your route definitions (not wrapped in withAuth)
	r.HandleFunc("/ws-test", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("WebSocket test connection request received from: %s", r.RemoteAddr)

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Upgrade connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Printf("Error upgrading connection: %v", err)
			return
		}

		logger.Printf("Test WebSocket connection established successfully")

		// Send welcome message
		welcomeMsg := map[string]interface{}{
			"type": "welcome",
			"payload": map[string]string{
				"message": "Test connection established",
			},
		}
		welcomeBytes, _ := json.Marshal(welcomeMsg)
		conn.WriteMessage(websocket.TextMessage, welcomeBytes)

		// Simple echo handler
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				logger.Printf("WebSocket read error: %v", err)
				break
			}
			logger.Printf("WebSocket test received message: %s", string(message))

			// Echo the message back
			if err := conn.WriteMessage(messageType, message); err != nil {
				logger.Printf("WebSocket write error: %v", err)
				break
			}
		}
	})

	// WebSocket endpoint
	r.HandleFunc("/ws", withAuth(handleWebSocket))

	// Health check
	r.HandleFunc("/health", handleHealthCheck).Methods("GET")

	r.Use(corsMiddleware) // Add CORS middleware first

	// Start cleanup goroutine for expired sessions
	go cleanupSessions()

	// Setup graceful shutdown
	srv := &http.Server{
		Addr:         ":" + config.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      r,
	}

	// Run server in a goroutine
	go func() {
		logger.Printf("Starting server on port %s...\n", config.Port)
		if config.UseHTTPS {
			logger.Fatal(srv.ListenAndServeTLS(config.CertFile, config.KeyFile))
		} else {
			logger.Fatal(srv.ListenAndServe())
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Create deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown server
	srv.Shutdown(ctx)
	logger.Println("Server gracefully shutting down...")
}

// Initialize database connection
func initDB() error {
	// Get database connection info from environment
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "plexsync")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Construct DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Connect to database
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	logger.Println("Connected to PostgreSQL database")

	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// Update the corsMiddleware function in main.go
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all responses
		// In development mode, allow all origins
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Plex-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Add Vary header to indicate to browsers that response may vary based on the Origin header
		w.Header().Set("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")

		// Log request details for debugging
		logger.Printf("CORS request: %s %s from %s", r.Method, r.URL.Path, origin)

		// Handle preflight OPTIONS requests immediately
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Add these log statements in the validatePlexToken function
func validatePlexToken(token string) (bool, string, error) {
	logger.Printf("Validating token: %s (first 8 chars only)", token[:min(8, len(token))])

	// In a real implementation, you would:
	// 1. Make a GET request to https://plex.tv/users/account with the token
	// 2. Check if the response is valid

	// For demo, check if token exists in database
	user, err := getUserByToken(token)
	if err != nil {
		logger.Printf("Error getting user by token: %v", err)
		return false, "", err
	}

	if user == nil {
		logger.Printf("No user found with token")
		return false, "", nil
	}

	logger.Printf("Token valid for user: %s", user.Username)
	return true, user.ID, nil
}

// Add this helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Create database tables
func createTables() error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			plex_token VARCHAR(255) NOT NULL,
			last_seen TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create sessions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS media_sessions (
			media_key VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			position INTEGER NOT NULL,
			duration INTEGER NOT NULL,
			state VARCHAR(50) NOT NULL,
			client_id VARCHAR(255) NOT NULL,
			last_updated TIMESTAMP WITH TIME ZONE NOT NULL,
			PRIMARY KEY (media_key, user_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create media_sessions table: %w", err)
	}

	// Create session metadata table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS session_metadata (
			media_key VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			key VARCHAR(255) NOT NULL,
			value TEXT NOT NULL,
			PRIMARY KEY (media_key, user_id, key),
			FOREIGN KEY (media_key, user_id) REFERENCES media_sessions(media_key, user_id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create session_metadata table: %w", err)
	}

	logger.Println("Database tables created successfully")
	return nil
}

// Close database connection
func closeDB() {
	if db != nil {
		db.Close()
	}
}

// StoreUser stores a user in the database
func storeUser(user *User) error {
	query := `
		INSERT INTO users (id, username, plex_token, last_seen)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			username = $2,
			plex_token = $3,
			last_seen = $4
	`
	_, err := db.Exec(query, user.ID, user.Username, user.PlexToken, user.LastSeen)
	if err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}
	return nil
}

// GetUserByID retrieves a user from the database by ID
func getUserByID(id string) (*User, error) {
	query := `
		SELECT id, username, plex_token, last_seen
		FROM users
		WHERE id = $1
	`
	row := db.QueryRow(query, id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PlexToken, &user.LastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByToken retrieves a user from the database by Plex token
func getUserByToken(token string) (*User, error) {
	query := `
		SELECT id, username, plex_token, last_seen
		FROM users
		WHERE plex_token = $1
	`
	row := db.QueryRow(query, token)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PlexToken, &user.LastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by token: %w", err)
	}

	return &user, nil
}

// UpdateUserLastSeen updates the last_seen timestamp for a user
func updateUserLastSeen(userID string) error {
	query := `
		UPDATE users
		SET last_seen = $1
		WHERE id = $2
	`
	_, err := db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user last_seen: %w", err)
	}
	return nil
}

// DeleteMediaSession deletes a media session from the database
func deleteMediaSession(mediaKey, userID string) error {
	query := `
		DELETE FROM media_sessions
		WHERE media_key = $1 AND user_id = $2
	`
	_, err := db.Exec(query, mediaKey, userID)
	if err != nil {
		return fmt.Errorf("failed to delete media session: %w", err)
	}
	return nil
}

// Update the cleanup sessions function to use the database
func cleanupSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		logger.Println("Running session cleanup")

		// Delete expired sessions
		err := deleteExpiredSessions(config.SessionTimeout)
		if err != nil {
			logger.Printf("Error deleting expired sessions: %v", err)
		}

		// Delete expired users
		err = deleteExpiredUsers(config.SessionTimeout)
		if err != nil {
			logger.Printf("Error deleting expired users: %v", err)
		}
	}
}

// DeleteExpiredSessions deletes sessions that haven't been updated for a while
func deleteExpiredSessions(timeout time.Duration) error {
	query := `
		DELETE FROM media_sessions
		WHERE last_updated < $1
	`
	cutoff := time.Now().Add(-timeout)
	_, err := db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// StoreMediaSession stores a media session in the database
func storeMediaSession(session *MediaSession) error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert or update session
	query := `
		INSERT INTO media_sessions (media_key, user_id, position, duration, state, client_id, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (media_key, user_id) DO UPDATE SET
			position = $3,
			duration = $4,
			state = $5,
			client_id = $6,
			last_updated = $7
	`
	_, err = tx.Exec(query, session.MediaKey, session.UserID, session.Position, session.Duration, session.State, session.ClientID, session.LastUpdated)
	if err != nil {
		return fmt.Errorf("failed to store media session: %w", err)
	}

	// Clear existing metadata
	if len(session.Metadata) > 0 {
		_, err = tx.Exec(`DELETE FROM session_metadata WHERE media_key = $1 AND user_id = $2`, session.MediaKey, session.UserID)
		if err != nil {
			return fmt.Errorf("failed to clear session metadata: %w", err)
		}

		// Insert new metadata
		insertMetaQuery := `
			INSERT INTO session_metadata (media_key, user_id, key, value)
			VALUES ($1, $2, $3, $4)
		`
		for key, value := range session.Metadata {
			_, err = tx.Exec(insertMetaQuery, session.MediaKey, session.UserID, key, value)
			if err != nil {
				return fmt.Errorf("failed to insert session metadata: %w", err)
			}
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetMediaSession retrieves a media session from the database
func getMediaSession(mediaKey, userID string) (*MediaSession, error) {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get session
	query := `
		SELECT media_key, user_id, position, duration, state, client_id, last_updated
		FROM media_sessions
		WHERE media_key = $1 AND user_id = $2
	`
	row := tx.QueryRow(query, mediaKey, userID)

	var session MediaSession
	err = row.Scan(
		&session.MediaKey,
		&session.UserID,
		&session.Position,
		&session.Duration,
		&session.State,
		&session.ClientID,
		&session.LastUpdated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to get media session: %w", err)
	}

	// Get metadata
	metaQuery := `
		SELECT key, value
		FROM session_metadata
		WHERE media_key = $1 AND user_id = $2
	`
	rows, err := tx.Query(metaQuery, mediaKey, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query session metadata: %w", err)
	}
	defer rows.Close()

	session.Metadata = make(map[string]string)
	for rows.Next() {
		var key, value string
		err = rows.Scan(&key, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata row: %w", err)
		}
		session.Metadata[key] = value
	}

	// Check for errors from iterating over rows
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating metadata rows: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &session, nil
}

// GetUserMediaSessions retrieves all media sessions for a user
func getUserMediaSessions(userID string) ([]*MediaSession, error) {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get sessions
	query := `
		SELECT media_key, user_id, position, duration, state, client_id, last_updated
		FROM media_sessions
		WHERE user_id = $1
		ORDER BY last_updated DESC
	`
	rows, err := tx.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user media sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*MediaSession
	for rows.Next() {
		var session MediaSession
		err = rows.Scan(
			&session.MediaKey,
			&session.UserID,
			&session.Position,
			&session.Duration,
			&session.State,
			&session.ClientID,
			&session.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		session.Metadata = make(map[string]string)
		sessions = append(sessions, &session)
	}

	// Check for errors from iterating over rows
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	// Get metadata for all sessions
	for _, session := range sessions {
		metaQuery := `
			SELECT key, value
			FROM session_metadata
			WHERE media_key = $1 AND user_id = $2
		`
		metaRows, err := tx.Query(metaQuery, session.MediaKey, session.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to query session metadata: %w", err)
		}

		for metaRows.Next() {
			var key, value string
			err = metaRows.Scan(&key, &value)
			if err != nil {
				metaRows.Close()
				return nil, fmt.Errorf("failed to scan metadata row: %w", err)
			}
			session.Metadata[key] = value
		}
		metaRows.Close()

		// Check for errors from iterating over metadata rows
		err = metaRows.Err()
		if err != nil {
			return nil, fmt.Errorf("error iterating metadata rows: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sessions, nil
}

// GetContinueWatchingSessions retrieves sessions that are in progress for a user
func getContinueWatchingSessions(userID string) ([]*MediaSession, error) {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get sessions that are in progress (position > 0 and < 95% complete)
	query := `
		SELECT media_key, user_id, position, duration, state, client_id, last_updated
		FROM media_sessions
		WHERE user_id = $1 AND position > 0
		  AND (duration = 0 OR (position::float / duration::float) < 0.95)
		ORDER BY last_updated DESC
	`
	rows, err := tx.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query continue watching sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*MediaSession
	for rows.Next() {
		var session MediaSession
		err = rows.Scan(
			&session.MediaKey,
			&session.UserID,
			&session.Position,
			&session.Duration,
			&session.State,
			&session.ClientID,
			&session.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		session.Metadata = make(map[string]string)
		sessions = append(sessions, &session)
	}

	// Check for errors from iterating over rows
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	// Get metadata for all sessions
	for _, session := range sessions {
		metaQuery := `
			SELECT key, value
			FROM session_metadata
			WHERE media_key = $1 AND user_id = $2
		`
		metaRows, err := tx.Query(metaQuery, session.MediaKey, session.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to query session metadata: %w", err)
		}

		for metaRows.Next() {
			var key, value string
			err = metaRows.Scan(&key, &value)
			if err != nil {
				metaRows.Close()
				return nil, fmt.Errorf("failed to scan metadata row: %w", err)
			}
			session.Metadata[key] = value
		}
		metaRows.Close()

		// Check for errors from iterating over metadata rows
		err = metaRows.Err()
		if err != nil {
			return nil, fmt.Errorf("error iterating metadata rows: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sessions, nil
}

// Update playback position
func handleUpdatePosition(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	vars := mux.Vars(r)
	mediaKey := vars["mediaKey"]

	var data struct {
		Position int    `json:"position"`
		Duration int    `json:"duration"`
		State    string `json:"state"`
		ClientID string `json:"clientId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing session or create new one
	session, err := getMediaSession(mediaKey, userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session == nil {
		// Get media metadata
		metadata, err := getMediaMetadata(userID, mediaKey)
		if err != nil {
			metadata = make(map[string]string)
		}

		session = &MediaSession{
			MediaKey:    mediaKey,
			Position:    data.Position,
			Duration:    data.Duration,
			State:       data.State,
			UserID:      userID,
			ClientID:    data.ClientID,
			Metadata:    metadata,
			LastUpdated: time.Now(),
		}
	} else {
		// Update existing session
		session.Position = data.Position
		if data.Duration > 0 {
			session.Duration = data.Duration
		}
		if data.State != "" {
			session.State = data.State
		}
		session.ClientID = data.ClientID
		session.LastUpdated = time.Now()
	}

	// Save session to database
	err = storeMediaSession(session)
	if err != nil {
		logger.Printf("Error storing media session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Get playback position
func handleGetPosition(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	vars := mux.Vars(r)
	mediaKey := vars["mediaKey"]

	// Get session from database
	session, err := getMediaSession(mediaKey, userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session == nil {
		// No session found, return zero position
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"position": 0,
			"duration": 0,
			"state":    "stopped",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"position":   session.Position,
		"duration":   session.Duration,
		"state":      session.State,
		"lastClient": session.ClientID,
	})
}

// Get recent watch history for continue watching
func handleContinueWatching(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Get continue watching sessions from database
	sessions, err := getContinueWatchingSessions(userID)
	if err != nil {
		logger.Printf("Error getting continue watching sessions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// Then, modify the handleWebSocket function
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Add extensive debug logging
	logger.Printf("WebSocket connection request received from: %s, Origin: %s",
		r.RemoteAddr, r.Header.Get("Origin"))

	// Add CORS headers BEFORE upgrading the connection
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// If this is a preflight OPTIONS request, handle it and return
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID := r.Context().Value("userID").(string)
	clientID := r.URL.Query().Get("clientId")
	token := r.URL.Query().Get("token")

	logger.Printf("WebSocket connection parameters - clientId: %s, userID: %s, token length: %d",
		clientID, userID, len(token))

	if clientID == "" {
		logger.Printf("Missing clientId parameter")
		http.Error(w, "Missing clientId", http.StatusBadRequest)
		return
	}

	// Upgrade the connection with the permissive upgrader
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Printf("ERROR upgrading connection: %v, Headers: %+v", err, r.Header)
		return
	}

	logger.Printf("WebSocket connection successfully established for client: %s", clientID)

	// Add connection to user's connections
	userConnectionsMutex.Lock()
	if _, exists := userConnections[userID]; !exists {
		userConnections[userID] = make([]*websocket.Conn, 0)
	}
	userConnections[userID] = append(userConnections[userID], conn)
	userConnectionsMutex.Unlock()

	// Send a welcome message to confirm connection is working
	welcomeMsg := WebSocketMessage{
		Type: "welcome",
		Payload: map[string]string{
			"message":  "Connection established",
			"clientId": clientID,
		},
	}
	welcomeBytes, _ := json.Marshal(welcomeMsg)
	conn.WriteMessage(websocket.TextMessage, welcomeBytes)

	// Remove connection when closed
	defer func() {
		conn.Close()
		userConnectionsMutex.Lock()
		connections := userConnections[userID]
		for i, c := range connections {
			if c == conn {
				userConnections[userID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
		userConnectionsMutex.Unlock()
	}()

	// WebSocket message handler
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Printf("Error parsing WebSocket message: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "update_position":
			handlePositionUpdate(userID, clientID, msg.Payload)
		case "play":
			handlePlayEvent(userID, clientID, msg.Payload)
		case "pause":
			handlePauseEvent(userID, clientID, msg.Payload)
		case "stop":
			handleStopEvent(userID, clientID, msg.Payload)
		case "get_sessions":
			handleGetSessionsEvent(conn, userID)
		}
	}
}

// Handle position update from WebSocket
func handlePositionUpdate(userID, clientID string, payload interface{}) {
	var data struct {
		MediaKey string `json:"mediaKey"`
		Position int    `json:"position"`
	}

	payloadBytes, _ := json.Marshal(payload)
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return
	}

	// Get existing session or create new one
	session, err := getMediaSession(data.MediaKey, userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		return
	}

	if session == nil {
		session = &MediaSession{
			MediaKey:    data.MediaKey,
			Position:    data.Position,
			State:       "paused",
			UserID:      userID,
			ClientID:    clientID,
			Metadata:    make(map[string]string),
			LastUpdated: time.Now(),
		}
	} else {
		session.Position = data.Position
		session.ClientID = clientID
		session.LastUpdated = time.Now()
	}

	// Save session to database
	err = storeMediaSession(session)
	if err != nil {
		logger.Printf("Error storing media session: %v", err)
		return
	}

	// Broadcast to other clients of the same user
	broadcastPositionUpdate(userID, clientID, data.MediaKey, data.Position)
}

// Broadcast position update to other clients of the same user
func broadcastPositionUpdate(userID, excludeClientID, mediaKey string, position int) {
	userConnectionsMutex.RLock()
	connections := userConnections[userID]
	userConnectionsMutex.RUnlock()

	message := WebSocketMessage{
		Type: "position_update",
		Payload: map[string]interface{}{
			"mediaKey": mediaKey,
			"position": position,
			"clientId": excludeClientID,
		},
	}

	messageBytes, _ := json.Marshal(message)

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			logger.Printf("Error sending position update: %v", err)
		}
	}
}

// Handle play event
func handlePlayEvent(userID, clientID string, payload interface{}) {
	var data struct {
		MediaKey string `json:"mediaKey"`
		Position int    `json:"position"`
	}

	payloadBytes, _ := json.Marshal(payload)
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return
	}

	// Get existing session or create new one
	session, err := getMediaSession(data.MediaKey, userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		return
	}

	if session == nil {
		session = &MediaSession{
			MediaKey:    data.MediaKey,
			Position:    data.Position,
			State:       "playing",
			UserID:      userID,
			ClientID:    clientID,
			Metadata:    make(map[string]string),
			LastUpdated: time.Now(),
		}
	} else {
		session.Position = data.Position
		session.State = "playing"
		session.ClientID = clientID
		session.LastUpdated = time.Now()
	}

	// Save session to database
	err = storeMediaSession(session)
	if err != nil {
		logger.Printf("Error storing media session: %v", err)
		return
	}

	// Broadcast to other clients of the same user
	broadcastPlayEvent(userID, clientID, data.MediaKey, data.Position)
}

// Broadcast play event to other clients of the same user
func broadcastPlayEvent(userID, excludeClientID, mediaKey string, position int) {
	userConnectionsMutex.RLock()
	connections := userConnections[userID]
	userConnectionsMutex.RUnlock()

	message := WebSocketMessage{
		Type: "play_event",
		Payload: map[string]interface{}{
			"mediaKey": mediaKey,
			"position": position,
			"clientId": excludeClientID,
		},
	}

	messageBytes, _ := json.Marshal(message)

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			logger.Printf("Error sending play event: %v", err)
		}
	}
}

// Handle pause event
func handlePauseEvent(userID, clientID string, payload interface{}) {
	var data struct {
		MediaKey string `json:"mediaKey"`
		Position int    `json:"position"`
	}

	payloadBytes, _ := json.Marshal(payload)
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return
	}

	// Get existing session
	session, err := getMediaSession(data.MediaKey, userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		return
	}

	if session != nil {
		session.Position = data.Position
		session.State = "paused"
		session.ClientID = clientID
		session.LastUpdated = time.Now()

		// Save session to database
		err = storeMediaSession(session)
		if err != nil {
			logger.Printf("Error storing media session: %v", err)
			return
		}
	}

	// Broadcast to other clients of the same user
	broadcastPauseEvent(userID, clientID, data.MediaKey, data.Position)
}

// Broadcast pause event to other clients of the same user
func broadcastPauseEvent(userID, excludeClientID, mediaKey string, position int) {
	userConnectionsMutex.RLock()
	connections := userConnections[userID]
	userConnectionsMutex.RUnlock()

	message := WebSocketMessage{
		Type: "pause_event",
		Payload: map[string]interface{}{
			"mediaKey": mediaKey,
			"position": position,
			"clientId": excludeClientID,
		},
	}

	messageBytes, _ := json.Marshal(message)

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			logger.Printf("Error sending pause event: %v", err)
		}
	}
}

// Improved handler functions for Plex-style paths with better authentication logging

// Handler for getting position using Plex-style paths
func handleGetPlexStylePosition(w http.ResponseWriter, r *http.Request) {
	// Log the auth header for debugging
	authHeader := r.Header.Get("Authorization")
	logger.Printf("Auth header in handleGetPlexStylePosition: %s", authHeader[:10]+"...")

	vars := mux.Vars(r)
	id := vars["id"]
	mediaKey := fmt.Sprintf("/library/metadata/%s", id)

	// Log the request details
	logger.Printf("Processing Plex-style position GET request for ID: %s, mediaKey: %s", id, mediaKey)

	// Get user from context - should be set by withAuth middleware
	userID := r.Context().Value("userID")
	if userID == nil {
		logger.Printf("ERROR: No userID in context. Authorization failed or middleware didn't set userID")
		http.Error(w, "Unauthorized - no user ID", http.StatusUnauthorized)
		return
	}

	// Convert to string
	userIDStr, ok := userID.(string)
	if !ok {
		logger.Printf("ERROR: userID is not a string: %v", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get session from database
	session, err := getMediaSession(mediaKey, userIDStr)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session == nil {
		// No session found, return zero position
		logger.Printf("No session found for mediaKey: %s", mediaKey)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"position": 0,
			"duration": 0,
			"state":    "stopped",
		})
		return
	}

	logger.Printf("Found session for mediaKey: %s, position: %d", mediaKey, session.Position)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"position":   session.Position,
		"duration":   session.Duration,
		"state":      session.State,
		"lastClient": session.ClientID,
	})
}

// Handler for updating position using Plex-style paths
func handleUpdatePlexStylePosition(w http.ResponseWriter, r *http.Request) {
	// Log the auth header for debugging
	authHeader := r.Header.Get("Authorization")
	logger.Printf("Auth header in handleUpdatePlexStylePosition: %s", authHeader[:10]+"...")

	vars := mux.Vars(r)
	id := vars["id"]
	mediaKey := fmt.Sprintf("/library/metadata/%s", id)

	// Log the request details
	logger.Printf("Processing Plex-style position UPDATE request for ID: %s, mediaKey: %s", id, mediaKey)

	// Get user from context - should be set by withAuth middleware
	userID := r.Context().Value("userID")
	if userID == nil {
		logger.Printf("ERROR: No userID in context. Authorization failed or middleware didn't set userID")
		http.Error(w, "Unauthorized - no user ID", http.StatusUnauthorized)
		return
	}

	// Convert to string
	userIDStr, ok := userID.(string)
	if !ok {
		logger.Printf("ERROR: userID is not a string: %v", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var data struct {
		Position int    `json:"position"`
		Duration int    `json:"duration"`
		State    string `json:"state"`
		ClientID string `json:"clientId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Printf("Request data: position=%d, duration=%d, state=%s, clientID=%s",
		data.Position, data.Duration, data.State, data.ClientID)

	// Get existing session or create new one
	session, err := getMediaSession(mediaKey, userIDStr)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session == nil {
		logger.Printf("Creating new session for mediaKey: %s", mediaKey)
		// Get media metadata
		metadata, err := getMediaMetadata(userIDStr, mediaKey)
		if err != nil {
			logger.Printf("Error getting media metadata: %v", err)
			metadata = make(map[string]string)
		}

		session = &MediaSession{
			MediaKey:    mediaKey,
			Position:    data.Position,
			Duration:    data.Duration,
			State:       data.State,
			UserID:      userIDStr,
			ClientID:    data.ClientID,
			Metadata:    metadata,
			LastUpdated: time.Now(),
		}
	} else {
		logger.Printf("Updating existing session for mediaKey: %s", mediaKey)
		// Update existing session
		session.Position = data.Position
		if data.Duration > 0 {
			session.Duration = data.Duration
		}
		if data.State != "" {
			session.State = data.State
		}
		session.ClientID = data.ClientID
		session.LastUpdated = time.Now()
	}

	// Save session to database
	err = storeMediaSession(session)
	if err != nil {
		logger.Printf("Error storing media session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Printf("Successfully updated session for mediaKey: %s", mediaKey)
	w.WriteHeader(http.StatusOK)
}

func handleStreamRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers again here to ensure they're included even if there's a redirect
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Plex-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	userID := r.Context().Value("userID").(string)
	vars := mux.Vars(r)
	mediaID := vars["id"]

	// Use the user's token from the database, not from environment
	plexToken := getEnv("PLEX_TOKEN", "")

	// Get user from database
	user, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get server information
	server, err := getPlexServer(plexToken)
	if err != nil {
		http.Error(w, "Failed to get Plex server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create streaming URL
	streamURL := fmt.Sprintf("%s/video/:/transcode/universal/start?path=%s&X-Plex-Token=%s",
		server.URL, fmt.Sprintf("library/metadata/%s", mediaID), user.PlexToken)

	// Log the URL for debugging
	logger.Printf("Stream URL created: %s", streamURL)

	// Return stream URL to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"streamUrl": streamURL,
	})
}

// Handle stop event
func handleStopEvent(userID, clientID string, payload interface{}) {
	var data struct {
		MediaKey string `json:"mediaKey"`
		Position int    `json:"position"`
	}

	payloadBytes, _ := json.Marshal(payload)
	if err := json.Unmarshal(payloadBytes, &data); err != nil {
		return
	}

	// Get existing session
	session, err := getMediaSession(data.MediaKey, userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
		return
	}

	if session != nil {
		session.Position = data.Position
		session.State = "stopped"
		session.ClientID = clientID
		session.LastUpdated = time.Now()

		// Save session to database
		err = storeMediaSession(session)
		if err != nil {
			logger.Printf("Error storing media session: %v", err)
			return
		}
	}

	// Broadcast to other clients of the same user
	broadcastStopEvent(userID, clientID, data.MediaKey, data.Position)
}

// Broadcast stop event to other clients of the same user
func broadcastStopEvent(userID, excludeClientID, mediaKey string, position int) {
	userConnectionsMutex.RLock()
	connections := userConnections[userID]
	userConnectionsMutex.RUnlock()

	message := WebSocketMessage{
		Type: "stop_event",
		Payload: map[string]interface{}{
			"mediaKey": mediaKey,
			"position": position,
			"clientId": excludeClientID,
		},
	}

	messageBytes, _ := json.Marshal(message)

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			logger.Printf("Error sending stop event: %v", err)
		}
	}
}

// Handle get sessions event
func handleGetSessionsEvent(conn *websocket.Conn, userID string) {
	// Get all sessions for user from database
	sessions, err := getUserMediaSessions(userID)
	if err != nil {
		logger.Printf("Error getting user media sessions: %v", err)
		return
	}

	message := WebSocketMessage{
		Type:    "sessions",
		Payload: sessions,
	}

	messageBytes, _ := json.Marshal(message)
	conn.WriteMessage(websocket.TextMessage, messageBytes)
}

// Get Plex server info
func getPlexServer(token string) (struct{ URL string }, error) {
	// In a real implementation, you would:
	// 1. Make a GET request to https://plex.tv/api/resources
	// 2. Parse the response to get server info

	// This is a mock implementation for demonstration
	return struct{ URL string }{
		URL: config.PlexURL,
	}, nil
}

// Get media metadata from Plex
func getMediaMetadata(userID, mediaKey string) (map[string]string, error) {
	// Check if user exists
	user, err := getUserByID(userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Normally you would implement a call to the Plex API here
	// This is a placeholder
	return map[string]string{
		"title":     "Unknown",
		"thumbnail": "",
		"type":      "unknown",
	}, nil
}

// Get Plex libraries
func handleGetLibraries(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	logger.Printf("Getting libraries for user: %s", userID)

	// Check if user exists
	user, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Libraries matching what we saw in the screenshot
	libraries := []map[string]interface{}{
		{
			"key":   "1",
			"title": "Movies",
			"type":  "movie",
			"agent": "Ryzen-Win",
		},
		{
			"key":   "2",
			"title": "TV Shows",
			"type":  "show",
			"agent": "Ryzen-Win",
		},
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	logger.Printf("Returning libraries: %+v", libraries)
	json.NewEncoder(w).Encode(libraries)
}

// Get Plex library items
func handleGetLibraryItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	vars := mux.Vars(r)
	libraryKey := vars["libraryKey"]

	// Use the user's token from the database, not from environment
	plexToken := getEnv("PLEX_TOKEN", "")

	logger.Printf("Getting items for library %s for user %s", libraryKey, userID)

	// Check if user exists
	user, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get Plex server information
	server, err := getPlexServer(plexToken)
	if err != nil {
		logger.Printf("Error getting Plex server: %v", err)
		http.Error(w, "Failed to connect to Plex server", http.StatusInternalServerError)
		return
	}

	// Add logging to verify server URL
	logger.Printf("Using Plex server URL: %s", server.URL)

	// Build URL to fetch library items from Plex
	plexURL := fmt.Sprintf("%s/library/sections/%s/all", server.URL, libraryKey)

	// Create HTTP request to Plex
	req, err := http.NewRequest("GET", plexURL, nil)
	if err != nil {
		logger.Printf("Error creating request to Plex: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add required headers - use the user's token
	req.Header.Add("X-Plex-Token", plexToken)
	req.Header.Add("Accept", "application/json")

	// Send request to Plex
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Error sending request to Plex: %v", err)
		http.Error(w, "Failed to connect to Plex server", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Printf("Plex returned status code %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf("Plex server returned error: %s", resp.Status), resp.StatusCode)
		return
	}

	// Parse Plex response
	var plexResp struct {
		MediaContainer struct {
			Metadata []map[string]interface{} `json:"Metadata"`
			Title1   string                   `json:"title1"` // Library title
		} `json:"MediaContainer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&plexResp); err != nil {
		logger.Printf("Error parsing Plex response: %v", err)
		http.Error(w, "Failed to parse Plex response", http.StatusInternalServerError)
		return
	}

	// Transform Plex data to our format
	items := make([]map[string]interface{}, 0, len(plexResp.MediaContainer.Metadata))

	for _, plexItem := range plexResp.MediaContainer.Metadata {
		// Construct proper thumbnail URL with authentication token
		var thumbnailURL string

		// First, check if thumb exists
		thumb, hasThumb := plexItem["thumb"].(string)
		if hasThumb && thumb != "" {
			// Log the original thumbnail path
			logger.Printf("Original thumbnail path: %s", thumb)

			// Make sure the thumbnail URL is absolute
			if strings.HasPrefix(thumb, "/") {
				// It's a relative URL, so prepend the server URL and add token
				thumbnailURL = fmt.Sprintf("%s%s?X-Plex-Token=%s", server.URL, thumb, plexToken)
			} else {
				// It's already an absolute URL
				thumbnailURL = thumb
				// Still need to add the token if it doesn't have one
				if !strings.Contains(thumbnailURL, "X-Plex-Token") {
					if strings.Contains(thumbnailURL, "?") {
						thumbnailURL += "&X-Plex-Token=" + plexToken
					} else {
						thumbnailURL += "?X-Plex-Token=" + plexToken
					}
				}
			}

			// Log the constructed URL (truncate token for security)
			tokenStart := strings.Index(thumbnailURL, "X-Plex-Token=")
			if tokenStart > 0 {
				logURL := thumbnailURL[:tokenStart+13] + "REDACTED"
				logger.Printf("Constructed thumbnail URL: %s", logURL)
			} else {
				logger.Printf("Constructed thumbnail URL: %s", thumbnailURL)
			}
		} else {
			logger.Printf("No thumbnail found for item: %s", plexItem["title"])
		}

		// Convert Plex item to our format
		item := map[string]interface{}{
			"key":                 plexItem["key"],
			"title":               plexItem["title"],
			"type":                plexItem["type"],
			"thumbnail":           thumbnailURL,
			"librarySectionTitle": plexResp.MediaContainer.Title1,
		}

		// Add year if available
		if year, ok := plexItem["year"]; ok {
			item["year"] = year
		}

		// Add duration if available
		if duration, ok := plexItem["duration"]; ok {
			item["duration"] = duration
		}

		// Add viewOffset if available (for continue watching)
		if viewOffset, ok := plexItem["viewOffset"]; ok {
			item["viewOffset"] = viewOffset
		}

		// Add additional metadata based on content type
		switch plexItem["type"] {
		case "movie":
			if summary, ok := plexItem["summary"]; ok {
				item["summary"] = summary
			}
			if rating, ok := plexItem["rating"]; ok {
				item["rating"] = rating
			}
		case "show":
			if summary, ok := plexItem["summary"]; ok {
				item["summary"] = summary
			}
			if childCount, ok := plexItem["childCount"]; ok {
				item["seasonCount"] = childCount
			}
		case "episode":
			if index, ok := plexItem["index"]; ok {
				item["episodeNumber"] = index
			}
			if parentIndex, ok := plexItem["parentIndex"]; ok {
				item["seasonNumber"] = parentIndex
			}
			if grandparentTitle, ok := plexItem["grandparentTitle"]; ok {
				item["showTitle"] = grandparentTitle
			}
		}

		items = append(items, item)
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	logger.Printf("Returning %d items for library %s", len(items), libraryKey)
	json.NewEncoder(w).Encode(items)
}

// Get media info
func handleGetMediaInfo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	vars := mux.Vars(r)
	mediaKey := vars["mediaKey"]

	// Check if user exists
	user, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get media info from Plex
	// In a real implementation, you would make an API call to Plex
	// This is a mock implementation
	mediaInfo := map[string]interface{}{
		"key":           mediaKey,
		"title":         "Sample Media",
		"year":          2023,
		"type":          "movie",
		"summary":       "This is a sample movie description.",
		"duration":      7200000, // 2 hours in milliseconds
		"thumbnail":     "/library/metadata/12345/thumb/123456",
		"genres":        []string{"Action", "Drama"},
		"directors":     []string{"Sample Director"},
		"actors":        []string{"Actor 1", "Actor 2"},
		"contentRating": "PG-13",
		"rating":        7.5,
	}

	// Check if we have a session for this media
	session, err := getMediaSession(mediaKey, userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
	}

	if session != nil {
		mediaInfo["viewOffset"] = session.Position
		mediaInfo["state"] = session.State
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mediaInfo)
}

// Get Plex servers
func handleGetServers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Check if user exists
	user, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get servers from Plex
	// In a real implementation, you would make an API call to Plex
	// This is a mock implementation
	servers := []map[string]interface{}{
		{
			"name":     "My Plex Server",
			"url":      config.PlexURL,
			"version":  "1.32.5.7328",
			"platform": "Linux",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// Add this new handler function
func handleGetMediaInfoByPlexPath(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mediaID := vars["id"]

	logger.Printf("Handling Plex-style media request for ID: %s", mediaID)

	// Get user from context
	userID := r.Context().Value("userID").(string)

	// Get user from database
	user, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Use the same logic as the existing handler
	// Get media info from Plex
	// In a real implementation, you would make an API call to Plex
	mediaInfo := map[string]interface{}{
		"key":           fmt.Sprintf("/library/metadata/%s", mediaID),
		"title":         "Sample Media",
		"year":          2023,
		"type":          "movie",
		"summary":       "This is a sample movie description.",
		"duration":      7200000, // 2 hours in milliseconds
		"thumbnail":     fmt.Sprintf("/library/metadata/%s/thumb/123456", mediaID),
		"genres":        []string{"Action", "Drama"},
		"directors":     []string{"Sample Director"},
		"actors":        []string{"Actor 1", "Actor 2"},
		"contentRating": "PG-13",
		"rating":        7.5,
	}

	// Check if we have a session for this media
	session, err := getMediaSession(fmt.Sprintf("/library/metadata/%s", mediaID), userID)
	if err != nil {
		logger.Printf("Error getting media session: %v", err)
	}

	if session != nil {
		mediaInfo["viewOffset"] = session.Position
		mediaInfo["state"] = session.State
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mediaInfo)
}

// Get recently added media from Plex
func handleRecentlyAdded(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Check if user exists
	user, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get recently added from Plex
	// In a real implementation, you would make an API call to Plex
	// This is a mock implementation
	recentlyAdded := []map[string]interface{}{
		{
			"key":       "/library/metadata/12345",
			"title":     "Sample Movie",
			"type":      "movie",
			"thumbnail": "/library/metadata/12345/thumb/123456",
			"addedAt":   time.Now().Unix(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recentlyAdded)
}

// syncMediaProgress syncs a specific media's progress to Plex server
func syncMediaProgress(userID, mediaKey string) error {
	// Get session from database
	session, err := getMediaSession(mediaKey, userID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return fmt.Errorf("session not found")
	}

	// Get user information
	user, err := getUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// In a real implementation, you would:
	// 1. Make a POST request to Plex server to update the watch progress
	// 2. Use the appropriate Plex API endpoint

	// This is a mock implementation
	logger.Printf("Syncing progress for media %s: %d ms", mediaKey, session.Position)

	return nil
}

// startSyncWorker starts a background worker to sync sessions to Plex periodically
func startSyncWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Query all users from database
		rows, err := db.Query(`SELECT id FROM users`)
		if err != nil {
			logger.Printf("Error querying users: %v", err)
			continue
		}

		var userIDs []string
		for rows.Next() {
			var userID string
			if err := rows.Scan(&userID); err != nil {
				logger.Printf("Error scanning user ID: %v", err)
				continue
			}
			userIDs = append(userIDs, userID)
		}
		rows.Close()

		// For each user, sync their sessions
		for _, userID := range userIDs {
			// Query all sessions for this user
			sessionRows, err := db.Query(`
				SELECT media_key FROM media_sessions 
				WHERE user_id = $1
			`, userID)

			if err != nil {
				logger.Printf("Error querying sessions: %v", err)
				continue
			}

			var mediaKeys []string
			for sessionRows.Next() {
				var mediaKey string
				if err := sessionRows.Scan(&mediaKey); err != nil {
					logger.Printf("Error scanning media key: %v", err)
					continue
				}
				mediaKeys = append(mediaKeys, mediaKey)
			}
			sessionRows.Close()

			// Sync each session
			for _, mediaKey := range mediaKeys {
				go syncMediaProgress(userID, mediaKey)
			}

			// Also sync between devices
			go syncUserSessions(userID)
		}
	}
}

// syncUserSessions synchronizes sessions between devices for a user
func syncUserSessions(userID string) {
	// Get user information
	user, err := getUserByID(userID)
	if err != nil || user == nil {
		logger.Printf("Error getting user for sync: %v", err)
		return
	}

	// Get all sessions for this user
	sessions, err := getUserMediaSessions(userID)
	if err != nil {
		logger.Printf("Error getting sessions for sync: %v", err)
		return
	}

	// Get active connections for this user
	userConnectionsMutex.RLock()
	connections := userConnections[userID]
	userConnectionsMutex.RUnlock()

	if len(connections) == 0 {
		return // No active connections
	}

	// Send sessions to all connected clients
	message := WebSocketMessage{
		Type:    "sessions_sync",
		Payload: sessions,
	}

	messageBytes, _ := json.Marshal(message)

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			logger.Printf("Error sending sessions sync: %v", err)
		}
	}
}

// Initialize the sync worker when starting the server
func init() {
	// This will run when the package is initialized
	go func() {
		// Wait a bit for the server to start up
		time.Sleep(5 * time.Second)
		startSyncWorker()
	}()
}

// getEnv retrieves an environment variable or returns a default value if not present
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool retrieves a boolean environment variable or returns a default value if not present
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// Improved withAuth middleware with better error handling and logging
func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add detailed logging
		logger.Printf("withAuth middleware processing request: %s %s", r.Method, r.URL.Path)

		// Special handling for WebSocket endpoints
		if r.URL.Path == "/ws" {
			// For WebSocket, get token from URL query parameter
			token := r.URL.Query().Get("token")
			if token == "" {
				logger.Printf("Missing token in WebSocket request")
				http.Error(w, "Unauthorized - missing token", http.StatusUnauthorized)
				return
			}

			// Validate token
			valid, userID, err := validatePlexToken(token)
			if err != nil {
				logger.Printf("Error validating token: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !valid {
				logger.Printf("Invalid token")
				http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
				return
			}

			// Set user ID in request context
			ctx := context.WithValue(r.Context(), "userID", userID)
			logger.Printf("WebSocket authentication successful for user: %s", userID)
			next(w, r.WithContext(ctx))
			return
		}

		// Regular HTTP authentication for non-WebSocket endpoints
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		logger.Printf("Authorization header: %v", authHeader[:10]+"...") // Log only first part for security

		if authHeader == "" {
			logger.Printf("Missing Authorization header")
			http.Error(w, "Unauthorized - missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Printf("Invalid Authorization header format: %s", parts[0])
			http.Error(w, "Unauthorized - invalid Authorization format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			logger.Printf("Empty token")
			http.Error(w, "Unauthorized - empty token", http.StatusUnauthorized)
			return
		}

		// Look up user by token
		user, err := getUserByToken(token)
		if err != nil {
			logger.Printf("Error looking up user by token: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			logger.Printf("Invalid token - no user found")
			http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
			return
		}

		// Update last seen time
		user.LastSeen = time.Now()
		err = updateUserLastSeen(user.ID)
		if err != nil {
			logger.Printf("Error updating user last seen: %v", err)
			// Don't fail the request for this error
		}

		// Set user ID in request context
		ctx := context.WithValue(r.Context(), "userID", user.ID)
		logger.Printf("Authentication successful for user: %s, path: %s", user.ID, r.URL.Path)
		next(w, r.WithContext(ctx))
	}
}

// authenticateWithPlex authenticates a user with Plex and returns their token and user ID
func authenticateWithPlex(username, password string) (string, string, error) {
	// In a real implementation, you would:
	// 1. Make a POST request to https://plex.tv/users/sign_in.json
	// 2. Include the X-Plex-Client-Identifier header
	// 3. Parse the response to get the auth token

	// This is a mock implementation for demonstration
	if username == "" || password == "" {
		return "", "", fmt.Errorf("username and password are required")
	}

	// Generate a deterministic user ID and token based on username
	h := sha256.New()
	h.Write([]byte(username))
	userID := fmt.Sprintf("%x", h.Sum(nil))[:16]

	h.Reset()
	h.Write([]byte(username + password))
	token := fmt.Sprintf("%x", h.Sum(nil))

	return token, userID, nil
}

// loggingMiddleware logs each HTTP request
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		// Log the request details
		logger.Printf(
			"%s %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			rw.statusCode,
			time.Since(start),
		)
	})
}

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// Override WriteHeader to capture status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Health check endpoint
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	err := db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Database connection failed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "1.0.0",
	})
}

// Handle login request
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate with Plex
	token, userID, err := authenticateWithPlex(loginReq.Username, loginReq.Password)
	if err != nil {
		http.Error(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Store user in database
	user := &User{
		ID:        userID,
		Username:  loginReq.Username,
		PlexToken: token,
		LastSeen:  time.Now(),
	}

	err = storeUser(user)
	if err != nil {
		logger.Printf("Error storing user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return token to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":    token,
		"userId":   userID,
		"username": loginReq.Username,
	})
}

// Handle registration request
func handleRegister(w http.ResponseWriter, r *http.Request) {
	var regReq RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&regReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In a real implementation, you would:
	// 1. Validate the registration data
	// 2. Check if the username is available
	// 3. Register the user with Plex or your own auth system

	// For this implementation, we'll use the same authentication
	// flow as login since we're mocking the Plex authentication
	token, userID, err := authenticateWithPlex(regReq.Username, regReq.Password)
	if err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, err := getUserByID(userID)
	if err != nil {
		logger.Printf("Error checking for existing user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Store user in database
	user := &User{
		ID:        userID,
		Username:  regReq.Username,
		PlexToken: token,
		LastSeen:  time.Now(),
	}

	err = storeUser(user)
	if err != nil {
		logger.Printf("Error storing user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return token to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":    token,
		"userId":   userID,
		"username": regReq.Username,
	})
}

// DeleteExpiredUsers deletes users that haven't been seen for a while
func deleteExpiredUsers(timeout time.Duration) error {
	query := `
		DELETE FROM users
		WHERE last_seen < $1
	`
	cutoff := time.Now().Add(-timeout)
	_, err := db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete expired users: %w", err)
	}
	return nil
}
