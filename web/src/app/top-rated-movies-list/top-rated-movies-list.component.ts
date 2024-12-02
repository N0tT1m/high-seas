import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router'
import { Movie } from '../http-service/http-service.component';

@Component({
  selector: 'app-top-rated-movie-list',
  standalone: true,
  imports: [RouterModule, NgFor],
  template: `
  <div>
    <section class="listing" *ngFor="let movie of movie?.results; index as i;">
      <a [routerLink]="['/top-rated/movies/details', movie.id, this.page]">
        <div class="show-image">
          <img [src]="movie.poster_path" alt="Show Poster" class="poster-image">
        </div>
      </a>
      <a [routerLink]="['/top-rated/movies/details', movie.id, this.page]">
        <div>
            <h2 class="show-name">{{ movie.title }}</h2>
        </div>
      </a>
      <p class="show-overview">{{ movie.overview}}</p>
      <p class="show-vote-average">The vote average for this show is: {{movie.vote_average }} </p>
    </section>
  </div>
  `,
  styleUrls: ['./top-rated-movies-list.component.sass', '../../styles.sass']
})
export class TopRatedMoviesListComponent {

  @Input() movie!: Movie;
  @Input() page!: number;
  @Input() query!: string;

  constructor() {
  }

}
