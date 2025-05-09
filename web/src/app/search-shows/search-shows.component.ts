import { CommonModule } from '@angular/common';
import { Component, inject, ViewChild, OnInit, AfterViewInit } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule } from 'ng-gallery';
import { TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { TvShowService, TvShowFilterOptions } from '../tv-service.service';
import { SearchShowListComponent } from '../search-show-list/search-show-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginator } from '@angular/material/paginator';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatSliderModule } from '@angular/material/slider';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';

@Component({
  selector: 'app-search-shows',
  standalone: true,
  imports: [
    GalleryModule,
    CommonModule,
    RouterModule,
    SearchShowListComponent,
    FormsModule,
    MatPaginatorModule,
    MatExpansionModule,
    MatSliderModule,
    MatSelectModule,
    MatCheckboxModule
  ],
  providers: [TvShowService, NgModel],
  template: `
    <div class="container">
      <!-- Add loading indicator -->
      <div class="loading-indicator" *ngIf="isLoading">
        <div class="spinner"></div>
        <span>Loading shows...</span>
      </div>

      <section class="header-section">
        <form class="search-form" (ngSubmit)="getShows(1)">
          <input type="text" [(ngModel)]="showSearch" name="showSearch" id="showSearch" placeholder="Find Show by Name" #filter />

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

                <!-- For SearchShowsComponent -->
                <div class="filter-group">
                  <label for="year">First Air Year:</label>
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
                    <mat-option value="first_air_date.desc">First Air Date (Newest)</mat-option>
                    <mat-option value="first_air_date.asc">First Air Date (Oldest)</mat-option>
                    <mat-option value="name.asc">Name (A-Z)</mat-option>
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
                  <label for="status">Status:</label>
                  <mat-select [(ngModel)]="status" name="status" id="status">
                    <mat-option value="">Any</mat-option>
                    <mat-option value="returning series">Returning Series</mat-option>
                    <mat-option value="ended">Ended</mat-option>
                    <mat-option value="canceled">Canceled</mat-option>
                  </mat-select>
                </div>

                <div class="filter-group">
                  <label for="networkFilter">Network:</label>
                  <mat-select [(ngModel)]="networkFilter" name="networkFilter" id="networkFilter">
                    <mat-option value="">Any</mat-option>
                    <mat-option value="213">Netflix</mat-option>
                    <mat-option value="1024">Amazon Prime</mat-option>
                    <mat-option value="2552">Apple TV+</mat-option>
                    <mat-option value="49">HBO</mat-option>
                    <mat-option value="2739">Disney+</mat-option>
                    <mat-option value="67">Showtime</mat-option>
                    <mat-option value="4">BBC</mat-option>
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
        <div class="results" [class.loading]="isLoading" *ngIf="this.filteredShowsList.length != 0">
          <div class="show-item" *ngFor="let showItem of this.filteredShowsList; index as i;">
            <div class="show-info">
              <app-search-show-list
                [tvShow]="showItem" [page]="showItem.page" [query]="filter.value">
              </app-search-show-list>
            </div>
          </div>
        </div>
      </section>

      <!-- No results message -->
      <div class="no-results" *ngIf="!isLoading && this.filteredShowsList.length === 0">
        <p>No shows found matching your criteria. Try adjusting your search filters.</p>
      </div>

      <footer class="paginator-container">
        <mat-paginator
          [length]="this.totalShows"
          [pageSize]="this.showsLength"
          [pageIndex]="page - 1"
          aria-label="Select page"
          (page)="onPageChange($event)">
        </mat-paginator>
      </footer>
    </div>
  `,
  styleUrls: ['./search-shows.component.sass', '../../styles.sass']
})
export class SearchShowsComponent implements OnInit, AfterViewInit {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;
  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public fetchedShows: TvShow[] = [{
    page: 0,
    results: [{
      adult: false,
      backdrop_path: "",
      genre_ids: [],
      id: 0,
      name: "",
      first_air_date: "",
      original_language: "",
      original_name: "",
      overview: "",
      popularity: 0,
      poster_path: "",
      vote_average: 0,
      vote_count: 0,
      video: false
    }],
    total_pages: 0,
    total_result: 0
  }];

  public filteredShowsList: TvShow[] = [];
  public showNames = [{}];
  public allShows: TvShow[] = [{
    page: 0,
    results: [{
      adult: false,
      backdrop_path: "",
      genre_ids: [],
      id: 0,
      name: "",
      first_air_date: "",
      original_language: "",
      original_name: "",
      overview: "",
      popularity: 0,
      poster_path: "",
      vote_average: 0,
      vote_count: 0,
      video: false
    }],
    total_pages: 0,
    total_result: 0
  }];

  public showsLength: number;
  public totalShows: number;
  public releaseYear: string[] = [""];
  public endYear: string[] = [""];
  public showSearch: string = '';
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
  public status: string = '';
  public networkFilter: string = '';

  public tvShowService: TvShowService = inject(TvShowService);

  // Optional loading state
  public isLoading: boolean = false;

  constructor() {
    this.tvShowService.getGenres().subscribe((resp) => {
      if (resp && resp.genres) {
        this.genreDetails = resp.genres;
      }
    });
  }

  ngOnInit() {
    // Initialize with popular shows
    this.getShows(1);
  }

  ngAfterViewInit() {
    // Initialize paginator with current page if available
    if (this.paginator) {
      this.paginator.pageIndex = this.page - 1;
      this.paginator.length = this.totalShows || 0;
      this.paginator.pageSize = this.showsLength || 20;
    }
  }

  // Fixed onPageChange method
  onPageChange(event: PageEvent) {
    console.log('Page changed:', event);
    this.page = event.pageIndex + 1;

    // Call getShows with the new page number
    this.getShows(this.page);

    // Scroll to top of the page
    window.scrollTo({top: 0, behavior: 'smooth'});
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredShowsList = this.allShows;
    }

    return this.filteredShowsList = this.allShows.filter((show) =>
      show.results[0]?.name.toLowerCase().includes(text.toLowerCase())
    );
  }

  // Fixed getShows method
  async getShows(page: number) {
    console.log('getShows called with page:', page);

    // Set loading state
    this.isLoading = true;

    // Clear arrays by reassigning them instead of using pop()
    this.fetchedShows = [];
    this.showNames = [];

    // Build filter options
    const filterOptions: TvShowFilterOptions = {
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

    if (this.status) {
      filterOptions.status = this.status;
    }

    if (this.networkFilter) {
      filterOptions.withNetworks = [parseInt(this.networkFilter)];
    }

    // Choose which API to call based on search query
    let observable;
    if (this.showSearch) {
      observable = this.tvShowService.searchShows(this.showSearch, filterOptions);
    } else if (Object.keys(filterOptions).length > 2) {
      // If we have filters beyond page and includeAdult, use discover
      observable = this.tvShowService.discoverShows(filterOptions);
    } else {
      // Default to popular shows if no search or filters
      observable = this.tvShowService.getPopular(filterOptions);
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
        this.showsLength = resp['results'].length;
        this.totalShows = resp['total_result'] || resp['total_results'];

        resp['results'].forEach((show) => {
          let result: TvShowResult[] = [{
            adult: show['adult'],
            backdrop_path: show['backdrop_path'],
            genre_ids: show['genre_ids'],
            id: show['id'],
            name: show['name'],
            first_air_date: show['first_air_date'],
            original_language: show['original_language'],
            original_name: show['original_name'],
            overview: show['overview'],
            popularity: show['popularity'],
            poster_path: this.baseUrl + show['poster_path'],
            vote_average: show['vote_average'],
            vote_count: show['vote_count'],
            video: show['video']
          }];

          this.fetchedShows.push({
            page: resp['page'],
            results: result,
            total_pages: resp['total_pages'],
            total_result: resp['total_result'] || resp['total_results']
          });

          if (this.pages) {
            this.pages.push(resp['page']);
          }

          if (show['first_air_date'] && this.releaseYear) {
            this.releaseYear.push(show['first_air_date']);
          }
        });

        // Simple assignment for lists
        this.allShows = [...this.fetchedShows];
        this.filteredShowsList = this.fetchedShows;

        // Update paginator if available
        if (this.paginator) {
          this.paginator.pageIndex = this.page - 1;
          this.paginator.length = this.totalShows;
        }

        // Log status for debugging
        console.log('Shows processed:', this.fetchedShows.length);
        console.log('Total shows:', this.totalShows);
        console.log('Current page:', this.page);

        // End loading state
        this.isLoading = false;

        // Scroll to top
        window.scrollTo({top: 0, behavior: 'auto'});
      },
      error: (error) => {
        console.error('Error fetching shows:', error);
        this.isLoading = false;
      }
    });
  }
}
