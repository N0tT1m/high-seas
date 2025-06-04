package metrics

import (
	"sync"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	mutex sync.RWMutex

	// Search metrics
	TotalSearches     int64         `json:"total_searches"`
	SuccessfulSearch  int64         `json:"successful_searches"`
	FailedSearches    int64         `json:"failed_searches"`
	AverageSearchTime time.Duration `json:"average_search_time"`
	TotalSearchTime   time.Duration `json:"total_search_time"`

	// Cache metrics
	CacheHits   int64 `json:"cache_hits"`
	CacheMisses int64 `json:"cache_misses"`

	// Download metrics
	TotalDownloads      int64 `json:"total_downloads"`
	SuccessfulDownloads int64 `json:"successful_downloads"`
	FailedDownloads     int64 `json:"failed_downloads"`

	// API metrics
	TotalRequests       int64         `json:"total_requests"`
	TotalErrors         int64         `json:"total_errors"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TotalResponseTime   time.Duration `json:"total_response_time"`

	// Service health
	JackettStatus   string    `json:"jackett_status"`
	DelugeStatus    string    `json:"deluge_status"`
	PlexStatus      string    `json:"plex_status"`
	LastHealthCheck time.Time `json:"last_health_check"`

	// Uptime
	StartTime time.Time `json:"start_time"`

	// Popular searches
	PopularQueries map[string]int64 `json:"popular_queries"`

	// Quality distribution
	QualityStats map[string]int64 `json:"quality_stats"`

	// Type distribution
	TypeStats map[string]int64 `json:"type_stats"`
}

// Stats represents formatted metrics for API responses
type Stats struct {
	// Search stats
	TotalSearches      int64   `json:"total_searches"`
	SuccessfulSearches int64   `json:"successful_searches"`
	FailedSearches     int64   `json:"failed_searches"`
	SearchSuccessRate  float64 `json:"search_success_rate"`
	AverageSearchTime  string  `json:"average_search_time"`

	// Cache stats
	CacheHits    int64   `json:"cache_hits"`
	CacheMisses  int64   `json:"cache_misses"`
	CacheHitRate float64 `json:"cache_hit_rate"`

	// Download stats
	TotalDownloads      int64   `json:"total_downloads"`
	SuccessfulDownloads int64   `json:"successful_downloads"`
	FailedDownloads     int64   `json:"failed_downloads"`
	DownloadSuccessRate float64 `json:"download_success_rate"`

	// API stats
	TotalRequests       int64   `json:"total_requests"`
	TotalErrors         int64   `json:"total_errors"`
	ErrorRate           float64 `json:"error_rate"`
	AverageResponseTime string  `json:"average_response_time"`

	// Service health
	Services        map[string]string `json:"services"`
	LastHealthCheck string            `json:"last_health_check"`

	// Uptime
	Uptime string `json:"uptime"`

	// Popular searches (top 10)
	PopularQueries []QueryStat `json:"popular_queries"`

	// Quality distribution
	QualityDistribution []StatItem `json:"quality_distribution"`

	// Type distribution
	TypeDistribution []StatItem `json:"type_distribution"`
}

// QueryStat represents a popular query statistic
type QueryStat struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

// StatItem represents a generic statistic item
type StatItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

var (
	globalMetrics *Metrics
	once          sync.Once
)

// New creates a new metrics instance
func New() *Metrics {
	return &Metrics{
		StartTime:      time.Now(),
		PopularQueries: make(map[string]int64),
		QualityStats:   make(map[string]int64),
		TypeStats:      make(map[string]int64),
		JackettStatus:  "unknown",
		DelugeStatus:   "unknown",
		PlexStatus:     "unknown",
	}
}

// GetGlobalMetrics returns the global metrics instance
func GetGlobalMetrics() *Metrics {
	once.Do(func() {
		globalMetrics = New()
	})
	return globalMetrics
}

// IncrementSearches increments the search counter
func (m *Metrics) IncrementSearches() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.TotalSearches++
}

// IncrementSuccessfulSearches increments the successful search counter
func (m *Metrics) IncrementSuccessfulSearches() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.SuccessfulSearch++
}

// IncrementFailedSearches increments the failed search counter
func (m *Metrics) IncrementFailedSearches() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.FailedSearches++
}

// RecordSearchTime records a search time
func (m *Metrics) RecordSearchTime(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.TotalSearchTime += duration
	if m.TotalSearches > 0 {
		m.AverageSearchTime = m.TotalSearchTime / time.Duration(m.TotalSearches)
	}
}

// IncrementCacheHits increments the cache hit counter
func (m *Metrics) IncrementCacheHits() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.CacheHits++
}

// IncrementCacheMisses increments the cache miss counter
func (m *Metrics) IncrementCacheMisses() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.CacheMisses++
}

// IncrementDownloads increments the download counter
func (m *Metrics) IncrementDownloads() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.TotalDownloads++
}

// IncrementSuccessfulDownloads increments the successful download counter
func (m *Metrics) IncrementSuccessfulDownloads() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.SuccessfulDownloads++
}

// IncrementFailedDownloads increments the failed download counter
func (m *Metrics) IncrementFailedDownloads() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.FailedDownloads++
}

// IncrementRequests increments the request counter
func (m *Metrics) IncrementRequests() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.TotalRequests++
}

// IncrementErrors increments the error counter
func (m *Metrics) IncrementErrors() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.TotalErrors++
}

// RecordResponseTime records a response time
func (m *Metrics) RecordResponseTime(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.TotalResponseTime += duration
	if m.TotalRequests > 0 {
		m.AverageResponseTime = m.TotalResponseTime / time.Duration(m.TotalRequests)
	}
}

// RecordQuery records a search query for popularity tracking
func (m *Metrics) RecordQuery(query string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.PopularQueries[query]++
}

// RecordQuality records a quality preference
func (m *Metrics) RecordQuality(quality string) {
	if quality == "" {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.QualityStats[quality]++
}

// RecordType records a content type
func (m *Metrics) RecordType(contentType string) {
	if contentType == "" {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.TypeStats[contentType]++
}

// UpdateServiceStatus updates service health status
func (m *Metrics) UpdateServiceStatus(service, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch service {
	case "jackett":
		m.JackettStatus = status
	case "deluge":
		m.DelugeStatus = status
	case "plex":
		m.PlexStatus = status
	}

	m.LastHealthCheck = time.Now()
}

// GetStats returns formatted statistics
func (m *Metrics) GetStats() Stats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := Stats{
		TotalSearches:       m.TotalSearches,
		SuccessfulSearches:  m.SuccessfulSearch,
		FailedSearches:      m.FailedSearches,
		AverageSearchTime:   m.AverageSearchTime.String(),
		CacheHits:           m.CacheHits,
		CacheMisses:         m.CacheMisses,
		TotalDownloads:      m.TotalDownloads,
		SuccessfulDownloads: m.SuccessfulDownloads,
		FailedDownloads:     m.FailedDownloads,
		TotalRequests:       m.TotalRequests,
		TotalErrors:         m.TotalErrors,
		AverageResponseTime: m.AverageResponseTime.String(),
		Services: map[string]string{
			"jackett": m.JackettStatus,
			"deluge":  m.DelugeStatus,
			"plex":    m.PlexStatus,
		},
		LastHealthCheck: m.LastHealthCheck.Format(time.RFC3339),
		Uptime:          time.Since(m.StartTime).String(),
	}

	// Calculate rates
	if m.TotalSearches > 0 {
		stats.SearchSuccessRate = float64(m.SuccessfulSearch) / float64(m.TotalSearches) * 100
	}

	if m.CacheHits+m.CacheMisses > 0 {
		stats.CacheHitRate = float64(m.CacheHits) / float64(m.CacheHits+m.CacheMisses) * 100
	}

	if m.TotalDownloads > 0 {
		stats.DownloadSuccessRate = float64(m.SuccessfulDownloads) / float64(m.TotalDownloads) * 100
	}

	if m.TotalRequests > 0 {
		stats.ErrorRate = float64(m.TotalErrors) / float64(m.TotalRequests) * 100
	}

	// Get popular queries (top 10)
	stats.PopularQueries = m.getTopQueries(10)

	// Get quality distribution
	stats.QualityDistribution = m.getQualityDistribution()

	// Get type distribution
	stats.TypeDistribution = m.getTypeDistribution()

	return stats
}

// getTopQueries returns the most popular queries
func (m *Metrics) getTopQueries(limit int) []QueryStat {
	type queryCount struct {
		query string
		count int64
	}

	var queries []queryCount
	for query, count := range m.PopularQueries {
		queries = append(queries, queryCount{query, count})
	}

	// Simple bubble sort for top queries
	for i := 0; i < len(queries)-1; i++ {
		for j := 0; j < len(queries)-i-1; j++ {
			if queries[j].count < queries[j+1].count {
				queries[j], queries[j+1] = queries[j+1], queries[j]
			}
		}
	}

	// Limit results
	if len(queries) > limit {
		queries = queries[:limit]
	}

	var result []QueryStat
	for _, q := range queries {
		result = append(result, QueryStat{
			Query: q.query,
			Count: q.count,
		})
	}

	return result
}

// getQualityDistribution returns quality statistics
func (m *Metrics) getQualityDistribution() []StatItem {
	var result []StatItem
	for quality, count := range m.QualityStats {
		result = append(result, StatItem{
			Name:  quality,
			Count: count,
		})
	}
	return result
}

// getTypeDistribution returns type statistics
func (m *Metrics) getTypeDistribution() []StatItem {
	var result []StatItem
	for contentType, count := range m.TypeStats {
		result = append(result, StatItem{
			Name:  contentType,
			Count: count,
		})
	}
	return result
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.TotalSearches = 0
	m.SuccessfulSearch = 0
	m.FailedSearches = 0
	m.AverageSearchTime = 0
	m.TotalSearchTime = 0
	m.CacheHits = 0
	m.CacheMisses = 0
	m.TotalDownloads = 0
	m.SuccessfulDownloads = 0
	m.FailedDownloads = 0
	m.TotalRequests = 0
	m.TotalErrors = 0
	m.AverageResponseTime = 0
	m.TotalResponseTime = 0
	m.StartTime = time.Now()
	m.PopularQueries = make(map[string]int64)
	m.QualityStats = make(map[string]int64)
	m.TypeStats = make(map[string]int64)
}

// Global metric functions

// IncrementSearches increments the global search counter
func IncrementSearches() {
	GetGlobalMetrics().IncrementSearches()
}

// IncrementSuccessfulSearches increments the global successful search counter
func IncrementSuccessfulSearches() {
	GetGlobalMetrics().IncrementSuccessfulSearches()
}

// IncrementFailedSearches increments the global failed search counter
func IncrementFailedSearches() {
	GetGlobalMetrics().IncrementFailedSearches()
}

// RecordSearchTime records a search time in global metrics
func RecordSearchTime(duration time.Duration) {
	GetGlobalMetrics().RecordSearchTime(duration)
}

// IncrementCacheHits increments the global cache hit counter
func IncrementCacheHits() {
	GetGlobalMetrics().IncrementCacheHits()
}

// IncrementCacheMisses increments the global cache miss counter
func IncrementCacheMisses() {
	GetGlobalMetrics().IncrementCacheMisses()
}

// IncrementDownloads increments the global download counter
func IncrementDownloads() {
	GetGlobalMetrics().IncrementDownloads()
}

// IncrementSuccessfulDownloads increments the global successful download counter
func IncrementSuccessfulDownloads() {
	GetGlobalMetrics().IncrementSuccessfulDownloads()
}

// IncrementFailedDownloads increments the global failed download counter
func IncrementFailedDownloads() {
	GetGlobalMetrics().IncrementFailedDownloads()
}

// IncrementRequests increments the global request counter
func IncrementRequests() {
	GetGlobalMetrics().IncrementRequests()
}

// IncrementErrors increments the global error counter
func IncrementErrors() {
	GetGlobalMetrics().IncrementErrors()
}

// RecordResponseTime records a response time in global metrics
func RecordResponseTime(duration time.Duration) {
	GetGlobalMetrics().RecordResponseTime(duration)
}

// RecordQuery records a search query in global metrics
func RecordQuery(query string) {
	GetGlobalMetrics().RecordQuery(query)
}

// RecordQuality records a quality preference in global metrics
func RecordQuality(quality string) {
	GetGlobalMetrics().RecordQuality(quality)
}

// RecordType records a content type in global metrics
func RecordType(contentType string) {
	GetGlobalMetrics().RecordType(contentType)
}

// UpdateServiceStatus updates service health status in global metrics
func UpdateServiceStatus(service, status string) {
	GetGlobalMetrics().UpdateServiceStatus(service, status)
}

// GetStats returns formatted global statistics
func GetStats() Stats {
	return GetGlobalMetrics().GetStats()
}

// Reset resets all global metrics
func Reset() {
	GetGlobalMetrics().Reset()
}
