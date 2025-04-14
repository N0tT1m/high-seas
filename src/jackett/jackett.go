package jackett

import (
	"context"
	"fmt"
	jackett "github.com/webtor-io/go-jackett"
	"high-seas/src/deluge"
	"high-seas/src/logger"
	"high-seas/src/utils"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var apiKey = utils.EnvVar("JACKETT_API_KEY", "")
var ip = utils.EnvVar("JACKETT_IP", "")
var port = utils.EnvVar("JACKETT_PORT", "")

// Result scoring constants - adjusted to prioritize exact matches
const (
	TITLE_MATCH_WEIGHT   = 0.5 // Increased to prioritize exact title matches
	YEAR_MATCH_WEIGHT    = 0.1 // New weight for year matching
	SEEDERS_WEIGHT       = 0.2 // Decreased slightly
	QUALITY_WEIGHT       = 0.1 // Decreased slightly
	SIZE_WEIGHT          = 0.1 // Kept the same
	MIN_ACCEPTABLE_SCORE = 0.5 // Increased to require better matches
)

// Common constants and categories
const (
	searchDelay = 500 * time.Millisecond
)

var (
	animeMovieCategories  = []uint{2000, 2010, 100001}
	animeSeriesCategories = []uint{100060, 140679, 5070}
)

type searchResult struct {
	result *jackett.Result
	score  float64
}

// Make sure MakeMovieQuery uses the same pattern as MakeShowQuery
func MakeMovieQuery(query string, tmdbID int, quality string, year int) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	// Add year to query for more specific matching
	queryString := fmt.Sprintf("%s %d %s", query, year, quality)
	logger.WriteInfo(fmt.Sprintf("Searching for movie: %s", queryString))

	resp, err := j.Fetch(ctx, &jackett.FetchRequest{
		Categories: []uint{2000, 2010, 2020, 2030, 2040, 2050, 2060, 2070, 2080}, // Movie categories
		Query:      queryString,
	})
	if err != nil {
		logger.WriteFatal("Failed to fetch from Jackett.", err)
		return err
	}

	results := processResults(resp.Results, tmdbID, quality, query, year)
	if len(results) > 0 {
		if addTorrentToDeluge(results[0].result) {
			return nil
		}
		return fmt.Errorf("failed to add movie torrent to Deluge")
	}

	// Fallback to search without year if no results found
	queryString = fmt.Sprintf("%s %s", query, quality)
	logger.WriteInfo(fmt.Sprintf("Fallback search for movie without year: %s", queryString))

	resp, err = j.Fetch(ctx, &jackett.FetchRequest{
		Categories: []uint{2000, 2010, 2020, 2030, 2040, 2050, 2060, 2070, 2080},
		Query:      queryString,
	})

	if err != nil {
		return fmt.Errorf("failed to fetch from Jackett: %v", err)
	}

	results = processResults(resp.Results, tmdbID, quality, query, year)
	if len(results) > 0 {
		if addTorrentToDeluge(results[0].result) {
			return nil
		}
	}

	return fmt.Errorf("no suitable matches found for movie: %s", query)
}

func MakeShowQuery(query string, seasons []int, tmdbID int, quality string, year int) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	totalSeasons := len(seasons)
	logger.WriteInfo(fmt.Sprintf("Starting search for %s (%d) with %d total seasons", query, year, totalSeasons))

	// Print out the episode counts for each season for debugging
	for i, count := range seasons {
		logger.WriteInfo(fmt.Sprintf("Season %d has %d episodes", i+1, count))
	}

	// Step 1: Try complete series bundle with year
	if searchFullSeriesBundle(ctx, j, query, seasons, tmdbID, quality, year) {
		return nil
	}

	// Step 2: Try season bundles or individual episodes
	currentSeason := 1
	for currentSeason <= totalSeasons {
		// Try to find season pack first with year
		if found := searchCompleteSeason(ctx, j, query, currentSeason, seasons[currentSeason-1], tmdbID, quality, year); !found {
			// If season pack not found, search episode by episode
			logger.WriteInfo(fmt.Sprintf("No season pack found for season %d, searching %d individual episodes",
				currentSeason, seasons[currentSeason-1]))
			searchSeasonEpisodesByOne(ctx, j, query, currentSeason, tmdbID, quality, seasons[currentSeason-1], year)
		}
		currentSeason++
	}

	return nil
}

func searchCompleteSeason(ctx context.Context, j *jackett.Jackett, query string, season, episodeCount, tmdbID int, quality string, year int) bool {
	seasonFormat := fmt.Sprintf("S%02d", season)
	searchQueries := []string{
		// Include year in search queries for more specificity
		fmt.Sprintf("%s (%d) %s season %s", query, year, seasonFormat, quality),
		fmt.Sprintf("%s %s season %s", query, seasonFormat, quality),
		fmt.Sprintf("%s complete %s %s", query, seasonFormat, quality),
		fmt.Sprintf("%s %s complete %s", query, seasonFormat, quality),
	}

	for _, queryString := range searchQueries {
		logger.WriteInfo(fmt.Sprintf("Searching for complete season %d (%d episodes) with query: %s",
			season, episodeCount, queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: []uint{5000, 5020, 5030, 5040, 5045},
			Query:      queryString,
		})
		if err != nil {
			continue
		}

		results := processResults(resp.Results, tmdbID, quality, query, year)
		if len(results) > 0 {
			seasonResults := filterSeasonPacks(results, season, episodeCount)
			if len(seasonResults) > 0 {
				bestResult := selectBestResult(seasonResults)
				if bestResult != nil && addTorrentToDeluge(bestResult) {
					logger.WriteInfo(fmt.Sprintf("Successfully added complete season %d (Size: %.2f GB)",
						season, float64(bestResult.Size)/1024/1024/1024))
					return true
				}
			}
		}
	}

	return false
}

