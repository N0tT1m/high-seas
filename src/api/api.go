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
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

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

func QueryAnimeRequest(c *gin.Context) {
	respBody := c.Request.Body

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.AnimeRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	jackett.MakeAnimeQuery(request.Query, request.Episodes, request.Name, request.Year, request.Description)

	logger.WriteCMDInfo("Read body complete.", "Success")

	c.JSON(http.StatusOK, gin.H{
		"message": "Query Request was successfully run.",
	})
}

func QueryTopRatedTvShows(c *gin.Context) {
	reqHeader := c.Request.Header
	header := reqHeader.Get("Authorization")

	logger.WriteInfo(fmt.Sprintf("Completed getting the header with the value: %s", header))

	reqBody := c.Request.Body
	body, err := io.ReadAll(reqBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.TMDbRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	url := strings.Trim(request.Url, " ")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.WriteError("Failed to create a new request.", err)
	}
	req.Header.Add("Authorization", header)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.WriteError("Failed to make a request.", err)
	}
	defer resp.Body.Close()

	logger.WriteInfo(fmt.Sprintf("Received %s back from '%s'", resp.Status, url))

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var response db.TMDbResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	logger.WriteInfo(fmt.Sprintf("Received: %v", response))

	c.JSON(http.StatusOK, response)
}

// Gets the first set of top rated shows
func QueryInitialTopRatedTvShows(c *gin.Context) {
	reqHeader := c.Request.Header
	header := reqHeader.Get("Authorization")

	logger.WriteInfo(fmt.Sprintf("Completed getting the header with the value: %s", header))

	reqBody := c.Request.Body
	body, err := io.ReadAll(reqBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.TMDbRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	url := strings.Trim(request.Url, " ")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.WriteError("Failed to create a new request.", err)
	}
	req.Header.Add("Authorization", header)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.WriteError("Failed to make a request.", err)
	}
	defer resp.Body.Close()

	logger.WriteInfo(fmt.Sprintf("Received %s back from '%s'", resp.Status, url))

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var response db.TMDbResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	logger.WriteInfo(fmt.Sprintf("Received: %v", response))

	c.JSON(http.StatusOK, response)
}

func QueryOnTheAirTvShows(c *gin.Context) {
	reqHeader := c.Request.Header
	header := reqHeader.Get("Authorization")

	logger.WriteInfo(fmt.Sprintf("Completed getting the header with the value: %s", header))

	reqBody := c.Request.Body
	body, err := io.ReadAll(reqBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.TMDbRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	url := strings.Trim(request.Url, " ")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.WriteError("Failed to create a new request.", err)
	}
	req.Header.Add("Authorization", header)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.WriteError("Failed to make a request.", err)
	}
	defer resp.Body.Close()

	logger.WriteInfo(fmt.Sprintf("Received %s back from '%s'", resp.Status, url))

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var response db.TMDbResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	logger.WriteInfo(fmt.Sprintf("Received: %v", response))

	c.JSON(http.StatusOK, response)
}

func QueryInitialOnTheAirTvShows(c *gin.Context) {
	reqHeader := c.Request.Header
	header := reqHeader.Get("Authorization")

	logger.WriteInfo(fmt.Sprintf("Completed getting the header with the value: %s", header))

	reqBody := c.Request.Body
	body, err := io.ReadAll(reqBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.TMDbRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	url := strings.Trim(request.Url, " ")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.WriteError("Failed to create a new request.", err)
	}
	req.Header.Add("Authorization", header)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.WriteError("Failed to make a request.", err)
	}
	defer resp.Body.Close()

	logger.WriteInfo(fmt.Sprintf("Received %s back from '%s'", resp.Status, url))

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var response db.TMDbResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	logger.WriteInfo(fmt.Sprintf("Received: %v", response))

	c.JSON(http.StatusOK, response)
}

func QueryPopularTvShows(c *gin.Context) {
	reqHeader := c.Request.Header
	header := reqHeader.Get("Authorization")

	logger.WriteInfo(fmt.Sprintf("Completed getting the header with the value: %s", header))

	reqBody := c.Request.Body
	body, err := io.ReadAll(reqBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.TMDbRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	url := strings.Trim(request.Url, " ")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.WriteError("Failed to create a new request.", err)
	}
	req.Header.Add("Authorization", header)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.WriteError("Failed to make a request.", err)
	}
	defer resp.Body.Close()

	logger.WriteInfo(fmt.Sprintf("Received %s back from '%s'", resp.Status, url))

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var response db.TMDbResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	logger.WriteInfo(fmt.Sprintf("Received: %v", response))

	c.JSON(http.StatusOK, response)
}

func QueryInitialPopularTvShows(c *gin.Context) {
	reqHeader := c.Request.Header
	header := reqHeader.Get("Authorization")

	logger.WriteInfo(fmt.Sprintf("Completed getting the header with the value: %s", header))

	reqBody := c.Request.Body
	body, err := io.ReadAll(reqBody)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var request db.TMDbRequest

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	url := strings.Trim(request.Url, " ")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.WriteError("Failed to create a new request.", err)
	}
	req.Header.Add("Authorization", header)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.WriteError("Failed to make a request.", err)
	}
	defer resp.Body.Close()

	logger.WriteInfo(fmt.Sprintf("Received %s back from '%s'", resp.Status, url))

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.WriteError("Failed to read the response body.", err)
	}

	var response db.TMDbResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.WriteError("Failed to Unmarshal JSON.", err)
	}

	logger.WriteInfo(fmt.Sprintf("Received: %v", response))

	c.JSON(http.StatusOK, response)
}
