import { Component, inject, ViewChild, OnInit, AfterViewInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Movie, MovieResult, Genre } from '../http-service/http-service.component';
import { MovieService, MovieFilterOptions } from '../movies.service';
import { SearchMovieListComponent } from '../search-movie-list/search-movie-list.component';
import { MatPaginator } from '@angular/material/paginator';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatSliderModule } from '@angular/material/slider';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';

@Component({
  selector: 'app-search-movies',
  standalone: true,
  imports: [
    CommonModule,
    SearchMovieListComponent,
    MatPaginatorModule,
    FormsModule,
    MatExpansionModule,
    MatSliderModule,
    MatSelectModule,
    MatCheckboxModule
  ],
  providers: [NgModel],
  template: `
    <!-- SearchMoviesComponent -->
    <div class="container">
      <section class="header-section">
        <form class="search-form" (ngSubmit)="getMovies(1)">
          <input type="text" [(ngModel)]="movieSearch" name="movieSearch" id="movieSearch" placeholder="Find Movie by Title" #filter />

          <!-- Advanced Search Toggle -->
          <mat-expansion-panel class="advanced-search-panel">
            <mat-expansion-panel-header>
              <mat-panel-title>
                Advanced Search
              </mat-panel-title>
            </mat-expansion-panel-header>

            <!-- Advanced Search Options -->
            <div class="advanced-filters">
              <div class="filter-row">
                <div class="filter-group">
                  <label for="genres">Genres:</label>
                  <mat-select [(ngModel)]="selectedGenres" name="genres" id="genres" multiple>
                    <mat-option *ngFor="let genre of genreDetails" [value]="genre.id">{{genre.name}}</mat-option>
                  </mat-select>
                </div>

                <!-- For SearchMoviesComponent -->
                <div class="filter-group">
                  <label for="year">Release Year:</label>
                  <input type="number" [(ngModel)]="yearFilter" placeholder="YYYY" name="year" id="year" min="1900" max="2099">
                  <span class="input-hint">Enter year only (e.g., 2022)</span>
                </div>
              </div>

              <div class="filter-row">
                <div class="filter-group">
                  <label for="yearRangeStart">Year Range Start:</label>
                  <input type="number" [(ngModel)]="yearRangeStart" name="yearRangeStart" id="yearRangeStart" min="1900" max="2099">
                </div>

                <div class="filter-group">
                  <label for="yearRangeEnd">Year Range End:</label>
                  <input type="number" [(ngModel)]="yearRangeEnd" name="yearRangeEnd" id="yearRangeEnd" min="1900" max="2099">
                </div>
              </div>

              <div class="filter-row">
                <div class="filter-group">
                  <label for="minRating">Minimum Rating:</label>
                  <input type="number" [(ngModel)]="minRating" name="minRating" id="minRating" min="0" max="10" step="0.1">
                </div>

                <div class="filter-group">
                  <label for="maxRating">Maximum Rating:</label>
                  <input type="number" [(ngModel)]="maxRating" name="maxRating" id="maxRating" min="0" max="10" step="0.1">
                </div>
              </div>

              <div class="filter-row">
                <div class="filter-group">
                  <label for="sortBy">Sort By:</label>
                  <mat-select [(ngModel)]="sortBy" name="sortBy" id="sortBy">
                    <mat-option value="popularity.desc">Popularity (Descending)</mat-option>
                    <mat-option value="popularity.asc">Popularity (Ascending)</mat-option>
                    <mat-option value="vote_average.desc">Rating (Descending)</mat-option>
                    <mat-option value="vote_average.asc">Rating (Ascending)</mat-option>
                    <mat-option value="release_date.desc">Release Date (Newest)</mat-option>
                    <mat-option value="release_date.asc">Release Date (Oldest)</mat-option>
                    <mat-option value="revenue.desc">Revenue (Descending)</mat-option>
                    <mat-option value="title.asc">Title (A-Z)</mat-option>
                  </mat-select>
                </div>

                <div class="filter-group">
                  <label for="language">Language:</label>
                  <mat-select [(ngModel)]="language" name="language" id="language">
                    <mat-option value="">Any</mat-option>
                    <mat-option value="en">English</mat-option>
                    <mat-option value="fr">French</mat-option>
                    <mat-option value="es">Spanish</mat-option>
                    <mat-option value="de">German</mat-option>
                    <mat-option value="ja">Japanese</mat-option>
                    <mat-option value="ko">Korean</mat-option>
                    <mat-option value="zh">Chinese</mat-option>
                  </mat-select>
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

        <!-- Update results section to include loading state -->
        <div class="results" [class.loading]="isLoading" *ngIf="this.filteredMovieList.length != 0">
          <div class="movie-item" *ngFor="let movieItem of this.filteredMovieList; index as i;">
            <div class="movie-info">
              <app-search-movie-list
                [movieItem]="movieItem" [page]="movieItem.page" [query]="filter.value">
              </app-search-movie-list>
            </div>
          </div>
        </div>
      </section>

      <div class="no-results" *ngIf="!isLoading && this.filteredMovieList.length === 0">
        <p>No movies found matching your criteria. Try adjusting your search filters.</p>
      </div>

      <footer class="paginator-container">
        <mat-paginator
          [length]="this.totalMovies"
          [pageSize]="this.moviesLength"
          [pageIndex]="page - 1"
          aria-label="Select page"
          (page)="onPageChange($event)">
        </mat-paginator>
      </footer>
    </div>
  `,
  styleUrls: ['./search-movies.component.sass', '../../styles.sass']
})
export class SearchMoviesComponent implements OnInit, AfterViewInit {
  @ViewChild(MatPaginator) paginator: MatPaginator;
  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  public movieTitles = [{}];
  public fetchedMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0 }]
  public filteredMovieList: Movie[] = [];
  public allMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0 }]
  public moviesLength: number;
  public totalMovies: number;
  public releaseYear: string[] = [""];
  public endYear: string[] = [""];
  public movieSearch: string = '';
  public genreDetails: Genre[] = [];
  public pages: number[] = [0];
  public page: number = 1;

  // Advanced search filters
  public selectedGenres: number[] = [];
  public yearFilter: number | null = null;
  public yearRangeStart: number | null = null;
  public yearRangeEnd: number | null = null;
  public minRating: number | null = null;
  public maxRating: number | null = null;
  public sortBy: string = 'popularity.desc';
  public language: string = '';
  public includeAdult: boolean = false;

  // Optional loading state
  public isLoading: boolean = false;

  public movieService: MovieService = inject(MovieService)

  constructor() {
    this.movieService.getGenres().subscribe((resp) => {
      if (resp && resp.genres) {
        this.genreDetails = resp.genres;
      }
    });
  }

  ngOnInit() {
    // Initialize with popular movies
    this.getMovies(1);
  }

  ngAfterViewInit() {
    // Initialize paginator with current page if available
    if (this.paginator) {
      this.paginator.pageIndex = this.page - 1;
      this.paginator.length = this.totalMovies || 0;
      this.paginator.pageSize = this.moviesLength || 20;
    }
  }

  // Fixed onPageChange method
  onPageChange(event: PageEvent) {
    console.log('Page changed:', event);
    this.page = event.pageIndex + 1;

    // Call getMovies with the new page number
    this.getMovies(this.page);

    // Scroll to top of the page
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  // Fixed getMovies method
  async getMovies(page: number) {
    console.log('getMovies called with page:', page);

    // Set loading state
    this.isLoading = true;

    // Clear arrays by reassigning them instead of using pop()
    this.fetchedMovies = [];
    this.movieTitles = [];

    // Build filter options
    const filterOptions: MovieFilterOptions = {
      page: page,
      includeAdult: this.includeAdult,
      sortBy: this.sortBy
    };

    // Add optional filters if they have values
    if (this.selectedGenres && this.selectedGenres.length > 0) {
      filterOptions.genres = this.selectedGenres;
    }

    if (this.yearFilter) {
      filterOptions.year = this.yearFilter;
    } else if (this.yearRangeStart || this.yearRangeEnd) {
      filterOptions.yearRange = {};
      if (this.yearRangeStart) {
        filterOptions.yearRange.start = this.yearRangeStart;
      }
      if (this.yearRangeEnd) {
        filterOptions.yearRange.end = this.yearRangeEnd;
      }
    }

    // Only add minRating and maxRating if they are not null
    if (this.minRating !== null) {
      filterOptions.minRating = this.minRating;
    }

    if (this.maxRating !== null) {
      filterOptions.maxRating = this.maxRating;
    }

    if (this.language) {
      filterOptions.language = this.language;
    }

    // Choose which API to call based on search query
    let observable;
    if (this.movieSearch) {
      observable = this.movieService.searchMovies(this.movieSearch, filterOptions);
    } else if (Object.keys(filterOptions).length > 2) {
      // If we have filters beyond page and includeAdult, use discover
      observable = this.movieService.discoverMovies(filterOptions);
    } else {
      // Default to popular movies if no search or filters
      observable = this.movieService.getPopular(filterOptions);
    }

    // Use modern subscription pattern
    observable.subscribe({
      next: (resp) => {
        console.log('API response:', resp);

        // Check for valid response
        if (!resp || !resp['results']) {
          console.error('Invalid response format:', resp);
          this.isLoading = false;
          return;
        }

        // Use bracket notation for accessing ALL properties
        this.moviesLength = resp['results'].length;
        this.totalMovies = resp['total_result'] || resp['total_results'];

        resp['results'].forEach((movie) => {
          let result: MovieResult[] = [{
            adult: movie['adult'],
            backdrop_path: movie['backdrop_path'],
            genre_ids: movie['genre_ids'],
            id: movie['id'],
            title: movie['title'],
            release_date: movie['release_date'],
            original_language: movie['original_language'],
            original_title: movie['original_title'],
            overview: movie['overview'],
            popularity: movie['popularity'],
            poster_path: this.baseUrl + movie['poster_path'],
            vote_average: movie['vote_average'],
            vote_count: movie['vote_count'],
            video: movie['video']
          }];

          this.fetchedMovies.push({
            page: resp['page'],
            results: result,
            total_pages: resp['total_pages'],
            total_result: resp['total_result'] || resp['total_results']
          });

          if (this.pages) {
            this.pages.push(resp['page']);
          }

          if (movie['release_date']) {
            if (this.releaseYear) {
              this.releaseYear.push(movie['release_date']);
            }
            if (this.endYear) {
              this.endYear.push(movie['release_date']);
            }
          }
        });

        // Simple assignment for lists
        this.allMovies = [...this.fetchedMovies];
        this.filteredMovieList = this.fetchedMovies;

        // Update paginator if available
        if (this.paginator) {
          this.paginator.pageIndex = this.page - 1;
          this.paginator.length = this.totalMovies;
        }

        // Log status for debugging
        console.log('Movies processed:', this.fetchedMovies.length);
        console.log('Total movies:', this.totalMovies);
        console.log('Current page:', this.page);

        // End loading state
        this.isLoading = false;

        // Scroll to top
        window.scrollTo({ top: 0, behavior: 'auto' });
      },
      error: (error) => {
        console.error('Error fetching movies:', error);
        this.isLoading = false;
      }
    });
  }

  protected readonly Math = Math;
}
