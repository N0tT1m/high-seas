// ===============================
// Flexible HTTP/HTTPS routes package
package routes

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"high-seas/src/api"
	"high-seas/src/logger"
	"high-seas/src/metrics"
	"high-seas/src/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	metricsCollector = metrics.New()
)

func init() {
	metricsCollector = metrics.New()
}

// Enhanced CORS middleware
func setupCORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}

// Metrics middleware
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		metricsCollector.RecordSearchTime(time.Since(start))

		// Log request
		logger.WriteInfoWithData("Request processed", map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   time.Since(start),
			"user_agent": c.Request.UserAgent(),
		})
	}
}

// Rate limiting middleware
func rateLimitMiddleware() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use Redis or similar
	return func(c *gin.Context) {
		// For now, just pass through
		// Implement actual rate limiting as needed
		c.Next()
	}
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	tlsEnabled := utils.EnvVarBool("ENABLE_TLS", false)
	mode := "http"
	if tlsEnabled {
		mode = "https"
	}

	health := gin.H{
		"status":      "healthy",
		"timestamp":   time.Now(),
		"version":     "2.0.0",
		"service":     "high-seas",
		"mode":        mode,
		"tls_enabled": tlsEnabled,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

// Metrics endpoint
func getMetrics(c *gin.Context) {
	stats := metricsCollector.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// Enhanced configuration endpoint
func getConfig(c *gin.Context) {
	tlsEnabled := utils.EnvVarBool("ENABLE_TLS", false)

	config := gin.H{
		"jackett": gin.H{
			"host": utils.EnvVar("JACKETT_IP", ""),
			"port": utils.EnvVar("JACKETT_PORT", ""),
		},
		"deluge": gin.H{
			"host": utils.EnvVar("DELUGE_IP", ""),
			"port": utils.EnvVar("DELUGE_PORT", ""),
		},
		"features": gin.H{
			"cache_enabled":   utils.EnvVarBool("ENABLE_CACHE", true),
			"metrics_enabled": true,
			"rate_limiting":   utils.EnvVarBool("ENABLE_RATE_LIMIT", false),
			"tls_enabled":     tlsEnabled,
		},
		"server": gin.H{
			"mode":    utils.EnvVar("SERVER_MODE", "http"),
			"address": utils.EnvVar("SERVER_ADDR", ":8782"),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

func SetupRouter() {
	// Validate configuration first
	if err := utils.ValidateConfig(); err != nil {
		logger.WriteFatal("Configuration validation failed", err)
	}

	// Set Gin mode based on environment
	if utils.EnvVar("GIN_MODE", "release") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Enhanced middleware stack
	r.Use(gin.Recovery())
	r.Use(setupCORS())
	r.Use(metricsMiddleware())
	r.Use(rateLimitMiddleware())

	// System endpoints
	system := r.Group("/system")
	{
		system.GET("/health", healthCheck)
		system.GET("/metrics", getMetrics)
		system.GET("/config", getConfig)
	}

	// Legacy API endpoints (maintain backward compatibility)
	r.POST("/movie/query", api.QueryMovieRequest)
	r.POST("/show/query", api.QueryShowRequest)
	r.POST("/anime/movie/query", api.QueryAnimeMovieRequest)
	r.POST("/anime/show/query", api.MakeAnimeShowQuery)

	// Enhanced search endpoints with better error handling
	v2 := r.Group("/v2")
	{
		search := v2.Group("/search")
		{
			search.POST("/movie", api.EnhancedMovieSearch)
			search.POST("/tv", api.EnhancedTVSearch)
			search.POST("/anime/movie", api.EnhancedAnimeMovieSearch)
			search.POST("/anime/tv", api.EnhancedAnimeTVSearch)
			search.POST("/batch", api.BatchSearch)
		}

		download := v2.Group("/download")
		{
			download.POST("/movie", api.DownloadMovie)
			download.POST("/tv", api.DownloadTV)
			download.POST("/anime", api.DownloadAnime)
		}

		status := v2.Group("/status")
		{
			status.GET("/deluge", api.DelugeStatus)
			status.GET("/jackett", api.JackettStatus)
		}
	}

	// Existing TV show routes
	tmdbShow := r.Group("/tmdb/show")
	{
		tmdbShow.POST("/top-rated-tv-shows", api.QueryTopRatedTvShows)
		tmdbShow.POST("/initial-top-rated-tv-shows", api.QueryInitialTopRatedTvShows)
		tmdbShow.POST("/on-the-air-tv-shows", api.QueryOnTheAirTvShows)
		tmdbShow.POST("/initial-on-the-air-tv-shows", api.QueryInitialOnTheAirTvShows)
		tmdbShow.POST("/popular-tv-shows", api.QueryPopularTvShows)
		tmdbShow.POST("/initial-popular-tv-shows", api.QueryInitialPopularTvShows)
		tmdbShow.POST("/airing-today-tv-shows", api.QueryAiringTodayTvShows)
		tmdbShow.POST("/initial-airing-today-tv-shows", api.QueryInitialAiringTodayTvShows)
		tmdbShow.POST("/all-shows", api.QueryAllTvShows)
		tmdbShow.POST("/all-show-details", api.QueryInitialAllTvShows)
		tmdbShow.POST("/tv-show-details", api.QueryDetailedTopRatedTvShows)
		tmdbShow.POST("/genres", api.QueryShowGenres)
		tmdbShow.POST("/all-tv-show-details", api.QueryAllShowsForDetails)
		tmdbShow.POST("/all-shows-from-date", api.QueryAllShowsFromSelectedDate)
		tmdbShow.POST("/seasons", api.QueryTvShowSeasons)
		tmdbShow.POST("/recommendations", api.QueryTvShowRecommendations)
		tmdbShow.POST("/similar", api.QuerySimilarTvShows)
		tmdbShow.POST("/by-genre", api.QueryShowsByGenre)
		tmdbShow.POST("/search", api.QueryShowSearch)
	}

	// Existing movie routes
	tmdbMovie := r.Group("/tmdb/movie")
	{
		tmdbMovie.POST("/top-rated", api.QueryTopRatedMovies)
		tmdbMovie.POST("/popular", api.QueryPopularMovies)
		tmdbMovie.POST("/now-playing", api.QueryNowPlayingMovies)
		tmdbMovie.POST("/upcoming", api.QueryUpcomingMovies)
		tmdbMovie.POST("/details", api.QueryMovieDetails)
		tmdbMovie.POST("/by-genre", api.QueryMoviesByGenre)
		tmdbMovie.POST("/search", api.QueryMovieSearch)
		tmdbMovie.POST("/genres", api.QueryMovieGenres)
		tmdbMovie.POST("/recommendations", api.QueryMovieRecommendations)
		tmdbMovie.POST("/similar", api.QuerySimilarMovies)
		tmdbMovie.POST("/all-movies", api.QueryAllMovies)
		tmdbMovie.POST("/all-movie-details", api.QueryAllMoviesForDetails)
		tmdbMovie.POST("/all-movies-from-date", api.QueryAllMoviesFromSelectedDate)
	}

	// Start server with appropriate protocol
	startServer(r)
}

func startServer(handler *gin.Engine) {
	serverAddr := utils.EnvVar("SERVER_ADDR", ":8782")

	startHTTPServer(handler, serverAddr)

	// Check multiple ways to determine if TLS should be enabled
	//enableTLS := shouldEnableTLS()

	// if enableTLS {
	//	startHTTPSServer(handler, serverAddr)
	//} else {
	//	startHTTPServer(handler, serverAddr)
	//}
}

// Determine if TLS should be enabled based on multiple factors
func shouldEnableTLS() bool {
	// Method 1: Explicit environment variable
	if utils.EnvVarBool("ENABLE_TLS", false) {
		return true
	}

	// Method 2: Check SERVER_MODE environment variable
	serverMode := utils.EnvVar("SERVER_MODE", "http")
	if serverMode == "https" || serverMode == "tls" {
		return true
	}

	// Method 3: Auto-detect based on certificate files
	certFile := utils.EnvVar("TLS_CERT_FILE", "./fullchain.pem")
	keyFile := utils.EnvVar("TLS_KEY_FILE", "./privkey.pem")

	if utils.EnvVarBool("AUTO_DETECT_TLS", true) {
		if _, err := os.Stat(certFile); err == nil {
			if _, err := os.Stat(keyFile); err == nil {
				logger.WriteInfo("TLS certificates detected, enabling HTTPS")
				return true
			}
		}
	}

	return false
}

func startHTTPSServer(handler *gin.Engine, serverAddr string) {
	certFile := utils.EnvVar("TLS_CERT_FILE", "./fullchain.pem")
	keyFile := utils.EnvVar("TLS_KEY_FILE", "./privkey.pem")

	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.WriteFatal("Failed to load TLS certificate", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.WriteInfo(fmt.Sprintf("Starting HTTPS server on %s", serverAddr))

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			logger.WriteFatal("Failed to start HTTPS server", err)
		}
	}()

	gracefulShutdown(server)
}

func startHTTPServer(handler *gin.Engine, serverAddr string) {
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.WriteInfo(fmt.Sprintf("Starting HTTP server on %s", serverAddr))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WriteFatal("Failed to start HTTP server", err)
		}
	}()

	gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.WriteInfo("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WriteError("Server forced to shutdown", err)
	}

	logger.WriteInfo("Server exited")
}
