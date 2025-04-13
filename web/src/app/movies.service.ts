import { Injectable, inject } from '@angular/core';
import {
  Movie,
  MovieResult,
  MovieDetails,
  GenreRequest,
  QueryRequest
} from './http-service/http-service.component';
import { Observable } from 'rxjs';
import { HttpClient, HttpParams } from '@angular/common/http';
import { environment } from './environments/environment.prod';

// Interface for movie filter options
export interface MovieFilterOptions {
  genres?: number[];       // Array of genre IDs
  year?: number;           // Specific release year
  yearRange?: {            // Range of release years
    start?: number;
    end?: number;
  };
  minRating?: number;      // Minimum user rating (0-10)
  maxRating?: number;      // Maximum user rating (0-10)
  language?: string;       // Original language (ISO 639-1 code)
  includeAdult?: boolean;  // Include adult content
  sortBy?: string;         // Sort criteria (popularity.desc, vote_average.desc, etc.)
  withCompanies?: number[]; // Filter by production companies
  keywords?: string[];     // Keywords or tags
  releaseDateRange?: {     // Specific release date range
    start?: string;        // YYYY-MM-DD
    end?: string;          // YYYY-MM-DD
  };
  watchProviders?: number[]; // Filter by available watch providers
  page?: number;           // Page number for pagination
  region?: string;         // ISO 3166-1 country code
  runtime?: {              // Runtime in minutes
    min?: number;
    max?: number;
  };
  withType?: string;       // Movie type
  voteCount?: {            // Filter by vote count
    min?: number;
    max?: number;
  };
}

@Injectable({
  providedIn: 'root'
})
export class MovieService {
  http: HttpClient = inject(HttpClient);

  // API configuration
  private headers = {
    'Authorization': `Bearer ${environment.envVar.authorization}`,
    'accept': 'application/json',
  };

  private baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  private apiBaseUrl = `${environment.envVar.transport}${environment.envVar.ip}:${environment.envVar.port}`;

  // Enhanced request headers with CORS support
  private requestHeaders = {
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Credentials': 'true',
    'Access-Control-Allow-Headers': 'Content-Type',
    'Access-Control-Allow-Methods': 'POST,DELETE'
  };