// New function to filter for actual season packs
func filterSeasonPacks(results []searchResult, season, expectedEpisodeCount int) []searchResult {
	var seasonPacks []searchResult
	seasonStr := fmt.Sprintf("S%02d", season)
	episodePattern := regexp.MustCompile(fmt.Sprintf("S%02dE\\d+", season))

	for _, result := range results {
		title := strings.ToLower(result.result.Title)
		// Check if it contains the season number and doesn't contain episode numbers
		if strings.Contains(title, strings.ToLower(seasonStr)) && !episodePattern.MatchString(title) {
			// Look for season pack indicators
			if strings.Contains(title, "complete") ||
				strings.Contains(title, "season") ||
				!strings.Contains(title, "e") { // No episode marker
				seasonPacks = append(seasonPacks, result)
			}
		}
	}

	return seasonPacks
}

func searchFullSeriesBundle(ctx context.Context, j *jackett.Jackett, query string, seasons []int, tmdbID int, quality string, year int) bool {
	searchQueries := []string{
		// Include year in search queries
		fmt.Sprintf("%s (%d) complete series %s", query, year, quality),
		fmt.Sprintf("%s complete series %s", query, quality),
		fmt.Sprintf("%s (%d) season 1-%d %s", query, year, len(seasons), quality),
		fmt.Sprintf("%s season 1-%d %s", query, len(seasons), quality),
	}

	for _, queryString := range searchQueries {
		logger.WriteInfo(fmt.Sprintf("Searching for complete series with query: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: []uint{5000, 5020, 5030, 5040, 5045},
			Query:      queryString,
		})
		if err != nil {
			continue
		}

		results := processResults(resp.Results, tmdbID, quality, query, year)
		if len(results) > 0 {
			bestResult := selectBestResult(results)
			if bestResult != nil && addTorrentToDeluge(bestResult) {
				logger.WriteInfo("Successfully added complete series")
				return true
			}
		}
	}

	return false
}

// Modified searchSeasonEpisodesByOne to ensure we get every episode
func searchSeasonEpisodesByOne(ctx context.Context, j *jackett.Jackett, query string, season, tmdbID int, quality string, episodeCount int, year int) bool {
	logger.WriteInfo(fmt.Sprintf("Searching for %d individual episodes of season %d", episodeCount, season))
	successCount := 0
	var missingEpisodes []int

	for episode := 1; episode <= episodeCount; episode++ {
		episodeFormat := fmt.Sprintf("S%02dE%02d", season, episode)
		// More specific query with year and double quotes for exact matching
		queryString := fmt.Sprintf("\"%s\" %s %s", query, episodeFormat, quality)
		yearQueryString := fmt.Sprintf("\"%s\" (%d) %s %s", query, year, episodeFormat, quality)

		// Try with year first
		logger.WriteInfo(fmt.Sprintf("Searching for episode with year: %s", yearQueryString))
		if tryFetchEpisode(ctx, j, yearQueryString, tmdbID, quality, query, year, episodeFormat, &successCount, episodeCount) {
			continue
		}

		// Fallback to without year
		logger.WriteInfo(fmt.Sprintf("Searching for episode without year: %s", queryString))
		if !tryFetchEpisode(ctx, j, queryString, tmdbID, quality, query, year, episodeFormat, &successCount, episodeCount) {
			missingEpisodes = append(missingEpisodes, episode)
		}

		time.Sleep(500 * time.Millisecond)
	}

	if len(missingEpisodes) > 0 {
		logger.WriteWarning(fmt.Sprintf("Missing episodes for season %d: %v", season, missingEpisodes))
	}

	logger.WriteInfo(fmt.Sprintf("Found %d/%d episodes for season %d",
		successCount, episodeCount, season))
	return successCount > 0
}

func tryFetchEpisode(ctx context.Context, j *jackett.Jackett, queryString string, tmdbID int, quality string,
	query string, year int, episodeFormat string, successCount *int, episodeCount int) bool {
	resp, err := j.Fetch(ctx, &jackett.FetchRequest{
		Categories: []uint{5000, 5020, 5030, 5040, 5045},
		Query:      queryString,
	})
	if err != nil {
		return false
	}

	results := processResults(resp.Results, tmdbID, quality, query, year)
	if len(results) > 0 {
		// Filter results to ensure exact episode match
		var episodeResults []searchResult
		for _, result := range results {
			if isExactEpisodeMatch(result.result.Title, episodeFormat, query) {
				episodeResults = append(episodeResults, result)
			}
		}

		if len(episodeResults) > 0 {
			bestResult := selectBestResult(episodeResults)
			if bestResult != nil && addTorrentToDeluge(bestResult) {
				*successCount++
				logger.WriteInfo(fmt.Sprintf("Successfully added %s (%d/%d)",
					episodeFormat, *successCount, episodeCount))
				return true
			}
		}
	}

	return false
}

