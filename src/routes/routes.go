package routes

import (
	"crypto/tls"
	"high-seas/src/api"

	"github.com/gin-gonic/gin"

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

	// Existing movie query routes
	r.POST("/movie/query", api.QueryMovieRequest)
	r.POST("/show/query", api.QueryShowRequest)
	r.POST("/anime/movie/query", api.QueryAnimeMovieRequest)
	r.POST("/anime/show/query", api.MakeAnimeShowQuery)

	// Existing TV show routes
	r.POST("/tmdb/show/top-rated-tv-shows", api.QueryTopRatedTvShows)
	r.POST("/tmdb/show/initial-top-rated-tv-shows", api.QueryInitialTopRatedTvShows)
	r.POST("/tmdb/show/on-the-air-tv-shows", api.QueryOnTheAirTvShows)
	r.POST("/tmdb/show/initial-on-the-air-tv-shows", api.QueryInitialOnTheAirTvShows)
	r.POST("/tmdb/show/popular-tv-shows", api.QueryPopularTvShows)
	r.POST("/tmdb/show/initial-popular-tv-shows", api.QueryInitialPopularTvShows)
	r.POST("/tmdb/show/airing-today-tv-shows", api.QueryAiringTodayTvShows)
	r.POST("/tmdb/show/initial-airing-today-tv-shows", api.QueryInitialAiringTodayTvShows)
	r.POST("/tmdb/show/all-shows", api.QueryAllTvShows)
	r.POST("/tmdb/show/all-show-details", api.QueryInitialAllTvShows)
	r.POST("/tmdb/show/tv-show-details", api.QueryDetailedTopRatedTvShows)
	r.POST("/tmdb/show/genres", api.QueryShowGenres)
	r.POST("/tmdb/show/all-tv-show-details", api.QueryAllShowsForDetails)
	r.POST("/tmdb/show/all-shows-from-date", api.QueryAllShowsFromSelectedDate)

	// Existing movie routes
	r.POST("/tmdb/movie/top-rated", api.QueryTopRatedMovies)
	r.POST("/tmdb/movie/popular", api.QueryPopularMovies)
	r.POST("/tmdb/movie/now-playing", api.QueryNowPlayingMovies)
	r.POST("/tmdb/movie/upcoming", api.QueryUpcomingMovies)
	r.POST("/tmdb/movie/details", api.QueryMovieDetails)
	r.POST("/tmdb/movie/by-genre", api.QueryMoviesByGenre)
	r.POST("/tmdb/movie/search", api.QueryMovieSearch)
	r.POST("/tmdb/movie/genres", api.QueryMovieGenres)

	// NEW TV show routes to match movie capabilities
	r.POST("/tmdb/show/seasons", api.QueryTvShowSeasons)
	r.POST("/tmdb/show/recommendations", api.QueryTvShowRecommendations)
	r.POST("/tmdb/show/similar", api.QuerySimilarTvShows)
	r.POST("/tmdb/show/by-genre", api.QueryShowsByGenre)
	r.POST("/tmdb/show/search", api.QueryShowSearch)

	// NEW movie routes to match TV capabilities
	r.POST("/tmdb/movie/recommendations", api.QueryMovieRecommendations)
	r.POST("/tmdb/movie/similar", api.QuerySimilarMovies)
	r.POST("/tmdb/movie/all-movies", api.QueryAllMovies)
	r.POST("/tmdb/movie/all-movie-details", api.QueryAllMoviesForDetails)
	r.POST("/tmdb/movie/all-movies-from-date", api.QueryAllMoviesFromSelectedDate)

	// // Define routes based on the domain name
	// r.POST("/movie/query", func(c *gin.Context) {
	// 	host := c.Request.Host
	// 	if host == "www.cinemacloud.tv" {
	// 		api.QueryMovieRequest(c)
	// 	} else {
	// 		c.AbortWithStatus(http.StatusNotFound)
	// 	}
	// })

	// r.POST("/show/query", func(c *gin.Context) {
	// 	host := c.Request.Host
	// 	if host == "www.cinemacloud.tv" {
	// 		api.QueryShowRequest(c)
	// 	} else {
	// 		c.AbortWithStatus(http.StatusNotFound)
	// 	}
	// })

	// r.POST("/anime/query", func(c *gin.Context) {
	// 	host := c.Request.Host
	// 	if host == "www.cinemacloud.tv" {
	// 		api.QueryAnimeRequest(c)
	// 	} else {
	// 		c.AbortWithStatus(http.StatusNotFound)
	// 	}
	// })

	// r.Run(":8782")

	// The commented TLS code below requires the "crypto/tls" import
	// If you uncomment this code, you'll need to uncomment the import too

	// Load the Let's Encrypt certificate and key
	certFile := "./fullchain.pem"
	keyFile := "./privkey.pem"
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
		Addr:      "192.168.1.71:8782",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	// Start the server with TLS
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		panic(err)
	}
}
