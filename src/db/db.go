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

type TMDbResponse struct {
	Page         uint          `json:"page"`
	Results      []TMDbResults `json:"results"`
	TotalPages   uint          `json:"total_pages"`
	TotalResults uint          `json:"total_results"`
}

type TMDbResults struct {
	Adult            bool   `json:"adult"`
	BackdropPath     string `json:"backdrop_path"`
	FirstAirDate     string `json:"first_air_date"`
	GenreIds         []uint `json:"genre_ids"`
	Id               uint   `json:"id"`
	Name             string `json:"name"`
	OriginalLanguage string `json:"original_language"`
	OriginalName     string `json:"original_name"`
	Overview         string `json:"overview"`
	Popularity       uint   `json:"popularity"`
	PosterPath       string `json:"poster_path"`
	VoteAverage      uint   `json:"vote_average"`
	VoteCount        uint   `json:"vote_count"`
	Video            bool   `json:"video"`
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
