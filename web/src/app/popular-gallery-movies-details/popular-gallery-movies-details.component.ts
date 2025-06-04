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
  <article class="movie-details" *ngFor="let movie of this.fetchedMovie?.results; index as i;">
    <div class="movie-header">
      <img *ngIf="movie!.poster_path" class="movie-poster" [src]="movie!.poster_path" alt="Poster of {{movie!.title}}" />
      <div class="movie-info">
        <h2 class="movie-title">{{movie!.title}}</h2>
        <p class="movie-overview">{{movie!.overview}}</p>
      </div>
    </div>
    <section class="movie-details-section">
      <h3 class="section-heading">About this movie:</h3>
      <div class="movie-meta">
        <div class="movie-meta-item">
          <div class="movie-meta-label">Original Language:</div>
          <div class="movie-meta-value">{{movie!.original_language}}</div>
        </div>
        <div class="movie-meta-item">
          <div class="movie-meta-label">Original Title:</div>
          <div class="movie-meta-value">{{movie!.original_title}}</div>
        </div>
        <div class="movie-meta-item">
          <div class="movie-meta-label">Popularity:</div>
          <div class="movie-meta-value">{{movie!.popularity}}</div>
        </div>
        <div class="movie-meta-item">
          <div class="movie-meta-label">Release Date:</div>
          <div class="movie-meta-value">{{movie!.release_date}}</div>
        </div>
        <div class="movie-meta-item">
          <div class="movie-meta-label">IMDb ID:</div>
          <div class="movie-meta-value">{{this.imdbId}}</div>
        </div>
        <div class="movie-meta-item">
          <div class="movie-meta-label">Budget for {{ movie.title }}:</div>
          <div class="movie-meta-value">{{ this.budget }}</div>
        </div>
      </div>
      <div class="movie-homepage">
        <span class="movie-homepage-label">Homepage:</span>
        <a class="movie-homepage-link" href="{{this.homepage}}" target="_blank">{{this.homepage}}</a>
      </div>
      <div class="movie-tagline">
        <span class="movie-tagline-label">Tagline for {{movie.title}}:</span>
        <span class="movie-tagline-value">{{this.tagline}}</span>
      </div>
      <div class="movie-video" *ngIf="movie!.video != false">
        <span class="movie-video-label">Is a video</span>
      </div>
    </section>
    <div class="movie-meta-item" *ngIf="this.in_plex != false">
      <div class="movie-meta-label">Status:</div>
      <div class="movie-meta-value">
        <span class="plex-badge">Available in Plex</span>
      </div>
    </div>
    <div class="movie-actions">
      <div class="movie-download-quality">
        <label for="quality" class="movie-download-quality-label">Download Quality:</label>
        <select [(ngModel)]="quality" name="quality" id="quality" class="movie-download-quality-select">
          <option value="4k">4k</option>
          <option value="2k">2k</option>
          <option value="1080p" selected>1080p</option>
          <option value="720p">720p</option>
          <option value="480p">480p</option>
          <option value="240p">240p</option>
        </select>
      </div>
      <button class="movie-download-button" (click)="downloadMovie(movie.title, movie.title, movie.release_date, this.quality, movie.original_language)">Download Movie</button>
    </div>
  </article>
    `,
  styleUrls: ['./popular-gallery-movies-details.component.sass', '../../styles.sass'],
})

export class PopularGalleryMoviesDetailsComponent {
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
  public quality = '1080p'; // Default download quality
  public tmdbId: number = 0;
  public in_plex = false;

  constructor() {

  }

  ngOnInit() {
    const movieId = parseInt(this.route.snapshot.params['id'], 10);
    const page = parseInt(this.route.snapshot.params['page'], 10);

    this.movieService.getPopularMovies(page).subscribe((resp) => {
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
        let in_plex = resp['in_plex'];

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
      this.in_plex = movie.in_plex;
    })
  }

  downloadMovie(title: string, name: string, year: string, quality: string, lang: string) {
    if (lang === 'ja') {
      console.log('ANIME');
      this.movieService.makeAnimeMovieDownloadRequest(title, name, this.releaseDate, this.quality, Number(this.tmdbId), this.overview).subscribe(request => console.log(request))
    } else {
      console.log('Movie');
      this.movieService.makeMovieDownloadRequest(title, this.quality, Number(this.tmdbId), this.overview).subscribe(
        request => {
          console.log(request);
          // Show the pop-up when the request is successful
          alert('Download request submitted successfully!');
        },
        error => {
          console.error(error);
          // Show an error message if the request fails
          alert('An error occurred while submitting the download request.');
        }
      );
    }
  }
}