  /**
   * Get movie genres list
   */
  getGenres(): Observable<GenreRequest> {
    const url = `${this.apiBaseUrl}/tmdb/movie/genres`;
    const tmdbUrl = 'https://api.themoviedb.org/3/genre/movie/list';
    return this.http.post<GenreRequest>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  /**
   * Get all movies using search query
   * @param page Page number
   * @param query Search query
   */
  getAllMovies(page: number, query: string): Observable<Movie> {
    if (!query) {
      return this.getPopular({ page: page });
    } else {
      return this.searchMovies(query, { page: page });
    }
  }

  /**
   * Get initial page of search results for pagination
   * @param query Search query
   */
  getInitialPage(query: string): Observable<Movie> {
    return this.searchMovies(query, { page: 1 });
  }

  /**
   * Get top rated movies with optional filters
   * @param filters Optional filter criteria
   */
  getTopRated(filters: MovieFilterOptions = {}): Observable<Movie> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/movie/top-rated-movies`;

    let tmdbUrl = `https://api.themoviedb.org/3/movie/top_rated?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get now playing movies with optional filters
   * @param filters Optional filter criteria
   */
  getNowPlaying(filters: MovieFilterOptions = {}): Observable<Movie> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/movie/now-playing-movies`;

    let tmdbUrl = `https://api.themoviedb.org/3/movie/now_playing?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get popular movies with optional filters
   * @param filters Optional filter criteria
   */
  getPopular(filters: MovieFilterOptions = {}): Observable<Movie> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/movie/popular-movies`;

    let tmdbUrl = `https://api.themoviedb.org/3/movie/popular?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get upcoming movies with optional filters
   * @param filters Optional filter criteria
   */
  getUpcoming(filters: MovieFilterOptions = {}): Observable<Movie> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/movie/upcoming-movies`;

    let tmdbUrl = `https://api.themoviedb.org/3/movie/upcoming?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Search movies by query string with advanced filters
   * @param query Search query
   * @param filters Optional filter criteria
   */
  searchMovies(query: string, filters: MovieFilterOptions = {}): Observable<Movie> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/movie/all-movies`;

    let tmdbUrl = `https://api.themoviedb.org/3/search/movie?query=${encodeURIComponent(query)}&include_adult=${filters.includeAdult || false}&language=en-US&page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Discover movies with comprehensive filtering options
   * @param filters Filter criteria
   */
  discoverMovies(filters: MovieFilterOptions = {}): Observable<Movie> {
    const page = filters.page || 1;
    // Using all-movies endpoint for discover functionality
    const url = `${this.apiBaseUrl}/tmdb/movie/all-movies`;

    let tmdbUrl = `https://api.themoviedb.org/3/discover/movie?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get movies by genre with enhanced filtering
   * @param filters Filter criteria (must include genres)
   */
  getMoviesByFilters(filters: MovieFilterOptions): Observable<Movie> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/movie/all-movies`;

    // Generate the base URL for discover
    let tmdbUrl = `https://api.themoviedb.org/3/discover/movie?page=${page}&sort_by=${filters.sortBy || 'popularity.desc'}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get detailed movie information
   * @param id TMDb movie ID
   * @param appendToResponse Additional data to append to the response (videos, credits, similar, etc.)
   */
  getMovieDetails(id: number, appendToResponse: string[] = []): Observable<MovieDetails> {
    const url = `${this.apiBaseUrl}/tmdb/movie/movie-details`;

    let tmdbUrl = `https://api.themoviedb.org/3/movie/${id}?language=en-US`;

    // Add optional append_to_response parameter
    if (appendToResponse.length > 0) {
      tmdbUrl += `&append_to_response=${appendToResponse.join(',')}`;
    }

    return this.http.post<MovieDetails>(
      url,
      { url: tmdbUrl, request_id: id },
      { headers: this.headers }
    );
  }

  /**
   * Get movie recommendations based on a movie ID
   * @param id TMDb movie ID
   * @param page Page number
   */
  getMovieRecommendations(id: number, page: number = 1): Observable<Movie> {
    const url = `${this.apiBaseUrl}/tmdb/movie/recommendations`;
    const tmdbUrl = `https://api.themoviedb.org/3/movie/${id}/recommendations?page=${page}`;

    return this.http.post<Movie>(url, { url: tmdbUrl, movie_id: id }, { headers: this.headers });
  }

  /**
   * Get similar movies based on a movie ID
   * @param id TMDb movie ID
   * @param page Page number
   */
  getSimilarMovies(id: number, page: number = 1): Observable<Movie> {
    const url = `${this.apiBaseUrl}/tmdb/movie/similar`;
    const tmdbUrl = `https://api.themoviedb.org/3/movie/${id}/similar?page=${page}`;

    return this.http.post<Movie>(url, { url: tmdbUrl, movie_id: id }, { headers: this.headers });
  }

  /**
   * Get all movies since a specific release date with filtering options
   * @param genre Genre ID (optional)
   * @param releaseYear Release year start (optional)
   * @param endYear Release year end (optional)
   * @param page Page number (optional)
   * @returns Observable<Movie>
   */
  getAllMoviesFromSelectedDate(genre: number, releaseYear: string, endYear: string, page: number = 1): Observable<Movie> {
    const filters: MovieFilterOptions = {
      page: page,
      includeAdult: false,
      sortBy: 'popularity.desc'
    };

    // Add genre filter if specified and not 0
    if (genre && genre !== 0) {
      filters.genres = [genre];
    }

    // Add release date range if specified
    if (releaseYear || endYear) {
      filters.yearRange = {};

      if (releaseYear) {
        filters.yearRange.start = parseInt(releaseYear);
      }

      if (endYear) {
        filters.yearRange.end = parseInt(endYear);
      }

      // Also set the releaseDateRange for direct date filtering if needed
      filters.releaseDateRange = {};

      if (releaseYear) {
        filters.releaseDateRange.start = `${releaseYear}-01-01`;
      }

      if (endYear) {
        filters.releaseDateRange.end = `${endYear}-12-31`;
      }
    }

    // Use the existing method that supports the MovieFilterOptions interface
    return this._getAllMoviesFromSelectedDate(filters);
  }

  /**
   * Private implementation of getting movies from date with filter options
   * @param filters Filter options including releaseDateRange
   */
  private _getAllMoviesFromSelectedDate(filters: MovieFilterOptions): Observable<Movie> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/movie/all-movies-from-date`;

    // Generate the base URL for discover
    let tmdbUrl = `https://api.themoviedb.org/3/discover/movie?page=${page}&include_adult=${filters.includeAdult || false}&sort_by=${filters.sortBy || 'popularity.desc'}`;

    if (filters.releaseDateRange?.start) {
      tmdbUrl += `&primary_release_date.gte=${filters.releaseDateRange.start}`;
    }

    if (filters.releaseDateRange?.end) {
      tmdbUrl += `&primary_release_date.lte=${filters.releaseDateRange.end}`;
    }

    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<Movie>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Make a request to download a movie
   * @param query Movie title
   * @param quality Video quality (e.g., "1080p")
   * @param tmdb TMDb ID
   * @param description Movie description
   * @param year Release year (extracted from details if available)
   */
  makeMovieDownloadRequest(
    query: string,
    quality: string,
    tmdb: number,
    description: string,
    year?: number
  ): Observable<QueryRequest> {
    const url = `${this.apiBaseUrl}/movie/query`;

    // Extract year from query or description if not explicitly provided
    if (!year) {
      const yearMatch = query.match(/\((\d{4})\)/) || description.match(/\((\d{4})\)/);
      if (yearMatch) {
        year = parseInt(yearMatch[1], 10);
      }
    }

    return this.http.post<QueryRequest>(
      url,
      {
        query,
        quality,
        TMDb: tmdb,
        description,
        year: year // Add year to the request
      },
      { headers: this.requestHeaders }
    );
  }

  /**
   * Get the Plex data for movies
   */
  makePlexRequest(): Observable<QueryRequest> {
    const movieUrl = 'http://127.0.0.1:5000/movies/';

    return this.http.get<QueryRequest>(movieUrl, {
      headers: this.requestHeaders,
      withCredentials: true,
    });
  }

  /**
   * Make a request to download an anime movie
   * @param query Movie title
   * @param name Original anime name
   * @param year Release date
   * @param quality Video quality (e.g., "1080p")
   * @param tmdb TMDb ID
   * @param description Movie description
   */
  makeAnimeMovieDownloadRequest(
    query: string,
    name: string,
    year: string,
    quality: string,
    tmdb: number,
    description: string
  ): Observable<QueryRequest> {
    const url = `${this.apiBaseUrl}/anime/movie/query`;

    // Extract release year if it's a full date
    let releaseYear: number | undefined;
    if (year) {
      const yearMatch = year.match(/^(\d{4})/);
      if (yearMatch) {
        releaseYear = parseInt(yearMatch[1], 10);
      }
    }

    return this.http.post<QueryRequest>(
      url,
      {
        query,
        name,
        quality,
        TMDb: tmdb,
        description,
        year: releaseYear
      },
      { headers: this.requestHeaders }
    );
  }

  /**
   * Helper method to append filter parameters to a TMDb URL
   * @param baseUrl Base TMDb URL
   * @param filters Filter criteria
   * @returns Updated URL with filter parameters
   */
  private appendFilterParams(baseUrl: string, filters: MovieFilterOptions): string {
    // Start with the base URL
    let url = baseUrl;

    // Add genre filter
    if (filters.genres && filters.genres.length > 0) {
      url += `&with_genres=${filters.genres.join(',')}`;
    }

    // Add year filter
    if (filters.year) {
      url += `&primary_release_year=${filters.year}`;
    }

    // Add year range filter
    if (filters.yearRange) {
      if (filters.yearRange.start) {
        url += `&primary_release_date.gte=${filters.yearRange.start}-01-01`;
      }
      if (filters.yearRange.end) {
        url += `&primary_release_date.lte=${filters.yearRange.end}-12-31`;
      }
    }

    // Add specific release date range filter
    if (filters.releaseDateRange) {
      if (filters.releaseDateRange.start && !url.includes('primary_release_date.gte')) {
        url += `&primary_release_date.gte=${filters.releaseDateRange.start}`;
      }
      if (filters.releaseDateRange.end) {
        url += `&primary_release_date.lte=${filters.releaseDateRange.end}`;
      }
    }

    // Add rating filter
    if (filters.minRating) {
      url += `&vote_average.gte=${filters.minRating}`;
    }
    if (filters.maxRating) {
      url += `&vote_average.lte=${filters.maxRating}`;
    }

    // Add vote count filter
    if (filters.voteCount) {
      if (filters.voteCount.min) {
        url += `&vote_count.gte=${filters.voteCount.min}`;
      }
      if (filters.voteCount.max) {
        url += `&vote_count.lte=${filters.voteCount.max}`;
      }
    }

    // Add language filter
    if (filters.language) {
      url += `&with_original_language=${filters.language}`;
    }

    // Add sort parameter
    if (filters.sortBy) {
      url += `&sort_by=${filters.sortBy}`;
    }

    // Add production companies filter
    if (filters.withCompanies && filters.withCompanies.length > 0) {
      url += `&with_companies=${filters.withCompanies.join('|')}`;
    }

    // Add keywords filter
    if (filters.keywords && filters.keywords.length > 0) {
      url += `&with_keywords=${filters.keywords.join('|')}`;
    }

    // Add runtime filter
    if (filters.runtime) {
      if (filters.runtime.min) {
        url += `&with_runtime.gte=${filters.runtime.min}`;
      }
      if (filters.runtime.max) {
        url += `&with_runtime.lte=${filters.runtime.max}`;
      }
    }

    // Add watch providers filter
    if (filters.watchProviders && filters.watchProviders.length > 0) {
      url += `&with_watch_providers=${filters.watchProviders.join('|')}`;
    }

    // Add region filter
    if (filters.region) {
      url += `&watch_region=${filters.region}`;
    }

    return url;
  }

  /**
   * Get the initial page of now playing movies
   */
  getInitialNowPlayingPage(): Observable<Movie> {
    return this.getNowPlayingMovies(1);
  }

  /**
   * Get now playing movies by page
   * @param page Page number
   */
  getNowPlayingMovies(page: number): Observable<Movie> {
    return this.getNowPlaying({ page });
  }

  /**
   * Get the initial page of top rated movies
   */
  getInitialTopRatedPage(): Observable<Movie> {
    return this.getTopRatedMovies(1);
  }

  /**
   * Get top rated movies by page
   * @param page Page number
   */
  getTopRatedMovies(page: number): Observable<Movie> {
    return this.getTopRated({ page });
  }

  /**
   * Get the initial page of popular movies
   */
  getInitialPopularPage(): Observable<Movie> {
    return this.getPopularMovies(1);
  }

  /**
   * Get popular movies by page
   * @param page Page number
   */
  getPopularMovies(page: number): Observable<Movie> {
    return this.getPopular({ page });
  }

  /**
   * Get the initial page of upcoming movies
   */
  getInitialUpcomingPage(): Observable<Movie> {
    return this.getUpcomingMovies(1);
  }

  /**
   * Get upcoming movies by page
   * @param page Page number
   */
  getUpcomingMovies(page: number): Observable<Movie> {
    return this.getUpcoming({ page });
  }

  /**
   * Get all movies by genre, release year, end year, and page
   * @param genre Genre ID
   * @param releaseYear Release year start
   * @param endYear Release year end
   * @param page Page number
   */
  getAllMoviesByGenre(genre: number, releaseYear: string, endYear: string, page: number): Observable<Movie> {
    if (genre === 0 && !releaseYear && !endYear) {
      return this.getPopular({ page });
    } else {
      const filters: MovieFilterOptions = { page };

      if (genre !== 0) {
        filters.genres = [genre];
      }

      if (releaseYear || endYear) {
        filters.yearRange = {};

        if (releaseYear) {
          filters.yearRange.start = parseInt(releaseYear);
        }

        if (endYear) {
          filters.yearRange.end = parseInt(endYear);
        }
      }

      // Using all-movies endpoint instead of discover
      return this.getMoviesByFilters(filters);
    }
  }

  /**
   * Get movie details for the discover section
   * @param genre Genre ID
   * @param releaseYear Release year start
   * @param endYear Release year end
   * @param page Page number
   */
  getAllMoviesForDetails(genre: number, releaseYear: string, endYear: string, page: number): Observable<Movie> {
    return this.getAllMoviesByGenre(genre, releaseYear, endYear, page);
  }
}
