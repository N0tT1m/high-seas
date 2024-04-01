import { CommonModule } from '@angular/common';
import { Component, inject, ViewChild } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule, Gallery, GalleryRef } from 'ng-gallery';
import {TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';
import { ShowListComponent } from '../discover-show-list/show-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatPaginator } from '@angular/material/paginator';

@Component({
  selector: 'app-discover-shows',
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule, ShowListComponent, FormsModule, MatPaginatorModule],
  providers: [MovieService, TvShowService, NgModel],
  template: `
  <div class="container">
    <section class="header-section">
      <div id="filters">
        <label for="filters" id='filter-label'>Search for Shows:</label>
        <form class='filters-form' (ngSubmit)="getGenre(1)">
          <div class="filters-div">
            <label for="genre">Genre:</label>
            <select [(ngModel)]="genre" name="genres" id="genres" (ngModelChange)="getGenre(1)" class='select-section'>
              <option id="genre" *ngFor="let genre of genreDetails" value="{{genre.id}}">{{genre.name}}</option>
            </select>
          </div>

          <div class="filters-div">
            <label for="releaseYear">AirDate:</label>
            <input type='text' [(ngModel)]="airDate" name="airDate" id="airDate" class='select-section' />
          </div>

          <button class="button big-btn filter-button" type="submit">Search</button>
        </form>
      </div>

      <label for="filters" id='filter-label'>Filter Shows:</label>
      <form class="search-form">
        <input type="text" placeholder="Filter Show by Name" #filter>
        <button class="big-btn filter-button" type="button" (click)="filterResults(filter.value)">Filter</button>
      </form>
    </section>

    <div class="results" *ngIf="genre != 0">
      <div class="movie-item" *ngFor="let showItem of filteredShowsList">
        <div class="movie-info">
          <app-show-list
            [tvShow]="showItem" [page]="this.page" [airDate]="this.airDate" [genre]="this.genre">
          </app-show-list>
        </div>
      </div>
    </div>

    <footer>
      <mat-paginator [length]=this.totalShows
                [pageSize]=this.showsLength
                aria-label="Select page"
                (page)="onPageChange($event)">
      </mat-paginator>
    </footer>
  </div>
  `,
  styleUrls: ['./discover-shows.component.sass', '../../styles.sass'],
})
export class DiscoverShowsComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;

  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public show: TvShow;
  public showNames = [{}];
  public fetchedTvShows: TvShow[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public allShows: TvShow[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  filteredShowsList: TvShow[] = [];
  public galleryShowsRef: GalleryRef;
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public airDate: string;
  public showsLength: number;
  public totalShows: number;
  public page: number = 1;

  public tvShowsService: TvShowService = inject(TvShowService)

  constructor(private gallery: Gallery) {
    this.tvShowsService.getGenres().subscribe((resp) => {
      resp['genres'].forEach((genre) => {
        var item = {id: genre.id, name: genre.name}

        this.genreDetails.push(item)
      })
    })
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredShowsList = this.fetchedTvShows;
    }

    return this.filteredShowsList = this.fetchedTvShows.filter((show) => show.results[0]?.name.toLowerCase().includes(text.toLowerCase()));
  }

  ngOnInit() {
    // Get the galleryRef by id
    this.galleryShowsRef = this.gallery.ref('showGallery');

    this.galleryShowsRef.play()
  }

  onPageChange(event?: PageEvent) {
    if (event === null) {

    } else {
      this.page = event!.pageIndex + 1;
      this.getGenre(this.page);
    }
  }

  getMoviesFromDate() {
    while (this.page <= this.totalShows) {
      this.tvShowsService.getAllShowsFromSelectedDate(this.genre, this.airDate, this.page).subscribe((resp) => {
        console.log(resp['results']);

        this.showsLength = resp['results'].length;
        this.totalShows = resp['total_pages'];

        resp['results'].forEach((show) => {
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
          let totalResult = resp['total_result'];

          let result: TvShowResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, name: name, first_air_date: firstAirDate, original_language: originalLanguage, original_name: originalName, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

          this.allShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });
        })
      })

      this.page++;
    }

    this.allShows.splice(0, 1);
  }

  async getGenre(page: number) {
    while (this.fetchedTvShows.length > 0) {
      this.fetchedTvShows.pop()
    }

    while (this.showNames.length > 0) {
      this.showNames.pop()
    }

    this.tvShowsService.getAllTvShows(this.genre, this.airDate, page).subscribe((resp) => {
      this.showsLength = resp['results'].length;
      this.totalShows = resp['total_results'];

      resp['results'].forEach((show) => {
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
        let totalResult = resp['total_result'];

        let result: TvShowResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, name: name, first_air_date: firstAirDate, original_language: originalLanguage, original_name: originalName, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

        this.fetchedTvShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });
      })
    })

    this.fetchedTvShows.splice(0, 1);

    if (this.filteredShowsList.length > 0) {
      while (this.filteredShowsList.length > 0) {
        this.filteredShowsList.pop();
      }
      this.filteredShowsList = this.fetchedTvShows;
    } else {
      this.filteredShowsList = this.fetchedTvShows;
    }
  }
}