// New function to check for exact episode matches
func isExactEpisodeMatch(resultTitle, episodeFormat, showTitle string) bool {
	cleanTitle := cleanTitleForComparison(resultTitle)
	cleanShowTitle := cleanExactTitle(showTitle)

	// Check if title contains both the show name and the exact episode format
	return strings.Contains(cleanTitle, cleanShowTitle) &&
		strings.Contains(strings.ToLower(resultTitle), strings.ToLower(episodeFormat))
}

// Modify processResults to include more logging and year matching
func processResults(results []jackett.Result, tmdbID int, quality string, exactTitle string, year int) []searchResult {
	var scoredResults []searchResult

	logger.WriteInfo(fmt.Sprintf("Processing %d results", len(results)))

	for _, result := range results {
		// Skip results that don't match the exact show title
		if !isExactTitleMatch(result.Title, exactTitle) {
			logger.WriteInfo(fmt.Sprintf("Skipping non-matching title: %s", result.Title))
			continue
		}

		// Calculate score including year match
		score := calculateScore(&result, tmdbID, quality, exactTitle, year)
		if score >= MIN_ACCEPTABLE_SCORE {
			scoredResults = append(scoredResults, searchResult{
				result: &result,
				score:  score,
			})
			logger.WriteInfo(fmt.Sprintf("Added to candidates: %s (Score: %.2f)", result.Title, score))
		} else {
			logger.WriteInfo(fmt.Sprintf("Score too low (%.2f < %.2f): %s", score, MIN_ACCEPTABLE_SCORE, result.Title))
		}
	}

	// Sort by score in descending order, then by size
	sort.Slice(scoredResults, func(i, j int) bool {
		if math.Abs(scoredResults[i].score-scoredResults[j].score) < 0.05 {
			// If scores are very close, prefer larger file sizes
			return scoredResults[i].result.Size > scoredResults[j].result.Size
		}
		return scoredResults[i].score > scoredResults[j].score
	})

	if len(scoredResults) > 0 {
		logger.WriteInfo(fmt.Sprintf("Selected best match: %s (Score: %.2f, Size: %.2f GB)",
			scoredResults[0].result.Title,
			scoredResults[0].score,
			float64(scoredResults[0].result.Size)/1024/1024/1024))
	}

	return scoredResults
}

// Enhanced title matching with more strict comparisons
func isExactTitleMatch(resultTitle, exactTitle string) bool {
	// Clean both titles for comparison
	cleanResultTitle := cleanTitleForComparison(resultTitle)
	cleanExactTitle := cleanExactTitle(exactTitle)

	// Get words from the exact title as a set for checking
	exactTitleWords := getSignificantWords(cleanExactTitle)

	// The result title should contain all the significant words from the exact title
	// and ideally in the same order (using HasPrefix)
	allWordsPresent := true
	for word := range exactTitleWords {
		if !strings.Contains(cleanResultTitle, word) {
			allWordsPresent = false
			break
		}
	}

	// Either all words are present OR the cleaned result starts with the exact title
	// This handles both exact matches and minor variations
	return allWordsPresent || strings.HasPrefix(cleanResultTitle, cleanExactTitle)
}

// Get significant words from a title (no common words/articles)
func getSignificantWords(title string) map[string]bool {
	words := strings.Fields(title)
	result := make(map[string]bool)

	// Articles and common words to skip
	skipWords := map[string]bool{
		"the": true, "a": true, "an": true,
		"and": true, "or": true, "of": true,
		"in": true, "on": true, "at": true,
	}

	for _, word := range words {
		word = strings.ToLower(word)
		if !skipWords[word] && len(word) > 1 {
			result[word] = true
		}
	}

	return result
}

func cleanTitleForComparison(title string) string {
	// Convert to lowercase
	title = strings.ToLower(title)

	// Remove common suffixes and quality indicators
	suffixesToRemove := []string{
		"2160p", "1080p", "720p", "480p",
		"web-dl", "webrip", "hdtv", "bluray", "blu-ray",
		"x264", "x265", "hevc", "h264", "h265",
		"proper", "repack",
		"amzn", "hulu", "nf", "dsnp", // streaming service indicators
	}

	for _, suffix := range suffixesToRemove {
		title = strings.ReplaceAll(title, suffix, "")
	}

	// Remove year patterns (e.g., "(2020)" or "2020")
	yearPattern := regexp.MustCompile(`\(\d{4}\)|\d{4}`)
	title = yearPattern.ReplaceAllString(title, "")

	// Remove season/episode information (e.g., S01E01, 1x01)
	seasonEpPattern := regexp.MustCompile(`\b[Ss]\d{1,2}[Ee]\d{1,2}\b|\b\d{1,2}x\d{1,2}\b`)
	title = seasonEpPattern.ReplaceAllString(title, "")

	// Remove special characters and extra spaces
	title = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(title, " ")
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return strings.TrimSpace(title)
}

func cleanExactTitle(title string) string {
	// Convert to lowercase and trim spaces
	title = strings.ToLower(strings.TrimSpace(title))

	// Remove any special characters, keeping only letters, numbers, and spaces
	title = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(title, " ")
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return strings.TrimSpace(title)
}

func selectBestResult(results []searchResult) *jackett.Result {
	if len(results) == 0 {
		return nil
	}

	// Already sorted by score, return the highest
	return results[0].result
}

