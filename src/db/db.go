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

type AnimeRequest struct {
	ID          uint   `gorm:"primaryKey"`
	Query       string `json:"query"`
	Episodes    int    `json:"episodes"`
	Name        string `json:"name"`
	Year        string `json:"year"`
	Description string `json:"description"`
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
