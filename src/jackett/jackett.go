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

	queryString := fmt.Sprintf("%s %s", query, quality)
	logger.WriteInfo(fmt.Sprintf("Searching for movie: %s", queryString))

	resp, err := j.Fetch(ctx, &jackett.FetchRequest{
		Categories: []uint{2000, 2010, 2020, 2030, 2040, 2050, 2060, 2070, 2080}, // Movie categories
		Query:      queryString,
	})
	if err != nil {
		logger.WriteFatal("Failed to fetch from Jackett.", err)
		return err
	}

	results := processResults(resp.Results, tmdbID, quality, query)
	if len(results) > 0 {
		if addTorrentToDeluge(results[0].result) {
			return nil
		}
		return fmt.Errorf("failed to add movie torrent to Deluge")
	}

	return fmt.Errorf("no suitable matches found for movie: %s", query)
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

	// The result title should start with the exact title we're looking for
	return strings.HasPrefix(cleanResultTitle, cleanExactTitle)
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

func MakeAnimeQuery(query string, episodes int, name string, year string, description string) {
	//	name = strings.Replace(name, ":", "", -1)
	//
	//	ctx := context.Background()
	//	j := jackett.NewJackett(&jackett.Settings{
	//		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
	//		ApiKey: fmt.Sprintf("%s", apiKey),
	//	})

	//for i := 0; i < episodes; i++ {
	//	count := 1
	//
	//	for count <= episodes {
	//		var sizeOfTorrent []uint
	//
	//		season := i + 1
	//
	//		fmt.Println("season: ", season)
	//		queryString := fmt.Sprintf("%s %d", query, count)
	//
	//		logger.WriteInfo(queryString)
	//
	//		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
	//			Categories: []uint{100060, 140679, 5070, 127720},
	//			Query:      queryString,
	//		})
	//		if err != nil {
	//			logger.WriteFatal("Failed to fetch from Jackett.", err)
	//		}
	//
	//		for i := 0; i < len(resp.Results); i++ {
	//			if !slices.Contains(sizeOfTorrent, resp.Results[i].Seeders) {
	//				sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
	//			}
	//		}
	//
	//		for _, r := range resp.Results {
	//			if isCorrectAnime(r, name, year) {
	//				if strings.Contains(query, "One Piece") {
	//					if !strings.Contains(query, fmt.Sprintf("%d", 2023)) {
	//						logger.WriteInfo(r.Title)
	//					}
	//				} else {
	//					link := r.Link
	//					logger.WriteInfo(link)
	//				}
	//			}
	//		}
	//
	//		count++
	//	}
	//}
}
