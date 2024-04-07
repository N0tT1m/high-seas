import { CommonModule } from '@angular/common';
import { Component, inject, ViewChild } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule, Gallery, GalleryRef } from 'ng-gallery';
import {TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';
import { AiringTodayTvShowsListComponent } from '../airing-today-tv-shows-list/airing-today-tv-shows-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatPaginator } from '@angular/material/paginator';

@Component({
  selector: 'app-discover-shows',
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule, AiringTodayTvShowsListComponent, FormsModule, MatPaginatorModule],
  providers: [MovieService, TvShowService, NgModel],
  template: `
  <div class="container">
    <section class="header-section">
    <div class="results" *ngIf="this.allShows.length != 0">
      <div class="show-item" *ngFor="let showItem of this.allShows; index as i;">
        <div class="show-info">
          <app-airing-today-tv-shows-list
            [tvShow]="showItem" [page]="showItem.page">
          </app-airing-today-tv-shows-list>
        </div>
      </div>

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
  styleUrls: ['./airing-today-tv-shows.component.sass', '../../styles.sass'],
})
export class AiringTodayTvShowsComponent {
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
  public pages: number[] = [0];

  public tvShowService: TvShowService = inject(TvShowService)

  constructor() {
    this.tvShowService.getInitialAiringTodayPage().subscribe((resp) => {
      this.showsLength = resp['results'].length;
      this.totalShows = resp['total_pages'];
    })
  }

  filterResults(text: string) {
    if (!text) {
      return this.filteredShowsList = this.fetchedTvShows;
    }

    return this.filteredShowsList = this.fetchedTvShows.filter((show) => show.results[0]?.name.toLowerCase().includes(text.toLowerCase()));
  }

  ngOnInit() {
    this.tvShowService.getInitialAiringTodayPage().subscribe((resp) => {
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

        this.allShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });



        this.pages.push(this.page)
      })

      this.allShows.splice(0, 1);
    })

    for (var p = 0; p < this.allShows.length; p++) {
      for (var j = 0; j < this.allShows[p].results.length; j++) {
        if (this.showNames.includes(this.allShows[p].results[j].name)) {
          continue
        } else {
          // this.galleryTvShowsRef.addImage({ src: this.allShows[p].results[j].poster_path, thumb: this.allShows[p].results[j].poster_path })
        }
      }
    }

    for (var p = 0; p < this.allShows.length; p++) {
      for (var j = 0; j < this.allShows[p].results.length; j++) {
        if (this.showNames.includes(this.allShows[p].results[j].name)) {
          continue
        } else {
          this.showNames.push({ 'name': this.allShows[p].results[j].name, 'id': this.allShows[p].results[j].id, 'page': this.allShows[p].page })
        }
      }
    }
  }

  getNextPage(page: number) {
    this.tvShowService.getAiringToday(page).subscribe((resp) => {
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

        this.allShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });



        this.pages.push(this.page)
      })

      this.allShows.splice(0, 1);
    })

    for (var p = 0; p < this.allShows.length; p++) {
      for (var j = 0; j < this.allShows[p].results.length; j++) {
        if (this.showNames.includes(this.allShows[p].results[j].name)) {
          continue
        } else {
          // this.galleryTvShowsRef.addImage({ src: this.allShows[p].results[j].poster_path, thumb: this.allShows[p].results[j].poster_path })
        }
      }
    }

    for (var p = 0; p < this.allShows.length; p++) {
      for (var j = 0; j < this.allShows[p].results.length; j++) {
        if (this.showNames.includes(this.allShows[p].results[j].name)) {
          continue
        } else {
          this.showNames.push({ 'name': this.allShows[p].results[j].name, 'id': this.allShows[p].results[j].id, 'page': this.allShows[p].page })
        }
      }
    }
  }

  onPageChange(event?: PageEvent) {
    if (event === null) {

    } else {
      this.page = event!.pageIndex + 1;
      this.getNextPage(this.page);
    }
  }
}
