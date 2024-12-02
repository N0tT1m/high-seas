import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router'
import { TvShow } from '../http-service/http-service.component';

@Component({
  selector: 'app-discover-show-list',
  standalone: true,
  imports: [RouterModule, NgFor],
  template: `
  <div>
    <section class="listing" *ngFor="let show of tvShow?.results; index as i;">
      <a [routerLink]="['/discover-show', show.id, this.page, this.airDate, this.genre]">
        <div class="show-image">
          <img [src]="show.poster_path" alt="Movie Poster" class="poster-image">
        </div>
      </a>
      <a [routerLink]="['/discover-show', show.id, this.page, this.airDate, this.genre]">
        <div>
            <h2 class="show-name">{{ show.name }}</h2>
        </div>
      </a>
      <p class="show-overview">{{ show.overview}}</p>
      <p class="show-vote-average">The vote average for this show is: {{show.vote_average }} </p>
    </section>
  </div>
  `,
  styleUrls: ['./show-list.component.sass', '../../styles.sass']
})
export class ShowListComponent {

  @Input() tvShow!: TvShow;
  @Input() genre!: number;
  @Input() airDate!: string;
  @Input() page!: number;

  constructor() {
  }

}
