import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router'
import { Movie } from '../http-service/http-service.component';
import { DiscoverMovieDetailsComponent } from '../discover-movie-details/discover-movie-details.component';

@Component({
  selector: 'app-movie-list',
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
      <a [routerLink]="['/discover-movie', movie.id, this.page, this.releaseYear, this.endYear, this.genre]">Link to {{ movie.title }}</a>
    </section>
  </div>
  `,
  styleUrls: ['./discover-movie-list.component.sass']
})
export class MovieListComponent {

  @Input() movieItem!: Movie;
  @Input() genre!: number;
  @Input() releaseYear!: string;
  @Input() endYear!: string;
  @Input() page!: number;

  constructor() {
  }
}
