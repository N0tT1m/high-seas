package jackett

import (
	"bufio"
	"context"
	"fmt"
	"high-seas/src/deluge"
	"high-seas/src/logger"
	"high-seas/src/utils"
	"io"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/webtor-io/go-jackett"

	"os"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"

	"net/http"
)

var apiKey = utils.EnvVar("JACKETT_API_KEY", "")
var ip = utils.EnvVar("JACKETT_IP", "")
var port = utils.EnvVar("JACKETT_PORT", "")

func MakeMovieQuery(query string, title string, year string, Imdb uint, description string) {
	var sizeOfTorrent []uint
	var qualityOfTorrent []uint

	title = strings.Replace(title, ":", "", -1)
	years := strings.Split(year, "-")

	logger.WriteInfo(title)
	logger.WriteInfo(years)

	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	resp, err := j.Fetch(ctx, &jackett.FetchRequest{
		// Categories: []uint{2000, 2010, 2020, 2030, 2040, 2050, 2060, 2070, 2080},
		Query: query,
	})
	if err != nil {
		logger.WriteFatal("Failed to fetch from Jackett.", err)
	}

	for i := 0; i < len(resp.Results); i++ {
		sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
		qualityOfTorrent = append(qualityOfTorrent, resp.Results[i].Size)
	}

	maxSeeders := slices.Max(sizeOfTorrent)
	var selectedTorrent *jackett.Result
	var maxScore float64

	for i := 0; i < len(resp.Results); i++ {
		if isCorrectMovie(resp.Results[i], title, description, years[0], Imdb) {
			// Calculate a score based on seeders and size
			score := calculateScore(resp.Results[i].Seeders, resp.Results[i].Size)

			logger.WriteInfo(fmt.Sprintf("This is the Jackett Title ==> %s. This is the TMDb Title ==> %s. This is the Jackett Seeders ==> %s, The score is ==> %.2f", resp.Results[i].Title, title, resp.Results[i].Seeders, score))

			if score > maxScore {
				maxScore = score
				selectedTorrent = &resp.Results[i]
			}
		}
	}

	if selectedTorrent != nil {
		link := selectedTorrent.Link
		logger.WriteInfo(link)

		// Try adding the torrent with the highest seeder value
		err := saveFileToRemotePC(selectedTorrent)
		if err != nil {
			logger.WriteError("Failed to add torrent with the highest seeder value.", err)

			// If adding the torrent fails, try the next highest seeder value
			sortedTorrents := sortTorrentsBySeeders(resp.Results)
			for _, torrent := range sortedTorrents {
				if isCorrectMovie(torrent, title, description, years[0], Imdb) && torrent.Seeders < maxSeeders { // isHighQuality(torrent.Size) &&
					link := torrent.Link
					logger.WriteInfo(link)

					err := saveFileToRemotePC(selectedTorrent)
					if err == nil {
						break
					} else {
						logger.WriteError("Failed to add torrent with the next highest seeder value.", err)
					}
				}
			}
		}
	} else {
		logger.WriteInfo("No matching torrent found with the maximum number of seeders and high quality.")
	}
}

func saveFileToRemotePC(torrent *jackett.Result) error {
	// Implement the logic to save the file to the remote PC
	// This is a placeholder function and needs to be implemented based on your specific requirements
	fileName := fmt.Sprintf("%s.torrent", torrent.Title)
	remoteFilePath := fmt.Sprintf("/home/timmy/torrents/%s", fileName)

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(torrent.Link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Example: Use SSH to copy the file to the remote PC
	// You'll need to implement this part based on your specific setup
	err = copyFileToRemotePC(fileName, remoteFilePath)
	if err != nil {
		logger.WriteError("Failed to save file to remote PC", err)
		return err
		// Implement fallback logic or retry mechanism if needed
	} else {
		logger.WriteInfo(fmt.Sprintf("File saved successfully: %s", remoteFilePath))
	}

	return nil
}

func copyFileToRemotePC(sourceURL, destinationPath string) error {
	sshConfig, err := auth.PasswordKey("timmy", "B@bycakes15!", nil)
	if err != nil {
		return err
	}

	scpClient := scp.NewClient("192.168.1.92:22", &sshConfig)

	err = scpClient.Connect()
	if err != nil {
		return err
	}

	fileData, err := os.Open("/Users/local-user-name/Desktop/test.txt")
	if err != nil {
		return err
	}

	ctx := context.Background()

	file, err := os.Open(sourceURL)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)

	scpClient.CopyFile(ctx, reader, destinationPath, "0655")

	defer scpClient.Close()
	defer fileData.Close()

	return nil
}
func calculateScore(seeders uint, size uint) float64 {
	// Normalize the seeders and size values
	normalizedSeeders := float64(seeders) / 1000.0
	normalizedSize := float64(size) / (1024.0 * 1024.0 * 1024.0) // Convert size to GB

	// Assign weights to seeders and size
	seederWeight := 0.7
	sizeWeight := 0.3

	// Calculate the score
	score := seederWeight*normalizedSeeders + sizeWeight*normalizedSize
	return score
}

