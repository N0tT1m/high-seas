import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from './environments/environment.prod';
import { Movie, MovieResult, GenreRequest, QueryRequest, MovieDetails } from './http-service/http-service.component';

@Injectable({
  providedIn: 'root'
})
export class MovieService {
  private http: HttpClient = inject(HttpClient);

  private headers = {
    'Authorization': `Bearer ${environment.envVar.authorization}`,
    'accept': 'application/json'
  };

  private baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  private apiBaseUrl = `${environment.envVar.transport}${environment.envVar.ip}:${environment.envVar.port}`;

  getGenres() {
    const url = `${this.apiBaseUrl}/tmdb/movie/genres`;
    const tmdbUrl = 'https://api.themoviedb.org/3/genre/movie/list?language=en';
    return this.http.post<GenreRequest>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getTopRated(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/movie/top-rated`;
    const tmdbUrl = `https://api.themoviedb.org/3/movie/top_rated?page=${page}`;
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialTopRatedPage() {
    const url = `${this.apiBaseUrl}/tmdb/movie/top-rated`;
    const tmdbUrl = 'https://api.themoviedb.org/3/movie/top_rated?page=1';
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getUpcoming(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/movie/upcoming`;
    const tmdbUrl = `https://api.themoviedb.org/3/movie/upcoming?page=${page}`;
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialUpcomingPage() {
    const url = `${this.apiBaseUrl}/tmdb/movie/upcoming`;
    const tmdbUrl = 'https://api.themoviedb.org/3/movie/upcoming?page=1';
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getPopular(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/movie/popular`;
    const tmdbUrl = `https://api.themoviedb.org/3/movie/popular?page=${page}`;
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialPopularPage() {
    const url = `${this.apiBaseUrl}/tmdb/movie/popular`;
    const tmdbUrl = 'https://api.themoviedb.org/3/movie/popular?page=1';
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getNowPlaying(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/movie/now-playing`;
    const tmdbUrl = `https://api.themoviedb.org/3/movie/now_playing?page=${page}`;
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialNowPlayingPage() {
    const url = `${this.apiBaseUrl}/tmdb/movie/now-playing`;
    const tmdbUrl = 'https://api.themoviedb.org/3/movie/now_playing?page=1';
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  searchMovies(page: number, query: string) {
    const url = `${this.apiBaseUrl}/tmdb/movie/search`;
    const tmdbUrl = `https://api.themoviedb.org/3/search/movie?query=${query}&include_adult=false&page=${page}`;
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialSearchPage(query: string) {
    const url = `${this.apiBaseUrl}/tmdb/movie/search`;
    const tmdbUrl = `https://api.themoviedb.org/3/search/movie?query=${query}&include_adult=false&page=1`;
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getMoviesByGenre(genre: number, releaseDateStart: string, releaseDateEnd: string, page: number) {
    const url = `${this.apiBaseUrl}/tmdb/movie/by-genre`;
    const tmdbUrl = `https://api.themoviedb.org/3/discover/movie?with_genres=${genre}&page=${page}&include_adult=false&include_null_first_air_dates=false&release_date.gte=${releaseDateStart}&release_date.lte=${releaseDateEnd}&sort_by=popularity.desc&with_release_type=1`;
    return this.http.post<Movie>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getMovieDetails(id: number) {
    const url = `${this.apiBaseUrl}/tmdb/movie/details`;
    const tmdbUrl = `https://api.themoviedb.org/3/movie/${id}`;
    return this.http.post<MovieDetails>(url, {
      url: tmdbUrl,
      request_id: id
    }, { headers: this.headers });
  }

  makeMovieDownloadRequest(query: string, quality: string, tmdb: number, description: string) {
    const url = `${this.apiBaseUrl}/movie/query`;
    const headers = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE'
    };

    return this.http.post<QueryRequest>(url, {
      query,
      quality,
      TMDb: tmdb,
      description
    }, {
      headers,
    });
  }

  makeAnimeMovieDownloadRequest(query: string, name: string, year: string, quality: string, tmdb: number, description: string) {
    const url = `${this.apiBaseUrl}/anime/movie/query`;
    const headers = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE'
    };

    return this.http.post<QueryRequest>(url, {
      query,
      name,
      year,
      quality,
      Imdb: tmdb,
      description
    }, { headers });
  }
}
