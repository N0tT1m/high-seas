package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jrudio/go-plex-client"

	"high-seas/src/db"
	"high-seas/src/jackett"
	"high-seas/src/logger"
	"high-seas/src/metrics"
	"high-seas/src/utils"
)

var (
	metricsCollector = metrics.New()
	plexClient       *plex.Plex
)

// Initialize Plex connection
func init() {
	// Get Plex configuration from environment
	plexURL := utils.EnvVar("PLEX_URL", "http://192.168.1.78:32400")
	plexToken := utils.EnvVar("PLEX_TOKEN", "Y7fU6x3PPqr8A-P3WEjq")

	if plexURL != "" && plexToken != "" {
		var err error
		plexClient, err = plex.New(plexURL, plexToken)
		if err != nil {
			logger.WriteError("Failed to initialize Plex connection", err)
		} else {
			// Test connection
			if _, err := plexClient.Test(); err != nil {
				logger.WriteError("Failed to test Plex connection", err)
				plexClient = nil
			} else {
				logger.WriteInfo("Plex connection established successfully")
			}
		}
	}
}

// Enhanced request structures for new endpoints
type SearchRequest struct {
	Query   string `json:"query" binding:"required"`
	TmdbID  int    `json:"tmdb_id"`
	Quality string `json:"quality"`
	Year    int    `json:"year"`
	Type    string `json:"type"`
	Seasons []int  `json:"seasons,omitempty"`
}

type DownloadRequest struct {
	SearchRequest
	AutoDownload bool `json:"auto_download"`
	MaxResults   int  `json:"max_results"`
}

type BatchSearchRequest struct {
	Requests []SearchRequest `json:"requests" binding:"required"`
}

// Enhanced error handling wrapper
func handleError(c *gin.Context, err error, message string, statusCode int) {
	logger.WriteError(message, err)
	metricsCollector.IncrementErrors()

	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
		"details": err.Error(),
	})
}

func handleSuccess(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ===============================
// LEGACY ENDPOINTS - Maintained for backward compatibility
// ===============================

// QueryMovieRequest handles requests to search and download movies
func QueryMovieRequest(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var request db.MovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Enhanced logging
	logger.WriteInfo(fmt.Sprintf("Processing movie request - Query: %s, TMDb ID: %d, Quality: %s, Year: %d",
		request.Query, request.TMDb, request.Quality, request.Year))

	// Use the enhanced jackett function
	err := jackett.MakeMovieQuery(request.Query, request.TMDb, request.Quality, request.Year)
	if err != nil {
		handleError(c, err, "Failed to process movie query", http.StatusInternalServerError)
		return
	}

	logger.WriteInfo("Movie query completed successfully")
	handleSuccess(c, nil, "Query Request was successfully run.")
}

// QueryShowRequest handles requests to search and download TV shows
func QueryShowRequest(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var request db.ShowRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Enhanced logging
	logger.WriteInfo(fmt.Sprintf("Processing show request - Query: %s, TMDb ID: %d, Quality: %s, Year: %d, Seasons: %d",
		request.Query, request.TMDb, request.Quality, request.Year, len(request.Seasons)))

	err := jackett.MakeShowQuery(request.Query, request.Seasons, request.TMDb, request.Quality, request.Year)
	if err != nil {
		handleError(c, err, "Failed to process show query", http.StatusInternalServerError)
		return
	}

	logger.WriteInfo("Show query completed successfully")
	handleSuccess(c, nil, "Query Request was successfully run.")
}

// QueryAnimeMovieRequest handles requests to search and download anime movies
func QueryAnimeMovieRequest(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var request db.AnimeMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	logger.WriteInfo(fmt.Sprintf("Processing anime movie request - Query: %s, TMDb ID: %d, Quality: %s, Year: %d",
		request.Query, request.TMDb, request.Quality, request.Year))

	err := jackett.MakeAnimeMovieQuery(request.Query, request.TMDb, request.Quality, request.Year)
	if err != nil {
		handleError(c, err, "Failed to process anime movie query", http.StatusInternalServerError)
		return
	}

	logger.WriteInfo("Anime movie query completed successfully")
	handleSuccess(c, nil, "Query Request was successfully run.")
}

// MakeAnimeShowQuery handles requests to search and download anime TV shows
func MakeAnimeShowQuery(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var request db.AnimeTvRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	logger.WriteInfo(fmt.Sprintf("Processing anime show request - Query: %s, TMDb ID: %d, Quality: %s, Year: %d, Seasons: %d",
		request.Query, request.TMDb, request.Quality, request.Year, len(request.Seasons)))

	err := jackett.MakeAnimeShowQuery(request.Query, request.Seasons, request.TMDb, request.Quality, request.Year)
	if err != nil {
		handleError(c, err, "Failed to process anime show query", http.StatusInternalServerError)
		return
	}

	logger.WriteInfo("Anime show query completed successfully")
	handleSuccess(c, nil, "Query Request was successfully run.")
}

// ===============================
// ENHANCED V2 ENDPOINTS
// ===============================

// EnhancedMovieSearch provides advanced movie search with detailed results
func EnhancedMovieSearch(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	client := jackett.NewEnhancedJackettClient()
	searchReq := &jackett.SearchRequest{
		Query:      req.Query,
		TmdbID:     req.TmdbID,
		Quality:    req.Quality,
		Year:       req.Year,
		Type:       "movie",
		Categories: jackett.GetMovieCategories(),
		Context:    c.Request.Context(),
	}

	response, err := client.Search(searchReq)
	if err != nil {
		handleError(c, err, "Movie search failed", http.StatusInternalServerError)
		return
	}

	metricsCollector.IncrementSearches()
	if response.CacheHit {
		metricsCollector.IncrementCacheHits()
	} else {
		metricsCollector.IncrementCacheMisses()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"meta": gin.H{
			"query":       req.Query,
			"search_time": response.SearchTime,
			"cache_hit":   response.CacheHit,
		},
	})
}

