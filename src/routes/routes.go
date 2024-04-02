package routes

import (
	"github.com/gin-gonic/gin"

	"high-seas/src/api"
	"net/http"
)

// CORS Middleware
func CORS(c *gin.Context) {

	// First, we add the headers with need to enable CORS
	// Make sure to adjust these headers to your needs
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	// Second, we handle the OPTIONS problem
	if c.Request.Method != "OPTIONS" {

		c.Next()

	} else {

		// Everytime we receive an OPTIONS request,
		// we just return an HTTP 200 Status Code
		// Like this, Angular can now do the real
		// request using any other method than OPTIONS
		c.AbortWithStatus(http.StatusOK)
	}
}

func SetupRouter() {

	r := gin.Default()

	r.Use(CORS)

	r.POST("/movie/query", api.QueryMovieRequest)

	r.POST("/show/query", api.QueryShowRequest)

	r.POST("/anime/query", api.QueryAnimeRequest)

	// Load the SSL/TLS certificate and key
	err := r.RunTLS(":443", "server.crt", "server.key")
	if err != nil {
		panic(err)
	}
}
