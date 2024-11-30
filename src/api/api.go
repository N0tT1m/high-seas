package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"high-seas/src/db"
	"high-seas/src/jackett"
	"high-seas/src/logger"

	"github.com/gin-gonic/gin"
)

func CheckPlexStatus(name string, firstAirDate string) bool {
	plex := login()
	shows := plex.library.section('TV Shows')
	results := shows.search(title=name, year=firstAirDate[:4])
	return len(results) > 0
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

	// Add Plex status for each show
	for i := range response.Results {
		response.Results[i].InPlex = CheckPlexStatus(
			response.Results[i].Name,
			response.Results[i].FirstAirDate,
		)
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