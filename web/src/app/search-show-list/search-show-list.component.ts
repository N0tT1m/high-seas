import { NgFor } from '@angular/common';
import { Component, Input } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { TvShow } from '../http-service/http-service.component';

@Component({
  selector: 'app-search-show-list',
  standalone: true,
  imports: [RouterModule, NgFor, CommonModule],
  template: `
    <div>
      <section class="listing" *ngFor="let show of tvShow?.results; index as i;">
        <a [routerLink]="['/search-show', show.id, this.page, this.query]">
          <div class="show-image">
            <img *ngIf="show!.poster_path" [src]="show.poster_path" alt="Show Poster" class="poster-image">
          </div>
        </a>
        <a [routerLink]="['/search-show', show.id, this.page, this.query]">
          <div>
            <h2 class="show-name">{{ show.name }}</h2>
          </div>
        </a>
        <p class="show-overview">{{ show.overview}}</p>
        <p class="show-vote-average">The vote average for this show is: {{show.vote_average }} </p>
      </section>
    </div>
  `,
  styleUrls: ['./search-show-list.component.sass', '../../styles.sass']
})
export class SearchShowListComponent {
  @Input() tvShow!: TvShow;
  @Input() page!: number;
  @Input() query!: string;

  constructor() {
  }
}
