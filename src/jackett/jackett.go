package jackett

import (
	"context"
	"fmt"
	"high-seas/src/deluge"
	"high-seas/src/logger"
	"high-seas/src/utils"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/webtor-io/go-jackett"
)

var apiKey = utils.EnvVar("JACKETT_API_KEY", "")
var ip = utils.EnvVar("JACKETT_IP", "")
var port = utils.EnvVar("JACKETT_PORT", "")

func MakeMovieQuery(query string, title string, year string, Imdb uint, description string) {
	var sizeOfTorrent []uint

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
		Categories: []uint{2000, 2010, 2020, 2030, 2040, 2050, 2060, 2070, 2080},
		Query:      query,
	})
	if err != nil {
		logger.WriteFatal("Failed to fetch from Jackett.", err)
	}

	for i := 0; i < len(resp.Results); i++ {
		sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
	}

	maxSeeders := slices.Max(sizeOfTorrent)
	var selectedTorrent *jackett.Result

	for _, r := range resp.Results {
		if isCorrectMovie(r, title, description, years[0], Imdb) {
			logger.WriteInfo(fmt.Sprintf("This is the Jackett Title ==> %s. This is the TMDb Title ==> %s. This is the Jackett Seeders ==> %s, The max seeders is ==> %d", r.Title, title, r.Seeders, maxSeeders))

			if r.Seeders == maxSeeders {
				selectedTorrent = &r
				break
			}
		}
	}

	if selectedTorrent != nil {
		link := selectedTorrent.Link
		logger.WriteInfo(link)

		// Try adding the torrent with the highest seeder value
		err := deluge.AddTorrent(link)
		if err != nil {
			logger.WriteError("Failed to add torrent with the highest seeder value.", err)

			// If adding the torrent fails, try the next highest seeder value
			sortedTorrents := sortTorrentsBySeeders(resp.Results)
			for _, torrent := range sortedTorrents {
				if isCorrectMovie(torrent, title, description, years[0], Imdb) && torrent.Seeders < maxSeeders {
					link := torrent.Link
					logger.WriteInfo(link)

					err := deluge.AddTorrent(link)
					if err == nil {
						break
					} else {
						logger.WriteError("Failed to add torrent with the next highest seeder value.", err)
					}
				}
			}
		}
	} else {
		logger.WriteInfo("No matching torrent found with the maximum number of seeders.")
	}
}

func sortTorrentsBySeeders(torrents []jackett.Result) []jackett.Result {
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Seeders > torrents[j].Seeders
	})
	return torrents
}

func MakeShowQuery(query string, seasons []int, name string, year string, description string) {
	name = strings.Replace(name, ":", "", -1)

	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	for i := 0; i < len(seasons); i++ {
		season := i + 1
		episodes := seasons[i]

		for episode := 1; episode <= episodes; episode++ {
			var sizeOfTorrent []uint

			if season < 10 {
				if episode < 10 {
					queryString := fmt.Sprintf("%s S0%dE0%d", query, season, episode)
					queryStringWSpace := fmt.Sprintf("%s S0%d E0%d", query, season, episode)
					searchQueries := []string{queryString, queryStringWSpace}

					for _, q := range searchQueries {
						logger.WriteInfo(q)

						resp, err := j.Fetch(ctx, &jackett.FetchRequest{
							Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
							Query:      q,
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
				} else {
					queryString := fmt.Sprintf("%s S0%dE%d", query, season, episode)
					queryStringWSpace := fmt.Sprintf("%s S0%d E%d", query, season, episode)
					searchQueries := []string{queryString, queryStringWSpace}

					for _, q := range searchQueries {
						logger.WriteInfo(q)

						resp, err := j.Fetch(ctx, &jackett.FetchRequest{
							Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
							Query:      q,
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
			} else {
				if episode < 10 {
					queryString := fmt.Sprintf("%s S%dE0%d", query, season, episode)
					queryStringWSpace := fmt.Sprintf("%s S%d E0%d", query, season, episode)
					searchQueries := []string{queryString, queryStringWSpace}

					for _, q := range searchQueries {
						logger.WriteInfo(q)

						resp, err := j.Fetch(ctx, &jackett.FetchRequest{
							Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
							Query:      q,
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
				} else {
					queryString := fmt.Sprintf("%s S%dE%d", query, season, episode)
					queryStringWSpace := fmt.Sprintf("%s S%d E%d", query, season, episode)
					searchQueries := []string{queryString, queryStringWSpace}

					for _, q := range searchQueries {
						logger.WriteInfo(q)

						resp, err := j.Fetch(ctx, &jackett.FetchRequest{
							Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
							Query:      q,
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
				Categories: []uint{100060, 140679, 5070, 127720},
				Query:      queryString,
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

	if !containsAnyPart(r.Title, versions) && !compareDescriptions(r.Description, description) && !checkEpisodeTitlesAndDescriptions(r.Title, name) && !checkExternalIDs(r.TVDBId, r.Imdb) && !checkProductionInfo(r.Category) && !matchGenre(r.Category) {
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
		nameParts := strings.Split(name, ":")
		mainName := strings.TrimSpace(nameParts[0])

		if strings.Contains(r.Title, mainName) {
			// Extract the substring after the main show name
			mainNameIndex := strings.Index(r.Title, mainName)
			substringAfterMainName := strings.TrimSpace(r.Title[mainNameIndex+len(mainName):])

			// Check if the substring after the main show name is present in the original name
			if !strings.Contains(name, substringAfterMainName) {
				return false
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

func isCorrectMovie(r jackett.Result, title, description, year string, imdbID uint) bool {
	// Check if the title and year match
	// if !strings.Contains(r.Title, title) || !strings.Contains(r.Title, year) {
	// 	return false
	// }

	// Compare plot summaries and descriptions
	versions := createStringVersions(title)

	if !containsAnyPart(r.Title, versions) && !compareDescriptions(r.Description, description) && !checkExternalIDs(r.TVDBId, r.Imdb) && r.Imdb != imdbID && !matchGenre(r.Category) {
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