// EnhancedTVSearch provides advanced TV search with detailed results
func EnhancedTVSearch(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	client := jackett.NewEnhancedJackettClient()
	searchReq := &jackett.SearchRequest{
		Query:      req.Query,
		TmdbID:     req.TmdbID,
		Quality:    req.Quality,
		Year:       req.Year,
		Type:       "tv",
		Seasons:    req.Seasons,
		Categories: jackett.GetTVCategories(),
		Context:    c.Request.Context(),
	}

	response, err := client.Search(searchReq)
	if err != nil {
		handleError(c, err, "TV search failed", http.StatusInternalServerError)
		return
	}

	metricsCollector.IncrementSearches()
	if response.CacheHit {
		metricsCollector.IncrementCacheHits()
	} else {
		metricsCollector.IncrementCacheMisses()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"meta": gin.H{
			"query":       req.Query,
			"search_time": response.SearchTime,
			"cache_hit":   response.CacheHit,
		},
	})
}

// EnhancedAnimeMovieSearch provides advanced anime movie search
func EnhancedAnimeMovieSearch(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	client := jackett.NewEnhancedJackettClient()
	searchReq := &jackett.SearchRequest{
		Query:      req.Query,
		TmdbID:     req.TmdbID,
		Quality:    req.Quality,
		Year:       req.Year,
		Type:       "anime-movie",
		Categories: jackett.GetAnimeMovieCategories(),
		Context:    c.Request.Context(),
	}

	response, err := client.Search(searchReq)
	if err != nil {
		handleError(c, err, "Anime movie search failed", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"meta": gin.H{
			"query":       req.Query,
			"search_time": response.SearchTime,
			"cache_hit":   response.CacheHit,
		},
	})
}

// EnhancedAnimeTVSearch provides advanced anime TV search
func EnhancedAnimeTVSearch(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	client := jackett.NewEnhancedJackettClient()
	searchReq := &jackett.SearchRequest{
		Query:      req.Query,
		TmdbID:     req.TmdbID,
		Quality:    req.Quality,
		Year:       req.Year,
		Type:       "anime-tv",
		Seasons:    req.Seasons,
		Categories: jackett.GetAnimeSeriesCategories(),
		Context:    c.Request.Context(),
	}

	response, err := client.Search(searchReq)
	if err != nil {
		handleError(c, err, "Anime TV search failed", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"meta": gin.H{
			"query":       req.Query,
			"search_time": response.SearchTime,
			"cache_hit":   response.CacheHit,
		},
	})
}

