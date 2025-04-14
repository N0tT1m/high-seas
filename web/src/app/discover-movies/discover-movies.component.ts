import { CommonModule, NgFor } from '@angular/common';
import { HttpClientModule  } from '@angular/common/http'
import { Component, OnInit, ViewChild, inject, Input, Output, EventEmitter } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule, Gallery, GalleryRef, ImageItem } from 'ng-gallery';
import { Movie, MovieResult, TvShow, TvShowResult, GenreRequest, Genre } from '../http-service/http-service.component';
import { Observable } from 'rxjs';
import { MovieService, MovieFilterOptions } from '../movies.service';
import { TvShowService } from '../tv-service.service';
import { MovieListComponent } from '../discover-movie-list/discover-movie-list.component';
import { MatPaginator } from '@angular/material/paginator';
import { NgModel, FormsModule } from '@angular/forms';
import {MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatSliderModule } from '@angular/material/slider';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';

@Component({
  selector: 'app-discover-movies',
  standalone: true,
  imports: [
    GalleryModule,
    CommonModule,
    RouterModule,
    MovieListComponent,
    FormsModule,
    MatPaginatorModule,
    MatExpansionModule,
    MatSliderModule,
    MatSelectModule,
    MatCheckboxModule
  ],
  providers: [MovieService, TvShowService, NgModel],
  template: `
    <!-- DiscoverMoviesComponent -->
    <div class="container">
      <div id="filters">
        <label for="filters" id='filter-label'>Search for Movies:</label>

        <form class='filters-form' (ngSubmit)="getGenre(1)">
          <!-- Basic Filters -->
          <div class="filters-div">
            <label for="genre">Genre:</label>
            <select [(ngModel)]="genre" name="genres" id="genres" class='select-section'>
              <option id="genre" *ngFor="let genre of genreDetails" value="{{genre.id}}">{{genre.name}}</option>
            </select>
          </div>

          <div class="filters-div">
            <label for="releaseYear">Release Year:</label>
            <input type='text' [(ngModel)]="releaseYear" name="releaseYear" id="releaseYear" placeholder="YYYY" class='select-section' />
            <span class="input-hint">Enter year (e.g., 2022)</span>
          </div>

          <div class="filters-div">
            <label for="endYear">End Year:</label>
            <input type='text' [(ngModel)]="endYear" name="endYear" id="endYear" placeholder="YYYY" class='select-section' />
            <span class="input-hint">Enter end year for range</span>
          </div>

          <!-- Advanced Search Toggle -->
          <mat-expansion-panel class="advanced-search-panel">
            <mat-expansion-panel-header>
              <mat-panel-title>
                Advanced Search Options
              </mat-panel-title>
            </mat-expansion-panel-header>

            <!-- Advanced Search Options -->
            <div class="advanced-filters">
              <div class="filter-row">
                <div class="filter-group">
                  <label for="minRating">Minimum Rating:</label>
                  <input type="number" [(ngModel)]="minRating" name="minRating" id="minRating" min="0" max="10" step="0.1" class='select-section'>
                </div>

                <div class="filter-group">
                  <label for="maxRating">Maximum Rating:</label>
                  <input type="number" [(ngModel)]="maxRating" name="maxRating" id="maxRating" min="0" max="10" step="0.1" class='select-section'>
                </div>
              </div>

              <div class="filter-row">
                <div class="filter-group">
                  <label for="sortBy">Sort By:</label>
                  <select [(ngModel)]="sortBy" name="sortBy" id="sortBy" class='select-section'>
                    <option value="popularity.desc">Popularity (Descending)</option>
                    <option value="popularity.asc">Popularity (Ascending)</option>
                    <option value="vote_average.desc">Rating (Descending)</option>
                    <option value="vote_average.asc">Rating (Ascending)</option>
                    <option value="release_date.desc">Release Date (Newest)</option>
                    <option value="release_date.asc">Release Date (Oldest)</option>
                    <option value="revenue.desc">Revenue (Descending)</option>
                    <option value="title.asc">Title (A-Z)</option>
                  </select>
                </div>

                <div class="filter-group">
                  <label for="language">Language:</label>
                  <select [(ngModel)]="language" name="language" id="language" class='select-section'>
                    <option value="">Any</option>
                    <option value="en">English</option>
                    <option value="fr">French</option>
                    <option value="es">Spanish</option>
                    <option value="de">German</option>
                    <option value="ja">Japanese</option>
                    <option value="ko">Korean</option>
                    <option value="zh">Chinese</option>
                  </select>
                </div>
              </div>

              <div class="filter-row">
                <div class="filter-group">
                  <mat-checkbox [(ngModel)]="includeAdult" name="includeAdult" id="includeAdult">Include Adult Content</mat-checkbox>
                </div>
              </div>
            </div>
          </mat-expansion-panel>

          <button class="button big-btn filter-button" type="submit">Search</button>
        </form>
      </div>

      <label for="filters" id='filter-label'>Filters for Movies:</label>
      <section class="header-section">
        <form class="search-form">
          <input type="text" placeholder="Filter Movie by Title" #filter>
          <button class="big-btn filter-button" type="button" (click)="filterResults(filter.value)">Filter</button>
        </form>
      </section>

      <div class="results" *ngIf="filteredMovieList.length > 0">
        <div class="movie-item" *ngFor="let movieItem of filteredMovieList">
          <div class="movie-info">
            <app-discover-movie-list
              [movieItem]="movieItem" [page]="this.page" [releaseYear]="this.releaseYear" [endYear]="this.endYear" [genre]="this.genre">
            </app-discover-movie-list>
          </div>
        </div>
      </div>

      <footer class="paginator-container">
        <mat-paginator
          [length]="this.totalMovies"
          [pageSize]="this.moviesLength"
          aria-label="Select page"
          (page)="onPageChange($event)">
        </mat-paginator>
      </footer>
    </div>
  `,
  styleUrls: ['./discover-movies.component.sass', '../../styles.sass']
})
export class DiscoverMoviesComponent implements OnInit {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;
  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public movie: Movie;
  public movieTitles = [{}];
  public fetchedMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0 }]
  public allMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0 }]
  public filteredMovieList: Movie[] = [];
  public galleryMoviesRef: GalleryRef;
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public releaseYear: string = '';
  public endYear: string = '';
  public moviesLength: number;
  public totalMovies: number;
  public page: number = 1;

  // Advanced search filters
  public minRating: number | null = null;
  public maxRating: number | null = null;
  public sortBy: string = 'popularity.desc';
  public language: string = '';
  public includeAdult: boolean = false;

  public movieService: MovieService = inject(MovieService);

  constructor(private gallery: Gallery) {
    this.movieService.getGenres().subscribe((resp) => {
      if (resp && resp.genres) {
        this.genreDetails = [{ id: 0, name: "None" }];
        resp.genres.forEach((genre) => {
          this.genreDetails.push(genre);
        });
      } else {
        // Handle legacy response format
        resp['genres']?.forEach((genre) => {
          var item = {id: genre.id, name: genre.name};
          this.genreDetails.push(item);
        });
      }
    });
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredMovieList = this.fetchedMovies;
    }

    return this.filteredMovieList = this.fetchedMovies.filter((show) =>
      show.results[0]?.title.toLowerCase().includes(text.toLowerCase())
    );
  }

  ngOnInit() {
    // Get the galleryRef by id
    this.galleryMoviesRef = this.gallery.ref('movieGallery');
    this.galleryMoviesRef.play();
  }

  // EXACTLY match the working pattern
  onPageChange(event: PageEvent) {
    this.page = event.pageIndex + 1;
    this.getGenre(this.page);
  }

  // Alternative getGenre implementation using getAllMoviesByGenre

  async getGenre(page: number) {
    // Clear current movies
    while (this.fetchedMovies.length > 0) {
      this.fetchedMovies.pop();
    }

    while (this.movieTitles.length > 0) {
      this.movieTitles.pop();
    }

    // Use getAllMoviesByGenre which is more similar to the working getAllTvShows
    this.movieService.getAllMoviesByGenre(this.genre, this.releaseYear, this.endYear, page).subscribe((resp) => {
      this.moviesLength = resp['results'].length;
      this.totalMovies = resp['total_results'] || resp['total_result'];

      resp['results'].forEach((movie) => {
        // Extract movie details using bracket notation
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
        let totalResult = resp['total_results'] || resp['total_result'];

        // Create a result object
        let result: MovieResult[] = [{
          adult: isAdult,
          backdrop_path: backdropPath,
          genre_ids: genreIds,
          id: id,
          title: title,
          release_date: releaseDate,
          original_language: originalLanguage,
          original_title: originalTitle,
          overview: overview,
          popularity: popularity,
          poster_path: posterPath,
          vote_average: voteAverage,
          vote_count: voteCount,
          video: video
        }];

        // Add to fetchedMovies array
        this.fetchedMovies.push({
          page: page,
          results: result,
          total_pages: totalPages,
          total_result: totalResult
        });
      });

      // Update filtered list
      if (this.filteredMovieList.length > 0) {
        while (this.filteredMovieList.length > 0) {
          this.filteredMovieList.pop();
        }
      }
      this.filteredMovieList = this.fetchedMovies;

      // Scroll to top
      window.scrollTo({ top: 0, behavior: 'auto' });
    });
  }
}
