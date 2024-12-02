import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router'
import { TvShow } from '../http-service/http-service.component';

@Component({
  selector: 'app-on-the-air-tv-shows-list',
  standalone: true,
  imports: [RouterModule, NgFor],
  template: `
  <div>
    <section class="listing" *ngFor="let show of tvShow?.results; index as i;">
      <a [routerLink]="['/on-the-air/shows/details', show.id, this.page]">
        <div class="show-image">
          <img [src]="show.poster_path" alt="Show Poster" class="poster-image">
        </div>
      </a>
      <a [routerLink]="['/on-the-air/shows/details', show.id, this.page]">
        <div>
            <h2 class="show-name">{{ show.name }}</h2>
        </div>
      </a>
      <p class="show-overview">{{ show.overview}}</p>
      <p class="show-vote-average">The vote average for this show is: {{show.vote_average }} </p>
    </section>
  </div>
  `,
  styleUrls: ['./on-the-air-tv-shows-list.component.sass', '../../styles.sass']
})
export class OnTheAirTvShowsListComponent {

  @Input() tvShow!: TvShow;
  @Input() page!: number;
  @Input() query!: string;

  constructor() {
  }

}