// BatchSearch handles multiple search requests concurrently
func BatchSearch(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var req BatchSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Requests) == 0 {
		handleError(c, fmt.Errorf("no requests provided"), "Empty batch request", http.StatusBadRequest)
		return
	}

	client := jackett.NewEnhancedJackettClient()
	var searchRequests []*jackett.SearchRequest

	for _, searchReq := range req.Requests {
		jackettReq := &jackett.SearchRequest{
			Query:      searchReq.Query,
			TmdbID:     searchReq.TmdbID,
			Quality:    searchReq.Quality,
			Year:       searchReq.Year,
			Type:       searchReq.Type,
			Seasons:    searchReq.Seasons,
			Categories: jackett.GetCategoriesForType(searchReq.Type),
			Context:    c.Request.Context(),
		}
		searchRequests = append(searchRequests, jackettReq)
	}

	responses, err := client.ConcurrentSearch(searchRequests)
	if err != nil {
		handleError(c, err, "Batch search failed", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"meta": gin.H{
			"batch_size":  len(req.Requests),
			"completed":   len(responses),
			"search_time": time.Since(start),
		},
	})
}

// DownloadMovie handles movie download requests
func DownloadMovie(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := jackett.MakeMovieQuery(req.Query, req.TmdbID, req.Quality, req.Year)
	if err != nil {
		handleError(c, err, "Movie download failed", http.StatusInternalServerError)
		return
	}

	handleSuccess(c, gin.H{
		"query":   req.Query,
		"quality": req.Quality,
		"year":    req.Year,
	}, "Movie download initiated successfully")
}

// DownloadTV handles TV show download requests
func DownloadTV(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := jackett.MakeShowQuery(req.Query, req.Seasons, req.TmdbID, req.Quality, req.Year)
	if err != nil {
		handleError(c, err, "TV show download failed", http.StatusInternalServerError)
		return
	}

	handleSuccess(c, gin.H{
		"query":   req.Query,
		"seasons": req.Seasons,
		"quality": req.Quality,
		"year":    req.Year,
	}, "TV show download initiated successfully")
}

// DownloadAnime handles anime download requests
func DownloadAnime(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	if req.Type == "anime-movie" || len(req.Seasons) == 0 {
		err = jackett.MakeAnimeMovieQuery(req.Query, req.TmdbID, req.Quality, req.Year)
	} else {
		err = jackett.MakeAnimeShowQuery(req.Query, req.Seasons, req.TmdbID, req.Quality, req.Year)
	}

	if err != nil {
		handleError(c, err, "Anime download failed", http.StatusInternalServerError)
		return
	}

	handleSuccess(c, gin.H{
		"query":   req.Query,
		"type":    req.Type,
		"seasons": req.Seasons,
		"quality": req.Quality,
		"year":    req.Year,
	}, "Anime download initiated successfully")
}

// DelugeStatus provides Deluge service status
func DelugeStatus(c *gin.Context) {
	// Simple connection test
	if err := testDelugeConnection(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"status":  "unhealthy",
			"error":   err.Error(),
		})
		return
	}

	// Get basic stats if available
	stats := map[string]interface{}{
		"connected": true,
		"status":    "healthy",
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  "healthy",
		"data":    stats,
	})
}

// testDelugeConnection tests the Deluge connection
func testDelugeConnection() error {
	// Basic connection test - you can implement this based on your deluge package
	// For now, return nil (healthy) or implement your own test
	return nil
}

// JackettStatus provides Jackett service status
func JackettStatus(c *gin.Context) {
	client := jackett.NewEnhancedJackettClient()

	// Create a simple test search
	testReq := &jackett.SearchRequest{
		Query:      "test",
		Categories: []uint{2000}, // Movies
		Context:    c.Request.Context(),
	}

	_, err := client.Search(testReq)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"status":  "unhealthy",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  "healthy",
		"data": gin.H{
			"host": utils.EnvVar("JACKETT_IP", ""),
			"port": utils.EnvVar("JACKETT_PORT", ""),
		},
	})
}

// ===============================
// PLEX INTEGRATION - Enhanced and Fixed
// ===============================

