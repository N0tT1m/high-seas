import { Injectable, inject, NgModule, Output, EventEmitter, Component } from '@angular/core';
import { Movie, MovieResult, TvShow, TvShowResult, GenreRequest, QueryRequest, ShowDetails } from './http-service/http-service.component';
import { Observable, map, of } from 'rxjs';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { environment } from './environments/environment.prod';

@Injectable({
  providedIn: 'root'
})

export class TvShowService {
  http: HttpClient = inject(HttpClient)

  public genreUrl = 'https://api.themoviedb.org/3/genre/tv/list';
  public   headers = {
    'Authorization': "Bearer " + environment.envVar.authorization,
      'accept': 'application/json',
  }

  private baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  private apiBaseUrl = `${environment.envVar.transport}${environment.envVar.ip}:${environment.envVar.port}`;

  public results: TvShowResult[] = [];
  public tvShowsList: TvShow[] = [{page: 0, results: [{adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0}]
  public singleTvShow: TvShow | undefined
  public respData: TvShowResult[] = [];
  public  tvShow: any;
  public  tvShowToBeAdded: Movie

  TvShowService() {
  }

  getAllShows(page: number, query: string) {
    // TODO: Add multiple genres filter.
    // with_genres=Action%20AND%20Comedy
    // TODO: Add multiple people filter.
    // with_people=Shah%20Rukh%20Khan%20AND%20Shah%20Rukh%20Khan

    var tmdbUrl = 'https://api.themoviedb.org/3/search/tv?query=' + query + '&include_adult=false&language=en-US&&page=' + page.toString();
    var tvShowUrl = environment.envVar.transport + environment.envVar.ip + ':' + environment.envVar.port + '/tmdb/show/all-shows';

    setTimeout(function () {
    }, 4000000);

    return this.http.post<TvShow>(tvShowUrl, {"url": tmdbUrl}, { headers: this.headers });
  }

  getInitialPage(query: string) {
    var tmdbUrl = 'https://api.themoviedb.org/3/search/tv?query=' + query + '&include_adult=false&language=en-US&&page=1';
    var tvShowUrl = environment.envVar.transport + environment.envVar.ip + ':' + environment.envVar.port + '/tmdb/show/initial-all-shows';

    setTimeout(function () {
    }, 4000000);

    return this.http.post<TvShow>(tvShowUrl, {"url": tmdbUrl}, { headers: this.headers });
  }

  makePlexRequest() {
    var queryApiHeaders = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE',
    };

    var tvShowUrl = 'http://127.0.0.1:5000/shows/';

    const options = {
      headers: queryApiHeaders,
      rejectUnauthorized: false,
    };

    return this.http.get<QueryRequest>(tvShowUrl, options);
  }

  getShowDetails(id) {
    var tmdbUrl = 'https://api.themoviedb.org/3/tv/'+ id +'?language=en-US';
    var tvShowUrl = environment.envVar.transport + environment.envVar.ip + ':' + environment.envVar.port + '/tmdb/show/tv-show-details';

     return this.http.post<ShowDetails>(tvShowUrl, {"url": tmdbUrl, "request_id": id}, { headers: this.headers })
  }

  // Movie[] {
  getAllTvShows(genre: number, airDate: string, page: number) {
    var tmdbUrl = 'https://api.themoviedb.org/3/discover/tv?with_genres=' + genre.toString() + '&page=' + page.toString() +  '&first_air_date.gte=' + airDate + '&include_adult=false&include_null_first_air_dates=false&sort_by=popularity.desc&with_release_type=4|5|6';
    var tvShowUrl = environment.envVar.transport + environment.envVar.ip + ':' + environment.envVar.port + '/tmdb/show/all-shows';


    return this.http.post<TvShow>(tvShowUrl, {"url": tmdbUrl}, { headers: this.headers })
  }

  getTvShowById(id: number, genre: number, page: number, airDate: string): TvShow | undefined {
    this.getAllTvShows(genre, airDate, page).subscribe((resp) => {
      resp['results'].forEach((show) => {
        let page = resp['page'];
        let isAdult = show['adult'];
        let backdropPath = show['backdrop_path'];
        let genreIds = show['genre_ids'];
        let id = show['id'];
        let firstAirDate = show['first_air_date'];
        let video = show['video'];
        let name = show['name'];
        let originalLanguage = show['original_language'];
        let originalName = show['original_name'];
        let overview = show['overview'];
        let popularity = show['popularity'];
        let posterPath = this.baseUrl + show['poster_path'];
        let voteAverage = show['vote_average'];
        let voteCount = show['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];

        let result: TvShowResult[] = [{adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, name: name, first_air_date: firstAirDate, original_language: originalLanguage, original_name: originalName, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video}]

        this.tvShowsList.push({ page: page, results: result,  total_pages: totalPages, total_result: totalResult });
      })

      this.tvShowsList.splice(0, 1);

      for (var i = 0; i < this.tvShowsList.length; i++) {


        this.singleTvShow = this.tvShowsList.find(movieResult => movieResult.results[i].id === id);
      }

      return this.singleTvShow;
    })

    return this.singleTvShow;
    }

  getGenres() {
    const url = `${this.apiBaseUrl}/tmdb/show/genres`;
    const tmdbUrl = 'https://api.themoviedb.org/3/genre/tv/list';
    return this.http.post<GenreRequest>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getTopRated(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/top-rated-tv-shows`;
    const tmdbUrl = `https://api.themoviedb.org/3/tv/top_rated?page=${page}`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialTopRatedPage() {
    const url = `${this.apiBaseUrl}/tmdb/show/initial-top-rated-tv-shows`;
    const tmdbUrl = 'https://api.themoviedb.org/3/tv/top_rated?page=1';
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getOnTheAir(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/on-the-air-tv-shows`;
    const tmdbUrl = `https://api.themoviedb.org/3/tv/on_the_air?page=${page}`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialOnTheAirPage() {
    const url = `${this.apiBaseUrl}/tmdb/show/initial-on-the-air-tv-shows`;
    const tmdbUrl = 'https://api.themoviedb.org/3/tv/on_the_air?page=1';
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getPopular(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/popular-tv-shows`;
    const tmdbUrl = `https://api.themoviedb.org/3/tv/popular?page=${page}`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialPopularPage() {
    const url = `${this.apiBaseUrl}/tmdb/show/initial-popular-tv-shows`;
    const tmdbUrl = 'https://api.themoviedb.org/3/tv/popular?page=1';
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getAiringToday(page: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/airing-today-tv-shows`;
    const tmdbUrl = `https://api.themoviedb.org/3/tv/airing_today?page=${page}`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialAiringTodayPage() {
    const url = `${this.apiBaseUrl}/tmdb/show/initial-airing-today-tv-shows`;
    const tmdbUrl = 'https://api.themoviedb.org/3/tv/airing_today?page=1';
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  searchShows(page: number, query: string) {
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows`;
    const tmdbUrl = `https://api.themoviedb.org/3/search/tv?query=${query}&include_adult=false&language=en-US&page=${page}`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getInitialSearchPage(query: string) {
    const url = `${this.apiBaseUrl}/tmdb/show/initial-all-shows`;
    const tmdbUrl = `https://api.themoviedb.org/3/search/tv?query=${query}&include_adult=false&language=en-US&page=1`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getTvShowsByGenre(genre: number, airDate: string, page: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows`;
    const tmdbUrl = `https://api.themoviedb.org/3/discover/tv?with_genres=${genre}&page=${page}&first_air_date.gte=${airDate}&include_adult=false&include_null_first_air_dates=false&sort_by=popularity.desc&with_release_type=4|5|6`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  getTvShowDetails(id: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/tv-show-details`;
    const tmdbUrl = `https://api.themoviedb.org/3/tv/${id}?language=en-US`;
    return this.http.post<ShowDetails>(url, {
      url: tmdbUrl,
      request_id: id
    }, { headers: this.headers });
  }

  getAllShowsFromSelectedDate(genre: number, airDate: string, page: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/all-shows-from-date`;
    const tmdbUrl = `https://api.themoviedb.org/3/discover/tv?with_genres=${genre}&page=${page}&first_air_date.gte=${airDate}&include_adult=false&include_null_first_air_dates=false&sort_by=popularity.desc&with_release_type=4|5|6`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }

  makeTvShowDownloadRequest(query: string, seasons: Array<number>, quality: string, tmdb: number, description: string) {
    const url = `${this.apiBaseUrl}/show/query`;
    const headers = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE'
    };

    console.log(`[TV Show Download] Starting download request for: ${query} (Quality: ${quality}, Seasons: [${seasons.join(', ')}], TMDb: ${tmdb})`);

    return this.http.post<QueryRequest>(url, {
      query,
      seasons,
      quality,
      TMDb: tmdb,
      description
    }, {
      headers,
      // rejectUnauthorized: false
    });
  }

  makeAnimeShowDownloadRequest(query: string, seasons: Array<number>, quality: string, tmdb: number, description: string) {
    const url = `${this.apiBaseUrl}/anime/show/query`;
    const headers = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE'
    };

    console.log(`[Anime Show Download] Starting download request for: ${query} (Quality: ${quality}, Seasons: [${seasons.join(', ')}], TMDb: ${tmdb})`);

    return this.http.post<QueryRequest>(url, {
      query,
      seasons,
      quality,
      TMDb: tmdb,
      description
    }, { headers });
  }

  getAllShowsForDetails(genre: number, airDate: string, page: number) {
    const url = `${this.apiBaseUrl}/tmdb/show/all-tv-show-details`;
    const tmdbUrl = `https://api.themoviedb.org/3/discover/tv?with_genres=${genre}&page=${page}&first_air_date.gte=${airDate}&include_adult=false&include_null_first_air_dates=false&language=en-US&sort_by=popularity.desc&with_release_type=4|5|6`;
    return this.http.post<TvShow>(url, { url: tmdbUrl }, { headers: this.headers });
  }
}
