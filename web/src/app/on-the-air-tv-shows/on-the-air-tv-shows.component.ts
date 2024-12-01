import { CommonModule } from '@angular/common';
import { Component, inject, ViewChild } from '@angular/core';
import { RouterModule } from '@angular/router';
import { GalleryModule, Gallery, GalleryRef } from 'ng-gallery';
import { TvShow, TvShowResult, Genre } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatPaginator } from '@angular/material/paginator';
import { OnTheAirTvShowsListComponent } from '../on-the-air-tv-shows-list/on-the-air-tv-shows-list.component';

@Component({
  selector: 'app-on-the-air-tv-shows',
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule, OnTheAirTvShowsListComponent, FormsModule, MatPaginatorModule],
  providers: [MovieService, TvShowService, NgModel],
  template: `
    <div class="container">
      <section class="header-section">
        <div class="results" *ngIf="this.allShows.length != 0">
          <div class="show-item" *ngFor="let showItem of this.allShows; index as i;">
            <div class="show-info">
              <app-on-the-air-tv-shows-list
                [tvShow]="showItem" [page]="showItem.page">
              </app-on-the-air-tv-shows-list>
            </div>
          </div>
        </div>

        <footer class="paginator-container">
          <mat-paginator
            [length]="this.totalShows"
            [pageSize]="this.pageSize"
            (page)="onPageChange($event)">
          </mat-paginator>
        </footer>
      </section>
    </div>
  `,
  styleUrls: ['./on-the-air-tv-shows.component.sass', '../../styles.sass'],
})
export class OnTheAirTvShowsComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;

  public show: TvShow;
  public showNames = [{}];
  public allShows: TvShow[] = [];
  public galleryShowsRef: GalleryRef;
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public airDate: string;
  public pageSize: number = 20;
  public totalShows: number;
  public currentPage: number = 1;

  public tvShowService: TvShowService = inject(TvShowService);

  constructor() {}

  ngOnInit() {
    this.getShows(this.currentPage);
  }

  getShows(page: number) {
    this.tvShowService.getOnTheAir(page).subscribe((resp) => {
      this.allShows = resp['results'].map((show) => ({
        page: resp['page'],
        results: [{
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
          video: show['video'],
        }],
        total_pages: resp['total_pages'],
        total_result: resp['total_results'],
      }));

      this.totalShows = resp['total_results'];
    });
  }

  onPageChange(event: PageEvent) {
    this.currentPage = event.pageIndex + 1;
    this.getShows(this.currentPage);
  }
}
