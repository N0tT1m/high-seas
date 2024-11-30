import { CommonModule } from '@angular/common';
import { Component, inject, ViewChild } from '@angular/core';
import { RouterModule } from '@angular/router';
import { GalleryModule, Gallery, GalleryRef } from 'ng-gallery';
import { Movie, MovieResult, Genre } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';
import { PopularMoviesListComponent } from '../popular-movies-list/popular-movies-list.component';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatPaginator } from '@angular/material/paginator';

@Component({
  selector: 'app-popular-movies',
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule, PopularMoviesListComponent, FormsModule, MatPaginatorModule],
  providers: [MovieService, TvShowService, NgModel],
  template: `
    <div class="container">
      <section class="header-section">
        <div class="results" *ngIf="this.allMovies.length != 0">
          <div class="show-item" *ngFor="let movieItem of this.allMovies; index as i;">
            <div class="show-info">
              <app-popular-movie-list
                [movie]="movieItem" [page]="movieItem.page">
              </app-popular-movie-list>
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
  styleUrls: ['./popular-movies.component.sass', '../../styles.sass'],
})
export class PopularMoviesComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  @ViewChild(MatPaginator) paginator: MatPaginator;

  public title: Movie;
  public movieTitles = [{}];
  public allMovies: Movie[] = [];
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public airDate: string;
  public pageSize: number = 20;
  public totalShows: number;
  public currentPage: number = 1;

  public movieService: MovieService = inject(MovieService);

  constructor() {}

  ngOnInit() {
    this.getMovies(this.currentPage);
  }

  getMovies(page: number) {
    this.movieService.getPopular(page).subscribe((resp) => {
      this.allMovies = resp['results'].map((movie) => ({
        page: resp['page'],
        results: [{
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
          video: movie['video'],
          in_plex: movie['in_plex']
        }],
        total_pages: resp['total_pages'],
        total_result: resp['total_results'],
      }));

      this.totalShows = resp['total_results'];
    });
  }

  onPageChange(event: PageEvent) {
    this.currentPage = event.pageIndex + 1;
    this.getMovies(this.currentPage);
  }
}
