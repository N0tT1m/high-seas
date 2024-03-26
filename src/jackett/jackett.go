package jackett

import (
	"context"
	"fmt"
	"high-seas/src/deluge"
	"high-seas/src/logger"
	"high-seas/src/utils"
	"slices"
	"strings"

	"github.com/webtor-io/go-jackett"
)

var apiKey = utils.EnvVar("JACKETT_API_KEY", "")
var ip = utils.EnvVar("JACKETT_IP", "")
var port = utils.EnvVar("JACKETT_PORT", "")

func MakeMovieQuery(query string, name string, year string, Imdb uint) {
	var sizeOfTorrent []uint

	name = strings.Replace(name, ":", "", -1)
	years := strings.Split(year, "-")

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
		if strings.Contains(r.Title, name) && strings.Contains(r.Title, years[0]) {

			logger.WriteInfo(fmt.Sprintf("This is the Jackett Imdb ID ==> %d. This is the TMDb ID ==> %d", r.Imdb, Imdb))

			if r.Imdb == Imdb {
				if r.Seeders == slices.Max(sizeOfTorrent) {

					link := r.Link

					logger.WriteInfo(link)

					// deluge.AddTorrent(link)
				}
			}
		}
	}
}

func MakeShowQuery(query string, seasons []int, name string, year string) {
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
					if strings.Contains(r.Title, name) {
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
					if strings.Contains(r.Title, name) {
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
					if strings.Contains(r.Title, name) {
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
					if strings.Contains(r.Title, name) {
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

func MakeAnimeQuery(query string, episodes int, name string, year string) {
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
				if strings.Contains(r.Title, name) {
					if strings.Contains(query, "One Piece") {
						if !strings.Contains(query, fmt.Sprintf("%d", 2023)) {
							// link := r.Link

							// logger.WriteInfo(link)

							// links = append(links, link)

							// deluge.AddTorrent(link)

							logger.WriteInfo(r.Title)
						}
					} else {
						link := r.Link

						logger.WriteInfo(link)

						// links = append(links, link)

						// deluge.AddTorrent(link)
					}
				}
			}

			count++
		}
	}
}
