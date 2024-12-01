interface TvShow {
  page: number,
  results: TvShowResult[]
  total_pages: number
  total_result: number
}

interface GenreRequest {
  genres: Genre[],
}

interface Genre {
  id: number,
  name: string,
}

interface TvShowResult {
  adult: boolean,
  backdrop_path: string,
  first_air_date: string,
  genre_ids: Array<number>,
  id: number,
  name: string,
  original_language: string,
  original_name: string,
  overview: string,
  popularity: number,
  poster_path: string,
  vote_average: number,
  vote_count: number,
  video: boolean,
}

interface Movie {
  page: number,
  results: MovieResult[]
  total_pages: number
  total_result: number
}

interface MovieResult {
  adult: boolean,
  backdrop_path: string,
  // first_air_date: string,
  genre_ids: Array<number>,
  id: number,
  title: string,
  release_date: string,
  original_language: string,
  original_title: string,
  overview: string,
  popularity: number,
  poster_path: string,
  vote_average: number,
  vote_count: number,
  video: boolean,
}

interface ShowDetails {
  adult: boolean,
  backdrop_path: string,
  created_by: Array<Map<string, any>>,
  episode_run_time: Array<any>,
  first_air_date: string,
  genres: Array<number>,
  homepage: string,
  id: number,
  in_production: string,
  languages: Array<string>,
  last_air_date: string,
  last_episode_to_air: Map<string, any>,
  name: string,
  next_episode_to_air: Map<string, any>,
  networks: Array<Map<string, any>>,
  number_of_episodes: number,
  number_of_seasons: number,
  origin_country: Array<string>,
  original_language: string,
  original_name: string,
  overview: string,
  popularity: number,
  poster_path: string,
  production_companires: Array<Map<string, any>>,
  production_countries: Array<Map<string, any>>,
  seasons: Array<Map<string, any>>,
  spoken_languages: Array<Map<string, any>>,
  status: string,
  tagline: string,
  type: string,
  vote_average: number,
  vote_count: number,
  in_plex: boolean,
}

interface MovieDetails {
  adult: boolean,
  backdrop_path: string,
  belongs_to_collection: boolean,
  budget: number,
  genres: Array<Map<string, any>>,
  homepage: string,
  id: number,
  imdb_id: string,
  original_language: string,
  original_title: string,
  overview: string,
  popularity: number,
  poster_path: string,
  production_companires: Array<Map<string, any>>,
  production_countries: Array<Map<string, any>>,
  release_date: string,
  tagline: string,
  title: string,
  video: boolean,
  vote_average: number,
  vote_count: number,
  in_plex: boolean,
}

interface QueryRequest {
  Query: string,
}

export {
  TvShow,
  TvShowResult,
  Movie,
  MovieResult,
  GenreRequest,
  Genre,
  QueryRequest,
  ShowDetails,
  MovieDetails,
}
