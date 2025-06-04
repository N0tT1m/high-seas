import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router'
import { TvShow } from '../http-service/http-service.component';

@Component({
  selector: 'app-search-anime-list',
  standalone: true,
  imports: [RouterModule, NgFor],
  template: `
  <div>
    <section class="listing" *ngFor="let anime of tvShow?.results; index as i;">
      <a [routerLink]="['/search-anime', anime.id, this.page, this.query]">
        <div class="anime-image">
          <img [src]="anime.poster_path" alt="Anime Poster" class="poster-image">
        </div>
      </a>
      <a [routerLink]="['/search-anime', anime.id, this.page, this.query]">
        <div>
            <h2 class="anime-name">{{ anime.name }}</h2>
            <h3 class="anime-original-name" *ngIf="anime.original_name !== anime.name">{{ anime.original_name }}</h3>
        </div>
      </a>
      <p class="anime-overview">{{ anime.overview}}</p>
      <div class="anime-meta">
        <p class="anime-vote-average">Rating: {{anime.vote_average }}/10</p>
        <p class="anime-first-air-date" *ngIf="anime.first_air_date">First aired: {{ anime.first_air_date }}</p>
      </div>
    </section>
  </div>
  `,
  styleUrls: ['./search-anime-list.component.sass']
})
export class SearchAnimeListComponent {
  @Input() tvShow!: TvShow;
  @Input() page!: number;
  @Input() query!: string;

  constructor() {
  }
}
