import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { Movie } from '../http-service/http-service.component';

@Component({
  selector: 'app-search-movie-list',
  standalone: true,
  imports: [RouterModule, NgFor, CommonModule],
  template: `
    <div>
      <section class="row listing" *ngFor="let movie of movieItem!.results; index as i;">
        <a [routerLink]="['/search-movie', movie.id, this.page, this.query]">
          <div class="movie-image">
            <img *ngIf="movie!.poster_path" [src]="movie.poster_path" alt="Movie Poster" class="poster-image">
          </div>
        </a>
        <a [routerLink]="['/search-movie', movie.id, this.page, this.query]">
          <div>
            <h2 class="column movie-name">{{ movie.title }}</h2>
          </div>
        </a>
        <p class="column movie-overview">{{ movie.overview}}</p>
        <p class="column movie-vote-average">The vote average for this movie is: {{movie.vote_average }} </p>
      </section>
    </div>
  `,
  styleUrls: ['./search-movie-list.component.sass', '../../styles.sass']
})
export class SearchMovieListComponent {
  @Input() movieItem!: Movie;
  @Input() page!: number;
  @Input() query!: string;

  constructor() {
  }
}
