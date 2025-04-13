package api

import (
	"encoding/json"
	"fmt"
	"github.com/jrudio/go-plex-client"
	"high-seas/src/db"
	"high-seas/src/jackett"
	"high-seas/src/logger"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// QueryMovieRequest handles requests to search and download movies
func QueryMovieRequest(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	var request db.MovieRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Use the Year field from the request for searching
	err = jackett.MakeMovieQuery(request.Query, request.TMDb, request.Quality, request.Year)
	if err != nil {
		logger.WriteError("Failed to query movie.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process movie query"})
		return
	}

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

// QueryShowRequest handles requests to search and download TV shows
func QueryShowRequest(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	var request db.ShowRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Log request details for debugging
	logger.WriteInfo(fmt.Sprintf("Processing show request: %s (TMDb ID: %d, Year: %d)",
		request.Query, request.TMDb, request.Year))

	// Use the Year field from the request for searching
	err = jackett.MakeShowQuery(request.Query, request.Seasons, request.TMDb, request.Quality, request.Year)
	if err != nil {
		logger.WriteError("Failed to query show.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process show query"})
		return
	}

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

// QueryAnimeMovieRequest handles requests to search and download anime movies
func QueryAnimeMovieRequest(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	var request db.AnimeMovieRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Use the Year field from the request for searching
	err = jackett.MakeAnimeMovieQuery(request.Query, request.TMDb, request.Quality, request.Year)
	if err != nil {
		logger.WriteError("Failed to Query Anime Movie Request.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process anime movie query"})
		return
	}

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

// MakeAnimeShowQuery handles requests to search and download anime TV shows
func MakeAnimeShowQuery(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	var request db.AnimeTvRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Use the Year field from the request for searching
	err = jackett.MakeAnimeShowQuery(request.Query, request.Seasons, request.TMDb, request.Quality, request.Year)
	if err != nil {
		logger.WriteError("Failed to process anime show query.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process anime show query"})
		return
	}

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

func login() *plex.Plex {
	plexConnection, err := plex.New("http://192.168.1.78:32400", "Y7fU6x3PPqr8A-P3WEjq")
	if err != nil {
		logger.WriteError("Failed to connect to Plex:", err)
	}

	_, err = plexConnection.Test()
	if err != nil {
		logger.WriteError("Failed to connect to Plex:", err)
	}

	return plexConnection
}

func CheckPlexStatus(title string) bool {
	plex := login()

	fmt.Println(fmt.Sprintf("Checking Plex Status: %s", title))

	searchResults, err := plex.Search(title)
	if err != nil {
		logger.WriteError("search failed for title: ", err)
	}

	for _, v := range searchResults.MediaContainer.Metadata {
		if v.ParentTitle == "" && v.GrandparentTitle == "" {
			return true
		}
	}

	//results := shows.search(title=name, year=firstAirDate[:4])
	//return len(results) > 0

	return false
}

func processTMDbRequest(c *gin.Context, url string) (*db.TMDbResponse, error) {
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

	return &response, nil
}

func processTMDbGenreRequest(c *gin.Context, url string) (*db.TMDbGenreResponse, error) {
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

	var response db.TMDbGenreResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
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

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryShowGenres(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbGenreRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllShowsForDetails(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryAllShowsFromSelectedDate(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func processDetailedTMDbRequest(c *gin.Context, url string, requestId int) (*db.TVShowDetails, error) {
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

	var response db.TVShowDetails
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	fmt.Println(response.InPlex)

	fmt.Println(response.ID, requestId)

	if response.ID == requestId {
		response.InPlex = CheckPlexStatus(response.Name)
	}

	fmt.Println(response.InPlex)

	fmt.Println(response)

	return &response, nil
}

func QueryDetailedTopRatedTvShows(c *gin.Context) {
	var request db.TMDbTvShowsRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processDetailedTMDbRequest(c, request.Url, request.RequestID)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func processMovieTMDbRequest(c *gin.Context, url string) (*db.TMDbMovieResponse, error) {
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

	fmt.Println(string(body))

	var response db.TMDbMovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &response, nil
}

func processDetailedMovieTMDbRequest(c *gin.Context, url string, requestID int) (*db.MovieDetails, error) {
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

	var response db.MovieDetails
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	if response.ID == requestID {
		response.InPlex = CheckPlexStatus(response.Title)
	}

	return &response, nil
}

// Movie endpoints
func QueryTopRatedMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryPopularMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryNowPlayingMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryUpcomingMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMovieDetails(c *gin.Context) {
	var request db.TMDbDetailedMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processDetailedMovieTMDbRequest(c, request.URL, request.RequestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMoviesByGenre(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMovieSearch(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func QueryMovieGenres(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbGenreRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// New TV show endpoint handlers

// QueryTvShowSeasons handles requests for TV show seasons and episodes
func QueryTvShowSeasons(c *gin.Context) {
	var request db.TMDbTvShowsRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processDetailedTMDbRequest(c, request.Url, request.RequestID)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QueryTvShowRecommendations handles requests for TV show recommendations
func QueryTvShowRecommendations(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QuerySimilarTvShows handles requests for similar TV shows
func QuerySimilarTvShows(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QueryShowsByGenre handles requests for TV shows filtered by genre
func QueryShowsByGenre(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QueryShowSearch handles search requests for TV shows
func QueryShowSearch(c *gin.Context) {
	var request db.TMDbRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processTMDbRequest(c, request.Url)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// New movie endpoint handlers

// QueryMovieRecommendations handles requests for movie recommendations
func QueryMovieRecommendations(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QuerySimilarMovies handles requests for similar movies
func QuerySimilarMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QueryAllMovies handles requests for all movies with filters
func QueryAllMovies(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QueryAllMoviesForDetails handles requests for movie details in discover section
func QueryAllMoviesForDetails(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QueryAllMoviesFromSelectedDate handles requests for movies from specific dates
func QueryAllMoviesFromSelectedDate(c *gin.Context) {
	var request db.TMDbMovieRequest
	if err := c.BindJSON(&request); err != nil {
		logger.WriteError("Failed to bind request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process TMDb URL with date filters
	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		logger.WriteError("Failed to process TMDb request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
