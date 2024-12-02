use std::iter::Map;
use serde::{Serialize, Deserialize};

#[derive(Debug, Serialize, Deserialize)]
struct MovieRequest {
    id: u32,
    query: String,
    tmdb: i32,
    quality: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct ShowRequest {
    id: u32,
    query: String,
    seasons: Vec<i32>,
    tmdb: i32,
    quality: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SeasonInfo {
    season_number: i32,
    episode_count: i32,
}

#[derive(Debug, Serialize, Deserialize)]
struct AnimeMovieRequest {
    id: u32,
    query: String,
    tmdb: i32,
    quality: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct AnimeTvRequest {
    id: u32,
    query: String,
    tmdb: i32,
    quality: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbRequest {
    id: u32,
    url: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbTvShowsRequest {
    id: u32,
    url: String,
    request_id: i32,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbResponse {
    page: u32,
    results: Vec<TMDbResults>,
    total_pages: u32,
    total_results: u32,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbResults {
    adult: bool,
    backdrop_path: String,
    first_air_date: String,
    genre_ids: Vec<u32>,
    id: i32,
    name: String,
    original_language: String,
    original_name: String,
    overview: String,
    popularity: f64,
    poster_path: String,
    vote_average: f64,
    vote_count: f64,
    video: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbMovieResponse {
    page: u32,
    results: Vec<TMDbMovieResults>,
    total_pages: u32,
    total_results: u32,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbMovieResults {
    adult: bool,
    backdrop_path: String,
    genre_ids: Vec<u32>,
    id: i32,
    title: String,
    original_language: String,
    original_title: String,
    overview: String,
    popularity: f64,
    poster_path: String,
    vote_average: f64,
    vote_count: f64,
    video: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbGenreResponse {
    genres: Vec<Genres>,
}

#[derive(Debug, Serialize, Deserialize)]
struct Genres {
    id: u32,
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct MovieDetails {
    id: u32,
    title: String,
    overview: String,
    release_date: String,
    vote_average: f64,
    in_plex: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbMovieRequest {
    url: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct TMDbDetailedMovieRequest {
    url: String,
    request_id: i32,
}

#[derive(Debug, Serialize, Deserialize)]
struct TvShow {
    page: u32,
    results: Vec<TvShowDetails>,
    total_pages: u32,
    total_results: u32,
}

#[derive(Debug, Serialize, Deserialize)]
struct TvShowDetails {
    adult: bool,
    backdrop_path: String,
}