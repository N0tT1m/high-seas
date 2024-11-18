package jackett

import (
	"context"
	"fmt"
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
	// var sizeOfTorrent []uint

	title = strings.Replace(title, ":", "", -1)
	// years := strings.Split(year, "-")

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
		fmt.Printf("Title: %s; TMDb: %d", resp.Results[i].Title, resp.Results[i].TMDb)
	}

	//for i := 0; i < len(resp.Results); i++ {
	//	sizeOfTorrent = append(sizeOfTorrent, resp.Results[i].Seeders)
	//}
	//
	//for _, r := range resp.Results {
	//	if isCorrectMovie(r, title, years[0], Imdb) {
	//		logger.WriteInfo(fmt.Sprintf("This is the Jackett Imdb ID ==> %d. This is the TMDb ID ==> %d", r.Imdb, Imdb))
	//
	//		if r.Imdb == Imdb {
	//			if r.Seeders == slices.Max(sizeOfTorrent) {
	//				link := r.Link
	//				logger.WriteInfo(link)
	//				// deluge.AddTorrent(link)
	//			}
	//		}
	//	}
	//}
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

	//var maxSeeders uint
	//if len(sizeOfTorrent) > 0 {
	//	maxSeeders = slices.Max(sizeOfTorrent)
	//}
	//
	//var selectedTorrent *jackett.Result
	//var maxScore float64

	for i := 0; i < len(resp.Results); i++ {
		respResults := fmt.Sprintf("Search Season Bundle: %v", resp.Results[i].PublishDate)
		logger.WriteInfo(respResults)
	}

	return false
}

func searchIndividualSeasons(ctx context.Context, j *jackett.Jackett, query string, seasons []int, name string, year string, description string) bool {
	for season := 1; season <= len(seasons); season++ {
		var sizeOfTorrent []uint
		seasonFormat := "%02d"
		if season >= 10 {
			seasonFormat = "%d"
		}
		queryString := fmt.Sprintf("%s S"+seasonFormat, query, season)
		logger.WriteInfo(queryString)

		resp, err := j.Fetch(ctx, &jackett.FetchRequest{
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

		//var maxSeeders uint
		//if len(sizeOfTorrent) > 0 {
		//	maxSeeders = slices.Max(sizeOfTorrent)
		//}
		//
		//var selectedTorrent *jackett.Result
		//var maxScore float64

		for i := 0; i < len(resp.Results); i++ {
			respResults := fmt.Sprintf("Search Individual Seasons: %v", resp.Results[i].PublishDate)
			logger.WriteInfo(respResults)
		}
	}

	return true
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

			// var maxSeeders uint
			//if len(sizeOfTorrent) > 0 {
			//	maxSeeders = slices.Max(sizeOfTorrent)
			//}

			// var selectedTorrent *jackett.Result
			// var maxScore float64

			for i := 0; i < len(resp.Results); i++ {
				respResults := fmt.Sprintf("Search Individual Episodes: %v", resp.Results[i].PublishDate)
				logger.WriteInfo(respResults)
			}

			//if selectedTorrent != nil {
			//	link := selectedTorrent.Link
			//	logger.WriteInfo(link)
			//
			//	// Try adding the torrent with the highest seeder value
			//	err := saveFileToRemotePC(selectedTorrent)
			//	if err != nil {
			//		logger.WriteError("Failed to add torrent with the highest seeder value.", err)
			//
			//		// If adding the torrent fails, try the next highest seeder value
			//		sortedTorrents := sortTorrentsBySeeders(resp.Results)
			//		for _, torrent := range sortedTorrents {
			//			if isCorrectShow(torrent, name, year, description) && torrent.Seeders < maxSeeders {
			//				link := torrent.Link
			//				logger.WriteInfo(link)
			//
			//				err := saveFileToRemotePC(selectedTorrent)
			//				if err == nil {
			//					break
			//				} else {
			//					logger.WriteError("Failed to add torrent with the next highest seeder value.", err)
			//				}
			//			}
			//		}
			//	}
			//} else {
			//	logger.WriteInfo("No matching torrent found with the maximum number of seeders and high quality.")
			//}
		}
	}
}

