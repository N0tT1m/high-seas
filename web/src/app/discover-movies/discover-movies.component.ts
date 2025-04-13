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
          <input type='text' [(ngModel)]="releaseYear" name="releaseYear" id="releaseYear" class='select-section' />
        </div>

        <div class="filters-div">
          <label for="endYear">End Year:</label>
          <input type='text' [(ngModel)]="endYear" name="endYear" id="endYear" class='select-section' />
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

  onPageChange(event?: PageEvent) {
    if (event === null) {
      // Handle null case
    } else {
      this.page = event!.pageIndex + 1;
      this.getGenre(this.page);
    }
  }

  getMoviesFromDate() {
    while (this.page <= this.totalMovies) {
      this.movieService.getAllMoviesFromSelectedDate(this.genre, this.releaseYear, this.endYear, this.page).subscribe((resp) => {
        this.moviesLength = resp.results.length;
        this.totalMovies = resp.total_pages;

        resp.results.forEach((movie) => {
          let result: MovieResult[] = [{
            adult: movie.adult,
            backdrop_path: movie.backdrop_path,
            genre_ids: movie.genre_ids,
            id: movie.id,
            title: movie.title,
            release_date: movie.release_date,
            original_language: movie.original_language,
            original_title: movie.original_title,
            overview: movie.overview,
            popularity: movie.popularity,
            poster_path: this.baseUrl + movie.poster_path,
            vote_average: movie.vote_average,
            vote_count: movie.vote_count,
            video: movie.video
          }];

          this.allMovies.push({
            page: resp.page,
            results: result,
            total_pages: resp.total_pages,
            total_result: resp.total_result
          });
        });
      });

      this.page++;
    }

    this.allMovies.splice(0, 1);
  }

  async getGenre(page: number) {
    // Clear current movies
    while (this.fetchedMovies.length > 0) {
      this.fetchedMovies.pop();
    }

    while (this.movieTitles.length > 0) {
      this.movieTitles.pop();
    }

    // Build filter options
    const filterOptions: MovieFilterOptions = {
      page: page,
      includeAdult: this.includeAdult,
      sortBy: this.sortBy
    };

    // Add genre if selected (not 0)
    if (this.genre !== 0) {
      filterOptions.genres = [this.genre];
    }

    // Add year range if specified
    if (this.releaseYear || this.endYear) {
      filterOptions.yearRange = {};
      if (this.releaseYear) {
        filterOptions.yearRange.start = parseInt(this.releaseYear);
      }
      if (this.endYear) {
        filterOptions.yearRange.end = parseInt(this.endYear);
      }
    }

    // Add rating filters if specified
    if (this.minRating) {
      filterOptions.minRating = this.minRating;
    }
    if (this.maxRating) {
      filterOptions.maxRating = this.maxRating;
    }

    // Add language filter if specified
    if (this.language) {
      filterOptions.language = this.language;
    }

    // Use the discover API with the filters
    this.movieService.discoverMovies(filterOptions).subscribe((resp) => {
      this.moviesLength = resp.results.length;
      this.totalMovies = resp.total_result;

      resp.results.forEach((movie) => {
        let result: MovieResult[] = [{
          adult: movie.adult,
          backdrop_path: movie.backdrop_path,
          genre_ids: movie.genre_ids,
          id: movie.id,
          title: movie.title,
          release_date: movie.release_date,
          original_language: movie.original_language,
          original_title: movie.original_title,
          overview: movie.overview,
          popularity: movie.popularity,
          poster_path: this.baseUrl + movie.poster_path,
          vote_average: movie.vote_average,
          vote_count: movie.vote_count,
          video: movie.video
        }];

        this.fetchedMovies.push({
          page: resp.page,
          results: result,
          total_pages: resp.total_pages,
          total_result: resp.total_result
        });
      });

      // Update filtered list
      if (this.filteredMovieList.length > 0) {
        this.filteredMovieList = [];
      }
      this.filteredMovieList = this.fetchedMovies;
    });
  }
}