func sortTorrentsBySeeders(torrents []jackett.Result) []jackett.Result {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Seeders > torrents[j].Seeders
	})
	return torrents
}

func isHighQuality(size uint) {
	// Check if the size indicates 1080p or 4k quality
	// You can modify this logic based on your specific criteria
	// return strings.Contains(size, "1080p") || strings.Contains(size, "4k")
}

func MakeShowQuery(query string, seasons []int, name string, year string, description string) {
	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	// Check if all seasons are available in a bundle
	bundleFound := searchSeasonBundle(ctx, j, query, seasons, name, year, description)

	if !bundleFound {
		// If a complete bundle is not found, search for individual seasons
		seasonFound := searchIndividualSeasons(ctx, j, query, seasons, name, year, description)

		if !seasonFound {
			searchIndividualEpisodes(ctx, j, query, seasons, name, year, description)
		}
	}
}

func searchSeasonBundle(ctx context.Context, j *jackett.Jackett, query string, seasons []int, name string, year string, description string) bool {
	seasonBundleFormat := "S%02d-S%02d"
	if len(seasons) >= 10 {
		seasonBundleFormat = "S%d-S%d"
	}
	seasonBundle := fmt.Sprintf(seasonBundleFormat, 1, len(seasons))
	queryString := fmt.Sprintf("%s %s", query, seasonBundle)
	logger.WriteInfo(queryString)

	resp, err := j.Fetch(ctx, &jackett.FetchRequest{
		// Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
		Query: queryString,
	})
	if err != nil {
		logger.WriteFatal("Failed to fetch from Jackett.", err)
	}

	var sizeOfTorrent []uint
	for i := 0; i < len(resp.Results); i++ {
		if !slices.Contains(sizeOfTorrent, resp.Results[i].Seeders) {
			sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
		}
	}

	sortedTorrents := sortTorrentsBySeeders(resp.Results)
	for _, r := range sortedTorrents {
		tmdbOutput := fmt.Sprintf("The TMDb from Jackett is --> %s.", r.Tracker)
		logger.WriteInfo(tmdbOutput)
		if compareBundle(r, name) {
			fmt.Println(r.Title)
			link := r.Link
			logger.WriteInfo(link)
			err := deluge.AddTorrent(link)
			if err == nil {
				return true
			} else {
				logger.WriteError("Failed to add torrent.", err)
			}
		}
	}

	return false
}

