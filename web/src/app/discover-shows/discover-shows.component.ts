import { CommonModule } from '@angular/common';
import { Component, inject, ViewChild, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule, Gallery, GalleryRef } from 'ng-gallery';
import { TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { TvShowService, TvShowFilterOptions } from '../tv-service.service';
import { ShowListComponent } from '../discover-show-list/show-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatPaginator } from '@angular/material/paginator';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatSliderModule } from '@angular/material/slider';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';

@Component({
  selector: 'app-discover-shows',
  standalone: true,
  imports: [
    GalleryModule,
    CommonModule,
    RouterModule,
    ShowListComponent,
    FormsModule,
    MatPaginatorModule,
    MatExpansionModule,
    MatSliderModule,
    MatSelectModule,
    MatCheckboxModule
  ],
  providers: [MovieService, TvShowService, NgModel],
  template: `
  <div class="container">
    <section class="header-section">
      <div id="filters">
        <label for="filters" id='filter-label'>Search for Shows:</label>
        <form class='filters-form' (ngSubmit)="getGenre(1)">
          <div class="filters-div">
            <label for="genre">Genre:</label>
            <select [(ngModel)]="genre" name="genres" id="genres" class='select-section'>
              <option id="genre" *ngFor="let genre of genreDetails" value="{{genre.id}}">{{genre.name}}</option>
            </select>
          </div>

          <div class="filters-div">
            <label for="releaseYear">First Air Date:</label>
            <input type='text' [(ngModel)]="airDate" name="airDate" id="airDate" class='select-section' placeholder="YYYY-MM-DD" />
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
                  <label for="yearStart">First Air Date Range (Start):</label>
                  <input type="number" [(ngModel)]="yearRangeStart" name="yearRangeStart" id="yearRangeStart" min="1900" max="2099" class='select-section'>
                </div>

                <div class="filter-group">
                  <label for="yearEnd">First Air Date Range (End):</label>
                  <input type="number" [(ngModel)]="yearRangeEnd" name="yearRangeEnd" id="yearRangeEnd" min="1900" max="2099" class='select-section'>
                </div>
              </div>

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
                    <option value="first_air_date.desc">First Air Date (Newest)</option>
                    <option value="first_air_date.asc">First Air Date (Oldest)</option>
                    <option value="name.asc">Name (A-Z)</option>
                  </select>
                </div>

                <div class="filter-group">
                  <label for="language">Original Language:</label>
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
                  <label for="status">Status:</label>
                  <select [(ngModel)]="status" name="status" id="status" class='select-section'>
                    <option value="">Any</option>
                    <option value="returning series">Returning Series</option>
                    <option value="ended">Ended</option>
                    <option value="canceled">Canceled</option>
                  </select>
                </div>

                <div class="filter-group">
                  <label for="networkFilter">Network:</label>
                  <select [(ngModel)]="networkFilter" name="networkFilter" id="networkFilter" class='select-section'>
                    <option value="">Any</option>
                    <option value="213">Netflix</option>
                    <option value="1024">Amazon Prime</option>
                    <option value="2552">Apple TV+</option>
                    <option value="49">HBO</option>
                    <option value="2739">Disney+</option>
                    <option value="67">Showtime</option>
                    <option value="4">BBC</option>
                  </select>
                </div>
              </div>

              <div class="filter-row">
                <div class="filter-group">
                  <label for="seasonCount">Season Count:</label>
                  <select [(ngModel)]="seasonCount" name="seasonCount" id="seasonCount" class='select-section'>
                    <option value="">Any</option>
                    <option value="1">1 Season</option>
                    <option value="2-3">2-3 Seasons</option>
                    <option value="4-6">4-6 Seasons</option>
                    <option value="7+">7+ Seasons</option>
                  </select>
                </div>

                <div class="filter-group">
                  <mat-checkbox [(ngModel)]="includeAdult" name="includeAdult" id="includeAdult">Include Adult Content</mat-checkbox>
                </div>
              </div>
            </div>
          </mat-expansion-panel>

          <button class="button big-btn filter-button" type="submit">Search</button>
        </form>
      </div>

      <label for="filters" id='filter-label'>Filter Results:</label>
      <form class="search-form">
        <input type="text" placeholder="Filter Show by Name" #filter>
        <button class="big-btn filter-button" type="button" (click)="filterResults(filter.value)">Filter</button>
      </form>
    </section>

    <div class="results" *ngIf="filteredShowsList.length > 0">
      <div class="movie-item" *ngFor="let showItem of filteredShowsList">
        <div class="movie-info">
          <app-discover-show-list
            [tvShow]="showItem" [page]="this.page" [airDate]="this.airDate" [genre]="this.genre">
          </app-discover-show-list>
        </div>
      </div>
    </div>

    <footer class="paginator-container">
      <!-- Explicit binding for MatPaginator -->
      <mat-paginator
        [length]="totalShows"
        [pageSize]="showsLength"
        [pageIndex]="page - 1"
        aria-label="Select page"
        (page)="onPageChange($event)">
      </mat-paginator>
    </footer>
  </div>
  `,
  styleUrls: ['./discover-shows.component.sass', '../../styles.sass'],
})
export class DiscoverShowsComponent implements OnInit {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;
  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public show: TvShow;
  public showNames = [{}];
  public fetchedTvShows: TvShow[] = [{
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

  filteredShowsList: TvShow[] = [];
  public galleryShowsRef: GalleryRef;
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public airDate: string = '';
  public showsLength: number;
  public totalShows: number;
  public page: number = 1;

  // Advanced search filters
  public yearRangeStart: number | null = null;
  public yearRangeEnd: number | null = null;
  public minRating: number | null = null;
  public maxRating: number | null = null;
  public sortBy: string = 'popularity.desc';
  public language: string = '';
  public includeAdult: boolean = false;
  public status: string = '';
  public networkFilter: string = '';
  public seasonCount: string = '';

  public tvShowsService: TvShowService = inject(TvShowService);

  constructor(private gallery: Gallery) {
    this.tvShowsService.getGenres().subscribe((resp) => {
      if (resp && resp.genres) {
        this.genreDetails = [{ id: 0, name: "None" }];
        resp.genres.forEach((genre) => {
          this.genreDetails.push(genre);
        });
      } else if (resp['genres']) {
        resp['genres'].forEach((genre) => {
          var item = {id: genre.id, name: genre.name};
          this.genreDetails.push(item);
        });
      }
    });
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredShowsList = this.fetchedTvShows;
    }

    return this.filteredShowsList = this.fetchedTvShows.filter((show) =>
      show.results[0]?.name.toLowerCase().includes(text.toLowerCase())
    );
  }

  ngOnInit() {
    // Get the galleryRef by id
    this.galleryShowsRef = this.gallery.ref('showGallery');
    this.galleryShowsRef.play();
  }

  getMoviesFromDate() {
    while (this.page <= this.totalShows) {
      this.tvShowsService.getAllShowsFromSelectedDate(this.genre, this.airDate, this.page).subscribe((resp) => {
        this.showsLength = resp.results.length;
        this.totalShows = resp.total_pages;

        resp.results.forEach((show) => {
          let result: TvShowResult[] = [{
            adult: show.adult,
            backdrop_path: show.backdrop_path,
            genre_ids: show.genre_ids,
            id: show.id,
            name: show.name,
            first_air_date: show.first_air_date,
            original_language: show.original_language,
            original_name: show.original_name,
            overview: show.overview,
            popularity: show.popularity,
            poster_path: this.baseUrl + show.poster_path,
            vote_average: show.vote_average,
            vote_count: show.vote_count,
            video: show.video
          }];

          this.allShows.push({
            page: resp.page,
            results: result,
            total_pages: resp.total_pages,
            total_result: resp.total_result
          });
        });
      });

      this.page++;
    }

    this.allShows.splice(0, 1);
  }

  // Replace your onPageChange and getGenre methods with these versions

  // Update these methods in your component

  /**
   * Handle page change events from the paginator
   */
  onPageChange(event: PageEvent) {
    // Update the page number (adding 1 since paginator is zero-indexed)
    this.page = event.pageIndex + 1;

    // Scroll to top of the page
    window.scrollTo({ top: 0, behavior: 'smooth' });

    // Call getGenre with the new page number
    this.getGenre(this.page);
  }

  /**
   * Get TV shows for the selected genre and air date at the specified page
   * @param page Page number to retrieve
   */
  async getGenre(page: number) {
    console.log('getGenre called with page:', page);
    console.log('Current genre:', this.genre);
    console.log('Current airDate:', this.airDate);

    // Clear existing shows
    this.fetchedTvShows = [];
    this.showNames = [];

    // Set loading state if you have one
    // this.isLoading = true;

    // Process year input if needed
    let processedAirDate = this.airDate;

    // Check if input is just a year (4 digits)
    if (this.airDate && /^\d{4}$/.test(this.airDate.trim())) {
      const year = this.airDate.trim();
      processedAirDate = year;
      console.log('Searching for year:', year);
    }

    // Use the original working method that fetches shows
    this.tvShowsService.getAllTvShows(this.genre, processedAirDate, page).subscribe({
      next: (resp) => {
        console.log('API response:', resp);

        // Extract response data using bracket notation (as in your working example)
        if (!resp || !resp['results']) {
          console.error('Invalid response format:', resp);
          return;
        }

        this.showsLength = resp['results'].length;
        this.totalShows = resp['total_results'] || resp['total_result'];

        // Process each show in the results
        resp['results'].forEach((show) => {
          // Extract show details using bracket notation
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
          let totalResult = resp['total_results'] || resp['total_result'];

          // Create a result object
          let result: TvShowResult[] = [{
            adult: isAdult,
            backdrop_path: backdropPath,
            genre_ids: genreIds,
            id: id,
            name: name,
            first_air_date: firstAirDate,
            original_language: originalLanguage,
            original_name: originalName,
            overview: overview,
            popularity: popularity,
            poster_path: posterPath,
            vote_average: voteAverage,
            vote_count: voteCount,
            video: video
          }];

          // Add to fetchedTvShows array
          this.fetchedTvShows.push({
            page: page,
            results: result,
            total_pages: totalPages,
            total_result: totalResult
          });
        });

        // Update the filtered list displayed to the user
        this.filteredShowsList = this.fetchedTvShows;

        // End loading state if you have one
        // this.isLoading = false;

        // Scroll to top of page
        window.scrollTo({ top: 0, behavior: 'auto' });

        console.log('Shows processed:', this.fetchedTvShows.length);
      },
      error: (error) => {
        console.error('Error fetching shows:', error);
        // this.isLoading = false;
      }
    });
  }

  // 3. Add or modify ngAfterViewInit to ensure paginator is properly initialized
  ngAfterViewInit() {
    // Initialize paginator with current values after view is ready
    if (this.paginator) {
      this.paginator.pageIndex = this.page - 1;
      this.paginator.length = this.totalShows;
      this.paginator.pageSize = this.showsLength;
    }
  }

  protected readonly Math = Math;
}
