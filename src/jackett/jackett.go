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
	"strings"
	"time"
)

var apiKey = utils.EnvVar("JACKETT_API_KEY", "")
var ip = utils.EnvVar("JACKETT_IP", "")
var port = utils.EnvVar("JACKETT_PORT", "")

// Result scoring constants
const (
	TITLE_MATCH_WEIGHT   = 0.35 // Reduced from 0.4
	SEEDERS_WEIGHT       = 0.35 // Increased from 0.3
	QUALITY_WEIGHT       = 0.2  // Kept the same
	SIZE_WEIGHT          = 0.1  // Kept the same
	MIN_ACCEPTABLE_SCORE = 0.5  // Reduced from 0.6
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
func MakeMovieQuery(query string, tmdbID int, quality string) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	// Try multiple search strategies for better results
	searchStrategies := []string{
		fmt.Sprintf("\"%s\" %s", query, quality), // Exact title with quotes
		fmt.Sprintf("%s %s", query, quality),      // Regular search
		query, // Title only as fallback
	}

	logger.WriteInfo(fmt.Sprintf("Searching for movie: %s", query))

	for i, queryString := range searchStrategies {
		logger.WriteInfo(fmt.Sprintf("Movie search strategy %d: %s", i+1, queryString))
		
		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: []uint{2000, 2010, 2020, 2030, 2040, 2050, 2060, 2070, 2080}, // Movie categories
			Query:      queryString,
		})
		if err != nil {
			logger.WriteError(fmt.Sprintf("Strategy %d failed", i+1), err)
			continue
		}

		results := processMovieResults(resp.Results, tmdbID, quality, query)
		if len(results) > 0 {
			if addTorrentToDeluge(results[0].result) {
				return nil
			}
		}
		
		// Add delay between strategies
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("no suitable matches found for movie: %s", query)
}

// Specialized function for processing movie results with better validation
func processMovieResults(results []jackett.Result, tmdbID int, quality string, exactTitle string) []searchResult {
	var scoredResults []searchResult

	logger.WriteInfo(fmt.Sprintf("Processing %d movie results", len(results)))

	for _, result := range results {
		// Skip results that don't match the exact movie title
		if !isExactMovieMatch(result.Title, exactTitle) {
			logger.WriteInfo(fmt.Sprintf("Skipping non-matching movie title: %s", result.Title))
			continue
		}

		score := calculateScore(&result, tmdbID, quality)
		if score >= MIN_ACCEPTABLE_SCORE {
			scoredResults = append(scoredResults, searchResult{
				result: &result,
				score:  score,
			})
			logger.WriteInfo(fmt.Sprintf("Added movie candidate: %s (Score: %.2f)", result.Title, score))
		}
	}

	// Sort by score first, then by seeders
	sort.Slice(scoredResults, func(i, j int) bool {
		if scoredResults[i].score == scoredResults[j].score {
			return scoredResults[i].result.Seeders > scoredResults[j].result.Seeders
		}
		return scoredResults[i].score > scoredResults[j].score
	})

	if len(scoredResults) > 0 {
		logger.WriteInfo(fmt.Sprintf("Selected best movie match: %s (Score: %.2f, Size: %.2f GB)",
			scoredResults[0].result.Title,
			scoredResults[0].score,
			float64(scoredResults[0].result.Size)/1024/1024/1024))
	}

	return scoredResults
}

// Specialized function for movie title matching
func isExactMovieMatch(resultTitle, exactTitle string) bool {
	// Clean both titles for comparison
	cleanResultTitle := cleanTitleForComparison(resultTitle)
	cleanExactTitle := cleanExactTitle(exactTitle)

	// Exact match check first
	if cleanResultTitle == cleanExactTitle {
		return true
	}

	// For movies, we're more strict about exact matches
	// The result title should start with the exact title we're looking for
	if !strings.HasPrefix(cleanResultTitle, cleanExactTitle) {
		return false
	}

	// Check remainder for movie-specific exclusions
	remainder := strings.TrimSpace(strings.TrimPrefix(cleanResultTitle, cleanExactTitle))
	
	if remainder != "" {
		// Movie-specific spinoff/sequel indicators
		movieSpinoffIndicators := []string{
			"ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x",
			"2", "3", "4", "5", "6", "7", "8", "9",
			"part 2", "part 3", "part ii", "part iii",
			"sequel", "prequel", "origins", "begins", "returns",
			"reloaded", "revolution", "resurrection", "redemption",
			"extended cut", "director cut", "unrated", "theatrical",
			"special edition", "ultimate edition", "collectors edition",
		}
		
		lowerRemainder := strings.ToLower(remainder)
		for _, indicator := range movieSpinoffIndicators {
			if strings.Contains(lowerRemainder, indicator) {
				return false
			}
		}
	}

	return true
}

