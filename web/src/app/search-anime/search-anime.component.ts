import { Component, inject, ViewChild, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { TvShowService, TvShowFilterOptions } from '../tv-service.service';
import { SearchAnimeListComponent } from '../search-anime-list/search-anime-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginator } from '@angular/material/paginator';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatSliderModule } from '@angular/material/slider';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';

@Component({
  selector: 'app-search-anime',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    SearchAnimeListComponent,
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
    <section class="header-section">
      <form class="search-form" (ngSubmit)="getAnime(1)">
        <input type="text" [(ngModel)]="animeSearch" name="animeSearch" id="animeSearch" placeholder="Find Anime by Name" #filter />

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

              <div class="filter-group">
                <label for="year">First Air Year:</label>
                <input type="number" [(ngModel)]="yearFilter" name="year" id="year" min="1900" max="2099">
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
                <label for="animeType">Anime Type:</label>
                <mat-select [(ngModel)]="animeType" name="animeType" id="animeType">
                  <mat-option value="">Any</mat-option>
                  <mat-option value="tv">TV Series</mat-option>
                  <mat-option value="movie">Movie</mat-option>
                  <mat-option value="ova">OVA</mat-option>
                  <mat-option value="special">Special</mat-option>
                </mat-select>
              </div>
            </div>

            <div class="filter-row">
              <div class="filter-group">
                <label for="status">Status:</label>
                <mat-select [(ngModel)]="status" name="status" id="status">
                  <mat-option value="">Any</mat-option>
                  <mat-option value="returning series">Currently Airing</mat-option>
                  <mat-option value="ended">Finished Airing</mat-option>
                  <mat-option value="canceled">Canceled</mat-option>
                </mat-select>
              </div>

              <div class="filter-group">
                <label for="seasonCount">Season Count:</label>
                <mat-select [(ngModel)]="seasonCount" name="seasonCount" id="seasonCount">
                  <mat-option value="">Any</mat-option>
                  <mat-option value="1">1 Season</mat-option>
                  <mat-option value="2-3">2-3 Seasons</mat-option>
                  <mat-option value="4-6">4-6 Seasons</mat-option>
                  <mat-option value="7+">7+ Seasons</mat-option>
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

      <div class="results" *ngIf="this.filteredAnimeList.length != 0">
        <div class="anime-item" *ngFor="let animeItem of this.filteredAnimeList; index as i;">
          <div class="anime-info">
            <app-search-anime-list
              [tvShow]="animeItem" [page]="animeItem.page" [query]="filter.value">
            </app-search-anime-list>
          </div>
        </div>
      </div>
    </section>

    <footer class="paginator-container">
      <mat-paginator
        [length]="this.totalAnime"
        [pageSize]="this.animeLength"
        aria-label="Select page"
        (page)="onPageChange($event)">
      </mat-paginator>
    </footer>
  </div>
  `,
  styleUrls: ['./search-anime.component.sass']
})
export class SearchAnimeComponent implements OnInit {
  @ViewChild(MatPaginator) paginator: MatPaginator;
  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  public fetchedAnime: TvShow[] = [{
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

  public filteredAnimeList: TvShow[] = [];
  public animeNames = [{}];
  public allAnime: TvShow[] = [{
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

  public animeLength: number;
  public totalAnime: number;
  public releaseYear: string[] = [""];
  public endYear: string[] = [""];
  public animeSearch: string = '';
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
  public includeAdult: boolean = false;
  public status: string = '';
  public animeType: string = '';
  public seasonCount: string = '';

  public tvShowService: TvShowService = inject(TvShowService);

  constructor() {
    this.tvShowService.getGenres().subscribe((resp) => {
      if (resp && resp.genres) {
        this.genreDetails = resp.genres;
      }
    });
  }

  ngOnInit() {
    // Initialize with popular anime
    this.getAnime(1);
  }

  onPageChange(event?: PageEvent) {
    if (event === null) {
      // Handle null event
    } else {
      this.page = event!.pageIndex + 1;
      this.getAnime(this.page);
    }
  }

  async getAnime(page: number) {
    // Clear current anime
    while (this.fetchedAnime.length > 0) {
      this.fetchedAnime.pop();
    }

    while (this.animeNames.length > 0) {
      this.animeNames.pop();
    }

    // Build filter options
    const filterOptions: TvShowFilterOptions = {
      page: page,
      language: 'ja', // Set Japanese language for anime
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

    if (this.minRating) {
      filterOptions.minRating = this.minRating;
    }

    if (this.maxRating) {
      filterOptions.maxRating = this.maxRating;
    }

    if (this.status) {
      filterOptions.status = this.status;
    }

    if (this.animeType) {
      filterOptions.withType = this.animeType;
    }

    // Handle season count filter
    if (this.seasonCount) {
      filterOptions.seasonCount = {};

      switch (this.seasonCount) {
        case '1':
          filterOptions.seasonCount.min = 1;
          filterOptions.seasonCount.max = 1;
          break;
        case '2-3':
          filterOptions.seasonCount.min = 2;
          filterOptions.seasonCount.max = 3;
          break;
        case '4-6':
          filterOptions.seasonCount.min = 4;
          filterOptions.seasonCount.max = 6;
          break;
        case '7+':
          filterOptions.seasonCount.min = 7;
          break;
      }
    }

    // Choose which API to call based on search query
    let observable;
    if (this.animeSearch) {
      observable = this.tvShowService.searchShows(this.animeSearch, filterOptions);
    } else if (Object.keys(filterOptions).length > 3) {
      // If we have filters beyond page, language, and includeAdult, use discover
      observable = this.tvShowService.discoverShows(filterOptions);
    } else {
      // Default to popular shows with Japanese language filter
      observable = this.tvShowService.getPopular(filterOptions);
    }

    observable.subscribe((resp) => {
      this.animeLength = resp.results.length;
      this.totalAnime = resp.total_result;

      resp.results.forEach((anime) => {
        // Only include Japanese content
        if (anime.original_language === 'ja') {
          let result: TvShowResult[] = [{
            adult: anime.adult,
            backdrop_path: anime.backdrop_path,
            genre_ids: anime.genre_ids,
            id: anime.id,
            name: anime.name,
            first_air_date: anime.first_air_date,
            original_language: anime.original_language,
            original_name: anime.original_name,
            overview: anime.overview,
            popularity: anime.popularity,
            poster_path: this.baseUrl + anime.poster_path,
            vote_average: anime.vote_average,
            vote_count: anime.vote_count,
            video: anime.video
          }];

          this.fetchedAnime.push({
            page: resp.page,
            results: result,
            total_pages: resp.total_pages,
            total_result: resp.total_result
          });

          this.pages.push(resp.page);
          this.releaseYear.push(anime.first_air_date);
        }
      });

      this.allAnime = [...this.fetchedAnime];
      this.filteredAnimeList = this.fetchedAnime;
    });
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredAnimeList = this.allAnime;
    }

    return this.filteredAnimeList = this.allAnime.filter((anime) =>
      anime.results[0]?.name.toLowerCase().includes(text.toLowerCase()) ||
      anime.results[0]?.original_name.toLowerCase().includes(text.toLowerCase())
    );
  }
}
