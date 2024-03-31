import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router'
import { Movie } from '../http-service/http-service.component';
import { DiscoverMovieDetailsComponent } from '../discover-movie-details/discover-movie-details.component';

@Component({
  selector: 'app-search-movie-list',
  standalone: true,
  imports: [RouterModule, NgFor, DiscoverMovieDetailsComponent],
  template: `
  <div>
    <section class="row listing" *ngFor="let movie of movieItem!.results; index as i;">
      <div class="movie-image">
        <img [src]="movie.poster_path" alt="Movie Poster" class="poster-image">
      </div>
      <h2 class="column movie-name">Movie title: {{ movie.title }}</h2>
      <p class="column movie-overview">{{ movie.overview}}</p>
      <p class="column movie-vote-average">The vote average for this movie is: {{movie.vote_average }} </p>
      <a [routerLink]="['/search-movie', movie.id, this.page, this.query]">Link to {{ movie.title }}</a>
    </section>
  </div>
  `,
  styleUrls: ['./search-movie-list.component.sass']
})
export class SearchMovieListComponent {

  @Input() movieItem!: Movie;
  @Input() page!: number;
  @Input() query!: string;

  constructor() {
  }
}