func calculateScore(result *jackett.Result, tmdbID int, quality string, exactTitle string, year int) float64 {
	score := 0.0

	// Title match score (enhanced)
	titleScore := calculateTitleMatch(result.Title, exactTitle)
	score += titleScore * TITLE_MATCH_WEIGHT

	// Year match score (new)
	yearScore := calculateYearMatch(result.Title, year)
	score += yearScore * YEAR_MATCH_WEIGHT

	// TMDb match bonus
	if result.TMDb > 0 && int(result.TMDb) == tmdbID {
		score += 0.1 // Reduced impact but still valued
	}

	// Seeders score
	seedersScore := math.Min(float64(result.Seeders)/500.0, 1.0)
	score += seedersScore * SEEDERS_WEIGHT

	// Quality match
	qualityScore := calculateQualityMatch(result.Title, quality)
	score += qualityScore * QUALITY_WEIGHT

	// Size score
	sizeScore := calculateSizeScore(result.Size, quality)
	score += sizeScore * SIZE_WEIGHT

	return score
}

// Enhanced title match function to be more precise
func calculateTitleMatch(title string, exactTitle string) float64 {
	// Remove common strings that don't affect title matching
	cleanTitle := cleanTitleForComparison(title)
	cleanExactTitle := cleanExactTitle(exactTitle)

	// Check for spinoffs and wrong shows by comparing title words
	// exactTitleWords := getSignificantWords(cleanExactTitle)

	// Calculate Levenshtein/edit distance ratio between titles
	// as a measure of similarity (simplified approach)
	titleSimilarity := calculateTitleSimilarity(cleanTitle, cleanExactTitle)

	// Basic scoring criteria
	score := titleSimilarity

	// Penalize for suspicious patterns
	if strings.Contains(cleanTitle, "sample") {
		score -= 0.3
	}

	if strings.Contains(cleanTitle, "trailer") {
		score -= 0.5
	}

	// Penalize for obviously wrong content
	if strings.Contains(cleanTitle, "ost") || strings.Contains(cleanTitle, "soundtrack") {
		score -= 0.8
	}

	// Check for spinoff indicators - these significantly reduce score
	spinoffIndicators := []string{"spinoff", "spin-off", "spin off", "special"}
	for _, indicator := range spinoffIndicators {
		if strings.Contains(cleanTitle, indicator) && !strings.Contains(cleanExactTitle, indicator) {
			score -= 0.5
			break
		}
	}

	return math.Max(0.0, score) // Ensure score doesn't go negative
}

// Calculate similarity ratio between two strings (simplified)
func calculateTitleSimilarity(str1, str2 string) float64 {
	// For exact matches
	if str1 == str2 {
		return 1.0
	}

	// If one is prefix of the other
	if strings.HasPrefix(str1, str2) || strings.HasPrefix(str2, str1) {
		// Calculate ratio based on length difference
		minLen := math.Min(float64(len(str1)), float64(len(str2)))
		maxLen := math.Max(float64(len(str1)), float64(len(str2)))
		return minLen / maxLen
	}

	// Count matching words
	words1 := getSignificantWords(str1)
	words2 := getSignificantWords(str2)

	matchCount := 0
	for word := range words1 {
		if words2[word] {
			matchCount++
		}
	}

	totalWords := len(words1) + len(words2) - matchCount
	if totalWords == 0 {
		return 0.0
	}

	return float64(matchCount) / float64(totalWords)
}

