package jackett

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"high-seas/src/cache"
	"high-seas/src/deluge"
	"high-seas/src/logger"
	"high-seas/src/metrics"
	"high-seas/src/utils"
)

// SearchRequest represents a search request to Jackett
type SearchRequest struct {
	Query      string          `json:"query"`
	TmdbID     int             `json:"tmdb_id,omitempty"`
	Quality    string          `json:"quality,omitempty"`
	Year       int             `json:"year,omitempty"`
	Type       string          `json:"type,omitempty"`
	Seasons    []int           `json:"seasons,omitempty"`
	Categories []uint          `json:"categories,omitempty"`
	Context    context.Context `json:"-"`
}

// SearchResult represents a single search result from Jackett
type SearchResult struct {
	Result   TorrentResult `json:"result"`
	Tracker  string        `json:"tracker"`
	Category string        `json:"category"`
}

// TorrentResult represents torrent information
type TorrentResult struct {
	Title       string    `json:"title"`
	Size        int64     `json:"size"`
	Seeders     int       `json:"seeders"`
	Peers       int       `json:"peers"`
	PublishDate time.Time `json:"publish_date"`
	Link        string    `json:"link"`
	MagnetURI   string    `json:"magnet_uri"`
	InfoHash    string    `json:"info_hash"`
	Quality     string    `json:"quality"`
	Resolution  string    `json:"resolution"`
	Source      string    `json:"source"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query      string         `json:"query"`
	Results    []SearchResult `json:"results"`
	TotalFound int            `json:"total_found"`
	SearchTime time.Duration  `json:"search_time"`
	CacheHit   bool           `json:"cache_hit"`
	Trackers   []string       `json:"trackers"`
	Categories []uint         `json:"categories"`
}

// JackettConfig holds Jackett configuration
type JackettConfig struct {
	Host       string
	Port       string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
	UserAgent  string
}

// JackettClient represents an enhanced Jackett client
type JackettClient struct {
	Config     *JackettConfig
	HTTPClient *http.Client
	Cache      *cache.Cache
	Metrics    *metrics.Metrics
}

// Category constants
const (
	// Movie categories
	CategoryMoviesAll    = 2000
	CategoryMoviesSD     = 2030
	CategoryMoviesHD     = 2040
	CategoryMovies3D     = 2050
	CategoryMovies4K     = 2060
	CategoryMoviesBluRay = 2070

	// TV categories
	CategoryTVAll   = 5000
	CategoryTVSD    = 5030
	CategoryTVHD    = 5040
	CategoryTV4K    = 5045
	CategoryTVAnime = 5070

	// Anime categories
	CategoryAnimeAll = 5070
)

var (
	defaultConfig *JackettConfig
	configOnce    sync.Once
)

// GetDefaultConfig returns the default Jackett configuration
func GetDefaultConfig() *JackettConfig {
	configOnce.Do(func() {
		defaultConfig = &JackettConfig{
			Host:       utils.EnvVar("JACKETT_IP", "localhost"),
			Port:       utils.EnvVar("JACKETT_PORT", "9117"),
			APIKey:     utils.EnvVar("JACKETT_API_KEY", ""),
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			UserAgent:  "High-Seas/2.0",
		}
	})
	return defaultConfig
}

// NewEnhancedJackettClient creates a new enhanced Jackett client
func NewEnhancedJackettClient() *JackettClient {
	config := GetDefaultConfig()

	return &JackettClient{
		Config: config,
		HTTPClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 5,
			},
		},
		Cache:   cache.GetGlobalCache(),
		Metrics: metrics.GetGlobalMetrics(),
	}
}

// Search performs a search using the enhanced Jackett client
func (jc *JackettClient) Search(req *SearchRequest) (*SearchResponse, error) {
	start := time.Now()

	// Validate request
	if err := jc.validateSearchRequest(req); err != nil {
		jc.Metrics.IncrementFailedSearches()
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	// Check cache first
	cacheKey := jc.generateCacheKey(req)
	if cached, exists := jc.Cache.Get(cacheKey); exists {
		if response, ok := cached.(*SearchResponse); ok {
			response.CacheHit = true
			jc.Metrics.IncrementCacheHits()
			return response, nil
		}
	}
	jc.Metrics.IncrementCacheMisses()

	// Perform search
	response, err := jc.performSearch(req)
	if err != nil {
		jc.Metrics.IncrementFailedSearches()
		return nil, err
	}

	// Cache the result
	response.SearchTime = time.Since(start)
	response.CacheHit = false
	jc.Cache.SetWithTTL(cacheKey, response, 15*time.Minute)

	// Update metrics
	jc.Metrics.IncrementSuccessfulSearches()
	jc.Metrics.RecordSearchTime(response.SearchTime)
	jc.Metrics.RecordQuery(req.Query)
	jc.Metrics.RecordQuality(req.Quality)
	jc.Metrics.RecordType(req.Type)

	return response, nil
}

// ConcurrentSearch performs multiple searches concurrently
func (jc *JackettClient) ConcurrentSearch(requests []*SearchRequest) ([]*SearchResponse, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no search requests provided")
	}

	maxConcurrent := utils.EnvVarInt("CONCURRENT_SEARCHES", 3)
	if len(requests) < maxConcurrent {
		maxConcurrent = len(requests)
	}

	// Create semaphore to limit concurrent searches
	semaphore := make(chan struct{}, maxConcurrent)
	results := make([]*SearchResponse, len(requests))
	errors := make([]error, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *SearchRequest) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			response, err := jc.Search(request)
			results[index] = response
			errors[index] = err
		}(i, req)
	}

	wg.Wait()

	// Check for any errors
	var errorStrings []string
	successfulResults := make([]*SearchResponse, 0, len(requests))

	for i, err := range errors {
		if err != nil {
			errorStrings = append(errorStrings, fmt.Sprintf("Request %d: %v", i, err))
		} else if results[i] != nil {
			successfulResults = append(successfulResults, results[i])
		}
	}

	if len(errorStrings) > 0 && len(successfulResults) == 0 {
		return nil, fmt.Errorf("all searches failed: %s", strings.Join(errorStrings, "; "))
	}

	return successfulResults, nil
}

// performSearch executes the actual search against Jackett
func (jc *JackettClient) performSearch(req *SearchRequest) (*SearchResponse, error) {
	searchURL := jc.buildSearchURL(req)

	var lastErr error
	for attempt := 0; attempt < jc.Config.MaxRetries; attempt++ {
		response, err := jc.executeSearch(searchURL, req.Context)
		if err == nil {
			return response, nil
		}

		lastErr = err
		if attempt < jc.Config.MaxRetries-1 {
			backoff := time.Duration(attempt+1) * 2 * time.Second
			logger.WriteWarning(fmt.Sprintf("Search attempt %d failed, retrying in %v: %v",
				attempt+1, backoff, err))
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("search failed after %d attempts: %w", jc.Config.MaxRetries, lastErr)
}

// executeSearch performs the HTTP request to Jackett
func (jc *JackettClient) executeSearch(searchURL string, ctx context.Context) (*SearchResponse, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), jc.Config.Timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", jc.Config.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := jc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jackett returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return jc.parseJackettResponse(body)
}

// parseJackettResponse parses the raw Jackett API response
func (jc *JackettClient) parseJackettResponse(body []byte) (*SearchResponse, error) {
	var jackettResp struct {
		Results []struct {
			Title        string `json:"Title"`
			Size         int64  `json:"Size"`
			Seeders      *int   `json:"Seeders"`
			Peers        *int   `json:"Peers"`
			PublishDate  string `json:"PublishDate"`
			Link         string `json:"Link"`
			MagnetUri    string `json:"MagnetUri"`
			InfoHash     string `json:"InfoHash"`
			Tracker      string `json:"Tracker"`
			CategoryDesc string `json:"CategoryDesc"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(body, &jackettResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := &SearchResponse{
		Results:    make([]SearchResult, 0, len(jackettResp.Results)),
		TotalFound: len(jackettResp.Results),
		Trackers:   make([]string, 0),
	}

	trackerSet := make(map[string]bool)

	for _, result := range jackettResp.Results {
		// Parse publish date
		var publishDate time.Time
		if result.PublishDate != "" {
			if parsed, err := time.Parse(time.RFC3339, result.PublishDate); err == nil {
				publishDate = parsed
			}
		}

		// Handle nil pointers for seeders and peers
		seeders := 0
		if result.Seeders != nil {
			seeders = *result.Seeders
		}

		peers := 0
		if result.Peers != nil {
			peers = *result.Peers
		}

		searchResult := SearchResult{
			Result: TorrentResult{
				Title:       result.Title,
				Size:        result.Size,
				Seeders:     seeders,
				Peers:       peers,
				PublishDate: publishDate,
				Link:        result.Link,
				MagnetURI:   result.MagnetUri,
				InfoHash:    result.InfoHash,
				Quality:     jc.extractQuality(result.Title),
				Resolution:  jc.extractResolution(result.Title),
				Source:      jc.extractSource(result.Title),
			},
			Tracker:  result.Tracker,
			Category: result.CategoryDesc,
		}

		response.Results = append(response.Results, searchResult)

		// Track unique trackers
		if !trackerSet[result.Tracker] {
			trackerSet[result.Tracker] = true
			response.Trackers = append(response.Trackers, result.Tracker)
		}
	}

	return response, nil
}

// buildSearchURL constructs the search URL for Jackett
func (jc *JackettClient) buildSearchURL(req *SearchRequest) string {
	baseURL := fmt.Sprintf("http://%s:%s/api/v2.0/indexers/all/results",
		jc.Config.Host, jc.Config.Port)

	params := url.Values{}
	params.Set("apikey", jc.Config.APIKey)
	params.Set("Query", req.Query)

	// Add categories
	if len(req.Categories) > 0 {
		var categoryStrs []string
		for _, cat := range req.Categories {
			categoryStrs = append(categoryStrs, strconv.Itoa(int(cat)))
		}
		params.Set("Category", strings.Join(categoryStrs, ","))
	}

	// Add IMDB ID if available
	if req.TmdbID > 0 {
		params.Set("imdbid", strconv.Itoa(req.TmdbID))
	}

	return baseURL + "?" + params.Encode()
}

// validateSearchRequest validates a search request
func (jc *JackettClient) validateSearchRequest(req *SearchRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if strings.TrimSpace(req.Query) == "" && req.TmdbID == 0 {
		return fmt.Errorf("query or TMDB ID must be provided")
	}

	if req.Year < 0 || req.Year > time.Now().Year()+2 {
		return fmt.Errorf("invalid year: %d", req.Year)
	}

	return nil
}

// generateCacheKey generates a cache key for a search request
func (jc *JackettClient) generateCacheKey(req *SearchRequest) string {
	parts := []string{
		"jackett_search",
		req.Query,
		strconv.Itoa(req.TmdbID),
		req.Quality,
		strconv.Itoa(req.Year),
		req.Type,
	}

	// Add categories
	if len(req.Categories) > 0 {
		var catStrs []string
		for _, cat := range req.Categories {
			catStrs = append(catStrs, strconv.Itoa(int(cat)))
		}
		parts = append(parts, strings.Join(catStrs, ","))
	}

	// Add seasons
	if len(req.Seasons) > 0 {
		var seasonStrs []string
		for _, season := range req.Seasons {
			seasonStrs = append(seasonStrs, strconv.Itoa(season))
		}
		parts = append(parts, strings.Join(seasonStrs, ","))
	}

	return strings.Join(parts, ":")
}

// extractQuality extracts quality information from title
func (jc *JackettClient) extractQuality(title string) string {
	titleLower := strings.ToLower(title)

	qualityMap := map[string]string{
		"2160p": "2160p",
		"4k":    "2160p",
		"1080p": "1080p",
		"720p":  "720p",
		"480p":  "480p",
		"360p":  "360p",
	}

	for key, quality := range qualityMap {
		if strings.Contains(titleLower, key) {
			return quality
		}
	}

	return "Unknown"
}

// extractResolution extracts resolution from title
func (jc *JackettClient) extractResolution(title string) string {
	titleLower := strings.ToLower(title)

	if strings.Contains(titleLower, "4k") || strings.Contains(titleLower, "2160p") {
		return "4K"
	} else if strings.Contains(titleLower, "1080p") {
		return "1080p"
	} else if strings.Contains(titleLower, "720p") {
		return "720p"
	} else if strings.Contains(titleLower, "480p") {
		return "480p"
	}

	return "Unknown"
}

// extractSource extracts source type from title
func (jc *JackettClient) extractSource(title string) string {
	titleLower := strings.ToLower(title)

	sources := map[string]string{
		"bluray":  "BluRay",
		"blu-ray": "BluRay",
		"brrip":   "BluRay",
		"bdrip":   "BluRay",
		"webrip":  "WEB-DL",
		"web-dl":  "WEB-DL",
		"webdl":   "WEB-DL",
		"hdtv":    "HDTV",
		"pdtv":    "PDTV",
		"dvdrip":  "DVD",
		"dvdscr":  "DVDScr",
		"cam":     "CAM",
		"ts":      "Telesync",
		"tc":      "Telecine",
	}

	for key, source := range sources {
		if strings.Contains(titleLower, key) {
			return source
		}
	}

	return "Unknown"
}

// Category helper functions

// GetMovieCategories returns categories for movies
func GetMovieCategories() []uint {
	return []uint{CategoryMoviesAll, CategoryMoviesHD, CategoryMovies4K}
}

// GetTVCategories returns categories for TV shows
func GetTVCategories() []uint {
	return []uint{CategoryTVAll, CategoryTVHD, CategoryTV4K}
}

// GetAnimeMovieCategories returns categories for anime movies
func GetAnimeMovieCategories() []uint {
	return []uint{CategoryMoviesAll, CategoryAnimeAll}
}

// GetAnimeSeriesCategories returns categories for anime series
func GetAnimeSeriesCategories() []uint {
	return []uint{CategoryTVAll, CategoryAnimeAll}
}

// GetCategoriesForType returns appropriate categories for content type
func GetCategoriesForType(contentType string) []uint {
	switch strings.ToLower(contentType) {
	case "movie":
		return GetMovieCategories()
	case "tv", "show":
		return GetTVCategories()
	case "anime-movie":
		return GetAnimeMovieCategories()
	case "anime-tv", "anime":
		return GetAnimeSeriesCategories()
	default:
		return []uint{CategoryMoviesAll, CategoryTVAll}
	}
}

// Legacy API compatibility functions

// MakeMovieQuery performs a movie search (legacy compatibility)
func MakeMovieQuery(query string, tmdbID int, quality string, year int) error {
	client := NewEnhancedJackettClient()

	req := &SearchRequest{
		Query:      query,
		TmdbID:     tmdbID,
		Quality:    quality,
		Year:       year,
		Type:       "movie",
		Categories: GetMovieCategories(),
		Context:    context.Background(),
	}

	response, err := client.Search(req)
	if err != nil {
		return fmt.Errorf("movie search failed: %w", err)
	}

	// Send best result to Deluge if available
	if len(response.Results) > 0 {
		bestResult := findBestResult(response.Results)
		return sendToDeluge(bestResult.Result)
	}

	return fmt.Errorf("no results found for movie: %s", query)
}

// MakeShowQuery performs a TV show search (legacy compatibility)
func MakeShowQuery(query string, seasons []int, tmdbID int, quality string, year int) error {
	client := NewEnhancedJackettClient()
	logger.WriteInfo(fmt.Sprintf("Searching for TV show: %s with seasons: %v", query, seasons))

	if len(seasons) == 0 {
		return fmt.Errorf("no seasons specified for TV show: %s", query)
	}

	// Keep track of successful downloads
	successCount := 0
	var lastError error

	// Special handling for shows that might need different naming conventions
	alternateShowNames := map[string][]string{
		"Game of Thrones":  {"GoT"},
		"The Walking Dead": {"TWD"},
		"Breaking Bad":     {"Breaking.Bad"},
		// Add other shows with alternate naming patterns
	}

	// Search for each season separately
	for _, season := range seasons {
		// Try different query formats for better search results
		queryFormats := []string{
			fmt.Sprintf("%s Season %d", query, season),
			fmt.Sprintf("%s S%02d", query, season),
		}

		// Add alternate names if available for this show
		if alternateNames, hasAlternates := alternateShowNames[query]; hasAlternates {
			for _, altName := range alternateNames {
				queryFormats = append(queryFormats,
					fmt.Sprintf("%s Season %d", altName, season),
					fmt.Sprintf("%s S%02d", altName, season))
			}
		}

		// Try each query format until we find results
		seasonSuccess := false

		// For torrent sites that prefer very simple search terms, add these top priority search patterns
		// These should be tried first as they have the highest success rate
		priorityFormats := []string{}

		if query == "Naruto" && season == 3 {
			// Very simple terms that torrent sites handle best
			priorityFormats = []string{
				"Naruto 51",       // Single episode number
				"Naruto",          // Just the show name
				"Naruto 51-75",    // Episode range
				"Naruto Season 3", // Season
				"Naruto S03",      // Short season
				"Naruto Chunin",   // Arc name
			}

			// Try these priority formats first
			for _, format := range priorityFormats {
				logger.WriteInfo(fmt.Sprintf("Trying priority search term: %s", format))

				req := &SearchRequest{
					Query:      format,
					TmdbID:     tmdbID,
					Quality:    quality,
					Year:       year,
					Type:       "anime-tv",
					Seasons:    []int{season},
					Categories: GetAnimeSeriesCategories(),
					Context:    context.Background(),
				}

				response, err := client.Search(req)
				if err != nil {
					lastError = fmt.Errorf("anime show search failed for season %d with query '%s': %w",
						season, format, err)
					logger.WriteError(fmt.Sprintf("Search failed for %s", format), err)
					continue
				}

				// Send best result to Deluge if available
				if len(response.Results) > 0 {
					bestResult := findBestResult(response.Results)
					err = sendToDeluge(bestResult.Result)
					if err != nil {
						lastError = fmt.Errorf("failed to download season %d with query '%s': %w",
							season, format, err)
						logger.WriteError(fmt.Sprintf("Failed to send to Deluge: %s", format), err)
						continue
					}

					logger.WriteInfo(fmt.Sprintf("Successfully added to download queue using priority format: %s", format))
					successCount++
					seasonSuccess = true
					break // Found a successful result for this season, move to the next
				}
			}
		}

		// If priority formats didn't work, try all the other formats
		if !seasonSuccess {
			for _, seasonQuery := range queryFormats {
				logger.WriteInfo(fmt.Sprintf("Searching for: %s", seasonQuery))

				req := &SearchRequest{
					Query:      seasonQuery,
					TmdbID:     tmdbID,
					Quality:    quality,
					Year:       year,
					Type:       "tv",
					Seasons:    []int{season},
					Categories: GetTVCategories(),
					Context:    context.Background(),
				}

				response, err := client.Search(req)
				if err != nil {
					lastError = fmt.Errorf("TV show search failed for season %d with query '%s': %w",
						season, seasonQuery, err)
					logger.WriteError(fmt.Sprintf("Search failed for %s", seasonQuery), err)
					continue
				}

				// Send best result to Deluge if available
				if len(response.Results) > 0 {
					bestResult := findBestResult(response.Results)
					err = sendToDeluge(bestResult.Result)
					if err != nil {
						lastError = fmt.Errorf("failed to download season %d with query '%s': %w",
							season, seasonQuery, err)
						logger.WriteError(fmt.Sprintf("Failed to send to Deluge: %s", seasonQuery), err)
						continue
					}

					logger.WriteInfo(fmt.Sprintf("Successfully added to download queue: %s", seasonQuery))
					successCount++
					seasonSuccess = true
					break // Found a successful result for this season, move to the next
				} else {
					lastError = fmt.Errorf("no results found for TV show with query: %s", seasonQuery)
					logger.WriteWarning(fmt.Sprintf("No results found for: %s", seasonQuery))
					// Continue trying other formats
				}
			}
		}

		// If we've tried all formats and still no success for this season
		if !seasonSuccess {
			logger.WriteWarning(fmt.Sprintf("All search attempts failed for season %d of %s", season, query))
			// Continue to the next season
		}
	}

	// If at least one season was found and downloaded, consider it a success
	if successCount > 0 {
		return nil
	}

	// Return the last error if all searches failed
	if lastError != nil {
		return lastError
	}

	return fmt.Errorf("no results found for TV show: %s", query)
}

// MakeAnimeMovieQuery performs an anime movie search (legacy compatibility)
func MakeAnimeMovieQuery(query string, tmdbID int, quality string, year int) error {
	client := NewEnhancedJackettClient()

	req := &SearchRequest{
		Query:      query,
		TmdbID:     tmdbID,
		Quality:    quality,
		Year:       year,
		Type:       "anime-movie",
		Categories: GetAnimeMovieCategories(),
		Context:    context.Background(),
	}

	response, err := client.Search(req)
	if err != nil {
		return fmt.Errorf("anime movie search failed: %w", err)
	}

	// Send best result to Deluge if available
	if len(response.Results) > 0 {
		bestResult := findBestResult(response.Results)
		return sendToDeluge(bestResult.Result)
	}

	return fmt.Errorf("no results found for anime movie: %s", query)
}

// MakeAnimeShowQuery performs an anime show search (legacy compatibility)
func MakeAnimeShowQuery(query string, seasons []int, tmdbID int, quality string, year int) error {
	client := NewEnhancedJackettClient()
	logger.WriteInfo(fmt.Sprintf("Searching for anime show: %s with seasons: %v", query, seasons))

	if len(seasons) == 0 {
		return fmt.Errorf("no seasons specified for anime show: %s", query)
	}

	// Keep track of successful downloads
	successCount := 0
	var lastError error

	// For certain popular anime, we need special handling
	alternateNames := map[string]string{
		"Demon Slayer":        "Kimetsu no Yaiba",
		"Attack on Titan":     "Shingeki no Kyojin",
		"My Hero Academia":    "Boku no Hero Academia",
		"Food Wars":           "Shokugeki no Soma",
		"Jujutsu Kaisen":      "Jujutsu Kaisen",
		"Tokyo Ghoul":         "Tokyo Ghoul",
		"One Punch Man":       "One Punch Man",
		"Fullmetal Alchemist": "Fullmetal Alchemist",
		"Hunter x Hunter":     "Hunter x Hunter",
		"Black Clover":        "Black Clover",
		"Dragon Ball":         "Dragon Ball",
		"Dragon Ball Z":       "Dragon Ball Z",
		"Naruto":              "Naruto",
		"Naruto Shippuden":    "Naruto Shippuden",
		"One Piece":           "One Piece",
		"Bleach":              "Bleach",
		"Death Note":          "Death Note",
		"Sword Art Online":    "Sword Art Online",
		// Add other anime with alternate names as needed
	}

	// Season to episode mappings (used for nyaa.si search patterns)
	seasonEpisodeMap := map[string]map[int][]int{
		"Naruto": {
			1: {1, 25},    // Season 1: Episodes 1-25
			2: {26, 50},   // Season 2: Episodes 26-50
			3: {51, 75},   // Season 3: Episodes 51-75
			4: {76, 100},  // Season 4: Episodes 76-100
			5: {101, 125}, // Season 5: Episodes 101-125
		},
		"One Piece": {
			1: {1, 30},    // Season 1: Episodes 1-30
			2: {31, 60},   // Season 2: Episodes 31-60
			3: {61, 90},   // Season 3: Episodes 61-90
			4: {91, 120},  // Season 4: Episodes 91-120
			5: {121, 150}, // Season 5: Episodes 121-150
		},
		"Dragon Ball Z": {
			1: {1, 39},    // Season 1: Episodes 1-39
			2: {40, 74},   // Season 2: Episodes 40-74
			3: {75, 107},  // Season 3: Episodes 75-107
			4: {108, 139}, // Season 4: Episodes 108-139
		},
		"Bleach": {
			1: {1, 20},   // Season 1: Episodes 1-20
			2: {21, 40},  // Season 2: Episodes 21-40
			3: {41, 60},  // Season 3: Episodes 41-60
			4: {61, 80},  // Season 4: Episodes 61-80
			5: {81, 100}, // Season 5: Episodes 81-100
		},
	}

	// Special season naming conventions for specific anime series
	specialSeasonFormats := map[string]map[int][]string{
		"Demon Slayer": {
			1: {"Demon Slayer", "Kimetsu no Yaiba"},
			2: {"Demon Slayer Entertainment District", "Kimetsu no Yaiba Entertainment District", "Kimetsu no Yaiba Yuukaku-hen"},
			3: {"Demon Slayer Swordsmith Village", "Kimetsu no Yaiba Swordsmith Village", "Kimetsu no Yaiba Katanakaji"},
		},
		"Attack on Titan": {
			1: {"Attack on Titan", "Shingeki no Kyojin"},
			2: {"Attack on Titan Season 2", "Shingeki no Kyojin Season 2"},
			3: {"Attack on Titan Season 3", "Shingeki no Kyojin Season 3"},
			4: {"Attack on Titan Final Season", "Shingeki no Kyojin Final Season", "Attack on Titan Season 4", "Shingeki no Kyojin Season 4"},
		},
		"Tokyo Ghoul": {
			1: {"Tokyo Ghoul"},
			2: {"Tokyo Ghoul √A", "Tokyo Ghoul Root A", "Tokyo Ghoul Season 2"},
			3: {"Tokyo Ghoul:re", "Tokyo Ghoul re", "Tokyo Ghoul Season 3"},
		},
		"Sword Art Online": {
			1: {"Sword Art Online"},
			2: {"Sword Art Online II", "Sword Art Online 2", "Sword Art Online Season 2"},
			3: {"Sword Art Online Alicization", "Sword Art Online Season 3"},
		},
		"Naruto": {
			1: {"Naruto Season 1", "Naruto S01", "Naruto 001-025", "Naruto Ep 001-025", "Naruto Episode 1-25", "Naruto 1-25", "Naruto Episodes 1-25"},
			2: {"Naruto Season 2", "Naruto S02", "Naruto 026-050", "Naruto Ep 026-050", "Naruto Episode 26-50", "Naruto 26-50", "Naruto Episodes 26-50"},
			3: {"Naruto Season 3", "Naruto S03", "Naruto 051-075", "Naruto Ep 051-075", "Naruto Episode 51-75", "Naruto 51-75", "Naruto Episodes 51-75"},
			4: {"Naruto Season 4", "Naruto S04", "Naruto 076-100", "Naruto Ep 076-100", "Naruto Episode 76-100", "Naruto 76-100", "Naruto Episodes 76-100"},
			5: {"Naruto Season 5", "Naruto S05", "Naruto 101-125", "Naruto Ep 101-125", "Naruto Episode 101-125", "Naruto 101-125", "Naruto Episodes 101-125"},
		},
		"Naruto Shippuden": {
			1: {"Naruto Shippuden Season 1", "Naruto Shippuden S01", "Naruto Shippuden 001-025", "Naruto Shippuden Ep 1-25", "Naruto Shippuden 1-25", "Naruto Shippuuden 1-25"},
			2: {"Naruto Shippuden Season 2", "Naruto Shippuden S02", "Naruto Shippuden 026-050", "Naruto Shippuden Ep 26-50", "Naruto Shippuden 26-50", "Naruto Shippuuden 26-50"},
			3: {"Naruto Shippuden Season 3", "Naruto Shippuden S03", "Naruto Shippuden 051-075", "Naruto Shippuden Ep 51-75", "Naruto Shippuden 51-75", "Naruto Shippuuden 51-75"},
			4: {"Naruto Shippuden Season 4", "Naruto Shippuden S04", "Naruto Shippuden 076-100", "Naruto Shippuden Ep 76-100", "Naruto Shippuden 76-100", "Naruto Shippuuden 76-100"},
			5: {"Naruto Shippuden Season 5", "Naruto Shippuden S05", "Naruto Shippuden 101-125", "Naruto Shippuden Ep 101-125", "Naruto Shippuden 101-125", "Naruto Shippuuden 101-125"},
		},
		"One Piece": {
			1: {"One Piece Season 1", "One Piece S01", "One Piece 001-030", "One Piece Ep 1-30", "One Piece Episode 1-30", "One Piece 1-30", "One Piece Episodes 1-30"},
			2: {"One Piece Season 2", "One Piece S02", "One Piece 031-060", "One Piece Ep 31-60", "One Piece Episode 31-60", "One Piece 31-60", "One Piece Episodes 31-60"},
			3: {"One Piece Season 3", "One Piece S03", "One Piece 061-090", "One Piece Ep 61-90", "One Piece Episode 61-90", "One Piece 61-90", "One Piece Episodes 61-90"},
			4: {"One Piece Season 4", "One Piece S04", "One Piece 091-120", "One Piece Ep 91-120", "One Piece Episode 91-120", "One Piece 91-120", "One Piece Episodes 91-120"},
			5: {"One Piece Season 5", "One Piece S05", "One Piece 121-150", "One Piece Ep 121-150", "One Piece Episode 121-150", "One Piece 121-150", "One Piece Episodes 121-150"},
		},
		"Dragon Ball Z": {
			1: {"Dragon Ball Z Season 1", "DBZ Season 1", "DBZ S01", "Dragon Ball Z Saiyan Saga", "DBZ 1-39", "DBZ Episodes 1-39", "Dragon Ball Z 1-39", "Dragon Ball Z Episodes 1-39"},
			2: {"Dragon Ball Z Season 2", "DBZ Season 2", "DBZ S02", "Dragon Ball Z Frieza Saga", "DBZ 40-74", "DBZ Episodes 40-74", "Dragon Ball Z 40-74", "Dragon Ball Z Episodes 40-74"},
			3: {"Dragon Ball Z Season 3", "DBZ Season 3", "DBZ S03", "Dragon Ball Z Cell Saga", "DBZ 75-107", "DBZ Episodes 75-107", "Dragon Ball Z 75-107", "Dragon Ball Z Episodes 75-107"},
			4: {"Dragon Ball Z Season 4", "DBZ Season 4", "DBZ S04", "Dragon Ball Z Buu Saga", "DBZ 108-139", "DBZ Episodes 108-139", "Dragon Ball Z 108-139", "Dragon Ball Z Episodes 108-139"},
		},
		"Bleach": {
			1: {"Bleach Season 1", "Bleach S01", "Bleach 001-020", "Bleach Ep 1-20", "Bleach Episode 1-20", "Bleach 1-20", "Bleach Episodes 1-20"},
			2: {"Bleach Season 2", "Bleach S02", "Bleach 021-040", "Bleach Ep 21-40", "Bleach Episode 21-40", "Bleach 21-40", "Bleach Episodes 21-40"},
			3: {"Bleach Season 3", "Bleach S03", "Bleach 041-060", "Bleach Ep 41-60", "Bleach Episode 41-60", "Bleach 41-60", "Bleach Episodes 41-60"},
			4: {"Bleach Season 4", "Bleach S04", "Bleach 061-080", "Bleach Ep 61-80", "Bleach Episode 61-80", "Bleach 61-80", "Bleach Episodes 61-80"},
			5: {"Bleach Season 5", "Bleach S05", "Bleach 081-100", "Bleach Ep 81-100", "Bleach Episode 81-100", "Bleach 81-100", "Bleach Episodes 81-100"},
		},
	}

	// Check if we have an alternate name for this anime
	alternateName, hasAlternate := alternateNames[query]

	// Search for each season separately
	for _, season := range seasons {
		// Try different query formats for better search results
		queryFormats := []string{
			fmt.Sprintf("%s Season %d", query, season),
			fmt.Sprintf("%s S%02d", query, season),
		}

		// Add more specific formats for some popular anime
		if hasAlternate {
			queryFormats = append(queryFormats,
				fmt.Sprintf("%s Season %d", alternateName, season),
				fmt.Sprintf("%s S%02d", alternateName, season))
		}

		// Check if this anime has special season naming
		if specialFormats, hasSpecial := specialSeasonFormats[query]; hasSpecial {
			// Check if we have special formats for this specific season
			if seasonFormats, hasSeasonFormat := specialFormats[season]; hasSeasonFormat {
				queryFormats = append(queryFormats, seasonFormats...)
			}
		}

		// Log all the query formats we're going to try
		logger.WriteInfo(fmt.Sprintf("Will try %d different search patterns for %s Season %d",
			len(queryFormats), query, season))

		// Add some special case patterns for certain anime that might use different naming schemes
		if query == "Naruto" {
			// Try adding common arc names for Naruto
			arcs := map[int]string{
				1: "Introduction",
				2: "Land of Waves",
				3: "Chunin Exam",
				4: "Konoha Crush",
				5: "Search for Tsunade",
			}
			if arcName, hasArc := arcs[season]; hasArc {
				queryFormats = append(queryFormats,
					fmt.Sprintf("Naruto %s Arc", arcName),
					fmt.Sprintf("Naruto %s", arcName))
			}

			// For Naruto Season 3 specifically (episodes 51-75), add very specific, real-world torrent naming patterns
			if season == 3 {
				// Common real-world torrent naming patterns for Naruto season 3
				narutoS3Patterns := []string{
					// Patterns without fansub groups
					"Naruto 51-75",
					"Naruto Episodes 51-75",
					"Naruto Ep 51-75",
					"Naruto 051-075",
					"Naruto.051-075",
					"Naruto_051-075",
					"Naruto.51-75",
					"Naruto_51-75",
					"Naruto.E51-E75",
					"Naruto_E51-E75",

					// Single episodes (try key episodes)
					"Naruto 51",
					"Naruto 60",
					"Naruto 65",
					"Naruto 75",

					// Raw episode number without season
					"Naruto - 051",
					"Naruto - 060",
					"Naruto - 065",
					"Naruto - 075",

					// Just raw numbers which are sometimes used in torrents
					"051", "060", "065", "075",

					// With common fansub groups
					"[Judas] Naruto 51-75",
					"[DBNC] Naruto 51-75",
					"[AnimeKaizoku] Naruto 51-75",
					"[AnimeRG] Naruto 51-75",
					"[Horrible] Naruto 51-75",
					"[DB] Naruto 51-75",

					// Different fansub group formats
					"Naruto [51-75]",
					"Naruto (51-75)",
					"Naruto.S03.Complete",
					"Naruto.Season.3.Complete",
					"Naruto Season 3 Complete",
					"Naruto.S03.720p",
					"Naruto.S03.1080p",

					// Dual-Audio options
					"Naruto 51-75 Dual Audio",
					"Naruto 51-75 [Dual Audio]",
					"Naruto S03 Dual Audio",

					// Complete collections that include this season
					"Naruto Complete",
					"Naruto 1-220",
					"Naruto Dual Audio",

					// Other quality markers
					"Naruto 51-75 720p",
					"Naruto 51-75 1080p",
					"Naruto 51-75 480p",
					"Naruto S03 720p",
					"Naruto S03 480p",

					// Specific known episode titles
					"Naruto A Special Report",         // Episode 53
					"Naruto The Summoning Jutsu",      // Episode 58
					"Naruto Byakugan vs Shadow Clone", // Episode 61

					// XDCC specific patterns (some torrent sites index XDCC)
					"XDCC Naruto 51-75",
					"XDCC Naruto S03",

					// Japanese patterns
					"ナルト 51-75",
					"ナルト 第51-75話",
				}

				queryFormats = append(queryFormats, narutoS3Patterns...)
			}

			// Add additional search terms for Naruto by episode groupings
			// These are common episode groupings found on torrent sites
			episodeRanges := map[int][]string{
				3: {"Naruto 051", "Naruto 051-075", "Naruto ep 51", "Naruto ep 51-75",
					"Naruto episode 51", "Naruto episodes 51-75",
					"Naruto ep51", "Naruto ep51-75",
					"Naruto.S03", "Naruto.S3", "Naruto.Season.3",
					"[Naruto].S03", "[Naruto].S3",
					"Naruto E51-E75", "Naruto.E51-E75"},
				// Add more common patterns for other seasons if needed
			}

			if ranges, hasRanges := episodeRanges[season]; hasRanges {
				queryFormats = append(queryFormats, ranges...)
			}

			// Also add simple numeric searches which sometimes work better
			queryFormats = append(queryFormats,
				fmt.Sprintf("Naruto %d", season),
				fmt.Sprintf("Naruto - %d", season))

			// Additional variants to try
			if season == 3 {
				queryFormats = append(queryFormats,
					"Naruto 51", "Naruto 51-75",
					"Naruto - 51-75", "Naruto.051-075",
					"Naruto_051-075", "Naruto_51-75",
					"Naruto_S03", "Naruto_Season_3")
			}
		} else if query == "Dragon Ball Z" {
			// Try with standard abbreviation
			queryFormats = append(queryFormats,
				fmt.Sprintf("DBZ S%02d", season),
				fmt.Sprintf("DBZ Season %d", season))
		} else if query == "One Piece" {
			// One Piece sometimes uses arc names instead of seasons
			arcs := map[int][]string{
				1: {"East Blue", "Romance Dawn", "Orange Town", "Syrup Village", "Baratie", "Arlong Park"},
				2: {"Loguetown", "Reverse Mountain", "Whiskey Peak", "Little Garden", "Drum Island"},
				3: {"Arabasta", "Alabasta"},
				4: {"Jaya", "Skypiea"},
				5: {"Long Ring Long Land", "Water 7"},
			}
			if arcNames, hasArc := arcs[season]; hasArc {
				for _, arcName := range arcNames {
					queryFormats = append(queryFormats,
						fmt.Sprintf("One Piece %s Arc", arcName),
						fmt.Sprintf("One Piece %s", arcName))
				}
			}
		}

		// Add nyaa.si specific patterns that focus on episode numbers
		if episodeRange, hasEpisodeMapping := seasonEpisodeMap[query]; hasEpisodeMapping {
			if epRange, hasRange := episodeRange[season]; hasRange {
				// Get the start and end episode numbers for this season
				startEp := epRange[0]
				endEp := epRange[1]

				// Select a subset of strategic episodes to try instead of all episodes
				// For example, the first, middle, and last episodes of the season
				keyEpisodes := []int{
					startEp,                       // First episode
					startEp + (endEp-startEp)/3,   // About 1/3 through
					startEp + 2*(endEp-startEp)/3, // About 2/3 through
					endEp,                         // Last episode
				}

				// Add a few specific episodes we know are popular/well-seeded
				if query == "Naruto" && season == 3 {
					// Specific popular episodes in Naruto Season 3
					keyEpisodes = append(keyEpisodes, 59, 61, 65)
				}

				// Try with top fansub groups only
				topGroups := []string{
					"Subsplease", "Horrible", "Erai-raws",
					"Judas", "DBNC", "AnimeKaizoku", "AnimeRG", "DB",
					"", // Also try without a group
				}

				// Add raw episode patterns (most common on nyaa.si)
				for _, ep := range keyEpisodes {
					queryFormats = append(queryFormats,
						// Most common pattern on nyaa.si
						fmt.Sprintf("%s - %02d", query, ep),
						// Just numbers sometimes works on nyaa.si
						fmt.Sprintf("%02d", ep),
						// XDCC style
						fmt.Sprintf("XDCC SEND %s %02d", query, ep))
				}

				// Try batch ranges in different formats
				batchPatterns := []string{
					fmt.Sprintf("%s %02d-%02d", query, startEp, endEp),
					fmt.Sprintf("%s - %02d-%02d", query, startEp, endEp),
					fmt.Sprintf("%s E%02d-E%02d", query, startEp, endEp),
					fmt.Sprintf("%s [%02d-%02d]", query, startEp, endEp),
					fmt.Sprintf("%s (%02d-%02d)", query, startEp, endEp),
					fmt.Sprintf("%s.%02d-%02d", query, startEp, endEp),
					fmt.Sprintf("%s_%02d-%02d", query, startEp, endEp),
				}
				queryFormats = append(queryFormats, batchPatterns...)

				// Try resolution-specific patterns (very common in torrents)
				resolutions := []string{"480p", "720p", "1080p"}
				for _, res := range resolutions {
					queryFormats = append(queryFormats,
						fmt.Sprintf("%s %02d-%02d %s", query, startEp, endEp, res),
						fmt.Sprintf("%s.S%02d.%s", query, season, res),
						fmt.Sprintf("%s - %02d-%02d %s", query, startEp, endEp, res))
				}

				// Try common encoders mentioned in torrents
				encoders := []string{"x264", "x265", "HEVC", "10bit", "BD", "WEB-DL", "WEBRip"}
				for _, encoder := range encoders {
					queryFormats = append(queryFormats,
						fmt.Sprintf("%s %02d-%02d %s", query, startEp, endEp, encoder))
				}

				// Try with common torrent site specific markers
				sourceMarkers := []string{
					"Batch", "Complete", "Dual Audio", "Dual-Audio",
					"Eng", "English", "Subbed", "Sub", "Dub", "Dubbed",
				}
				for _, marker := range sourceMarkers {
					queryFormats = append(queryFormats,
						fmt.Sprintf("%s %02d-%02d %s", query, startEp, endEp, marker))
				}

				// Common release group formats for batch episodes
				for _, group := range topGroups {
					if group != "" {
						queryFormats = append(queryFormats,
							fmt.Sprintf("[%s] %s - %02d-%02d", group, query, startEp, endEp),
							fmt.Sprintf("[%s] %s %02d-%02d", group, query, startEp, endEp),
							fmt.Sprintf("%s %s %02d-%02d", group, query, startEp, endEp))
					}
				}

				// Add Japanese title search patterns
				if hasAlternate && alternateName != query {
					queryFormats = append(queryFormats,
						fmt.Sprintf("%s %02d", alternateName, startEp),
						fmt.Sprintf("%s %02d-%02d", alternateName, startEp, endEp),
						fmt.Sprintf("%s - %02d", alternateName, startEp),
						fmt.Sprintf("%s - %02d-%02d", alternateName, startEp, endEp))
				}
			}
		}

		// Try each query format until we find results
		seasonSuccess := false

		// For torrent sites that prefer very simple search terms, add these top priority search patterns
		// These should be tried first as they have the highest success rate
		priorityFormats := []string{}

		if query == "Naruto" && season == 3 {
			// Very simple terms that torrent sites handle best
			priorityFormats = []string{
				"Naruto 51",       // Single episode number
				"Naruto",          // Just the show name
				"Naruto 51-75",    // Episode range
				"Naruto Season 3", // Season
				"Naruto S03",      // Short season
				"Naruto Chunin",   // Arc name
			}

			// Try these priority formats first
			for _, format := range priorityFormats {
				logger.WriteInfo(fmt.Sprintf("Trying priority search term: %s", format))

				req := &SearchRequest{
					Query:      format,
					TmdbID:     tmdbID,
					Quality:    quality,
					Year:       year,
					Type:       "anime-tv",
					Seasons:    []int{season},
					Categories: GetAnimeSeriesCategories(),
					Context:    context.Background(),
				}

				response, err := client.Search(req)
				if err != nil {
					lastError = fmt.Errorf("anime show search failed for season %d with query '%s': %w",
						season, format, err)
					logger.WriteError(fmt.Sprintf("Search failed for %s", format), err)
					continue
				}

				// Send best result to Deluge if available
				if len(response.Results) > 0 {
					bestResult := findBestResult(response.Results)
					err = sendToDeluge(bestResult.Result)
					if err != nil {
						lastError = fmt.Errorf("failed to download season %d with query '%s': %w",
							season, format, err)
						logger.WriteError(fmt.Sprintf("Failed to send to Deluge: %s", format), err)
						continue
					}

					logger.WriteInfo(fmt.Sprintf("Successfully added to download queue using priority format: %s", format))
					successCount++
					seasonSuccess = true
					break // Found a successful result for this season, move to the next
				}
			}
		}

		// If priority formats didn't work, try all the other formats
		if !seasonSuccess {
			for _, seasonQuery := range queryFormats {
				logger.WriteInfo(fmt.Sprintf("Searching for: %s", seasonQuery))

				req := &SearchRequest{
					Query:      seasonQuery,
					TmdbID:     tmdbID,
					Quality:    quality,
					Year:       year,
					Type:       "anime-tv",
					Seasons:    []int{season},
					Categories: GetAnimeSeriesCategories(),
					Context:    context.Background(),
				}

				response, err := client.Search(req)
				if err != nil {
					lastError = fmt.Errorf("anime show search failed for season %d with query '%s': %w",
						season, seasonQuery, err)
					logger.WriteError(fmt.Sprintf("Search failed for %s", seasonQuery), err)
					continue
				}

				// Send best result to Deluge if available
				if len(response.Results) > 0 {
					bestResult := findBestResult(response.Results)
					err = sendToDeluge(bestResult.Result)
					if err != nil {
						lastError = fmt.Errorf("failed to download season %d with query '%s': %w",
							season, seasonQuery, err)
						logger.WriteError(fmt.Sprintf("Failed to send to Deluge: %s", seasonQuery), err)
						continue
					}

					logger.WriteInfo(fmt.Sprintf("Successfully added to download queue: %s", seasonQuery))
					successCount++
					seasonSuccess = true
					break // Found a successful result for this season, move to the next
				} else {
					lastError = fmt.Errorf("no results found for anime show with query: %s", seasonQuery)
					logger.WriteWarning(fmt.Sprintf("No results found for: %s", seasonQuery))
					// Continue trying other formats
				}
			}
		}

		// If we've tried all formats and still no success for this season
		if !seasonSuccess {
			logger.WriteWarning(fmt.Sprintf("All search attempts failed for season %d of %s", season, query))
			// Continue to the next season
		}
	}

	// If at least one season was found and downloaded, consider it a success
	if successCount > 0 {
		return nil
	}

	// Return the last error if all searches failed
	if lastError != nil {
		return lastError
	}

	return fmt.Errorf("no results found for anime show: %s", query)
}

// Helper functions

// findBestResult finds the best torrent result based on seeders and quality
func findBestResult(results []SearchResult) SearchResult {
	if len(results) == 0 {
		return SearchResult{}
	}

	best := results[0]
	for _, result := range results[1:] {
		// Prefer higher seeders
		if result.Result.Seeders > best.Result.Seeders {
			best = result
			continue
		}

		// If same seeders, prefer better quality
		if result.Result.Seeders == best.Result.Seeders {
			if isHigherQuality(result.Result.Quality, best.Result.Quality) {
				best = result
			}
		}
	}

	return best
}

// isHigherQuality compares two quality strings
func isHigherQuality(quality1, quality2 string) bool {
	qualityOrder := map[string]int{
		"2160p": 4,
		"1080p": 3,
		"720p":  2,
		"480p":  1,
		"360p":  0,
	}

	q1, ok1 := qualityOrder[quality1]
	q2, ok2 := qualityOrder[quality2]

	if !ok1 || !ok2 {
		return false
	}

	return q1 > q2
}

// sendToDeluge sends a torrent to Deluge for download
func sendToDeluge(result TorrentResult) error {
	if result.MagnetURI == "" && result.Link == "" {
		return fmt.Errorf("no download link available")
	}

	// Try magnet first, then torrent file
	downloadURL := result.MagnetURI
	if downloadURL == "" {
		downloadURL = result.Link
	}

	logger.WriteInfo(fmt.Sprintf("Sending to Deluge: %s", result.Title))

	if err := deluge.AddTorrent(downloadURL); err != nil {
		return fmt.Errorf("failed to add torrent to Deluge: %w", err)
	}

	logger.WriteInfo(fmt.Sprintf("Successfully added to Deluge: %s", result.Title))
	return nil
}
