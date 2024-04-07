import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router'
import { Movie } from '../http-service/http-service.component';

@Component({
  selector: 'app-popular-movie-list',
  standalone: true,
  imports: [RouterModule, NgFor],
  template: `
  <div>
    <section class="listing" *ngFor="let movie of movie?.results; index as i;">
      <div class="show-image">
        <img [src]="movie.poster_path" alt="Show Poster" class="poster-image">
      </div>
      <h2 class="show-name">Show name: {{ movie.title }}</h2>
      <p class="show-overview">{{ movie.overview}}</p>
      <p class="show-vote-average">The vote average for this show is: {{movie.vote_average }} </p>
      <a [routerLink]="['/search-show', movie.id, this.page, this.query]">Link to {{ movie.title }}</a>
    </section>
  </div>
  `,
  styleUrls: ['./popular-movies-list.component.sass', '../../styles.sass']
})
export class PopularMoviesListComponent {

  @Input() movie!: Movie;
  @Input() page!: number;
  @Input() query!: string;

  constructor() {
  }

}