// New function to score year matches
func calculateYearMatch(title string, targetYear int) float64 {
	if targetYear <= 0 {
		return 0.5 // Neutral score if no year provided
	}

	// Look for year patterns in title
	yearPattern := regexp.MustCompile(`\((\d{4})\)|\b(\d{4})\b`)
	matches := yearPattern.FindAllStringSubmatch(title, -1)

	if len(matches) == 0 {
		return 0.3 // No year found, slightly negative
	}

	// Extract years found in the title
	for _, match := range matches {
		var yearStr string
		if match[1] != "" {
			yearStr = match[1] // (2020) format
		} else {
			yearStr = match[2] // 2020 format
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil {
			continue
		}

		// Exact year match
		if year == targetYear {
			return 1.0
		}

		// Close year match (within 1 year)
		if math.Abs(float64(year-targetYear)) <= 1 {
			return 0.8
		}

		// Wrong year but present
		return 0.1
	}

	return 0.3 // No valid year found
}

func calculateQualityMatch(title, targetQuality string) float64 {
	if strings.Contains(strings.ToLower(title), strings.ToLower(targetQuality)) {
		return 1.0
	}

	qualityMap := map[string]float64{
		"2160p": 1.0,
		"1080p": 0.8,
		"720p":  0.6,
		"480p":  0.4,
	}

	for quality, score := range qualityMap {
		if strings.Contains(strings.ToLower(title), strings.ToLower(quality)) {
			return score
		}
	}

	return 0.0
}

func calculateSizeScore(size uint, quality string) float64 {
	sizeGB := float64(size) / 1024 / 1024 / 1024

	// Adjusted size ranges for TV show episodes/seasons
	qualitySizeRanges := map[string]struct{ min, max, ideal float64 }{
		"2160p": {2.0, 20.0, 8.0},
		"1080p": {1.0, 10.0, 4.0},
		"720p":  {0.5, 4.0, 1.5},
		"480p":  {0.2, 2.0, 0.7},
	}

	if range_, exists := qualitySizeRanges[quality]; exists {
		if sizeGB < range_.min {
			return 0.3
		}
		if sizeGB > range_.max {
			return 0.5
		}
		// Higher score the closer to ideal size
		deviation := math.Abs(sizeGB-range_.ideal) / range_.ideal
		return math.Max(0.0, 1.0-deviation)
	}

	return 0.5
}

func addTorrentToDeluge(result *jackett.Result) bool {
	if result == nil {
		logger.WriteError("No valid result to add to Deluge", nil)
		return false
	}

	logger.WriteInfo(fmt.Sprintf("Attempting to add to Deluge: %s", result.Title))
	logger.WriteInfo(fmt.Sprintf("Torrent Link: %s", result.Link))
	logger.WriteInfo(fmt.Sprintf("Size: %.2f GB", float64(result.Size)/1024/1024/1024))
	logger.WriteInfo(fmt.Sprintf("Seeders: %d", result.Seeders))

	err := deluge.AddTorrent(result.Link)
	if err != nil {
		if strings.Contains(err.Error(), "Torrent already in session") {
			// If we get "already in session", consider it a success since it means
			// we already have this episode
			logger.WriteInfo(fmt.Sprintf("Torrent already exists in Deluge: %s", result.Title))
			return true
		}
		logger.WriteError(fmt.Sprintf("Failed to add torrent to Deluge for %s. Error: %v", result.Title, err), err)
		return false
	}

	logger.WriteInfo(fmt.Sprintf("Successfully sent to Deluge: %s", result.Title))
	return true
}

func tryAddTorrentWithFallback(results []searchResult) bool {
	for _, result := range results {
		if addTorrentToDeluge(result.result) {
			return true
		}
		// Wait a bit before trying the next result
		time.Sleep(1 * time.Second)
	}
	return false
}

// Anime functions below - updated with enhanced title matching
// MakeAnimeMovieQuery handles searching and downloading anime movies with improved validation
func MakeAnimeMovieQuery(query string, tmdbID int, quality string, year int) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	// Include year in query
	queryString := fmt.Sprintf("%s %d %s", query, year, quality)
	logger.WriteInfo(fmt.Sprintf("Searching for anime movie: %s", queryString))

	// First attempt with strict anime movie categories
	if err := searchAnimeMovie(ctx, j, queryString, tmdbID, quality, year); err == nil {
		return nil
	}

	// Fallback without year
	queryString = fmt.Sprintf("%s %s", query, quality)

	// Fallback to broader categories if needed
	fallbackCategories := [][]uint{
		{2000, 2010, 100001}, // Anime-specific
		{2000, 2010, 2020},   // General movies
	}

	for _, categories := range fallbackCategories {
		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: categories,
			Query:      queryString,
		})
		if err != nil {
			continue
		}

		results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
		if len(results) > 0 {
			// Try each result until we find one that works
			for _, result := range results {
				if validateAndAddAnimeTorrent(result.result) {
					return nil
				}
				time.Sleep(searchDelay)
			}
		}
	}

	return fmt.Errorf("no valid anime movie downloads found for: %s", query)
}

func MakeAnimeShowQuery(query string, seasons []int, tmdbID int, quality string, year int) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	totalEpisodes := 0
	for _, episodeCount := range seasons {
		totalEpisodes += episodeCount
	}

	logger.WriteInfo(fmt.Sprintf("Starting search for anime series: %s (%d) with %d total episodes",
		query, year, totalEpisodes))

	// Try batch downloads first
	if found := tryAnimeBatchDownloads(ctx, j, query, tmdbID, quality, year); found {
		return nil
	}

	// If batch download fails, try episode by episode
	return searchAnimeEpisodesByOne(ctx, j, query, tmdbID, quality, seasons, year)
}

func isAnimeTimeRelease(title string) bool {
	return strings.Contains(title, "[Anime Time]")
}