func containsEpisodeText(title string) bool {
	// Use regular expression to check if the title contains any episode text
	episodeRegex := regexp.MustCompile(`(?i)(?:e\d+|episode\s*\d+)`)
	return episodeRegex.MatchString(title)
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

//func isCorrectShow(r jackett.Result, name, year string) bool {
//	// Check if the name and year match
//	if !strings.Contains(r.Title, name) || !strings.Contains(r.Title, year) {
//		return false
//	}
//
//	// Compare plot summaries and descriptions
//	if !compareDescriptions(r.Description, name, year) {
//		return false
//	}
//
//	// Check episode titles and descriptions
//	if !checkEpisodeTitlesAndDescriptions(r.Title, name) {
//		return false
//	}
//
//	// Check air date
//	if !checkAirDate(r.PublishDate, year) {
//		return false
//	}
//
//	// Check TMDb, TVDB, or IMDb ID
//	if !checkExternalIDs(r.Tvdbid, r.Imdb) {
//		return false
//	}
//
//	// Check production company and country of origin
//	if !checkProductionInfo(r.Publisher, r.Categories) {
//		return false
//	}
//
//	// Match the genre
//	if !matchGenre(r.Categories) {
//		return false
//	}
//
//	return true
//}
//
//func isCorrectMovie(r jackett.Result, title, year string, imdbID uint) bool {
//	// Check if the title and year match
//	if !strings.Contains(r.Title, title) || !strings.Contains(r.Title, year) {
//		return false
//	}
//
//	// Compare plot summaries and descriptions
//	if !compareDescriptions(r.Description, title, year) {
//		return false
//	}
//
//	// Check rating
//	if !checkRating(r.Rating) {
//		return false
//	}
//
//	// Check cast
//	if !checkCast(r.Actors) {
//		return false
//	}
//
//	// Check release date
//	if !checkReleaseDate(r.PublishDate, year) {
//		return false
//	}
//
//	// Check TMDb, TVDB, or IMDb ID
//	if !checkExternalIDs(r.Tvdbid, r.Imdb) || r.Imdb != imdbID {
//		return false
//	}
//
//	// Check production company and country of origin
//	if !checkProductionInfo(r.Publisher, r.Categories) {
//		return false
//	}
//
//	// Match the genre
//	if !matchGenre(r.Categories) {
//		return false
//	}
//
//	return true
//}
//
//func isCorrectAnime(r jackett.Result, name, year string) bool {
//	// Check if the name and year match
//	if !strings.Contains(r.Title, name) || !strings.Contains(r.Title, year) {
//		return false
//	}
//
//	// Compare plot summaries and descriptions
//	if !compareDescriptions(r.Description, name, year) {
//		return false
//	}
//
//	// Check episode titles and descriptions
//	if !checkEpisodeTitlesAndDescriptions(r.Title, name) {
//		return false
//	}
//
//	// Check air date
//	if !checkAirDate(r.PublishDate, year) {
//		return false
//	}
//
//	// Check TMDb, TVDB, or IMDb ID
//	if !checkExternalIDs(r.Tvdbid, r.Imdb) {
//		return false
//	}
//
//	// Check production company and country of origin
//	if !checkProductionInfo(r.Publisher, r.Categories) {
//		return false
//	}
//
//	// Match the genre
//	if !matchGenre(r.Categories) {
//		return false
//	}
//
//	return true
//}
//
//// Helper functions for matching criteria (not implemented in this example)
//func compareDescriptions(description, title, year string) bool {
//	// Implement logic to compare plot summaries and descriptions
//	return true
//}
//
//func checkEpisodeTitlesAndDescriptions(title, name string) bool {
//	// Implement logic to check episode titles and descriptions
//	return true
//}
//
//func checkAirDate(publishDate, year string) bool {
//	// Implement logic to check air date
//	return true
//}
//
//func checkRating(rating float64) bool {
//	// Implement logic to check rating
//	return true
//}
//
//func checkCast(actors string) bool {
//	// Implement logic to check cast
//	return true
//}
//
//func checkReleaseDate(publishDate, year string) bool {
//	// Implement logic to check release date
//	return true
//}
//
//func checkExternalIDs(tvdbID, imdbID uint) bool {
//	// Implement logic to check TMDb, TVDB, or IMDb ID
//	return true
//}
//
//func checkProductionInfo(publisher string, categories []int) bool {
//	// Implement logic to check production company and country of origin
//	return true
//}
//
//func matchGenre(categories []int) bool {
//	// Implement logic to match genre
//	return true
//}
