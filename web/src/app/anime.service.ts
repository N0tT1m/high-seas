import { Injectable, inject } from '@angular/core';
import { Movie, MovieResult, TvShow, TvShowResult, GenreRequest, QueryRequest, ShowDetails, MovieDetails } from './http-service/http-service.component';
import { HttpClient } from '@angular/common/http';
import { environment } from './environments/environment.prod';

@Injectable({
  providedIn: 'root'
})
export class AnimeService {
  http: HttpClient = inject(HttpClient);

  public headers = {
    'Authorization': "Bearer " + environment.envVar.authorization,
    'accept': 'application/json',
  }
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  // Base URLs for anime-specific endpoints
  private animeMovieUrl = 'https://api.themoviedb.org/3/discover/movie';
  private animeTvUrl = 'https://api.themoviedb.org/3/discover/tv';

  // Anime-specific genre IDs from TMDB
  private animeGenreId = 16; // Animation genre ID
  private japaneseOriginCountry = 'JP';

  constructor() {}

  // Anime Movies
  getAnimeMovies(page: number = 1) {
    const url = `${this.animeMovieUrl}?with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&sort_by=popularity.desc`;
    return this.http.get<Movie>(url, { headers: this.headers });
  }

  getTopRatedAnimeMovies(page: number = 1) {
    const url = `${this.animeMovieUrl}?with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&sort_by=vote_average.desc&vote_count.gte=100`;
    return this.http.get<Movie>(url, { headers: this.headers });
  }

  getUpcomingAnimeMovies(page: number = 1) {
    const currentDate = new Date().toISOString().split('T')[0];
    const url = `${this.animeMovieUrl}?with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&release_date.gte=${currentDate}&sort_by=popularity.desc`;
    return this.http.get<Movie>(url, { headers: this.headers });
  }

  searchAnimeMovies(query: string, page: number = 1) {
    const url = `https://api.themoviedb.org/3/search/movie?query=${query}&with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&include_adult=false`;
    return this.http.get<Movie>(url, { headers: this.headers });
  }

  // Anime TV Shows
  getAnimeSeries(page: number = 1) {
    const url = `${this.animeTvUrl}?with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&sort_by=popularity.desc`;
    return this.http.get<TvShow>(url, { headers: this.headers });
  }

  getTopRatedAnimeSeries(page: number = 1) {
    const url = `${this.animeTvUrl}?with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&sort_by=vote_average.desc&vote_count.gte=100`;
    return this.http.get<TvShow>(url, { headers: this.headers });
  }

  getCurrentlyAiringAnime(page: number = 1) {
    const currentDate = new Date().toISOString().split('T')[0];
    const url = `${this.animeTvUrl}?with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&air_date.gte=${currentDate}&sort_by=popularity.desc`;
    return this.http.get<TvShow>(url, { headers: this.headers });
  }

  searchAnimeSeries(query: string, page: number = 1) {
    const url = `https://api.themoviedb.org/3/search/tv?query=${query}&with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&include_adult=false`;
    return this.http.get<TvShow>(url, { headers: this.headers });
  }

  // Details endpoints
  getAnimeMovieDetails(id: number) {
    const url = `https://api.themoviedb.org/3/movie/${id}`;
    return this.http.get<MovieDetails>(url, { headers: this.headers });
  }

  getAnimeSeriesDetails(id: number) {
    const url = `https://api.themoviedb.org/3/tv/${id}`;
    return this.http.get<ShowDetails>(url, { headers: this.headers });
  }

  // Download request methods (Updated to match improved backend)
  makeAnimeMovieDownloadRequest(query: string, quality: string, tmdb: number, description: string) {
    const url = `${environment.envVar.transport}${environment.envVar.ip}:${environment.envVar.port}/anime/movie/query`;
    const headers = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE'
    };

    console.log(`[Anime Movie Download] Starting download request for: ${query} (Quality: ${quality}, TMDb: ${tmdb})`);

    return this.http.post<QueryRequest>(url, {
      query,
      quality,
      TMDb: tmdb,
      description
    }, { headers });
  }

  makeAnimeSeriesDownloadRequest(query: string, seasons: Array<number>, quality: string, tmdb: number, description: string) {
    const url = `${environment.envVar.transport}${environment.envVar.ip}:${environment.envVar.port}/anime/show/query`;
    const headers = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE'
    };

    console.log(`[Anime Series Download] Starting download request for: ${query} (Quality: ${quality}, Seasons: [${seasons.join(', ')}], TMDb: ${tmdb})`);

    return this.http.post<QueryRequest>(url, {
      query,
      seasons,
      quality,
      TMDb: tmdb,
      description
    }, { headers });
  }

  // Helper method to filter anime by date range
  getAnimeByDateRange(isMovie: boolean, startDate: string, endDate: string, page: number = 1) {
    const baseUrl = isMovie ? this.animeMovieUrl : this.animeTvUrl;
    const dateField = isMovie ? 'release_date' : 'first_air_date';

    const url = `${baseUrl}?with_genres=${this.animeGenreId}&with_original_language=ja&page=${page}&${dateField}.gte=${startDate}&${dateField}.lte=${endDate}&sort_by=popularity.desc`;

    return this.http.get<Movie | TvShow>(url, { headers: this.headers });
  }
}
