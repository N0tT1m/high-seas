import { Component, inject, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { Movie, MovieResult } from '../http-service/http-service.component';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, NgModel } from '@angular/forms';
import { MovieService } from '../movies.service';

@Component({
  selector: 'app-now-playing-details',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    ReactiveFormsModule,
    FormsModule,
  ],
  providers: [MovieService, NgModel],
  template: `
  <article class="the-girls" *ngFor="let movie of this.fetchedMovie?.results; index as i;">
  <img class="movie-photo the-girls" [src]="movie!.poster_path"
    alt="Exterior photo of {{movie!.title}}"/>
  <section class="movie-description the-girls">
    <h2 class="movie-title">{{movie!.title}}</h2>
    <p class="movie-overview">{{movie!.overview}}</p>
  </section>
  <section class="movie-details the-girls">
    <h2 class="section-heading">About this movie</h2>
    <ul>
      <div class="movie-div">
        <li>Original Language: {{movie!.original_language}}</li>
        <li>Original Title: {{movie!.original_title}}</li>
        <li>popularity: {{movie!.popularity}}</li>
        <li>Release Date: {{movie!.release_date}}</li>
      </div>
      <div class="movie-div">
        <li>IMDB ID: {{this.imdbId}}</li>
        <li>Budget for {{movie!.title}}: {{this.budget}}</li>
        <li>Homepage for {{movie!.title}}: {{this.homepage}}</li>
        <li>Tagline for {{movie!.title}}: {{this.tagline}}</li>
      </div>
    </ul>

    <div class="download-quality">
      <select [(ngModel)]="quality" name="quality" id="quality">
        <option value="4k">4k</option>
        <option value="2k">2k</option>
        <option value="1080p">1080p</option>
        <option value="720p">720p</option>
        <option value="480p">480p</option>
        <option value="240p">240p</option>
      </select>
    </div>

    <button (click)="downloadMovie(movie.title, movie.title, movie.release_date, this.quality)">Download Movie</button>
  </section>
</article>
    `,
  styleUrls: ['./popular-movies-details.component.sass', '../../styles.sass'],
})

export class PopularMoviesDetailsComponent {
  private baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  public route: ActivatedRoute = inject(ActivatedRoute);
  public movieService = inject(MovieService);
  public fetchedMovie: Movie | undefined;
  public movieList: Movie[] = [{page: 0, results: [{adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0}]
  public tagline: string = "";
  public homepage: string = "";
  public releaseDate: string = "";
  public imdbId: string = "";
  public budget: number = 0;
  public belongsToCollection: boolean = false;
  public overview: string = "";
  public quality: string;
  public tmdbId: number = 0;

  constructor() {

  }

  ngOnInit() {
    const movieId = parseInt(this.route.snapshot.params['id'], 10);
    const page = parseInt(this.route.snapshot.params['page'], 10);

    this.movieService.getPopular(page).subscribe((resp) => {
      resp['results'].forEach((movie) => {
        let page = resp['page'];
        let isAdult = movie['adult'];
        let backdropPath = movie['backdrop_path'];
        let genreIds = movie['genre_ids'];
        let id = movie['id'];
        let releaseDate = movie['release_date'];
        let video = movie['video'];
        let title = movie['title'];
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
        this.fetchedMovie = this.movieList.find(movieResult => movieResult.results[i]!.id === movieId);
      }
    })

    this.movieService.getMovieDetails(movieId).subscribe(movie => {
      this.releaseDate = movie.release_date;
      this.homepage = movie.homepage;
      this.tagline = movie.tagline;
      this.imdbId = movie.imdb_id;
      this.tmdbId = movie.id;
      this.budget = movie.budget;
      this.belongsToCollection = movie.belongs_to_collection;
      this.overview = movie.overview;
    })
  }

  downloadMovie(title: string, name: string, year: string, quality: string) {
    this.movieService.makeDownloadRequest(title, name, year, quality, Number(this.imdbId)).subscribe(request => console.log(request));
  }
}
