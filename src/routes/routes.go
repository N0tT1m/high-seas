package routes

import (
	"crypto/tls"

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
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.Use(CORS)

	// Define routes based on the domain name
	r.POST("/movie/query", func(c *gin.Context) {
		host := c.Request.Host
		if host == "api.cinemacloud.tv" {
			api.QueryMovieRequest(c)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	r.POST("/show/query", func(c *gin.Context) {
		host := c.Request.Host
		if host == "api.cinemacloud.tv" {
			api.QueryShowRequest(c)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	r.POST("/anime/query", func(c *gin.Context) {
		host := c.Request.Host
		if host == "api.cinemacloud.tv" {
			api.QueryAnimeRequest(c)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	// r.Run(":8782")

	// Load the Let's Encrypt certificate and key
	certFile := "C:\\Certbot\\live\\api.cinemacloud.tv\\fullchain.pem"
	keyFile := "C:\\Certbot\\live\\api.cinemacloud.tv\\privkey.pem"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
	}

	// Create a custom TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Create a custom HTTP server with the TLS configuration
	server := &http.Server{
		Addr:      ":8782",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	// Start the server with TLS
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		panic(err)
	}
}