func MakeShowQuery(query string, seasons []int, tmdbID int, quality string) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	totalSeasons := len(seasons)
	logger.WriteInfo(fmt.Sprintf("Starting search for %s with %d total seasons", query, totalSeasons))

	// Print out the episode counts for each season for debugging
	for i, count := range seasons {
		logger.WriteInfo(fmt.Sprintf("Season %d has %d episodes", i+1, count))
	}

	// Step 1: Try complete series bundle
	if searchFullSeriesBundle(ctx, j, query, seasons, tmdbID, quality) {
		return nil
	}

	// Step 2: Try season bundles or individual episodes
	currentSeason := 1
	for currentSeason <= totalSeasons {
		// Try to find season pack first
		if found := searchCompleteSeason(ctx, j, query, currentSeason, seasons[currentSeason-1], tmdbID, quality); !found {
			// If season pack not found, search episode by episode
			logger.WriteInfo(fmt.Sprintf("No season pack found for season %d, searching %d individual episodes",
				currentSeason, seasons[currentSeason-1]))
			searchSeasonEpisodesByOne(ctx, j, query, currentSeason, tmdbID, quality, seasons[currentSeason-1])
		}
		currentSeason++
	}

	return nil
}

func searchCompleteSeason(ctx context.Context, j *jackett.Jackett, query string, season, episodeCount, tmdbID int, quality string) bool {
	seasonFormat := fmt.Sprintf("S%02d", season)
	searchQueries := []string{
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

		results := processResults(resp.Results, tmdbID, quality, query)
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

func searchFullSeriesBundle(ctx context.Context, j *jackett.Jackett, query string, seasons []int, tmdbID int, quality string) bool {
	searchQueries := []string{
		fmt.Sprintf("%s complete series %s", query, quality),
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

		results := processResults(resp.Results, tmdbID, quality, query)
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
func searchSeasonEpisodesByOne(ctx context.Context, j *jackett.Jackett, query string, season, tmdbID int, quality string, episodeCount int) bool {
	logger.WriteInfo(fmt.Sprintf("Searching for %d individual episodes of season %d", episodeCount, season))
	successCount := 0
	var missingEpisodes []int

	for episode := 1; episode <= episodeCount; episode++ {
		episodeFormat := fmt.Sprintf("S%02dE%02d", season, episode)
		// Add quotes around the show title to ensure exact matching
		queryString := fmt.Sprintf("\"%s\" %s %s", query, episodeFormat, quality)

		logger.WriteInfo(fmt.Sprintf("Searching for episode: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: []uint{5000, 5020, 5030, 5040, 5045},
			Query:      queryString,
		})
		if err != nil {
			missingEpisodes = append(missingEpisodes, episode)
			continue
		}

		results := processResults(resp.Results, tmdbID, quality, query)
		if len(results) > 0 {
			bestResult := selectBestResult(results)
			if bestResult != nil && addTorrentToDeluge(bestResult) {
				successCount++
				logger.WriteInfo(fmt.Sprintf("Successfully added %s (%d/%d)",
					episodeFormat, successCount, episodeCount))
			} else {
				missingEpisodes = append(missingEpisodes, episode)
			}
		} else {
			logger.WriteWarning(fmt.Sprintf("No results found for %s", episodeFormat))
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

// Modify processResults to include more logging
func processResults(results []jackett.Result, tmdbID int, quality string, exactTitle string) []searchResult {
	var scoredResults []searchResult

	logger.WriteInfo(fmt.Sprintf("Processing %d results", len(results)))

	for _, result := range results {
		// Skip results that don't match the exact show title
		if !isExactShowMatch(result.Title, exactTitle) {
			logger.WriteInfo(fmt.Sprintf("Skipping non-matching title: %s", result.Title))
			continue
		}

		score := calculateScore(&result, tmdbID, quality)
		if score >= MIN_ACCEPTABLE_SCORE {
			scoredResults = append(scoredResults, searchResult{
				result: &result,
				score:  score,
			})
			logger.WriteInfo(fmt.Sprintf("Added to candidates: %s (Score: %.2f)", result.Title, score))
		}
	}

	// Sort by size in descending order
	sort.Slice(scoredResults, func(i, j int) bool {
		return scoredResults[i].result.Size > scoredResults[j].result.Size
	})

	if len(scoredResults) > 0 {
		logger.WriteInfo(fmt.Sprintf("Selected best match: %s (Size: %.2f GB)",
			scoredResults[0].result.Title,
			float64(scoredResults[0].result.Size)/1024/1024/1024))
	}

	return scoredResults
}

func isExactShowMatch(resultTitle, exactTitle string) bool {
	// Clean both titles for comparison
	cleanResultTitle := cleanTitleForComparison(resultTitle)
	cleanExactTitle := cleanExactTitle(exactTitle)

	// Exact match check first
	if cleanResultTitle == cleanExactTitle {
		return true
	}

	// The result title should start with the exact title we're looking for
	if !strings.HasPrefix(cleanResultTitle, cleanExactTitle) {
		return false
	}

	// Additional validation to avoid spinoffs/sequels
	remainder := strings.TrimSpace(strings.TrimPrefix(cleanResultTitle, cleanExactTitle))
	
	// If there's a remainder, check if it's likely a spinoff/sequel
	if remainder != "" {
		// Common indicators of spinoffs/sequels/prequels that we want to avoid
		spinoffIndicators := []string{
			"origins", "legacy", "reloaded", "returns", "begins", "reborn",
			"resurrection", "revolution", "evolution", "redemption", "revenge",
			"rise of", "fall of", "war", "battle", "saga", "chronicles",
			"extended", "director", "special", "ultimate", "definitive",
			"prequel", "sequel", "spinoff", "spin off", "origins", "zero",
			"infinity", "unlimited", "maximum", "extreme", "deluxe",
		}
		
		lowerRemainder := strings.ToLower(remainder)
		for _, indicator := range spinoffIndicators {
			if strings.Contains(lowerRemainder, indicator) {
				return false
			}
		}
		
		// If remainder contains numbers (like "2", "3", "II", "III"), likely a sequel
		if containsSequelNumbers(remainder) {
			return false
		}
	}

	return true
}

// Helper function to detect sequel numbering
func containsSequelNumbers(text string) bool {
	text = strings.ToLower(text)
	
	// Roman numerals and common sequel patterns
	sequelPatterns := []string{
		" ii", " iii", " iv", " v", " vi", " vii", " viii", " ix", " x",
		" 2", " 3", " 4", " 5", " 6", " 7", " 8", " 9",
		"part 2", "part 3", "part 4", "part ii", "part iii",
		"volume 2", "volume 3", "vol 2", "vol 3",
		"season 2", "season 3", "season 4", // For different seasons being treated as sequels
	}
	
	for _, pattern := range sequelPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	
	return false
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

	// Sort by size in descending order
	sort.Slice(results, func(i, j int) bool {
		return results[i].result.Size > results[j].result.Size
	})

	return results[0].result
}

func calculateScore(result *jackett.Result, tmdbID int, quality string) float64 {
	score := 0.0

	// Title match score (NEW)
	titleScore := calculateTitleMatch(result.Title)
	score += titleScore * TITLE_MATCH_WEIGHT

	// TMDb match bonus (reduced since we now have title matching)
	if result.TMDb > 0 && int(result.TMDb) == tmdbID {
		score += 0.3 // Reduced from 0.5 since we now have title matching
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

// New function to calculate title match score
func calculateTitleMatch(title string) float64 {
	// Remove common strings that don't affect title matching
	cleanTitle := strings.ToLower(title)
	removeStrings := []string{
		"2160p", "1080p", "720p", "480p",
		"webrip", "web-dl", "bluray", "hdtv",
		"x264", "x265", "hevc", "h264", "h265",
		"dv", "hdr", "sdr",
		"aac", "ac3", "dts", "ddp", "eac3",
		"atmos",
	}

	for _, s := range removeStrings {
		cleanTitle = strings.ReplaceAll(cleanTitle, strings.ToLower(s), "")
	}

	// Remove special characters and extra spaces
	cleanTitle = strings.TrimSpace(cleanTitle)

	// Basic scoring criteria
	score := 1.0

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

	return math.Max(0.0, score) // Ensure score doesn't go negative
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

// MakeAnimeMovieQuery handles searching and downloading anime movies with improved validation
func MakeAnimeMovieQuery(query string, tmdbID int, quality string) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	// Try with specific anime movie categories
	queryString := fmt.Sprintf("%s %s", query, quality)
	logger.WriteInfo(fmt.Sprintf("Searching for anime movie: %s", queryString))

	// First attempt with strict anime movie categories
	if err := searchAnimeMovie(ctx, j, queryString, tmdbID, quality); err == nil {
		return nil
	}

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

		results := processAnimeResults(resp.Results, tmdbID, quality, query)
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

func MakeAnimeShowQuery(query string, seasons []int, tmdbID int, quality string) error {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	totalEpisodes := 0
	for _, episodeCount := range seasons {
		totalEpisodes += episodeCount
	}

	logger.WriteInfo(fmt.Sprintf("Starting search for anime series: %s with %d total episodes",
		query, totalEpisodes))

	// Try batch downloads first
	if found := tryAnimeBatchDownloads(ctx, j, query, tmdbID, quality); found {
		return nil
	}

	// If batch download fails, try episode by episode
	return searchAnimeEpisodesByOne(ctx, j, query, tmdbID, quality, seasons)
}

func isAnimeTimeRelease(title string) bool {
	return strings.Contains(title, "[Anime Time]")
}

func tryAnimeBatchDownloads(ctx context.Context, j *jackett.Jackett, query string, tmdbID int, quality string) bool {
	// Exact Anime Time pattern matching the format
	animeTimePatterns := []string{
		`[Anime Time] %s (Series+Movies) 2011 [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub] [Batch]`,
		// Slightly more generic fallbacks but still maintaining Anime Time format
		`[Anime Time] %s [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub] [Batch]`,
		`[Anime Time] %s [Batch]`,
	}

	// Try Anime Time patterns first
	for _, pattern := range animeTimePatterns {
		queryString := fmt.Sprintf(pattern, query)
		logger.WriteInfo(fmt.Sprintf("Trying Anime Time batch search with query: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeSeriesCategories,
			Query:      queryString,
		})

		if err == nil && len(resp.Results) > 0 {
			results := processAnimeResults(resp.Results, tmdbID, quality, query)
			for _, result := range results {
				if isAnimeTimeRelease(result.result.Title) && addTorrentMagnetToDeluge(result.result) {
					logger.WriteInfo(fmt.Sprintf("Successfully added Anime Time batch: %s", result.result.Title))
					return true
				}
			}
		}

		time.Sleep(searchDelay)
	}

	// Fallback patterns if Anime Time isn't found
	fallbackPatterns := []string{
		`[AnimeSkulls] %s (2011) [Batch] [Dual Audio][1080p][HEVC 10bit x265]`,
		`%s complete series`,
		`%s batch`,
	}

	for _, pattern := range fallbackPatterns {
		queryString := fmt.Sprintf(pattern, query)
		logger.WriteInfo(fmt.Sprintf("Trying fallback batch search with query: %s", queryString))

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			Categories: animeSeriesCategories,
			Query:      queryString,
		})

		if err != nil {
			continue
		}

		results := processAnimeResults(resp.Results, tmdbID, quality, query)
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

func searchAnimeEpisodesByOne(ctx context.Context, j *jackett.Jackett, query string, tmdbID int, quality string, seasons []int) error {
	logger.WriteInfo(fmt.Sprintf("Starting season-based anime search for %d seasons", len(seasons)))

	// Search by season and episode instead of sequential episodes
	for seasonNum := 1; seasonNum <= len(seasons); seasonNum++ {
		episodeCount := seasons[seasonNum-1]
		logger.WriteInfo(fmt.Sprintf("Searching season %d with %d episodes", seasonNum, episodeCount))

		for episode := 1; episode <= episodeCount; episode++ {
			// Anime Time episode patterns with proper season formatting
			animeTimePatterns := []string{
				`[Anime Time] %s S%02dE%02d [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]`,
				`[Anime Time] %s - S%02dE%02d [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]`,
				`[Anime Time] %s Season %d Episode %02d [Dual Audio][BD][1080p][HEVC 10bit x265][AAC][Eng Sub]`,
			}

			// Fallback patterns with season support
			fallbackPatterns := []string{
				`[AnimeSkulls] %s S%02dE%02d [1080p] [Dual.Audio] [x265] [RD]`,
				`%s S%02dE%02d`,
				`%s Season %d Episode %d`,
			}

			var foundEpisode bool

			// Try Anime Time patterns first
			for _, pattern := range animeTimePatterns {
				var queryString string
				if strings.Contains(pattern, "Season %d Episode") {
					queryString = fmt.Sprintf(pattern, query, seasonNum, episode)
				} else {
					queryString = fmt.Sprintf(pattern, query, seasonNum, episode)
				}
				logger.WriteInfo(fmt.Sprintf("Searching Anime Time for S%02dE%02d using query: %s",
					seasonNum, episode, queryString))

				resp, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: animeSeriesCategories,
					Query:      queryString,
				})

				if err == nil {
					results := processAnimeResults(resp.Results, tmdbID, quality, query)
					for _, result := range results {
						if isAnimeTimeRelease(result.result.Title) && addTorrentMagnetToDeluge(result.result) {
							logger.WriteInfo(fmt.Sprintf("Successfully added Anime Time S%02dE%02d: %s",
								seasonNum, episode, result.result.Title))
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
			// If Anime Time release not found, try fallback patterns
			if !foundEpisode {
				for _, pattern := range fallbackPatterns {
					var queryString string
					if strings.Contains(pattern, "Season %d Episode") {
						queryString = fmt.Sprintf(pattern, query, seasonNum, episode)
					} else {
						queryString = fmt.Sprintf(pattern, query, seasonNum, episode)
					}
					logger.WriteInfo(fmt.Sprintf("Searching fallback for S%02dE%02d using query: %s",
						seasonNum, episode, queryString))

					resp, err := j.Fetch(ctx, &jackett.FetchRequest{
						Categories: animeSeriesCategories,
						Query:      queryString,
					})

					if err != nil {
						continue
					}

					results := processAnimeResults(resp.Results, tmdbID, quality, query)
					if len(results) > 0 {
						for _, result := range results {
							if addTorrentMagnetToDeluge(result.result) {
								logger.WriteInfo(fmt.Sprintf("Successfully added S%02dE%02d: %s",
									seasonNum, episode, result.result.Title))
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
				logger.WriteWarning(fmt.Sprintf("No valid results found for S%02dE%02d", seasonNum, episode))
			}
		}
	}

	return nil
}

func searchAnimeMovie(ctx context.Context, j *jackett.Jackett, query string, tmdbID int, quality string) error {
	// Try different search patterns
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

		results := processAnimeResults(resp.Results, tmdbID, quality, query)
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

func processAnimeResults(results []jackett.Result, tmdbID int, quality string, query string) []searchResult {
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

		score := calculateAnimeScore(&result, tmdbID, quality)

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

	// Sort results by score and then seeders
	sort.Slice(scoredResults, func(i, j int) bool {
		if scoredResults[i].score == scoredResults[j].score {
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

func calculateAnimeScore(result *jackett.Result, tmdbID int, quality string) float64 {
	score := 0.0

	// Base score
	score += 0.1

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
