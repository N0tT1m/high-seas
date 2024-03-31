import { CommonModule, NgFor } from '@angular/common';
import { HttpClientModule  } from '@angular/common/http'
import { Component, OnInit, ViewChild, inject, Input, Output, EventEmitter } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule, Gallery, GalleryRef, ImageItem } from 'ng-gallery';
import { Movie, MovieResult, TvShow, TvShowResult, GenreRequest, Genre } from '../http-service/http-service.component';
import { Observable } from 'rxjs';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';
import { MovieListComponent } from '../discover-movie-list/discover-movie-list.component';
import { MatPaginator } from '@angular/material/paginator';
import { NgModel, FormsModule } from '@angular/forms';
import {MatPaginatorModule, PageEvent } from '@angular/material/paginator';

@Component({
  selector: 'app-discover-movies',
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule, MovieListComponent, FormsModule, MatPaginatorModule],
  providers: [MovieService, TvShowService, NgModel],
  template: `
  <!-- DiscoverMoviesComponent -->
  <div class="container">
    <section class="header-section">
      <form class="search-form">
        <input type="text" placeholder="Filter Movie by Title" #filter>
        <button class="big-btn filter-button" type="button" (click)="filterResults(filter.value)">Filter</button>
      </form>

      <div id="filters">
        <label for="filters" id='filter-label'>Filters for Movies:</label>
        <form class='filters-form' (ngSubmit)="getGenre(1)">
          <div class="filters-div">
            <label for="genre">Genre:</label>
            <select [(ngModel)]="genre" name="genres" id="genres" (ngModelChange)="getGenre(1)" class='select-section'>
              <option id="genre" *ngFor="let genre of genreDetails" value="{{genre.id}}">{{genre.name}}</option>
            </select>
          </div>

          <div class="filters-div">
            <label for="releaseYear">Release Year:</label>
            <input type='text' [(ngModel)]="releaseYear" name="releaseYear" id="releaseYear" class='select-section' />
          </div>

          <div class="filters-div">
            <label for="endYear">End Year:</label>
            <input type='text' [(ngModel)]="endYear" name="endYear" id="endYear" class='select-section' />
          </div>

          <button class="button big-btn filter-button" type="submit">Filter</button>
        </form>
      </div>
    </section>

    <div class="results" *ngIf="genre != 0">
      <div class="movie-item" *ngFor="let movieItem of filteredMovieList">
        <div class="movie-info">
          <app-movie-list
            [movieItem]="movieItem" [page]="this.page" [releaseYear]="this.releaseYear" [endYear]="this.endYear" [genre]="this.genre">
          </app-movie-list>
        </div>
      </div>
    </div>

    <footer>
      <mat-paginator [length]=this.totalMovies
                [pageSize]=this.moviesLength
                aria-label="Select page"
                (page)="onPageChange($event)">
      </mat-paginator>
    </footer>
  </div>
  `,
  styleUrls: ['./discover-movies.component.sass', '../../styles.sass']
})
export class DiscoverMoviesComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;

  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public movie: Movie;
  public movieTitles = [{}];
  public fetchedMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public allMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public filteredMovieList: Movie[] = [];
  public galleryMoviesRef: GalleryRef;
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public releaseYear: string;
  public endYear: string;
  public moviesLength: number;
  public totalMovies: number;
  public page: number = 1;

  public movieService: MovieService = inject(MovieService)


  constructor(private gallery: Gallery) {
    this.movieService.getGenres().subscribe((resp) => {
      resp['genres'].forEach((genre) => {
        var item = {id: genre.id, name: genre.name}

        this.genreDetails.push(item)
      })
    })
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredMovieList = this.fetchedMovies;
    }

    return this.filteredMovieList = this.fetchedMovies.filter((show) => show.results[0]?.title.toLowerCase().includes(text.toLowerCase()));
  }

  ngOnInit() {
    // Get the galleryRef by id
    this.galleryMoviesRef = this.gallery.ref('movieGallery');

    this.galleryMoviesRef.play()
  }

  filterFormSubmit() {

  }

  onPageChange(event?: PageEvent) {
    if (event === null) {

    } else {
      this.page = event!.pageIndex + 1;
      this.getGenre(this.page);
    }
  }

  getMoviesFromDate() {
    while (this.page <= this.totalMovies) {
      this.movieService.getAllMoviesFromSelectedDate(this.genre, this.releaseYear, this.endYear, this.page).subscribe((resp) => {
        console.log(resp['results']);

        this.moviesLength = resp['results'].length;
        this.totalMovies = resp['totalPages'];

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

          let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

          this.allMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });
        })
      })

      this.page++;
    }

    this.allMovies.splice(0, 1);
  }

  async getGenre(page: number) {
    while (this.fetchedMovies.length > 0) {
      this.fetchedMovies.pop()
    }

    // this.galleryMoviesRef.reset()

    while (this.movieTitles.length > 0) {
      this.movieTitles.pop()
    }

    this.movieService.getAllMoviesByGenre(this.genre, this.releaseYear.toString(), this.endYear.toString(), page).subscribe((resp) => {
      console.log(resp['results']);

      this.moviesLength = resp['results'].length;
      this.totalMovies = resp['total_results'];

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

        let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

        this.fetchedMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });
      })
    })

    this.fetchedMovies.splice(0, 1);

    if (this.filteredMovieList.length > 0) {
      while (this.filteredMovieList.length > 0) {
        this.filteredMovieList.pop();
      }
      this.filteredMovieList = this.fetchedMovies;
    } else {
      this.filteredMovieList = this.fetchedMovies;
    }
  }
}