func tryAnimeBatchDownloads(ctx context.Context, j *jackett.Jackett, query string, tmdbID int, quality string, year int) bool {
	// Try with year first
	yearQueryPatterns := []string{
		`[Anime Time] %s (%d) (Series+Movies) [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub] [Batch]`,
		`[Anime Time] %s (%d) [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub] [Batch]`,
		`[Anime Time] %s (%d) [Batch]`,
	}

	// Try Anime Time patterns with year first
	for _, pattern := range yearQueryPatterns {
		queryString := fmt.Sprintf(pattern, query, year)
		logger.WriteInfo(fmt.Sprintf("Trying Anime Time batch search with year: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeSeriesCategories,
			Query:      queryString,
		})

		if err == nil && len(resp.Results) > 0 {
			results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
			for _, result := range results {
				if isAnimeTimeRelease(result.result.Title) && addTorrentMagnetToDeluge(result.result) {
					logger.WriteInfo(fmt.Sprintf("Successfully added Anime Time batch with year: %s", result.result.Title))
					return true
				}
			}
		}

		time.Sleep(searchDelay)
	}

	// Exact Anime Time pattern matching the format - without year
	animeTimePatterns := []string{
		`[Anime Time] %s (Series+Movies) [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub] [Batch]`,
		`[Anime Time] %s [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub] [Batch]`,
		`[Anime Time] %s [Batch]`,
	}

	// Try Anime Time patterns without year
	for _, pattern := range animeTimePatterns {
		queryString := fmt.Sprintf(pattern, query)
		logger.WriteInfo(fmt.Sprintf("Trying Anime Time batch search without year: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeSeriesCategories,
			Query:      queryString,
		})

		if err == nil && len(resp.Results) > 0 {
			results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
			for _, result := range results {
				if isAnimeTimeRelease(result.result.Title) && addTorrentMagnetToDeluge(result.result) {
					logger.WriteInfo(fmt.Sprintf("Successfully added Anime Time batch: %s", result.result.Title))
					return true
				}
			}
		}

		time.Sleep(searchDelay)
	}

	// Fallback patterns with year
	yearFallbackPatterns := []string{
		`[AnimeSkulls] %s (%d) [Batch] [Dual Audio][1080p][HEVC 10bit x265]`,
		`%s (%d) complete series`,
		`%s (%d) batch`,
	}

	for _, pattern := range yearFallbackPatterns {
		queryString := fmt.Sprintf(pattern, query, year)
		logger.WriteInfo(fmt.Sprintf("Trying fallback batch search with year: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeSeriesCategories,
			Query:      queryString,
		})

		if err == nil && len(resp.Results) > 0 {
			results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
			if len(results) > 0 {
				for _, result := range results {
					if addTorrentMagnetToDeluge(result.result) {
						logger.WriteInfo(fmt.Sprintf("Successfully added batch with year: %s", result.result.Title))
						return true
					}
				}
			}
		}

		time.Sleep(searchDelay)
	}

	// Fallback patterns without year
	fallbackPatterns := []string{
		`[AnimeSkulls] %s [Batch] [Dual Audio][1080p][HEVC 10bit x265]`,
		`%s complete series`,
		`%s batch`,
	}

	for _, pattern := range fallbackPatterns {
		queryString := fmt.Sprintf(pattern, query)
		logger.WriteInfo(fmt.Sprintf("Trying fallback batch search without year: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeSeriesCategories,
			Query:      queryString,
		})

		if err != nil {
			continue
		}

		results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
		if len(results) > 0 {
			for _, result := range results {
				if addTorrentMagnetToDeluge(result.result) {
					logger.WriteInfo(fmt.Sprintf("Successfully added batch: %s", result.result.Title))
					return true
				}
			}
		}

		time.Sleep(searchDelay)
	}

	return false
}

func calculateAnimeSize(size uint, quality string) float64 {
	sizeGB := float64(size) / 1024 / 1024 / 1024

	// Size ranges for different qualities (in GB)
	ranges := map[string]struct{ min, max, ideal float64 }{
		"2160p": {2.0, 8.0, 4.0},  // 4K releases
		"1080p": {0.4, 2.0, 0.8},  // Standard episode size
		"720p":  {0.2, 1.0, 0.4},  // Smaller HD
		"480p":  {0.1, 0.5, 0.25}, // SD
	}

	if range_, exists := ranges[quality]; exists {
		if sizeGB < range_.min || sizeGB > range_.max {
			return 0.3 // Penalize sizes outside expected range
		}

		// Calculate how close we are to ideal size
		deviation := math.Abs(sizeGB-range_.ideal) / range_.ideal
		return math.Max(0.0, 1.0-deviation)
	}

	return 0.5 // Default score for unknown quality
}

func searchAnimeEpisodesByOne(ctx context.Context, j *jackett.Jackett, query string, tmdbID int, quality string, seasons []int, year int) error {
	episodeTotal := 0
	for _, count := range seasons {
		episodeTotal += count
	}

	for episode := 1; episode <= episodeTotal; episode++ {
		// Anime Time episode patterns with year
		animeTimeYearPatterns := []string{
			`[Anime Time] %s (%d) - %02d [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]`,
			`[Anime Time] %s (%d) - Episode %02d [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]`,
		}

		// Anime Time episode patterns without year
		animeTimePatterns := []string{
			`[Anime Time] %s - %02d [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]`,
			`[Anime Time] %s - Episode %02d [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]`,
		}

		// Fallback patterns with year
		yearFallbackPatterns := []string{
			`[AnimeSkulls] %s (%d) Episode %d [1080p] [Dual.Audio] [x265]`,
			`%s (%d) Episode %d`,
		}

		// Fallback patterns without year
		fallbackPatterns := []string{
			`[AnimeSkulls] %s Episode %d [1080p] [Dual.Audio] [x265]`,
			`%s Episode %d`,
		}

		var foundEpisode bool

		// Try Anime Time patterns with year first
		for _, pattern := range animeTimeYearPatterns {
			queryString := fmt.Sprintf(pattern, query, year, episode)
			logger.WriteInfo(fmt.Sprintf("Searching Anime Time for episode %d/%d with year: %s",
				episode, episodeTotal, queryString))

			resp, err := j.Fetch(ctx, &jackett.FetchRequest{
				Categories: animeSeriesCategories,
				Query:      queryString,
			})

			if err == nil {
				results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
				for _, result := range results {
					if isAnimeTimeRelease(result.result.Title) && addTorrentMagnetToDeluge(result.result) {
						logger.WriteInfo(fmt.Sprintf("Successfully added Anime Time episode %d with year: %s",
							episode, result.result.Title))
						foundEpisode = true
						break
					}
				}
			}

			if foundEpisode {
				break
			}
			time.Sleep(searchDelay)
		}

		// Try Anime Time patterns without year if not found
		if !foundEpisode {
			for _, pattern := range animeTimePatterns {
				queryString := fmt.Sprintf(pattern, query, episode)
				logger.WriteInfo(fmt.Sprintf("Searching Anime Time for episode %d/%d without year: %s",
					episode, episodeTotal, queryString))

				resp, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: animeSeriesCategories,
					Query:      queryString,
				})

				if err == nil {
					results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
					for _, result := range results {
						if isAnimeTimeRelease(result.result.Title) && addTorrentMagnetToDeluge(result.result) {
							logger.WriteInfo(fmt.Sprintf("Successfully added Anime Time episode %d: %s",
								episode, result.result.Title))
							foundEpisode = true
							break
						}
					}
				}

				if foundEpisode {
					break
				}
				time.Sleep(searchDelay)
			}
		}

		// Try fallback patterns with year if still not found
		if !foundEpisode {
			for _, pattern := range yearFallbackPatterns {
				queryString := fmt.Sprintf(pattern, query, year, episode)
				logger.WriteInfo(fmt.Sprintf("Searching fallback for episode %d/%d with year: %s",
					episode, episodeTotal, queryString))

				resp, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: animeSeriesCategories,
					Query:      queryString,
				})

				if err != nil {
					continue
				}

				results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
				if len(results) > 0 {
					for _, result := range results {
						if addTorrentMagnetToDeluge(result.result) {
							logger.WriteInfo(fmt.Sprintf("Successfully added episode %d with year: %s",
								episode, result.result.Title))
							foundEpisode = true
							break
						}
					}
				}

				if foundEpisode {
					break
				}
				time.Sleep(searchDelay)
			}
		}

		// Try fallback patterns without year as last resort
		if !foundEpisode {
			for _, pattern := range fallbackPatterns {
				queryString := fmt.Sprintf(pattern, query, episode)
				logger.WriteInfo(fmt.Sprintf("Searching fallback for episode %d/%d without year: %s",
					episode, episodeTotal, queryString))

				resp, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: animeSeriesCategories,
					Query:      queryString,
				})

				if err != nil {
					continue
				}

				results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
				if len(results) > 0 {
					for _, result := range results {
						if addTorrentMagnetToDeluge(result.result) {
							logger.WriteInfo(fmt.Sprintf("Successfully added episode %d: %s",
								episode, result.result.Title))
							foundEpisode = true
							break
						}
					}
				}

				if foundEpisode {
					break
				}
				time.Sleep(searchDelay)
			}
		}

		if !foundEpisode {
			logger.WriteWarning(fmt.Sprintf("No valid results found for episode %d", episode))
		}
	}

	return nil
}

