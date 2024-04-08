import { Injectable, inject, NgModule, Output, EventEmitter, Component } from '@angular/core';
import { Movie, MovieResult, GenreRequest, QueryRequest, MovieDetails } from './http-service/http-service.component';
import { Observable, map, of } from 'rxjs';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { environment } from './environments/environment.prod';

@Injectable({
  providedIn: 'root'
})

export class MovieService {
  http: HttpClient = inject(HttpClient)

  public genreUrl = 'https://api.themoviedb.org/3/genre/movie/list?language=en';
  public headers = {
    'Authorization': "Bearer " + environment.envVar.authorization,
    'accept': 'application/json',
  }
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  public results: MovieResult[] = [];
  public movieList: Movie[] = [{page: 0, results: [{adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0}]
  public singleMovie: Movie | undefined
  public respData: MovieResult[] = [];
  public  movie: any;
  public  movieToBeAdded: Movie


  MovieService() {
  }

  getGenres() {
    return this.http.get<GenreRequest>(this.genreUrl, { headers: this.headers })
  }

    // Movie[] {
  getAllMoviesByGenre(genre: number, releaseDateStart: string, releaseDateEnd: string, page: number) {
    // TODO: Add multiple genres filter.
    // with_genres=Action%20AND%20Comedy
    // TODO: Add multiple people filter.
    // with_people=Shah%20Rukh%20Khan%20AND%20Shah%20Rukh%20Khan
    var movieUrl = 'https://api.themoviedb.org/3/discover/movie?with_genres=' + genre.toString() + '&page=' + page.toString() + '&include_adult=false&include_null_first_air_dates=false&release_date.gte=' + releaseDateStart + '&release_date.lte=' + releaseDateEnd + '&sort_by=popularity.desc&with_release_type=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getTopRated(page: number) {
    var movieUrl = ' https://api.themoviedb.org/3/movie/top_rated?page=' + page.toString()

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getInitialTopRatedPage() {
    var movieUrl = ' https://api.themoviedb.org/3/movie/top_rated?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getUpcoming(page: number) {
    var movieUrl = 'https://api.themoviedb.org/3/movie/upcoming?page=' + page.toString()

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getInitialUpcomingPage() {
    var movieUrl = 'https://api.themoviedb.org/3/movie/upcoming?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getPopular(page: number) {
    var movieUrl = 'https://api.themoviedb.org/3/movie/popular?page=' + page.toString();

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }


  getInitialPopularPage() {
    var movieUrl = 'https://api.themoviedb.org/3/movie/popular?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getNowPlaying(page: number) {
    var movieUrl = 'https://api.themoviedb.org/3/movie/now_playing?page=' + page.toString()

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getInitialNowPlayingPage() {
    var movieUrl = 'https://api.themoviedb.org/3/movie/now_playing?page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getAllMovies(page: number, query: string) {
    // TODO: Add multiple genres filter.
    // with_genres=Action%20AND%20Comedy
    // TODO: Add multiple people filter.
    // with_people=Shah%20Rukh%20Khan%20AND%20Shah%20Rukh%20Khan

    var movieUrl = 'https://api.themoviedb.org/3/search/movie?query=' + query + '&include_adult=false&page=' + page.toString();

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getInitialPage(query: string) {
    var movieUrl = 'https://api.themoviedb.org/3/search/movie?query=' + query + '&include_adult=false&page=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  makeMovieDownloadRequest(query: string, name: string, year: string, quality: string, tmdb: number, description: string) {
    var queryApiHeaders = {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'POST,DELETE',
    };

    var movieUrl = environment.envVar.transport + environment.envVar.ip + ':' + environment.envVar.port + '/movie/query';

    const options = {
      headers: queryApiHeaders,
      rejectUnauthorized: false,
    };

    return this.http.post<QueryRequest>(movieUrl, { "query": query, 'name': name, 'year': year, 'quality': quality, 'Imdb': tmdb, 'description': description }, options);
  }

  getMovieDetails(id) {
    var movieUrl = 'https://api.themoviedb.org/3/movie/'+ id +'';

    return this.http.get<MovieDetails>(movieUrl, { headers: this.headers })
  }

  getAllMoviesFromSelectedDate(genre: number, releaseDateStart: string, releaseDateEnd: string, page: number) {
    // TODO: Add multiple genres filter.
    // with_genres=Action%20AND%20Comedy
    // TODO: Add multiple people filter.
    // with_people=Shah%20Rukh%20Khan%20AND%20Shah%20Rukh%20Khan
    var movieUrl = 'https://api.themoviedb.org/3/discover/movie?with_genres=' + genre.toString() + '&page=' + page.toString() + '&include_adult=false&include_null_first_air_dates=false&release_date.gte=' + releaseDateStart + '&release_date.lte=' + releaseDateEnd + '&sort_by=popularity.desc&with_release_type=1';

    setTimeout(function () {
    }, 4000000);

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getAllMoviesForDetails(genre: number, releaseDateStart: string, releaseDateEnd: string, page: number) {
    // TODO: Add multiple genres filter.
    // with_genres=Action%20AND%20Comedy
    // TODO: Add multiple people filter.
    // with_people=Shah%20Rukh%20Khan%20AND%20Shah%20Rukh%20Khan
    var movieUrl = 'https://api.themoviedb.org/3/discover/movie?with_genres=' + genre.toString() + '&page=' + page.toString() + '&include_adult=false&include_null_first_air_dates=false&release_date.gte=' + releaseDateStart + '&release_date.lte=' + releaseDateEnd + '&sort_by=popularity.desc&with_release_type=1';

    return this.http.get<Movie>(movieUrl, { headers: this.headers });
  }

  getMovieById(id: number, genre: number, releaseDateStart: string, releaseDateEnd: string, page: number): Movie | undefined {
    this.getAllMoviesForDetails(genre, releaseDateStart, releaseDateEnd, page).subscribe((resp) => {

      resp['results'].forEach((movie) => {
        let page = resp['page'];
        let isAdult = movie['adult'];
        let backdropPath = movie['backdrop_path'];
        let genreIds = movie['genre_ids'];
        let id = movie['id'];
        let releaseDate = movie['release_date'];
        let video = movie['video'];
        let title = movie['name'];
        let originalLanguage = movie['original_language'];
        let originalTitle = movie['original_title'];
        let overview = movie['overview'];
        let popularity = movie['popularity'];
        let posterPath = this.baseUrl + movie['poster_path'];
        let voteAverage = movie['vote_average'];
        let voteCount = movie['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];

        let result: MovieResult[] = [{adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video}]

        this.movieList.push({ page: page, results: result,  total_pages: totalPages, total_result: totalResult });
      })

      this.movieList.splice(0, 1);

      for (var i = 0; i < this.movieList.length; i++) {
        this.singleMovie = this.movieList.find(movieResult => movieResult.results[i].id === id);
      }

      return this.singleMovie;
    })

    return this.singleMovie;
    }
}
