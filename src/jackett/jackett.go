package jackett

import (
	"context"
	"fmt"
	"high-seas/src/deluge"
	"high-seas/src/logger"
	"high-seas/src/utils"
	"regexp"
	"slices"
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

	for _, r := range resp.Results {
		if isCorrectMovie(r, title, description, years[0], Imdb) {
			logger.WriteInfo(fmt.Sprintf("This is the Jackett Imdb ID ==> %d. This is the TMDb ID ==> %d", r.Imdb, Imdb))

			if r.Seeders == slices.Max(sizeOfTorrent) {
				link := r.Link
				logger.WriteInfo(link)
				deluge.AddTorrent(link)
			}
		}
	}
}

func MakeShowQuery(query string, seasons []int, name string, year string, description string) {
	name = strings.Replace(name, ":", "", -1)

	ctx := context.Background()
	j := jackett.NewJackett(&jackett.Settings{
		ApiURL: fmt.Sprintf("http://%s:%s/", ip, port),
		ApiKey: fmt.Sprintf("%s", apiKey),
	})

	for i := 0; i < len(seasons); i++ {
		count := 1

		for count <= seasons[i] {
			var sizeOfTorrent []uint

			season := i + 1

			if i < 10 {
				fmt.Println("COUNT", count)
				queryString := fmt.Sprintf("%s S0%d E0%d", query, season, count)
				queryStringWSpace := fmt.Sprintf("%s S0%dE0%d", query, season, count)

				logger.WriteInfo(queryString)
				logger.WriteInfo(queryStringWSpace)

				resp, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
					Query:      queryString,
				})
				if err != nil {
					logger.WriteFatal("Failed to fetch from Jackett.", err)
				}
				resp2, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
					Query:      queryStringWSpace,
				})
				if err != nil {
					logger.WriteFatal("Failed to fetch from Jackett.", err)
				}

				for i := 0; i < len(resp.Results); i++ {
					if !slices.Contains(sizeOfTorrent, resp.Results[i].Seeders) {
						sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
					}
				}
				for i := 0; i < len(resp2.Results); i++ {
					if !slices.Contains(sizeOfTorrent, resp2.Results[i].Seeders) {
						sizeOfTorrent = append(sizeOfTorrent, resp2.Results[i].Seeders)
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
				for _, r := range resp2.Results {
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
			} else if i >= 10 {
				queryString := fmt.Sprintf("%s S%d E0%d", query, season, count)
				queryStringWSpace := fmt.Sprintf("%s S0%dE0%d", query, season, count)

				logger.WriteInfo(queryString)
				logger.WriteInfo(queryStringWSpace)

				resp, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
					Query:      queryString,
				})
				if err != nil {
					logger.WriteFatal("Failed to fetch from Jackett.", err)
				}
				resp2, err := j.Fetch(ctx, &jackett.FetchRequest{
					Categories: []uint{5000, 5010, 5020, 5030, 5040, 5050, 5060, 5070, 5080},
					Query:      queryStringWSpace,
				})
				if err != nil {
					logger.WriteFatal("Failed to fetch from Jackett.", err)
				}

				for i := 0; i < len(resp.Results); i++ {
					if !slices.Contains(sizeOfTorrent, resp.Results[i].Seeders) {
						sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
					}
				}
				for i := 0; i < len(resp2.Results); i++ {
					if !slices.Contains(sizeOfTorrent, resp2.Results[i].Seeders) {
						sizeOfTorrent = append(sizeOfTorrent, resp2.Results[i].Seeders)
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
				for _, r := range resp2.Results {
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

			count++
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

	if !containsAnyPart(r.Title, versions) || !compareDescriptions(r.Description, description) && !checkEpisodeTitlesAndDescriptions(r.Title, name) && !checkExternalIDs(r.TVDBId, r.Imdb) && !checkProductionInfo(r.Category) && !matchGenre(r.Category) {
		return false
	}

	// // Check air date
	// if !checkAirDate(r.PublishDate, year) {
	// 	return false
	// }

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
	logger.WriteInfo(description)
	logger.WriteInfo(resultDescription)

	return strings.Contains(resultDescription, description)
}

func checkEpisodeTitlesAndDescriptions(title, name string) bool {
	// Basic implementation: return true if the title contains the name
	logger.WriteInfo(title)
	logger.WriteInfo(name)

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
