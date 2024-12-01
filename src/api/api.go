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

func QueryGenres(c *gin.Context) {
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

	response, err := processTMDbGenreRequest(c, request.Url)
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

	response, err := processTMDbGenreRequest(c, request.Url)
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

func processMovieTMDbRequest(c *gin.Context, url string) (*db.TMDbResponse, error) {
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

	response, err := processMovieTMDbRequest(c, request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