func searchIndividualSeasons(ctx context.Context, j *jackett.Jackett, query string, seasons []int, name string, year string, description string) bool {
	season := 1
	for season <= len(seasons) {
		var sizeOfTorrent []uint
		seasonFormat := "%02d"
		if season >= 10 {
			seasonFormat = "%d"
		}
		queryString := fmt.Sprintf("%s S"+seasonFormat, query, season)
		logger.WriteInfo(queryString)
		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
			// Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
			Query: queryString,
		})
		if err != nil {
			logger.WriteFatal("Failed to fetch from Jackett.", err)
		}
		for i := 0; i < len(resp.Results); i++ {
			if !slices.Contains(sizeOfTorrent, resp.Results[i].Seeders) {
				sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
			}
		}
		sortedTorrents := sortTorrentsBySeeders(resp.Results)
		found := false
		for _, r := range sortedTorrents {
			// Check if the result title contains any episode text
			if containsEpisodeText(r.Title) {
				continue // Skip this result and move to the next one
			}
			tmdbOutput := fmt.Sprintf("The TMDb from Jackett is --> %s.", r.Tracker)
			logger.WriteInfo(tmdbOutput)
			if isCorrectShow(r, name, year, description) {
				fmt.Println(r.Title)
				link := r.Link
				logger.WriteInfo(link)
				err := deluge.AddTorrent(link)

				season++

				if err != nil {
					logger.WriteError("Failed to add torrent.", err)
				} else {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func containsEpisodeText(title string) bool {
	// Use regular expression to check if the title contains any episode text
	episodeRegex := regexp.MustCompile(`(?i)(?:e\d+|episode\s*\d+)`)
	return episodeRegex.MatchString(title)
}

func searchIndividualEpisodes(ctx context.Context, j *jackett.Jackett, query string, seasons []int, name string, year string, description string) {
	var sizeOfTorrent []uint

	// Search for a bundle of seasons
	for startSeason := 1; startSeason <= len(seasons); startSeason++ {
		for endSeason := startSeason; endSeason <= len(seasons); endSeason++ {
			seasonBundleFormat := "S%02d-S%02d"
			if startSeason >= 10 || endSeason >= 10 {
				seasonBundleFormat = "S%d-S%d"
			}

			seasonBundle := fmt.Sprintf(seasonBundleFormat, startSeason, endSeason)
			queryString := fmt.Sprintf("%s %s", query, seasonBundle)

			logger.WriteInfo(queryString)

			resp, err := j.Fetch(ctx, &jackett.FetchRequest{
				// Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
				Query: queryString,
			})
			if err != nil {
				logger.WriteFatal("Failed to fetch from Jackett.", err)
			}

			for i := 0; i < len(resp.Results); i++ {
				if !slices.Contains(sizeOfTorrent, resp.Results[i].Seeders) {
					sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
				}
			}

			for _, r := range resp.Results {
				tmdbOutput := fmt.Sprintf("The TMDb from Jackett is --> %s.", r.Tracker)
				logger.WriteInfo(tmdbOutput)
				if isCorrectShow(r, name, year, description) {
					fmt.Println(r.Title)

					if r.Seeders == slices.Max(sizeOfTorrent) {
						link := r.Link
						logger.WriteInfo(link)
						deluge.AddTorrent(link)
					}
				}
			}
		}
	}
}

func MakeAnimeQuery(query string, episodes int, name string, year string, description string) {
	name = strings.Replace(name, ":", "", -1)

	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	for i := 0; i < episodes; i++ {
		count := 1

		for count <= episodes {
			var sizeOfTorrent []uint

			season := i + 1

			fmt.Println("season: ", season)
			queryString := fmt.Sprintf("%s %d", query, count)

			logger.WriteInfo(queryString)

			resp, err := j.Fetch(ctx, &jackett.FetchRequest{
				// Categories: []uint{100060, 140679, 5070, 127720},
				Query: queryString,
			})
			if err != nil {
				logger.WriteFatal("Failed to fetch from Jackett.", err)
			}

			for i := 0; i < len(resp.Results); i++ {
				if !slices.Contains(sizeOfTorrent, resp.Results[i].Seeders) {
					sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
				}
			}

			for _, r := range resp.Results {
				if isCorrectAnime(r, name, year, description) {
					if strings.Contains(query, "One Piece") {
						if !strings.Contains(query, fmt.Sprintf("%d", 2023)) {
							logger.WriteInfo(r.Title)
						}
					} else {
						link := r.Link
						logger.WriteInfo(link)
					}
				}
			}

			count++
		}
	}
}

func isCorrectShow(r jackett.Result, name, year, description string) bool {
	// Check if the name and year match
	versions := createStringVersions(name)
	if !containsAnyPart(r.Title, versions) || !compareDescriptions(r.Description, description) && !checkEpisodeTitlesAndDescriptions(r.Title, name) && !checkExternalIDs(r.TVDBId, r.Imdb) && !checkProductionInfo(r.Category) && !matchGenre(r.Category) {
		return false
	}

	// Check if the result title contains the exact show name or a variation
	exactMatch := false
	for _, version := range versions {
		if strings.Contains(r.Title, version) {
			exactMatch = true
			break
		}
	}

	if !exactMatch {
		// Check if the result title contains the main show name followed by extra text
		nameParts := strings.Fields(name)
		mainName := strings.TrimSpace(nameParts[0])
		if strings.Contains(r.Title, mainName) {
			// Extract the substring after the main show name
			mainNameIndex := strings.Index(r.Title, mainName)
			substringAfterMainName := strings.TrimSpace(r.Title[mainNameIndex+len(mainName):])

			// Check if the substring after the main show name is present in the original name
			if !strings.Contains(name, substringAfterMainName) {
				// Check if the extra text is separated by a colon, hyphen, or other common separators
				separators := []string{":", "-", "–", "|", "//", "."}
				isSeparated := false
				for _, sep := range separators {
					if strings.HasPrefix(substringAfterMainName, sep) {
						isSeparated = true
						break
					}
				}
				if !isSeparated {
					return false
				}
			}
		} else {
			// Check if the result title contains a different show with similar name
			if containsAnyPart(r.Title, nameParts) {
				return false
			}
		}
	}

	return true
}

func compareBundle(r jackett.Result, name string) bool {
	// Check if the result title contains the exact show name or a variation
	versions := createStringVersions(name)
	for _, version := range versions {
		if strings.Contains(r.Title, version) {
			return true
		}
	}

	return false
}

func isCorrectMovie(r jackett.Result, title, description, year string, imdbID uint) bool {
	// Check if the title and year match
	// if !strings.Contains(r.Title, title) || !strings.Contains(r.Title, year) {
	// 	return false
	// }

	// Compare plot summaries and descriptions
	versions := createStringVersions(title)

	if !containsAnyPart(r.Title, versions) || !compareDescriptions(r.Description, description) && !checkExternalIDs(r.TVDBId, r.Imdb) && r.Imdb != imdbID && !matchGenre(r.Category) {
		return false
	}

	// // Check release date
	// if !checkReleaseDate(r.PublishDate, year) {
	// 	return false
	// }

	return true
}

func isCorrectAnime(r jackett.Result, name, year, description string) bool {
	// Check if the name and year match
	// if !strings.Contains(r.Title, name) || !strings.Contains(r.Title, year) {
	// 	return false
	// }

	// Compare plot summaries and descriptions
	versions := createStringVersions(name)

	if !containsAnyPart(r.Title, versions) && !compareDescriptions(r.Description, description) && !checkEpisodeTitlesAndDescriptions(r.Title, name) && !checkExternalIDs(r.TVDBId, r.Imdb) && !matchGenre(r.Category) {
		return false
	}

	return true
}

// Helper functions for matching criteria
func compareDescriptions(resultDescription, description string) bool {
	// Basic implementation: return true if the description contains the title and year
	logger.WriteInfo("TMDb Description --> " + description)
	logger.WriteInfo("The Torrent Indexer Description --> " + resultDescription)

	return strings.Contains(resultDescription, description)
}

func checkEpisodeTitlesAndDescriptions(title, name string) bool {
	// Basic implementation: return true if the title contains the name
	logger.WriteInfo("The Torrent Indexer Title --> " + title)
	logger.WriteInfo("TMDb Description --> " + name)

	return strings.Contains(title, name)
}

func checkExternalIDs(tvdbID, imdbID uint) bool {
	// Basic implementation: return true if either tvdbID or imdbID is non-zero
	logger.WriteInfo(tvdbID)
	logger.WriteInfo(imdbID)

	return tvdbID != 0 || imdbID != 0
}

func checkProductionInfo(categories []uint) bool {
	// Basic implementation: always return true
	logger.WriteInfo(categories)

	return true
}

func matchGenre(categories []uint) bool {
	// Basic implementation: always return true
	logger.WriteInfo(categories)

	return true
}

func createStringVersions(str string) []string {
	// Create a slice to store the different versions
	versions := []string{str}

	// Generate versions with different combinations of '.', '-', and ' '
	separators := []string{".", "-", " "}
	for _, sep1 := range separators {
		for _, sep2 := range separators {
			for _, sep3 := range separators {
				version := strings.ReplaceAll(str, " ", sep1)
				version = strings.ReplaceAll(version, " ", sep2)
				version = strings.ReplaceAll(version, " ", sep3)
				versions = append(versions, version)
			}
		}
	}

	return versions
}

func containsAnyPart(str string, parts []string) bool {
	for _, part := range parts {
		// Check if the string contains ':'
		if strings.Contains(str, ":") {
			continue
		}

		// Remove numbers from the part if they are not present in the original part
		origPart := part
		hasNumber := containsNumber(part)
		part = removeNumbersIfNotPresent(part, origPart)

		if hasNumber {
			if containsNumber(str) && strings.Contains(str, part) {
				return true
			}
		} else if strings.Contains(str, part) {
			return true
		}
	}
	return false
}

func removeNumbersIfNotPresent(part, origPart string) string {
	// Check if the original part contains any numbers
	hasNumbers, _ := regexp.MatchString(`\d`, origPart)

	if !hasNumbers {
		// Remove numbers from the part using regular expression
		regex := regexp.MustCompile(`\d`)
		part = regex.ReplaceAllString(part, "")
	}
	return part
}

func containsNumber(str string) bool {
	// Check if the string contains any numbers using a regular expression
	hasNumber, _ := regexp.MatchString(`\d`, str)
	return hasNumber
}

// // Helper functions for matching criteria (not implemented in this example)
// func compareDescriptions(description, title, year string) bool {
// 	// Implement logic to compare plot summaries and descriptions
// 	return true
// }

// func checkEpisodeTitlesAndDescriptions(title, name string) bool {
// 	// Implement logic to check episode titles and descriptions
// 	return true
// }

// func checkAirDate(publishDate time.Time, year string) bool {
// 	// Implement logic to check air date
// 	return true
// }

// func checkRating(rating float64) bool {
// 	// Implement logic to check rating
// 	return true
// }

// func checkCast(actors string) bool {
// 	// Implement logic to check cast
// 	return true
// }

// func checkReleaseDate(publishDate, year string) bool {
// 	// Implement logic to check release date
// 	return true
// }

// func checkExternalIDs(tvdbID, imdbID uint) bool {
// 	// Implement logic to check TMDb, TVDB, or IMDb ID
// 	return true
// }

// func checkProductionInfo(categories []uint) bool {
// 	// Implement logic to check production company and country of origin
// 	return true
// }

// func matchGenre(categories []uint) bool {
// 	// Implement logic to match genre
// 	return true
// }
