import { CommonModule } from '@angular/common';
import { Component, inject, ViewChild } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule } from 'ng-gallery';
import {TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { TvShowService } from '../tv-service.service';
import { SearchShowListComponent } from '../search-show-list/search-show-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginator } from '@angular/material/paginator';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';


// TODO: FIX THE BOYS NOT BEING ABLE TO BE DOWNLOADED.

@Component({
  selector: 'app-search-movies',
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule, SearchShowListComponent, FormsModule, MatPaginatorModule],
  providers: [TvShowService, NgModel],

  template: `
  <div class="container">
    <section class="header-section">
      <form class="search-form" (ngSubmit)="getGenre(1)">
        <input type="text" [(ngModel)]="showSearch" name="showSearch" id="showSearch" placeholder="Find Show by Name" #filter />
        <button class="button big-btn filter-button" type="submit">Filter</button>
      </form>

      <div class="results" *ngIf="this.filteredShowsList.length != 0">
        <div class="show-item" *ngFor="let showItem of this.filteredShowsList; index as i;">
          <div class="show-info">
            <app-search-show-list
              [tvShow]="showItem" [page]="showItem.page" [query]="filter.value">
            </app-search-show-list>
          </div>
        </div>
      </div>
    </section>

    <footer class="paginator-container">
      <mat-paginator 
        [length]="this.totalShows"
        [pageSize]="this.showsLength"
        aria-label="Select page"
        (page)="onPageChange($event)">
      </mat-paginator>
    </footer>
  </div>
  `,
  styleUrls: ['./search-shows.component.sass', '../../styles.sass']
})
export class SearchShowsComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;

  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public fetchedShows: TvShow[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public filteredShowsList: TvShow[] = [];
  public showNames = [{}];
  public allShows: TvShow[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public showsLength: number;
  public totalShows: number;
  public releaseYear: string[] = [""];
  public endYear: string[] = [""];
  public showSearch: string;
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public pages: number[] = [0]
  public page: number = 1;

  public tvShowService: TvShowService = inject(TvShowService)

  constructor() {
  }

  ngOnInit() {

  }

  onPageChange(event?: PageEvent) {
    if (event === null) {

    } else {
      this.page = event!.pageIndex + 1;
      this.getGenre(this.page);
    }
  }

  async getGenre(page: number) {
    while (this.fetchedShows.length > 0) {
      this.fetchedShows.pop()
    }

    // this.galleryMoviesRef.reset()

    while (this.showNames.length > 0) {
      this.showNames.pop()
    }

    this.tvShowService.getAllShows(page, this.showSearch).subscribe((resp) => {
      console.log(resp['results']);

      this.showsLength = resp['results'].length;
      this.totalShows = resp['total_results'];

      resp['results'].forEach((show) => {
        let page = resp['page'];
        let isAdult = show['adult'];
        let backdropPath = show['backdrop_path'];
        let genreIds = show['genre_ids'];
        let id = show['id'];
        let firstAirDate = show['first_air'];
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

        this.fetchedShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });

        this.pages.push(page)
        this.releaseYear.push(firstAirDate)
        this.genre = genreIds[0]
      })

      this.allShows.splice(0, 1);
    })

    if (this.filteredShowsList.length > 0) {
      while (this.filteredShowsList.length > 0) {
        this.filteredShowsList.pop();
      }
      this.filteredShowsList = this.fetchedShows;
    } else {
      this.filteredShowsList = this.fetchedShows;
    }
    //return this.filteredMovieList = this.fetchedMovies.filter((movie) => movie.results[0]?.title.toLowerCase().includes(this.movieSearch.toLowerCase()));
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredShowsList = this.allShows;
    }
    let page = 1;
    this.tvShowService.getInitialPage(text).subscribe((resp) => {

      this.showsLength = resp['results'].length;
      this.totalShows = resp['total_pages'];
      console.log(resp['total_results']);
      console.log(this.totalShows);
      for (let i = 0; i < this.totalShows; i++) {
        console.log(this.totalShows);



        this.tvShowService.getAllShows(page, text).subscribe((resp) => {
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

            this.pages.push(page)
          })

        })

        this.allShows.splice(0, 1);
        this.pages.splice(0, 1);
        this.page++;
      }
    })

    return this.filteredShowsList = this.allShows.filter((show) => show.results[0]?.name.toLowerCase().includes(text.toLowerCase()));
  }
}