func searchAnimeMovie(ctx context.Context, j *jackett.Jackett, query string, tmdbID int, quality string, year int) error {
	// Try different search patterns with year
	yearSearchPatterns := []string{
		"%s (%d) %s [1080p]",
		"[Anime] %s (%d) %s",
		"%s (%d) [BD] %s",
	}

	for _, pattern := range yearSearchPatterns {
		formattedQuery := fmt.Sprintf(pattern, query, year, quality)
		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeMovieCategories,
			Query:      formattedQuery,
		})
		if err != nil {
			continue
		}

		results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
		for _, result := range results {
			if validateAndAddAnimeTorrent(result.result) {
				return nil
			}
		}
		time.Sleep(searchDelay)
	}

	// Fallback to patterns without year
	searchPatterns := []string{
		"%s %s [1080p]",
		"[Anime] %s %s",
		"%s [BD] %s",
	}

	for _, pattern := range searchPatterns {
		formattedQuery := fmt.Sprintf(pattern, query, quality)
		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeMovieCategories,
			Query:      formattedQuery,
		})
		if err != nil {
			continue
		}

		results := processAnimeResults(resp.Results, tmdbID, quality, query, year)
		for _, result := range results {
			if validateAndAddAnimeTorrent(result.result) {
				return nil
			}
		}
		time.Sleep(searchDelay)
	}

	return fmt.Errorf("no suitable matches found")
}

func validateAndAddAnimeTorrent(result *jackett.Result) bool {
	if result == nil {
		return false
	}

	// Log detailed information about the torrent
	logger.WriteInfo(fmt.Sprintf("Validating anime torrent: %s", result.Title))
	logger.WriteInfo(fmt.Sprintf("Size: %.2f GB", float64(result.Size)/1024/1024/1024))
	logger.WriteInfo(fmt.Sprintf("Seeders: %d", result.Seeders))

	// Check for valid size
	if result.Size == 0 {
		logger.WriteWarning(fmt.Sprintf("Skipping zero-size torrent: %s", result.Title))
		return false
	}

	// Prefer magnet links but fall back to regular links if needed
	downloadLink := result.MagnetUri
	if downloadLink == "" {
		downloadLink = result.Link
		if downloadLink == "" {
			logger.WriteWarning(fmt.Sprintf("No valid download link found for: %s", result.Title))
			return false
		}
	}

	// Additional validation for magnet links
	if strings.HasPrefix(downloadLink, "magnet:") {
		if !strings.Contains(downloadLink, "xt=urn:btih:") {
			logger.WriteWarning(fmt.Sprintf("Invalid magnet link format for: %s", result.Title))
			return false
		}
	}

	// Try to add the torrent
	err := deluge.AddTorrent(downloadLink)
	if err != nil {
		if strings.Contains(err.Error(), "Torrent already in session") {
			logger.WriteInfo(fmt.Sprintf("Torrent already exists in Deluge: %s", result.Title))
			return true
		}
		logger.WriteError(fmt.Sprintf("Failed to add torrent: %s, Error: %v", result.Title, err), err)
		return false
	}

	// Verify the torrent was added successfully
	logger.WriteInfo(fmt.Sprintf("Successfully added to Deluge: %s", result.Title))
	return true
}

