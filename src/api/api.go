package api

import (
	"encoding/json"
	"fmt"
	"github.com/jrudio/go-plex-client"
	"high-seas/src/jackett"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"high-seas/src/db"
	"high-seas/src/logger"

	"github.com/gin-gonic/gin"
)

func login() *plex.Plex {
	plexConnection, err := plex.New("http://192.168.1.78:32400", "Y7fU6x3PPqr8A-P3WEjq")
	if err != nil {
		logger.WriteError("Failed to connect to Plex:", err)
	}

	results, err := plexConnection.Test()
	if err != nil {
		logger.WriteError("Failed to connect to Plex:", err)
	}

	logger.WriteInfo(fmt.Sprintf("Plex results: %v", results))

	return plexConnection
}

func CheckPlexStatus(isMovie bool, title string) bool {
	if isMovie {
		plex := login()

		libraries, err := plex.GetLibraries()

		if err != nil {
			logger.WriteError("failed fetching libraries: ", err)
		}

		for _, dir := range libraries.MediaContainer.Directory {
			fmt.Println(dir.Title)
		}

		logger.WriteInfo(fmt.Sprintf("Libraries: %v", libraries))

		//results := shows.search(title=name, year=firstAirDate[:4])
		//return len(results) > 0

		return true
	} else {
		plex := login()

		searchResults, err := plex.Search(title)
		if err != nil {
			logger.WriteError("search failed for title: ", err)
		}

		logger.WriteInfo(fmt.Sprintf("Search results: %v", searchResults))

		//results := shows.search(title=name, year=firstAirDate[:4])
		//return len(results) > 0

		return true
	}
}

func QueryMovieRequest(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.MovieRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	jackett.MakeMovieQuery(request.Query, request.TMDb, request.Quality)

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

func QueryShowRequest(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.ShowRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	// fmt.Println("{}", request.Query, request.Seasons, request.Name, request.Year, request.Description)

	jackett.MakeShowQuery(request.Query, request.Seasons, request.TMDb, request.Quality)

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

func QueryAnimeMovieRequest(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.AnimeMovieRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	err = jackett.MakeAnimeMovieQuery(request.Query, request.TMDb, request.Quality)
	if err != nil {
		logger.WriteError("Failed to Query Anime Movie Request.", err)
	}

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

func MakeAnimeShowQuery(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.AnimeTvRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	jackett.MakeAnimeShowQuery(request.Query, request.Seasons, request.TMDb, request.Quality)

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

func processShowTMDbRequest(c *gin.Context, url string) (*db.TMDbResponse, error) {
	header := c.Request.Header.Get("Authorization")

	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest("GET", strings.TrimSpace(url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", header)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response db.TMDbResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	// Add Plex status for each show
	for i := range response.Results {
		response.Results[i].InPlex = CheckPlexStatus(false, response.Results[i].Name)
	}

	return &response, nil
}

func QueryTopRatedTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialTopRatedTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryOnTheAirTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialOnTheAirTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryPopularTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialPopularTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAiringTodayTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialAiringTodayTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryInitialAllTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processShowTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
