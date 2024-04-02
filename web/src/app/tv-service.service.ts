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
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  public results: TvShowResult[] = [];
  public tvShowsList: TvShow[] = [{page: 0, results: [{adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0}]
  public singleTvShow: TvShow | undefined
  public respData: TvShowResult[] = [];
  public  tvShow: any;
  public  tvShowToBeAdded: Movie

  TvShowService() {
  }

  getTopRated(page: number) {
    var tvShowUrl = ' https://api.themoviedb.org/3/tv/top_rated?page=' + page.toString()

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }

  getInitialTopRatedPage() {
    var tvShowUrl = ' https://api.themoviedb.org/3/tv/top_rated?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }

  getOnTheAir(page: number) {
    var tvShowUrl = 'https://api.themoviedb.org/3/tv/on_the_air?page=' + page.toString()

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }

  getInitialOnTheAirPage() {
    var tvShowUrl = 'https://api.themoviedb.org/3/tv/on_the_air?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }

  getPopular(page: number) {
    var tvShowUrl = 'https://api.themoviedb.org/3/tv/popular?page=' + page.toString();

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }


  getInitialPopularPage() {
    var tvShowUrl = 'https://api.themoviedb.org/3/tv/popular?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }

  getAiringToday(page: number) {
    var tvShowUrl = 'https://api.themoviedb.org/3/tv/airing_today?page=' + page.toString()

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }

  getInitialAiringTodayPage() {
    var tvShowUrl = 'https://api.themoviedb.org/3/tv/airing_today?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvShowUrl, { headers: this.headers });
  }

  getAllShows(page: number, query: string) {
    // TODO: Add multiple genres filter.
    // with_genres=Action%20AND%20Comedy
    // TODO: Add multiple people filter.
    // with_people=Shah%20Rukh%20Khan%20AND%20Shah%20Rukh%20Khan

    var tvShowUrl = 'https://api.themoviedb.org/3/search/tv?query=' + query + '&include_adult=false&language=en-US&&page=' + page.toString();

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(tvShowUrl, { headers: this.headers });
  }

  getInitialPage(query: string) {
    var tvShowUrl = 'https://api.themoviedb.org/3/search/tv?query=' + query + '&include_adult=false&language=en-US&&page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(tvShowUrl, { headers: this.headers });
  }

  makeTvShowDownloadRequest(query: string, seasons: Array<number>, quality: string, TMDb: number, description: string) {
    var queryApiHeaders = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE',
    };

    var tvShowUrl = 'https://' + environment.envVar.ip + ':' + environment.envVar.port + '/show/query';

    return this.http.post<QueryRequest>(tvShowUrl, { "query": query, "seasons": seasons, "quality": quality, "TMDb": TMDb, 'description': description }, { headers: queryApiHeaders });
  }

  makeAnimeDownloadRequest(query: string, episodes: number) {
    var queryApiHeaders = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Alohw-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE',
    };

    var tvShowUrl = 'https://127.0.0.1:8782/anime/query';

    return this.http.post<QueryRequest>(tvShowUrl, { "query": query, "episodes": episodes }, { headers: queryApiHeaders });
  }

  getGenres() {
    return this.http.get<GenreRequest>(this.genreUrl, { headers: this.headers })
  }

  getShowDetails(id) {
    var tvUrl = 'https://api.themoviedb.org/3/tv/'+ id +'?language=en-US';

    return this.http.get<ShowDetails>(tvUrl, { headers: this.headers })
  }

  // Movie[] {
  getAllTvShows(genre: number, airDate: string, page: number) {
      var tvUrl = 'https://api.themoviedb.org/3/discover/tv?with_genres=' + genre.toString() + '&page=' + page.toString() +  '&first_air_date.gte=' + airDate + '&include_adult=false&include_null_first_air_dates=false&sort_by=popularity.desc&with_release_type=4|5|6';

      return this.http.get<TvShow>(tvUrl, { headers: this.headers })
  }

  getAllShowsForDetails(genre: number, airDate: string, page: number) {
    var tvUrl = 'https://api.themoviedb.org/3/discover/tv?with_genres=' + genre.toString() + '&page=' + page.toString() +  '&first_air_date.gte=' + airDate + '&include_adult=false&include_null_first_air_dates=false&language=en-US&sort_by=popularity.desc&with_release_type=4|5|6';


    // const options = { method: 'GET', headers: this.headers };

    // const data = await fetch(`${this.movieUrl}`, options);
    // return await data.json() ?? {};

    return this.http.get<TvShow>(tvUrl, { headers: this.headers });
  }

  getAllShowsFromSelectedDate(genre: number, airDate: string, page: number) {
    console.log(genre);


    // TODO: Add multiple genres filter.
    // with_genres=Action%20AND%20Comedy
    // TODO: Add multiple people filter.
    // with_people=Shah%20Rukh%20Khan%20AND%20Shah%20Rukh%20Khan
    var tvUrl = 'https://api.themoviedb.org/3/discover/tv?with_genres=' + genre.toString() + '&page=' + page.toString() +  '&first_air_date.gte=' + airDate + '&include_adult=false&include_null_first_air_dates=false&sort_by=popularity.desc&with_release_type=4|5|6';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<TvShow>(tvUrl, { headers: this.headers });
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
}