func processAnimeResults(results []jackett.Result, tmdbID int, quality string, query string, year int) []searchResult {
	var scoredResults []searchResult
	logger.WriteInfo(fmt.Sprintf("Processing %d anime results", len(results)))

	for _, result := range results {
		// Skip invalid results
		if result.Size == 0 {
			logger.WriteInfo(fmt.Sprintf("Skipping zero-size result: %s", result.Title))
			continue
		}

		if result.MagnetUri == "" && result.Link == "" {
			logger.WriteInfo(fmt.Sprintf("Skipping result with no download link: %s", result.Title))
			continue
		}

		// Enhanced scoring with title and year matching
		score := calculateAnimeScore(&result, tmdbID, quality, query, year)

		// Additional scoring for preferred release groups
		score += getAnimeReleaseGroupScore(result.Title)

		if score >= 0.3 {
			scoredResults = append(scoredResults, searchResult{
				result: &result,
				score:  score,
			})
			logger.WriteInfo(fmt.Sprintf("Added candidate: %s (Score: %.2f, Size: %.2f GB, Seeders: %d)",
				result.Title, score, float64(result.Size)/1024/1024/1024, result.Seeders))
		}
	}

	// Sort results by score in descending order, then by seeders
	sort.Slice(scoredResults, func(i, j int) bool {
		if math.Abs(scoredResults[i].score-scoredResults[j].score) < 0.05 {
			return scoredResults[i].result.Seeders > scoredResults[j].result.Seeders
		}
		return scoredResults[i].score > scoredResults[j].score
	})

	return scoredResults
}

func getAnimeReleaseGroupScore(title string) float64 {
	title = strings.ToLower(title)

	// Preferred release groups and their scores
	releaseGroups := map[string]float64{
		"[anime time]":   0.7,
		"[animeskulls]":  0.5,
		"[subsplease]":   0.4,
		"[horriblesubs]": 0.3,
		"[erai-raws]":    0.3,
	}

	for group, score := range releaseGroups {
		if strings.Contains(title, group) {
			return score
		}
	}

	return 0.0
}

func calculateAnimeScore(result *jackett.Result, tmdbID int, quality string, query string, year int) float64 {
	score := 0.0

	// Base score
	score += 0.1

	// Title match score - much more important
	titleMatch := calculateAnimeTitleMatch(result.Title, query)
	score += titleMatch * 0.4

	// Year match - added for anime
	yearScore := calculateYearMatch(result.Title, year)
	score += yearScore * 0.1

	// Quality and format preferences
	if strings.Contains(strings.ToLower(result.Title), "1080p") {
		score += 0.2
	}
	if strings.Contains(strings.ToLower(result.Title), "x265") ||
		strings.Contains(strings.ToLower(result.Title), "hevc") {
		score += 0.1
	}
	if strings.Contains(strings.ToLower(result.Title), "dual.audio") ||
		strings.Contains(strings.ToLower(result.Title), "dual audio") {
		score += 0.2
	}

	// Seeder score (if available)
	seedersScore := math.Min(float64(result.Seeders)/50.0, 1.0)
	score += seedersScore * 0.2

	return score
}

// New function for anime title matching
func calculateAnimeTitleMatch(title string, exactTitle string) float64 {
	// Clean both titles
	cleanResultTitle := cleanTitleForComparison(title)
	cleanExactTitle := cleanExactTitle(exactTitle)

	// Check for exact match
	if cleanResultTitle == cleanExactTitle {
		return 1.0
	}

	// Check if result contains the exact title as a prefix/subset
	if strings.HasPrefix(cleanResultTitle, cleanExactTitle) {
		return 0.9
	}

	// Calculate similarity
	titleSimilarity := calculateTitleSimilarity(cleanResultTitle, cleanExactTitle)

	// Check for unwanted content
	if strings.Contains(strings.ToLower(title), "trailer") ||
		strings.Contains(strings.ToLower(title), "preview") ||
		strings.Contains(strings.ToLower(title), "pv") {
		titleSimilarity *= 0.2
	}

	return titleSimilarity
}

func addTorrentMagnetToDeluge(result *jackett.Result) bool {
	if result == nil {
		logger.WriteError("No valid result to add to Deluge", nil)
		return false
	}

	if result.MagnetUri == "" {
		logger.WriteError(fmt.Sprintf("No magnet URI available for: %s", result.Title), nil)
		return false
	}

	logger.WriteInfo(fmt.Sprintf("Attempting to add to Deluge: %s", result.Title))
	logger.WriteInfo(fmt.Sprintf("Magnet URI: %s", result.MagnetUri))
	logger.WriteInfo(fmt.Sprintf("Size: %.2f GB", float64(result.Size)/1024/1024/1024))
	logger.WriteInfo(fmt.Sprintf("Seeders: %d", result.Seeders))

	err := deluge.AddTorrent(result.MagnetUri)
	if err != nil {
		if strings.Contains(err.Error(), "Torrent already in session") {
			// If we get "already in session", consider it a success since it means
			// we already have this episode
			logger.WriteInfo(fmt.Sprintf("Torrent already exists in Deluge: %s", result.Title))
			return true
		}
		logger.WriteError(fmt.Sprintf("Failed to add magnet to Deluge for %s. Error: %v", result.Title, err), err)
		return false
	}

	logger.WriteInfo(fmt.Sprintf("Successfully sent to Deluge: %s", result.Title))
	return true
}
