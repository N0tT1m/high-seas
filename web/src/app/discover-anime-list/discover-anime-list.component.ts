import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { TvShow } from '../http-service/http-service.component';

@Component({
  selector: 'app-discover-anime-list',
  standalone: true,
  imports: [RouterModule, NgFor, CommonModule],
  template: `
  <div>
    <section class="listing" *ngFor="let anime of tvShow?.results; index as i;">
      <a [routerLink]="['/discover-anime', anime.id, this.page, this.releaseYear, this.endYear, this.genre]">
        <div class="anime-image">
          <img *ngIf="anime!.poster_path" [src]="anime.poster_path" alt="Anime Poster" class="poster-image">
        </div>
      </a>
      <a [routerLink]="['/discover-anime', anime.id, this.page, this.releaseYear, this.endYear, this.genre]">
        <div>
          <h2 class="anime-name">{{ anime.name }}</h2>
          <h3 class="anime-original-name" *ngIf="anime.original_name !== anime.name">{{ anime.original_name }}</h3>
        </div>
      </a>
      <div class="anime-meta">
        <p class="anime-air-date" *ngIf="anime.first_air_date">{{ anime.first_air_date | date:'yyyy' }}</p>
        <p class="anime-rating">Rating: {{ anime.vote_average }}/10</p>
      </div>
      <p class="anime-overview">{{ anime.overview }}</p>
    </section>
  </div>
  `,
  styleUrls: ['./discover-anime-list.component.sass']
})
export class DiscoverAnimeListComponent {
  @Input() tvShow!: TvShow;
  @Input() genre!: number;
  @Input() releaseYear!: string;
  @Input() endYear!: string;
  @Input() page!: number;

  constructor() {
  }
}
