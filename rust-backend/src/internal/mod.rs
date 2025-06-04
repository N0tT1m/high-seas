pub mod api {
    pub use once_cell::sync::Lazy;
    pub use std::fs::{File, OpenOptions};
    use std::sync::Mutex;
    use axum::response::IntoResponse;
    use reqwest;
    use reqwest::{Client, Error};
    use chrono::{Local};
    use std::io::Write;
    use tracing_subscriber::fmt::format;

    // Make the static variable public
    pub static LOG_FILE: Lazy<Mutex<File>> = Lazy::new(|| {
        let file = OpenOptions::new()
            .create(true)
            .append(true)
            .open("internal.log")
            .expect("Failed to create log file");
        Mutex::new(file)
    });

    fn log_to_file(message: &str) {
        if let Ok(mut file) = LOG_FILE.lock() {
            let timestamp = Local::now().format("%Y-%m-%d %H:%M:%S%.3f");
            let log_message = format!("[{}] {}\n", timestamp, message);
            let _ = file.write_all(log_message.as_bytes());
            println!("{}", log_message.trim()); // Also print to console
        }
    }

    pub async fn query_movie_request() -> impl IntoResponse {
        let client = Client::builder()
            .user_agent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
            .build();
        match client {
            Ok(client) => {
                log_to_file("Client was built.");
                
                println!("{:#?}", client)
            },
            Err(err) => {
                log_to_file(format!("Failed to build client: {err}").as_str());
            }
        }
    }
}

pub mod models {
    use std::any::Any;
    use std::collections::HashMap;
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
        created_by: HashMap<String, String>,
        episode_run_time: Vec<i32>,
        first_air_date: String,
        genres: Vec<Genres>,
        homepage: String,
        id: i32,
        in_production: bool,
        languages: Vec<String>,
        last_air_date: String,
        last_episode_to_air: HashMap<String, String>,
        name: String,
        next_episode_to_air: HashMap<String, String>,
        networks: HashMap<String, String>,
        number_of_episodes: i32,
        number_of_seasons: i32,
        origin_country: Vec<String>,
        original_language: String,
        original_name: String,
        overview: String,
        popularity: f64,
        poster_path: String,
        production_companies: HashMap<String, String>,
        production_countries: HashMap<String, String>,
        seasons: HashMap<String, String>,
        spoken_languages: HashMap<String, String>,
        status: String,
        tagline: String,
        #[serde(rename = "type")]
        tv_type: String,
        vote_average: f64,
        vote_count: i32,
        in_plex: bool,
    }
}

pub mod env {
    use std::fs::{File, OpenOptions};
    use once_cell::sync::Lazy;
    use tokio::sync::Mutex;
    use std::env;

    pub static LOG_FILE: Lazy<Mutex<File>> = Lazy::new(|| {
        let file = OpenOptions::new()
            .create(true)
            .append(true)
            .open("env.log")
            .expect("Failed to create log file");
        Mutex::new(file)
    });

    struct EnvReader {

    }

    impl EnvReader {
        fn new() -> Self {
            EnvReader {

            }
        }

        fn read_all_variables(&mut self) {
            for (key, value) in env::vars() {
                println!("{}: {}", key, value);
            }
        }
    }
}

pub mod db {
    pub use once_cell::sync::Lazy;
    pub use std::fs::{File, OpenOptions};
    use tokio::sync::Mutex;

    // Make the static variable public
    pub static LOG_FILE: Lazy<Mutex<File>> = Lazy::new(|| {
        let file = OpenOptions::new()
            .create(true)
            .append(true)
            .open("db.log")
            .expect("Failed to create log file");
        Mutex::new(file)
    });
}