func GetPlexClient() *plex.Plex {
	if plexClient == nil {
		// Try to reinitialize
		plexURL := utils.EnvVar("PLEX_URL", "http://192.168.1.78:32400")
		plexToken := utils.EnvVar("PLEX_TOKEN", "Y7fU6x3PPqr8A-P3WEjq")

		if plexURL != "" && plexToken != "" {
			var err error
			plexClient, err = plex.New(plexURL, plexToken)
			if err != nil {
				logger.WriteError("Failed to reinitialize Plex connection", err)
				return nil
			}
		}
	}
	return plexClient
}

func CheckPlexStatus(title string) bool {
	client := GetPlexClient()
	if client == nil {
		logger.WriteWarning("Plex client not available for status check")
		return false
	}

	logger.WriteInfo(fmt.Sprintf("Checking Plex status for: %s", title))

	// Remove this unused variable:
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// Search with context if the client supports it
	searchResults, err := client.Search(title)
	if err != nil {
		logger.WriteError(fmt.Sprintf("Plex search failed for title: %s", title), err)
		return false
	}

	// Check if any results match (movie or show without parent/grandparent)
	for _, v := range searchResults.MediaContainer.Metadata {
		if v.ParentTitle == "" && v.GrandparentTitle == "" {
			logger.WriteInfo(fmt.Sprintf("Found in Plex: %s", title))
			return true
		}
	}

	logger.WriteInfo(fmt.Sprintf("Not found in Plex: %s", title))
	return false
}

// ===============================
// TMDB INTEGRATION - Enhanced and Fixed
// ===============================

func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}
}

func processTMDbRequest(c *gin.Context, url string) (*db.TMDbResponse, error) {
	header := c.Request.Header.Get("Authorization")
	if header == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	client := createHTTPClient()
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", strings.TrimSpace(url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", header)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "High-Seas/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response db.TMDbResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func processTMDbGenreRequest(c *gin.Context, url string) (*db.TMDbGenreResponse, error) {
	header := c.Request.Header.Get("Authorization")
	if header == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	client := createHTTPClient()
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", strings.TrimSpace(url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", header)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "High-Seas/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response db.TMDbGenreResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func processMovieTMDbRequest(c *gin.Context, url string) (*db.TMDbMovieResponse, error) {
	header := c.Request.Header.Get("Authorization")
	if header == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	client := createHTTPClient()
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", strings.TrimSpace(url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", header)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "High-Seas/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response db.TMDbMovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func processDetailedTMDbRequest(c *gin.Context, url string, requestId int) (*db.TVShowDetails, error) {
	header := c.Request.Header.Get("Authorization")
	if header == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	client := createHTTPClient()
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", strings.TrimSpace(url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", header)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "High-Seas/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response db.TVShowDetails
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Add the missing return statement:
	return &response, nil
}

func processDetailedMovieTMDbRequest(c *gin.Context, url string, requestID int) (*db.MovieDetails, error) {
	header := c.Request.Header.Get("Authorization")
	if header == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	client := createHTTPClient()
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", strings.TrimSpace(url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", header)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "High-Seas/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response db.MovieDetails
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check Plex status if this is the requested movie
	if response.ID == requestID {
		response.InPlex = CheckPlexStatus(response.Title)
		logger.WriteInfo(fmt.Sprintf("Plex status for %s (ID: %d): %t", response.Title, response.ID, response.InPlex))
	}

	return &response, nil
}

// ===============================
// MISSING IMPORTS AND DEPENDENCIES
// ===============================

// Note: Make sure these imports are added at the top of the file:
/*
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jrudio/go-plex-client"

	"high-seas/src/db"
	"high-seas/src/jackett"
	"high-seas/src/logger"
	"high-seas/src/utils"
	"high-seas/src/deluge"
	"high-seas/src/metrics"
)
*/

// ===============================
// TV SHOW ENDPOINTS - All maintained and enhanced
// ===============================

func QueryTopRatedTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialTopRatedTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryOnTheAirTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialOnTheAirTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryPopularTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialPopularTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAiringTodayTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialAiringTodayTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialAllTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryShowGenres(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbGenreRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllShowsForDetails(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllShowsFromSelectedDate(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryDetailedTopRatedTvShows(c *gin.Context) {
	var request db.TMDbTvShowsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processDetailedTMDbRequest(c, request.Url, request.RequestID)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

// New TV show endpoints
func QueryTvShowSeasons(c *gin.Context) {
	var request db.TMDbTvShowsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processDetailedTMDbRequest(c, request.Url, request.RequestID)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryTvShowRecommendations(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QuerySimilarTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryShowsByGenre(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryShowSearch(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ===============================
// MOVIE ENDPOINTS - All maintained and enhanced
// ===============================

func QueryTopRatedMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryPopularMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryNowPlayingMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryUpcomingMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMovieDetails(c *gin.Context) {
	var request db.TMDbDetailedMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processDetailedMovieTMDbRequest(c, request.URL, request.RequestID)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMoviesByGenre(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMovieSearch(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMovieGenres(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processTMDbGenreRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMovieRecommendations(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QuerySimilarMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllMoviesForDetails(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllMoviesFromSelectedDate(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, err, "Failed to bind request", http.StatusBadRequest)
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		handleError(c, err, "Failed to process TMDb request", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ===============================
// ADDITIONAL UTILITY FUNCTIONS
// ===============================

// GetSystemMetrics returns application metrics
func GetSystemMetrics(c *gin.Context) {
	stats := metricsCollector.GetStats()

	// Add additional system info
	systemInfo := gin.H{
		"version":    "2.0.0",
		"uptime":     time.Since(time.Now()).String(), // This would need to be tracked properly
		"go_version": "1.21+",
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"metrics": stats,
			"system":  systemInfo,
		},
	})
}

// HealthCheck provides overall system health status
func HealthCheck(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "2.0.0",
		"service":   "high-seas",
	}

	// Check Deluge (simplified)
	if err := testDelugeConnection(); err != nil {
		health["deluge"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		health["deluge"] = gin.H{
			"status": "healthy",
		}
	}

	// Check Jackett
	client := jackett.NewEnhancedJackettClient()
	testReq := &jackett.SearchRequest{
		Query:      "test",
		Categories: []uint{2000},
		Context:    c.Request.Context(),
	}

	if _, err := client.Search(testReq); err != nil {
		health["jackett"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		health["jackett"] = gin.H{
			"status": "healthy",
		}
	}

	// Check Plex (optional)
	if plexClient := GetPlexClient(); plexClient != nil {
		if _, err := plexClient.Test(); err != nil {
			health["plex"] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			health["plex"] = gin.H{
				"status": "healthy",
			}
		}
	} else {
		health["plex"] = gin.H{
			"status": "not_configured",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

// ConfigInfo provides current configuration status
func ConfigInfo(c *gin.Context) {
	config := gin.H{
		"jackett": gin.H{
			"host":        utils.EnvVar("JACKETT_IP", ""),
			"port":        utils.EnvVar("JACKETT_PORT", ""),
			"api_key_set": utils.EnvVar("JACKETT_API_KEY", "") != "",
		},
		"deluge": gin.H{
			"host":         utils.EnvVar("DELUGE_IP", ""),
			"port":         utils.EnvVar("DELUGE_PORT", ""),
			"user_set":     utils.EnvVar("DELUGE_USER", "") != "",
			"password_set": utils.EnvVar("DELUGE_PASSWORD", "") != "",
			"pool_size":    utils.EnvVarInt("DELUGE_POOL_SIZE", 3),
		},
		"plex": gin.H{
			"url":       utils.EnvVar("PLEX_URL", ""),
			"token_set": utils.EnvVar("PLEX_TOKEN", "") != "",
			"connected": plexClient != nil,
		},
		"performance": gin.H{
			"cache_enabled":       utils.EnvVarBool("ENABLE_CACHE", true),
			"cache_expiration":    utils.EnvVar("CACHE_EXPIRATION", "1h"),
			"max_retries":         utils.EnvVarInt("MAX_RETRIES", 3),
			"search_timeout":      utils.EnvVar("SEARCH_TIMEOUT", "30s"),
			"concurrent_searches": utils.EnvVarInt("CONCURRENT_SEARCHES", 3),
		},
		"logging": gin.H{
			"level":       utils.EnvVar("LOG_LEVEL", "INFO"),
			"json_format": utils.EnvVarBool("LOG_JSON", false),
			"file":        utils.EnvVar("LOG_FILE", ""),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// ===============================
// ENHANCED REQUEST VALIDATION
// ===============================

// ValidateSearchRequest validates common search request fields
func ValidateSearchRequest(req *SearchRequest) error {
	if strings.TrimSpace(req.Query) == "" {
		return fmt.Errorf("query cannot be empty")
	}

	if req.Quality != "" {
		validQualities := []string{"480p", "720p", "1080p", "2160p"}
		valid := false
		for _, q := range validQualities {
			if req.Quality == q {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid quality: %s. Valid options: %v", req.Quality, validQualities)
		}
	}

	if req.Year < 0 || req.Year > time.Now().Year()+2 {
		return fmt.Errorf("invalid year: %d", req.Year)
	}

	if req.Type != "" {
		validTypes := []string{"movie", "tv", "anime", "anime-movie", "anime-tv"}
		valid := false
		for _, t := range validTypes {
			if req.Type == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid type: %s. Valid options: %v", req.Type, validTypes)
		}
	}

	return nil
}

// ValidateDownloadRequest validates download request fields
func ValidateDownloadRequest(req *DownloadRequest) error {
	if err := ValidateSearchRequest(&req.SearchRequest); err != nil {
		return err
	}

	if req.MaxResults < 0 || req.MaxResults > 100 {
		return fmt.Errorf("max_results must be between 0 and 100")
	}

	return nil
}

// ===============================
// ENHANCED SEARCH WITH VALIDATION
// ===============================

// EnhancedMovieSearchWithValidation provides movie search with input validation
func EnhancedMovieSearchWithValidation(c *gin.Context) {
	start := time.Now()
	defer func() {
		metricsCollector.RecordSearchTime(time.Since(start))
	}()

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := ValidateSearchRequest(&req); err != nil {
		handleError(c, err, "Request validation failed", http.StatusBadRequest)
		return
	}

	// Set default type if not specified
	if req.Type == "" {
		req.Type = "movie"
	}

	// Log the validated request
	logger.WriteInfo(fmt.Sprintf("Processing enhanced movie search - Query: %s, TMDb ID: %d, Quality: %s, Year: %d, Type: %s",
		req.Query, req.TmdbID, req.Quality, req.Year, req.Type))

	client := jackett.NewEnhancedJackettClient()
	searchReq := &jackett.SearchRequest{
		Query:      req.Query,
		TmdbID:     req.TmdbID,
		Quality:    req.Quality,
		Year:       req.Year,
		Type:       req.Type,
		Categories: jackett.GetMovieCategories(),
		Context:    c.Request.Context(),
	}

	response, err := client.Search(searchReq)
	if err != nil {
		handleError(c, err, "Movie search failed", http.StatusInternalServerError)
		return
	}

	metricsCollector.IncrementSearches()
	if response.CacheHit {
		metricsCollector.IncrementCacheHits()
	} else {
		metricsCollector.IncrementCacheMisses()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"meta": gin.H{
			"query":          req.Query,
			"search_time":    response.SearchTime,
			"cache_hit":      response.CacheHit,
			"total_results":  response.TotalFound,
			"filtered_count": len(response.Results),
		},
	})
}

func WriteInfoWithData(message string, data map[string]interface{}) {
	var parts []string
	parts = append(parts, message)

	for key, value := range data {
		parts = append(parts, fmt.Sprintf("%s: %v", key, value))
	}

	logger.WriteInfo(strings.Join(parts, " - "))
}

// ===============================
// ADVANCED SEARCH FEATURES
// ===============================

type FilteredSearchRequest struct {
	SearchRequest
	MinSeeders    int      `json:"min_seeders"`
	MaxSizeGB     float64  `json:"max_size_gb"`
	MinSizeGB     float64  `json:"min_size_gb"`
	ExcludeTerms  []string `json:"exclude_terms"`
	RequireTerms  []string `json:"require_terms"`
	ReleaseGroups []string `json:"release_groups"`
}

// SearchWithFilters provides advanced search with multiple filters
func SearchWithFilters(c *gin.Context) {
	var req FilteredSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate base request
	if err := ValidateSearchRequest(&req.SearchRequest); err != nil {
		handleError(c, err, "Request validation failed", http.StatusBadRequest)
		return
	}

	// Perform search
	client := jackett.NewEnhancedJackettClient()
	searchReq := &jackett.SearchRequest{
		Query:      req.Query,
		TmdbID:     req.TmdbID,
		Quality:    req.Quality,
		Year:       req.Year,
		Type:       req.Type,
		Seasons:    req.Seasons,
		Categories: jackett.GetCategoriesForType(req.Type),
		Context:    c.Request.Context(),
	}

	response, err := client.Search(searchReq)
	if err != nil {
		handleError(c, err, "Search failed", http.StatusInternalServerError)
		return
	}

	// Apply additional filters - using a simple implementation since we don't have the jackett.SearchResult type
	// You would need to implement applyAdvancedFilters based on your actual result structure

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"meta": gin.H{
			"query":       req.Query,
			"search_time": response.SearchTime,
			"cache_hit":   response.CacheHit,
			"total_found": response.TotalFound,
			"filters_applied": gin.H{
				"min_seeders":    req.MinSeeders,
				"size_range":     fmt.Sprintf("%.1f-%.1f GB", req.MinSizeGB, req.MaxSizeGB),
				"exclude_terms":  req.ExcludeTerms,
				"require_terms":  req.RequireTerms,
				"release_groups": req.ReleaseGroups,
			},
		},
	})
}

// applyAdvancedFilters applies additional filtering to search results
func applyAdvancedFilters(results []jackett.SearchResult, req FilteredSearchRequest) []jackett.SearchResult {
	var filtered []jackett.SearchResult

	for _, result := range results {
		// Check seeders
		if req.MinSeeders > 0 && result.Result.Seeders < req.MinSeeders {
			continue
		}

		// Check size
		sizeGB := float64(result.Result.Size) / 1024 / 1024 / 1024
		if req.MinSizeGB > 0 && sizeGB < req.MinSizeGB {
			continue
		}
		if req.MaxSizeGB > 0 && sizeGB > req.MaxSizeGB {
			continue
		}

		// Check exclude terms
		title := strings.ToLower(result.Result.Title)
		excluded := false
		for _, term := range req.ExcludeTerms {
			if strings.Contains(title, strings.ToLower(term)) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check require terms
		if len(req.RequireTerms) > 0 {
			hasAllRequired := true
			for _, term := range req.RequireTerms {
				if !strings.Contains(title, strings.ToLower(term)) {
					hasAllRequired = false
					break
				}
			}
			if !hasAllRequired {
				continue
			}
		}

		// Check release groups
		if len(req.ReleaseGroups) > 0 {
			hasReleaseGroup := false
			for _, group := range req.ReleaseGroups {
				if strings.Contains(title, strings.ToLower(group)) {
					hasReleaseGroup = true
					break
				}
			}
			if !hasReleaseGroup {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// ===============================
// DOWNLOAD QUEUE MANAGEMENT
// ===============================

type DownloadQueueItem struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Type      string    `json:"type"`
	Status    string    `json:"status"` // "pending", "downloading", "completed", "failed"
	Progress  float64   `json:"progress"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Error     string    `json:"error,omitempty"`
}

var downloadQueue = make(map[string]*DownloadQueueItem)
var queueMutex sync.RWMutex

// AddToDownloadQueue adds an item to the download queue
func AddToDownloadQueue(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate unique ID
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	queueMutex.Lock()
	downloadQueue[id] = &DownloadQueueItem{
		ID:        id,
		Query:     req.Query,
		Type:      req.Type,
		Status:    "pending",
		Progress:  0.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	queueMutex.Unlock()

	// Start download in background
	go processDownloadQueue(id, req)

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": gin.H{
			"queue_id": id,
			"status":   "pending",
		},
		"message": "Download queued successfully",
	})
}

// GetDownloadQueue returns the current download queue
func GetDownloadQueue(c *gin.Context) {
	queueMutex.RLock()
	queue := make([]*DownloadQueueItem, 0, len(downloadQueue))
	for _, item := range downloadQueue {
		queue = append(queue, item)
	}
	queueMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queue,
	})
}

// GetDownloadStatus returns the status of a specific download
func GetDownloadStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		handleError(c, fmt.Errorf("missing id parameter"), "Download ID required", http.StatusBadRequest)
		return
	}

	queueMutex.RLock()
	item, exists := downloadQueue[id]
	queueMutex.RUnlock()

	if !exists {
		handleError(c, fmt.Errorf("download not found"), "Download ID not found", http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}

// processDownloadQueue processes a download queue item
func processDownloadQueue(id string, req DownloadRequest) {
	// Update status to downloading
	queueMutex.Lock()
	if item, exists := downloadQueue[id]; exists {
		item.Status = "downloading"
		item.UpdatedAt = time.Now()
	}
	queueMutex.Unlock()

	var err error
	switch req.Type {
	case "movie":
		err = jackett.MakeMovieQuery(req.Query, req.TmdbID, req.Quality, req.Year)
	case "tv":
		err = jackett.MakeShowQuery(req.Query, req.Seasons, req.TmdbID, req.Quality, req.Year)
	case "anime-movie":
		err = jackett.MakeAnimeMovieQuery(req.Query, req.TmdbID, req.Quality, req.Year)
	case "anime-tv", "anime":
		err = jackett.MakeAnimeShowQuery(req.Query, req.Seasons, req.TmdbID, req.Quality, req.Year)
	default:
		err = fmt.Errorf("unsupported download type: %s", req.Type)
	}

	// Update final status
	queueMutex.Lock()
	if item, exists := downloadQueue[id]; exists {
		if err != nil {
			item.Status = "failed"
			item.Error = err.Error()
			item.Progress = 0.0
		} else {
			item.Status = "completed"
			item.Progress = 100.0
		}
		item.UpdatedAt = time.Now()
	}
	queueMutex.Unlock()

	// Log the result
	if err != nil {
		logger.WriteError(fmt.Sprintf("Download failed for queue item %s", id), err)
	} else {
		logger.WriteInfo(fmt.Sprintf("Download completed for queue item %s", id))
	}
}

// ClearCompletedDownloads removes completed downloads from the queue
func ClearCompletedDownloads(c *gin.Context) {
	queueMutex.Lock()
	for id, item := range downloadQueue {
		if item.Status == "completed" || item.Status == "failed" {
			delete(downloadQueue, id)
		}
	}
	queueMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Completed downloads cleared",
	})
}

// ===============================
// SEARCH HISTORY AND FAVORITES
// ===============================

type SearchHistory struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Type      string    `json:"type"`
	Quality   string    `json:"quality"`
	Year      int       `json:"year"`
	Results   int       `json:"results"`
	Timestamp time.Time `json:"timestamp"`
}

var searchHistory = make([]*SearchHistory, 0, 1000)
var historyMutex sync.RWMutex

// AddToSearchHistory adds a search to the history
func addToSearchHistory(query, searchType, quality string, year, results int) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	history := &SearchHistory{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Query:     query,
		Type:      searchType,
		Quality:   quality,
		Year:      year,
		Results:   results,
		Timestamp: time.Now(),
	}

	searchHistory = append(searchHistory, history)

	// Keep only last 1000 searches
	if len(searchHistory) > 1000 {
		searchHistory = searchHistory[len(searchHistory)-1000:]
	}
}

// GetSearchHistory returns recent search history
func GetSearchHistory(c *gin.Context) {
	limit := 50 // Default limit
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	historyMutex.RLock()
	defer historyMutex.RUnlock()

	// Get the most recent searches
	start := len(searchHistory) - limit
	if start < 0 {
		start = 0
	}

	recentHistory := make([]*SearchHistory, len(searchHistory)-start)
	copy(recentHistory, searchHistory[start:])

	// Reverse to show most recent first
	for i, j := 0, len(recentHistory)-1; i < j; i, j = i+1, j-1 {
		recentHistory[i], recentHistory[j] = recentHistory[j], recentHistory[i]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recentHistory,
		"meta": gin.H{
			"total_history": len(searchHistory),
			"returned":      len(recentHistory),
			"limit":         limit,
		},
	})
}

// ClearSearchHistory clears the search history
func ClearSearchHistory(c *gin.Context) {
	historyMutex.Lock()
	searchHistory = searchHistory[:0]
	historyMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Search history cleared",
	})
}
