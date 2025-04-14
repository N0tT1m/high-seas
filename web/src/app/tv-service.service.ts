import { Injectable, inject } from '@angular/core';
import {
  Movie,
  MovieResult,
  TvShow,
  TvShowResult,
  GenreRequest,
  QueryRequest,
  ShowDetails
} from './http-service/http-service.component';
import { Observable } from 'rxjs';
import { HttpClient, HttpParams } from '@angular/common/http';
import { environment } from './environments/environment.prod';

// Interface for TV show filter options
export interface TvShowFilterOptions {
  genres?: number[];       // Array of genre IDs
  year?: number;           // Specific first air year
  yearRange?: {            // Range of air years
    start?: number;
    end?: number;
  };
  minRating?: number;      // Minimum user rating (0-10)
  maxRating?: number;      // Maximum user rating (0-10)
  language?: string;       // Original language (ISO 639-1 code)
  includeAdult?: boolean;  // Include adult content
  sortBy?: string;         // Sort criteria (popularity.desc, vote_average.desc, etc.)
  withNetworks?: number[]; // Filter by networks (HBO, ABC, Netflix, etc.)
  withCompanies?: number[]; // Filter by production companies
  status?: string;         // Status (returning series, ended, cancelled)
  keywords?: string[];     // Keywords or tags
  airDateRange?: {         // Specific air date range
    start?: string;        // YYYY-MM-DD
    end?: string;          // YYYY-MM-DD
  };
  watchProviders?: number[]; // Filter by available watch providers
  page?: number;           // Page number for pagination
  region?: string;         // ISO 3166-1 country code
  runtime?: {              // Runtime in minutes per episode
    min?: number;
    max?: number;
  };
  withType?: string;       // TV show type (documentary, scripted, reality, etc.)
  includeNullFirstAirDates?: boolean; // Include shows with no first air date
  screened?: {             // TV content ratings
    US?: string[];         // e.g., ["TV-14", "TV-MA"]
    country?: string;      // ISO 3166-1 country code
  };
  seasonCount?: {          // Number of seasons
    min?: number;
    max?: number;
  };
}

