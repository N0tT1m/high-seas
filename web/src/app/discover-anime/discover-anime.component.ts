import { CommonModule, NgFor } from '@angular/common';
import { Component, OnInit, ViewChild, inject } from '@angular/core';
import { RouterModule } from '@angular/router';
import { GalleryModule, Gallery, GalleryRef } from 'ng-gallery';
import { TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { TvShowService, TvShowFilterOptions } from '../tv-service.service';
import { DiscoverAnimeListComponent } from '../discover-anime-list/discover-anime-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatPaginator } from '@angular/material/paginator';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatSliderModule } from '@angular/material/slider';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';

@Component({
  selector: 'app-discover-anime',
  standalone: true,
  imports: [
    GalleryModule,
    CommonModule,
    RouterModule,
    DiscoverAnimeListComponent,
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
    <div id="filters">
      <label for="filters" id='filter-label'>Search for Anime:</label>

      <form class='filters-form' (ngSubmit)="getAnime(1)">
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
                  <option value="first_air_date.desc">Release Date (Newest)</option>
                  <option value="first_air_date.asc">Release Date (Oldest)</option>
                  <option value="name.asc">Title (A-Z)</option>
                </select>
              </div>

              <div class="filter-group">
                <label for="animeType">Anime Type:</label>
                <select [(ngModel)]="animeType" name="animeType" id="animeType" class='select-section'>
                  <option value="">Any</option>
                  <option value="tv">TV Series</option>
                  <option value="movie">Movie</option>
                  <option value="ova">OVA</option>
                  <option value="special">Special</option>
                </select>
              </div>
            </div>

            <div class="filter-row">
              <div class="filter-group">
                <label for="status">Status:</label>
                <select [(ngModel)]="status" name="status" id="status" class='select-section'>
                  <option value="">Any</option>
                  <option value="returning series">Currently Airing</option>
                  <option value="ended">Finished Airing</option>
                  <option value="canceled">Canceled</option>
                </select>
              </div>

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

    <label for="filters" id='filter-label'>Filter Results:</label>
    <section class="header-section">
      <form class="search-form">
        <input type="text" placeholder="Filter Anime by Title" #filter>
        <button class="big-btn filter-button" type="button" (click)="filterResults(filter.value)">Filter</button>
      </form>
    </section>

    <div class="results" *ngIf="filteredAnimeList.length > 0">
      <div class="anime-item" *ngFor="let animeItem of filteredAnimeList">
        <div class="anime-info">
          <app-discover-anime-list
            [tvShow]="animeItem"
            [page]="this.page"
            [releaseYear]="this.releaseYear"
            [endYear]="this.endYear"
            [genre]="this.genre">
          </app-discover-anime-list>
        </div>
      </div>
    </div>

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
  styleUrls: ['./discover-anime.component.sass']
})
export class DiscoverAnimeComponent implements OnInit {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;
  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public animeNames = [{}];
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

  public filteredAnimeList: TvShow[] = [];
  public galleryAnimeRef: GalleryRef;
  public genreDetails: Genre[] = [{ id: 0, name: "All Genres" }];
  public genre: number = 0;
  public releaseYear: string = '';
  public endYear: string = '';
  public animeLength: number;
  public totalAnime: number;
  public page: number = 1;

  // Advanced search filters
  public minRating: number | null = null;
  public maxRating: number | null = null;
  public sortBy: string = 'popularity.desc';
  public includeAdult: boolean = false;
  public status: string = '';
  public animeType: string = '';
  public seasonCount: string = '';

  public tvShowService: TvShowService = inject(TvShowService);

  constructor(private gallery: Gallery) {
    this.tvShowService.getGenres().subscribe((resp) => {
      if (resp && resp.genres) {
        this.genreDetails = [{ id: 0, name: "All Genres" }];
        resp.genres.forEach((genre) => {
          this.genreDetails.push(genre);
        });
      } else if (resp['genres']) {
        // Legacy response format handling
        resp['genres'].forEach((genre) => {
          var item = {id: genre.id, name: genre.name};
          this.genreDetails.push(item);
        });
      }
    });
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredAnimeList = this.fetchedAnime;
    }

    return this.filteredAnimeList = this.fetchedAnime.filter((anime) =>
      anime.results[0]?.name.toLowerCase().includes(text.toLowerCase()) ||
      anime.results[0]?.original_name.toLowerCase().includes(text.toLowerCase())
    );
  }

  ngOnInit() {
    // Get the galleryRef by id
    this.galleryAnimeRef = this.gallery.ref('animeGallery');
    this.galleryAnimeRef.play();

    // Load initial anime list
    this.getAnime(1);
  }

  onPageChange(event?: PageEvent) {
    if (event === null) {
      // Handle null event case
    } else {
      this.page = event!.pageIndex + 1;
      this.getAnime(this.page);
    }
  }

  async getAnime(page: number) {
    // Clear current list
    while (this.fetchedAnime.length > 0) {
      this.fetchedAnime.pop();
    }

    // Clear anime names
    while (this.animeNames.length > 0) {
      this.animeNames.pop();
    }

    // Build filter object for advanced filtering
    const animeFilters: TvShowFilterOptions = {
      language: 'ja',  // Set Japanese language for anime
      page: page,
      includeAdult: this.includeAdult,
      sortBy: this.sortBy
    };

    // Add genre if selected (not 0)
    if (this.genre !== 0) {
      animeFilters.genres = [this.genre];
    }

    // Add year range if specified
    if (this.releaseYear || this.endYear) {
      animeFilters.yearRange = {};
      if (this.releaseYear) {
        animeFilters.yearRange.start = parseInt(this.releaseYear);
      }
      if (this.endYear) {
        animeFilters.yearRange.end = parseInt(this.endYear);
      }
    }

    // Add rating filters if specified
    if (this.minRating) {
      animeFilters.minRating = this.minRating;
    }
    if (this.maxRating) {
      animeFilters.maxRating = this.maxRating;
    }

    // Add status filter if specified
    if (this.status) {
      animeFilters.status = this.status;
    }

    // Add anime type filter if specified
    if (this.animeType) {
      animeFilters.withType = this.animeType;
    }

    // Add season count filter if specified
    if (this.seasonCount) {
      animeFilters.seasonCount = {};

      switch (this.seasonCount) {
        case '1':
          animeFilters.seasonCount.min = 1;
          animeFilters.seasonCount.max = 1;
          break;
        case '2-3':
          animeFilters.seasonCount.min = 2;
          animeFilters.seasonCount.max = 3;
          break;
        case '4-6':
          animeFilters.seasonCount.min = 4;
          animeFilters.seasonCount.max = 6;
          break;
        case '7+':
          animeFilters.seasonCount.min = 7;
          break;
      }
    }

    // Use the discover API with the filters
    this.tvShowService.discoverShows(animeFilters).subscribe((resp) => {
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
        }
      });

      // Update filtered list
      this.allAnime = [...this.fetchedAnime];
      this.filteredAnimeList = this.fetchedAnime;
    });
  }
}
