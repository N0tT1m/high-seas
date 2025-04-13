import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { Movie } from '../http-service/http-service.component';
import { DiscoverMovieDetailsComponent } from '../discover-movie-details/discover-movie-details.component';

@Component({
  selector: 'app-discover-movie-list',
  standalone: true,
  imports: [RouterModule, NgFor, CommonModule, DiscoverMovieDetailsComponent],
  template: `
    <div>
      <section class="row listing" *ngFor="let movie of movieItem!.results; index as i;">
        <a [routerLink]="['/discover-movie', movie.id, this.page, this.releaseYear, this.endYear, this.genre]">
          <div class="movie-image">
            <img *ngIf="movie!.poster_path" [src]="movie.poster_path" alt="Movie Poster" class="poster-image">
          </div>
        </a>
        <a [routerLink]="['/discover-movie', movie.id, this.page, this.releaseYear, this.endYear, this.genre]">
          <div>
            <h2 class="column movie-name">{{ movie.title }}</h2>
          </div>
        </a>
        <p class="column movie-overview">{{ movie.overview}}</p>
        <p class="column movie-vote-average">The vote average for this movie is: {{movie.vote_average }} </p>
      </section>
    </div>
  `,
  styleUrls: ['./discover-movie-list.component.sass', '../../styles.sass']
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