@Injectable({
  providedIn: 'root'
})
export class TvShowService {
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
   * Get TV show genres list
   */
  getGenres(): Observable<GenreRequest> {
    const url = `${this.apiBaseUrl}/tmdb/show/genres`;
    const tmdbUrl = 'https://api.themoviedb.org/3/genre/tv/list';
    return this.http.post<GenreRequest>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  /**
   * Get all TV shows using search query
   * @param page Page number
   * @param query Search query
   */
  getAllShows(page: number, query: string): Observable<TvShow> {
    if (!query) {
      return this.getPopular({ page: page });
    } else {
      return this.searchShows(query, { page: page });
    }
  }

  /**
   * Get initial page of search results for pagination
   * @param query Search query
   */
  getInitialPage(query: string): Observable<TvShow> {
    return this.searchShows(query, { page: 1 });
  }

  /**
   * Get top rated TV shows with optional filters
   * @param filters Optional filter criteria
   */
  getTopRated(filters: TvShowFilterOptions = {}): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/top-rated-tv-shows`;

    let tmdbUrl = `https://api.themoviedb.org/3/tv/top_rated?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get on the air TV shows with optional filters
   * @param filters Optional filter criteria
   */
  getOnTheAir(filters: TvShowFilterOptions = {}): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/on-the-air-tv-shows`;

    let tmdbUrl = `https://api.themoviedb.org/3/tv/on_the_air?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get popular TV shows with optional filters
   * @param filters Optional filter criteria
   */
  getPopular(filters: TvShowFilterOptions = {}): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/popular-tv-shows`;

    let tmdbUrl = `https://api.themoviedb.org/3/tv/popular?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get TV shows airing today with optional filters
   * @param filters Optional filter criteria
   */
  getAiringToday(filters: TvShowFilterOptions = {}): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/airing-today-tv-shows`;

    let tmdbUrl = `https://api.themoviedb.org/3/tv/airing_today?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Search TV shows by query string with advanced filters
   * @param query Search query
   * @param filters Optional filter criteria
   */
  searchShows(query: string, filters: TvShowFilterOptions = {}): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows`;

    let tmdbUrl = `https://api.themoviedb.org/3/search/tv?query=${encodeURIComponent(query)}&include_adult=${filters.includeAdult || false}&language=en-US&page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Discover TV shows with comprehensive filtering options
   * @param filters Filter criteria
   */
  discoverShows(filters: TvShowFilterOptions = {}): Observable<TvShow> {
    const page = filters.page || 1;
    // Using all-shows endpoint for discover functionality
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows`;

    let tmdbUrl = `https://api.themoviedb.org/3/discover/tv?page=${page}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get TV shows by genre with enhanced filtering
   * @param filters Filter criteria (must include genres)
   */
  getTvShowsByFilters(filters: TvShowFilterOptions): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows`;

    // Generate the base URL for discover
    let tmdbUrl = `https://api.themoviedb.org/3/discover/tv?page=${page}&include_null_first_air_dates=${filters.includeNullFirstAirDates || false}&sort_by=${filters.sortBy || 'popularity.desc'}`;
    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get detailed TV show information
   * @param id TMDb show ID
   * @param appendToResponse Additional data to append to the response (videos, credits, similar, etc.)
   */
  getTvShowDetails(id: number, appendToResponse: string[] = []): Observable<ShowDetails> {
    const url = `${this.apiBaseUrl}/tmdb/show/tv-show-details`;

    let tmdbUrl = `https://api.themoviedb.org/3/tv/${id}?language=en-US`;

    // Add optional append_to_response parameter
    if (appendToResponse.length > 0) {
      tmdbUrl += `&append_to_response=${appendToResponse.join(',')}`;
    }

    return this.http.post<ShowDetails>(
      url,
      { url: tmdbUrl, request_id: id },
      { headers: this.headers }
    );
  }

  /**
   * Get TV show recommendations based on a show ID
   * @param id TMDb show ID
   * @param page Page number
   */
  getTvShowRecommendations(id: number, page: number = 1): Observable<TvShow> {
    const url = `${this.apiBaseUrl}/tmdb/show/recommendations`;
    const tmdbUrl = `https://api.themoviedb.org/3/tv/${id}/recommendations?page=${page}`;

    return this.http.post<TvShow>(url, { url: tmdbUrl, show_id: id }, { headers: this.headers });
  }

  /**
   * Get similar TV shows based on a show ID
   * @param id TMDb show ID
   * @param page Page number
   */
  getSimilarTvShows(id: number, page: number = 1): Observable<TvShow> {
    const url = `${this.apiBaseUrl}/tmdb/show/similar`;
    const tmdbUrl = `https://api.themoviedb.org/3/tv/${id}/similar?page=${page}`;

    return this.http.post<TvShow>(url, { url: tmdbUrl, show_id: id }, { headers: this.headers });
  }

  /**
   * Get all TV shows from a selected date
   * @param filters Filter options including airDateRange
   */
  getAllShowsFromDate(filters: TvShowFilterOptions): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows-from-date`;

    // Generate the base URL for discover
    let tmdbUrl = `https://api.themoviedb.org/3/discover/tv?page=${page}&include_adult=${filters.includeAdult || false}&include_null_first_air_dates=${filters.includeNullFirstAirDates || false}&sort_by=${filters.sortBy || 'popularity.desc'}`;

    if (filters.airDateRange?.start) {
      tmdbUrl += `&first_air_date.gte=${filters.airDateRange.start}`;
    }

    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }

  /**
   * Get TV show seasons and episodes
   * @param id TMDb show ID
   * @param seasonNumber Season number (optional - if not provided, returns all seasons info)
   */
  getTvShowSeasons(id: number, seasonNumber?: number): Observable<any> {
    const url = `${this.apiBaseUrl}/tmdb/show/seasons`;
    let tmdbUrl = `https://api.themoviedb.org/3/tv/${id}`;

    if (seasonNumber !== undefined) {
      tmdbUrl = `https://api.themoviedb.org/3/tv/${id}/season/${seasonNumber}`;
    }

    return this.http.post<any>(url, {
      url: tmdbUrl,
      show_id: id,
      season_number: seasonNumber
    }, { headers: this.headers });
  }

  /**
   * Make a request to download a TV show
   * @param query Show title
   * @param seasons Array of episode counts per season
   * @param quality Video quality (e.g., "1080p")
   * @param tmdb TMDb ID
   * @param description Show description
   * @param year First air year (extracted from details if available)
   */
  makeTvShowDownloadRequest(
    query: string,
    seasons: number[],
    quality: string,
    tmdb: number,
    description: string,
    year?: number
  ): Observable<QueryRequest> {
    const url = `${this.apiBaseUrl}/show/query`;

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
        seasons,
        quality,
        TMDb: tmdb,
        description,
        year: year // Add year to the request
      },
      { headers: this.requestHeaders }
    );
  }

  /**
   * Get the Plex data
   */
  makePlexRequest(): Observable<QueryRequest> {
    const tvShowUrl = 'http://127.0.0.1:5000/shows/';

    return this.http.get<QueryRequest>(tvShowUrl, {
      headers: this.requestHeaders,
      // Angular HttpClient doesn't support rejectUnauthorized
      // We'll use withCredentials instead for CORS requests
      withCredentials: true,
    });
  }

  /**
   * Make a request to download an anime show
   * @param query Show title
   * @param seasons Array of episode counts per season
   * @param quality Video quality (e.g., "1080p")
   * @param tmdb TMDb ID
   * @param description Show description
   * @param year First air year (extracted from details if available)
   */
  makeAnimeShowDownloadRequest(
    query: string,
    seasons: number[],
    quality: string,
    tmdb: number,
    description: string,
    year?: number
  ): Observable<QueryRequest> {
    const url = `${this.apiBaseUrl}/anime/show/query`;

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
        seasons,
        quality,
        TMDb: tmdb,
        description,
        year: year // Add year to the request
      },
      { headers: this.requestHeaders }
    );
  }

  // Fix for TvShowService appendFilterParams method
  private appendFilterParams(baseUrl: string, filters: TvShowFilterOptions): string {
    // Start with the base URL
    let url = baseUrl;

    // Add genre filter
    if (filters.genres && filters.genres.length > 0) {
      url += `&with_genres=${filters.genres.join(',')}`;
    }

    // Add year filter - This is the critical part
    if (filters.year) {
      // If year is specified, we need to set both start and end dates to that year
      url += `&first_air_date.gte=${filters.year}-01-01&first_air_date.lte=${filters.year}-12-31`;
    }
    // Otherwise, use the year range if provided
    else if (filters.yearRange) {
      if (filters.yearRange.start) {
        url += `&first_air_date.gte=${filters.yearRange.start}-01-01`;
      }
      if (filters.yearRange.end) {
        url += `&first_air_date.lte=${filters.yearRange.end}-12-31`;
      }
    }

    // Add specific air date range filter
    if (filters.airDateRange) {
      if (filters.airDateRange.start && !url.includes('first_air_date.gte')) {
        url += `&first_air_date.gte=${filters.airDateRange.start}`;
      }
      if (filters.airDateRange.end) {
        url += `&first_air_date.lte=${filters.airDateRange.end}`;
      }
    }

    // Rest of the method remains the same...

    return url;
  }

  /**
   * Get the initial page of airing today TV shows
   */
  getInitialAiringTodayPage(): Observable<TvShow> {
    return this.getAiringTodayShows(1);
  }

  /**
   * Get TV shows airing today by page
   * @param page Page number
   */
  getAiringTodayShows(page: number): Observable<TvShow> {
    return this.getAiringToday({ page });
  }

  /**
   * Get the initial page of on the air TV shows
   */
  getInitialOnTheAirPage(): Observable<TvShow> {
    return this.getOnTheAirShows(1);
  }

  /**
   * Get on the air TV shows by page
   * @param page Page number
   */
  getOnTheAirShows(page: number): Observable<TvShow> {
    return this.getOnTheAir({ page });
  }

  /**
   * Get the initial page of top rated TV shows
   */
  getInitialTopRatedPage(): Observable<TvShow> {
    return this.getTopRatedShows(1);
  }

  /**
   * Get top rated TV shows by page
   * @param page Page number
   */
  getTopRatedShows(page: number): Observable<TvShow> {
    return this.getTopRated({ page });
  }

  /**
   * Get the initial page of popular TV shows
   */
  getInitialPopularPage(): Observable<TvShow> {
    return this.getPopularShows(1);
  }

  /**
   * Get popular TV shows by page
   * @param page Page number
   */
  getPopularShows(page: number): Observable<TvShow> {
    return this.getPopular({ page });
  }

  /**
   * Get all TV shows by genre, air date, and page
   * @param genre Genre ID
   * @param airDate Air date or year
   * @param page Page number
   */
  getAllTvShows(genre: number, airDate: string, page: number): Observable<TvShow> {
    if (genre === 0 && !airDate) {
      return this.getPopular({ page });
    } else {
      const filters: TvShowFilterOptions = { page };

      if (genre !== 0) {
        filters.genres = [genre];
      }

      if (airDate) {
        // Check if airDate is just a year (4 digits)
        if (/^\d{4}$/.test(airDate.trim())) {
          const year = parseInt(airDate.trim(), 10);
          // Use year range filter instead of specific date
          filters.yearRange = {
            start: year,
            end: year
          };
          console.log(`Searching for shows from ${year}`);
        } else {
          // It's a full date, use it as is
          filters.airDateRange = {
            start: airDate
          };
        }
      }

      // Using all-shows endpoint instead of discover
      return this.getTvShowsByFilters(filters);
    }
  }

  /**
   * Get TV show details for the discover section
   * @param genre Genre ID
   * @param airDate Air date
   * @param page Page number
   */
  getAllShowsForDetails(genre: number, airDate: string, page: number): Observable<TvShow> {
    return this.getAllTvShows(genre, airDate, page);
  }

  /**
   * Get TV show details
   * @param id TV show ID
   */
  getShowDetails(id: number): Observable<ShowDetails> {
    return this.getTvShowDetails(id);
  }

  /**
   * Get all TV shows from a selected date with optional genre filter
   * @param genre Genre ID (optional)
   * @param airDate Air date in YYYY-MM-DD format
   * @param page Page number
   * @returns Observable<TvShow>
   */
  getAllShowsFromSelectedDate(genre: number, airDate: string, page: number): Observable<TvShow> {
    const filters: TvShowFilterOptions = {
      page: page,
      includeAdult: false,
      sortBy: 'popularity.desc'
    };

    // Add genre filter if specified and not 0
    if (genre && genre !== 0) {
      filters.genres = [genre];
    }

    // Add air date if specified
    if (airDate) {
      filters.airDateRange = {
        start: airDate
      };
    }

    // Use the existing method that supports the TvShowFilterOptions interface
    return this._getAllShowsFromSelectedDate(filters);
  }

  /**
   * Private implementation of getting shows from date with filter options
   * @param filters Filter options including airDateRange
   */
  private _getAllShowsFromSelectedDate(filters: TvShowFilterOptions): Observable<TvShow> {
    const page = filters.page || 1;
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows-from-date`;

    // Generate the base URL for discover
    let tmdbUrl = `https://api.themoviedb.org/3/discover/tv?page=${page}&include_adult=${filters.includeAdult || false}&include_null_first_air_dates=${filters.includeNullFirstAirDates || false}&sort_by=${filters.sortBy || 'popularity.desc'}`;

    if (filters.airDateRange?.start) {
      tmdbUrl += `&first_air_date.gte=${filters.airDateRange.start}`;
    }

    if (filters.airDateRange?.end) {
      tmdbUrl += `&first_air_date.lte=${filters.airDateRange.end}`;
    }

    tmdbUrl = this.appendFilterParams(tmdbUrl, filters);

    return this.http.post<TvShow>(url, { url: tmdbUrl, filters }, { headers: this.headers });
  }
}
