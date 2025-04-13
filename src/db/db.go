package db

//
import (
	"fmt"
	"high-seas/src/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Existing models from the application

// TMDbRequest represents a request to forward to TMDb API for TV shows
type TMDbRequest struct {
	Url string `json:"url"`
}

// TMDbTvShowsRequest represents a request for detailed TV show info with ID
type TMDbTvShowsRequest struct {
	Url       string `json:"url"`
	RequestID int    `json:"request_id"`
}

// TMDbMovieRequest represents a request to forward to TMDb API for movies
type TMDbMovieRequest struct {
	URL     string                 `json:"url"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// TMDbDetailedMovieRequest represents a detailed movie request with ID
type TMDbDetailedMovieRequest struct {
	URL       string `json:"url"`
	RequestID int    `json:"request_id"`
}

// TMDbResponse represents a response from TMDb API for TV shows
type TMDbResponse struct {
	Page         int        `json:"page"`
	Results      []TvResult `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
}

// TMDbMovieResponse represents a response from TMDb API for movies
type TMDbMovieResponse struct {
	Page         int           `json:"page"`
	Results      []MovieResult `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

// TMDbGenreResponse represents a genre response from TMDb API
type TMDbGenreResponse struct {
	Genres []Genre `json:"genres"`
}

// Genre represents a genre from TMDb API
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TvResult represents a TV show result from TMDb API
type TvResult struct {
	BackdropPath     string   `json:"backdrop_path"`
	FirstAirDate     string   `json:"first_air_date"`
	GenreIDs         []int    `json:"genre_ids"`
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
}

// MovieResult represents a movie result from TMDb API
type MovieResult struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

// TVShowDetails represents detailed TV show information
type TVShowDetails struct {
	Adult            bool      `json:"adult"`
	BackdropPath     string    `json:"backdrop_path"`
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	OriginalLanguage string    `json:"original_language"`
	OriginalName     string    `json:"original_name"`
	Overview         string    `json:"overview"`
	PosterPath       string    `json:"poster_path"`
	MediaType        string    `json:"media_type"`
	Popularity       float64   `json:"popularity"`
	FirstAirDate     string    `json:"first_air_date"`
	VoteAverage      float64   `json:"vote_average"`
	VoteCount        int       `json:"vote_count"`
	InPlex           bool      `json:"in_plex"`
	Seasons          []Season  `json:"seasons,omitempty"`
	NumberOfSeasons  int       `json:"number_of_seasons"`
	NumberOfEpisodes int       `json:"number_of_episodes"`
	Status           string    `json:"status"`
	Genres           []Genre   `json:"genres"`
	Networks         []Network `json:"networks"`
}

// Network represents a TV network
type Network struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

// Season represents a TV show season
type Season struct {
	AirDate      string  `json:"air_date"`
	EpisodeCount int     `json:"episode_count"`
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	SeasonNumber int     `json:"season_number"`
	VoteAverage  float64 `json:"vote_average"`
}

// Episode represents a TV show episode
type Episode struct {
	AirDate        string  `json:"air_date"`
	EpisodeNumber  int     `json:"episode_number"`
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	ProductionCode string  `json:"production_code"`
	Runtime        int     `json:"runtime"`
	SeasonNumber   int     `json:"season_number"`
	ShowID         int     `json:"show_id"`
	StillPath      string  `json:"still_path"`
	VoteAverage    float64 `json:"vote_average"`
	VoteCount      int     `json:"vote_count"`
}

// MovieDetails represents detailed movie information
type MovieDetails struct {
	Adult         bool    `json:"adult"`
	BackdropPath  string  `json:"backdrop_path"`
	Budget        int     `json:"budget"`
	Genres        []Genre `json:"genres"`
	Homepage      string  `json:"homepage"`
	ID            int     `json:"id"`
	ImdbID        string  `json:"imdb_id"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	Popularity    float64 `json:"popularity"`
	PosterPath    string  `json:"poster_path"`
	ReleaseDate   string  `json:"release_date"`
	Revenue       int     `json:"revenue"`
	Runtime       int     `json:"runtime"`
	Status        string  `json:"status"`
	Tagline       string  `json:"tagline"`
	Title         string  `json:"title"`
	Video         bool    `json:"video"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	InPlex        bool    `json:"in_plex"`
}

// MovieRequest represents a request to download a movie
type MovieRequest struct {
	Query       string `json:"query"`
	Quality     string `json:"quality"`
	TMDb        int    `json:"TMDb"`
	Description string `json:"description"`
	Year        int    `json:"year,omitempty"` // Now includes year
}

// ShowRequest represents a request to download a TV show
type ShowRequest struct {
	Query       string `json:"query"`
	Seasons     []int  `json:"seasons"`
	Quality     string `json:"quality"`
	TMDb        int    `json:"TMDb"`
	Description string `json:"description"`
	Year        int    `json:"year,omitempty"` // Now includes year
}

// AnimeMovieRequest represents a request to download an anime movie
type AnimeMovieRequest struct {
	Query       string `json:"query"`
	Name        string `json:"name"`
	Quality     string `json:"quality"`
	TMDb        int    `json:"TMDb"`
	Description string `json:"description"`
	Year        int    `json:"year,omitempty"` // Now includes year
}

// AnimeTvRequest represents a request to download an anime TV show
type AnimeTvRequest struct {
	Query       string `json:"query"`
	Seasons     []int  `json:"seasons"`
	Quality     string `json:"quality"`
	TMDb        int    `json:"TMDb"`
	Description string `json:"description"`
	Year        int    `json:"year,omitempty"` // Now includes year
}

// TvShowSeasonRequest represents a request for TV show season information
type TvShowSeasonRequest struct {
	ShowID       int `json:"show_id"`
	SeasonNumber int `json:"season_number,omitempty"`
}

// TvShowRecommendationsRequest represents a request for TV show recommendations
type TvShowRecommendationsRequest struct {
	ShowID int `json:"show_id"`
	Page   int `json:"page,omitempty"`
}

// MovieRecommendationsRequest represents a request for movie recommendations
type MovieRecommendationsRequest struct {
	MovieID int `json:"movie_id"`
	Page    int `json:"page,omitempty"`
}

// FilterOptions represents filter options for TMDb API
type FilterOptions struct {
	Genres         []int    `json:"genres,omitempty"`
	Year           int      `json:"year,omitempty"`
	YearStart      int      `json:"year_start,omitempty"`
	YearEnd        int      `json:"year_end,omitempty"`
	MinRating      float64  `json:"min_rating,omitempty"`
	MaxRating      float64  `json:"max_rating,omitempty"`
	Language       string   `json:"language,omitempty"`
	IncludeAdult   bool     `json:"include_adult,omitempty"`
	SortBy         string   `json:"sort_by,omitempty"`
	WithNetworks   []int    `json:"with_networks,omitempty"`
	WithCompanies  []int    `json:"with_companies,omitempty"`
	Status         string   `json:"status,omitempty"`
	Keywords       []string `json:"keywords,omitempty"`
	WatchProviders []int    `json:"watch_providers,omitempty"`
	Page           int      `json:"page,omitempty"`
	Region         string   `json:"region,omitempty"`
	RuntimeMin     int      `json:"runtime_min,omitempty"`
	RuntimeMax     int      `json:"runtime_max,omitempty"`
	WithType       string   `json:"with_type,omitempty"`
}

//	type MovieRequest struct {
//		ID      uint   `gorm:"primaryKey"`
//		Query   string `json:"query"`
//		TMDb    int    `json:"TMDb"`
//		Quality string `json:"quality"`
//		Year    int    `json:"year"` // Added Year field
//	}
//
//	type ShowRequest struct {
//		ID      uint   `gorm:"primaryKey"`
//		Query   string `json:"query"`
//		Seasons []int  `json:"seasons"`
//		TMDb    int    `json:"TMDb"`
//		Quality string `json:"quality"`
//		Year    int    `json:"year"` // Added Year field
//	}
//
// // SeasonInfo struct to hold season and episode count
//
//	type SeasonInfo struct {
//		SeasonNumber int
//		EpisodeCount int
//	}
//
//	type AnimeMovieRequest struct {
//		ID      uint   `gorm:"primaryKey"`
//		Query   string `json:"query"`
//		TMDb    int    `json:"TMDb"`
//		Quality string `json:"quality"`
//		Year    int    `json:"year"` // Added Year field
//	}
//
//	type AnimeTvRequest struct {
//		ID      uint   `gorm:"primaryKey"`
//		Query   string `json:"query"`
//		Seasons []int  `json:"seasons"`
//		TMDb    int    `json:"TMDb"`
//		Quality string `json:"quality"`
//		Year    int    `json:"year"` // Added Year field
//	}
//
//	type TMDbRequest struct {
//		ID  uint   `gorm:"primaryKey"`
//		Url string `json:"url"`
//	}
//
//	type TMDbTvShowsRequest struct {
//		ID        uint   `gorm:"primaryKey"`
//		Url       string `json:"url"`
//		RequestID int    `json:"request_id"`
//	}
//
//	type TMDbResponse struct {
//		Page         uint          `json:"page"`
//		Results      []TMDbResults `json:"results"`
//		TotalPages   uint          `json:"total_pages"`
//		TotalResults uint          `json:"total_results"`
//	}
//
//	type TMDbMovieResponse struct {
//		Page         uint               `json:"page"`
//		Results      []TMDbMovieResults `json:"results"`
//		TotalPages   uint               `json:"total_pages"`
//		TotalResults uint               `json:"total_results"`
//	}
//
//	type TMDbGenreResponse struct {
//		Genres []Genre `json:"genres"`
//	}
//
//	type Genre struct {
//		ID   int    `json:"id"`
//		Name string `json:"name"`
//	}
//
// // MovieDetails represents the movie details response
//
//	type MovieDetails struct {
//		ID          int     `json:"id"`
//		Title       string  `json:"title"`
//		Overview    string  `json:"overview"`
//		ReleaseDate string  `json:"release_date"`
//		VoteAverage float64 `json:"vote_average"`
//		InPlex      bool    `json:"in_plex"`
//	}
//
// // TMDbMovieRequest represents the request structure for TMDb movie endpoints
//
//	type TMDbMovieRequest struct {
//		URL string `json:"url"`
//	}
//
// // TMDbDetailedMovieRequest includes request ID for detailed movie queries
//
//	type TMDbDetailedMovieRequest struct {
//		URL       string `json:"url"`
//		RequestID int    `json:"request_id"`
//	}
//
//	type TMDbResults struct {
//		Adult            bool    `json:"adult"`
//		BackdropPath     string  `json:"backdrop_path"`
//		FirstAirDate     string  `json:"first_air_date"`
//		GenreIds         []uint  `json:"genre_ids"`
//		ID               int     `json:"id"`
//		Name             string  `json:"name"`
//		OriginalLanguage string  `json:"original_language"`
//		OriginalName     string  `json:"original_name"`
//		Overview         string  `json:"overview"`
//		Popularity       float64 `json:"popularity"`
//		PosterPath       string  `json:"poster_path"`
//		VoteAverage      float64 `json:"vote_average"`
//		VoteCount        float64 `json:"vote_count"`
//		Video            bool    `json:"video"`
//	}
//
//	type TMDbMovieResults struct {
//		Adult               bool                     `json:"adult"`
//		BackdropPath        string                   `json:"backdrop_path"`
//		BelongsToCollect    bool                     `json:"belongs_to_collect"`
//		Budget              int                      `json:"budget"`
//		GenreIds            []uint                   `json:"genre_ids"`
//		Homepage            string                   `json:"homepage"`
//		ID                  int                      `json:"id"`
//		IMDbID              int                      `json:"imdb_id"`
//		Title               string                   `json:"title"`
//		ReleaseDate         string                   `json:"release_date"`
//		OriginalLanguage    string                   `json:"original_language"`
//		OriginalTitle       string                   `json:"original_title"`
//		Overview            string                   `json:"overview"`
//		Popularity          float64                  `json:"popularity"`
//		PosterPath          string                   `json:"poster_path"`
//		ProductionCompanies []map[string]interface{} `json:"production_companies"`
//		ProductionCountries []map[string]interface{} `json:"production_countries"`
//		Tagline             string                   `json:"tagline"`
//		VoteAverage         float64                  `json:"vote_average"`
//		VoteCount           float64                  `json:"vote_count"`
//		Video               bool                     `json:"video"`
//	}
//
//	type TVShow struct {
//		Page         uint            `json:"page"`
//		Results      []TVShowDetails `json:"results"`
//		TotalPages   uint            `json:"total_pages"`
//		TotalResults uint            `json:"total_results"`
//	}
//
//	type TVShowDetails struct {
//		Adult               bool                     `json:"adult"`
//		BackdropPath        string                   `json:"backdrop_path"`
//		CreatedBy           []map[string]interface{} `json:"created_by"`
//		EpisodeRunTime      []interface{}            `json:"episode_run_time"`
//		FirstAirDate        string                   `json:"first_air_date"`
//		Genres              []interface{}            `json:"genres"`
//		Homepage            string                   `json:"homepage"`
//		ID                  int                      `json:"id"`
//		InProduction        bool                     `json:"in_production"`
//		Languages           []string                 `json:"languages"`
//		LastAirDate         string                   `json:"last_air_date"`
//		LastEpisodeToAir    map[string]interface{}   `json:"last_episode_to_air"`
//		Name                string                   `json:"name"`
//		NextEpisodeToAir    map[string]interface{}   `json:"next_episode_to_air"`
//		Networks            []map[string]interface{} `json:"networks"`
//		NumberOfEpisodes    int                      `json:"number_of_episodes"`
//		NumberOfSeasons     int                      `json:"number_of_seasons"`
//		OriginCountry       []string                 `json:"origin_country"`
//		OriginalLanguage    string                   `json:"original_language"`
//		OriginalName        string                   `json:"original_name"`
//		Overview            string                   `json:"overview"`
//		Popularity          float64                  `json:"popularity"` // Assuming it can be a decimal
//		PosterPath          string                   `json:"poster_path"`
//		ProductionCompanies []map[string]interface{} `json:"production_companies"`
//		ProductionCountries []map[string]interface{} `json:"production_countries"`
//		Seasons             []map[string]interface{} `json:"seasons"`
//		SpokenLanguages     []map[string]interface{} `json:"spoken_languages"`
//		Status              string                   `json:"status"`
//		Tagline             string                   `json:"tagline"`
//		Type                string                   `json:"type"`
//		VoteAverage         float64                  `json:"vote_average"` // Assuming it can be a decimal
//		VoteCount           int                      `json:"vote_count"`
//		InPlex              bool                     `json:"in_plex"`
//	}
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
