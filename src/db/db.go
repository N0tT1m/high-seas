package db

import (
	"fmt"
	"high-seas/src/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MovieRequest struct {
	ID      uint   `gorm:"primaryKey"`
	Query   string `json:"query"`
	TMDb    int    `json:"TMDb"`
	Quality string `json:"quality"`
}

type ShowRequest struct {
	ID      uint   `gorm:"primaryKey"`
	Query   string `json:"query"`
	Seasons []int  `json:"seasons"`
	TMDb    int    `json:"TMDb"`
	Quality string `json:"quality"`
}

// SeasonInfo struct to hold season and episode count
type SeasonInfo struct {
	SeasonNumber int
	EpisodeCount int
}

type AnimeMovieRequest struct {
	ID      uint   `gorm:"primaryKey"`
	Query   string `json:"query"`
	TMDb    int    `json:"TMDb"`
	Quality string `json:"quality"`
}

type AnimeTvRequest struct {
	ID      uint   `gorm:"primaryKey"`
	Query   string `json:"query"`
	Seasons []int  `json:"seasons"`
	TMDb    int    `json:"TMDb"`
	Quality string `json:"quality"`
}

type TMDbRequest struct {
	ID  uint   `gorm:"primaryKey"`
	Url string `json:"url"`
}

type TMDbTvShowsRequest struct {
	ID        uint   `gorm:"primaryKey"`
	Url       string `json:"url"`
	RequestID int    `json:"request_id"`
}

type TMDbResponse struct {
	Page         uint          `json:"page"`
	Results      []TMDbResults `json:"results"`
	TotalPages   uint          `json:"total_pages"`
	TotalResults uint          `json:"total_results"`
}

type TMDbGenreResponse struct {
	Genres []Genre `json:"genres"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MovieDetails represents the movie details response
type MovieDetails struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	InPlex      bool    `json:"in_plex"`
}

// TMDbMovieRequest represents the request structure for TMDb movie endpoints
type TMDbMovieRequest struct {
	URL string `json:"url"`
}

// TMDbDetailedMovieRequest includes request ID for detailed movie queries
type TMDbDetailedMovieRequest struct {
	URL       string `json:"url"`
	RequestID int    `json:"request_id"`
}

type TMDbResults struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	FirstAirDate     string  `json:"first_air_date"`
	GenreIds         []uint  `json:"genre_ids"`
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	OriginalLanguage string  `json:"original_language"`
	OriginalName     string  `json:"original_name"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        float64 `json:"vote_count"`
	Video            bool    `json:"video"`
}

type TVShow struct {
	Page         uint            `json:"page"`
	Results      []TVShowDetails `json:"results"`
	TotalPages   uint            `json:"total_pages"`
	TotalResults uint            `json:"total_results"`
}

type TVShowDetails struct {
	Adult               bool                     `json:"adult"`
	BackdropPath        string                   `json:"backdrop_path"`
	CreatedBy           []map[string]interface{} `json:"created_by"`
	EpisodeRunTime      []interface{}            `json:"episode_run_time"`
	FirstAirDate        string                   `json:"first_air_date"`
	Genres              []interface{}            `json:"genres"`
	Homepage            string                   `json:"homepage"`
	ID                  int                      `json:"id"`
	InProduction        bool                     `json:"in_production"`
	Languages           []string                 `json:"languages"`
	LastAirDate         string                   `json:"last_air_date"`
	LastEpisodeToAir    map[string]interface{}   `json:"last_episode_to_air"`
	Name                string                   `json:"name"`
	NextEpisodeToAir    map[string]interface{}   `json:"next_episode_to_air"`
	Networks            []map[string]interface{} `json:"networks"`
	NumberOfEpisodes    int                      `json:"number_of_episodes"`
	NumberOfSeasons     int                      `json:"number_of_seasons"`
	OriginCountry       []string                 `json:"origin_country"`
	OriginalLanguage    string                   `json:"original_language"`
	OriginalName        string                   `json:"original_name"`
	Overview            string                   `json:"overview"`
	Popularity          float64                  `json:"popularity"` // Assuming it can be a decimal
	PosterPath          string                   `json:"poster_path"`
	ProductionCompanies []map[string]interface{} `json:"production_companies"`
	ProductionCountries []map[string]interface{} `json:"production_countries"`
	Seasons             []map[string]interface{} `json:"seasons"`
	SpokenLanguages     []map[string]interface{} `json:"spoken_languages"`
	Status              string                   `json:"status"`
	Tagline             string                   `json:"tagline"`
	Type                string                   `json:"type"`
	VoteAverage         float64                  `json:"vote_average"` // Assuming it can be a decimal
	VoteCount           int                      `json:"vote_count"`
	InPlex              bool                     `json:"in_plex"`
}

var user = utils.EnvVar("DB_USER", "")
var password = utils.EnvVar("DB_PASSWORD", "")
var ip = utils.EnvVar("DB_IP", "")
var port = utils.EnvVar("DB_PORT", "")

func ConnectToDb() (*gorm.DB, error) {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/cs?charset=utf8mb4&parseTime=True&loc=Local", user, password, ip, port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	return db, err
}
