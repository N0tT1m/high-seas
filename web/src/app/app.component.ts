import { Component, forwardRef, OnInit } from '@angular/core';
import { HomeComponent } from './home/home.component';
import { RouterLink, RouterOutlet } from '@angular/router';
import { MovieService } from './movies.service';
import { MovieResult } from './http-service/http-service.component';
import { Observable } from 'rxjs';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import {NgModule} from '@angular/core';
import { FormsModule } from '@angular/forms';
import {MatNativeDateModule} from '@angular/material/core';
import { HighSeasMaterialModule } from './material-module';
import { MAT_FORM_FIELD_DEFAULT_OPTIONS } from '@angular/material/form-field';
import { MatPaginator, MatPaginatorModule } from '@angular/material/paginator';

// /assets/es-cartoon.jpg

@Component({
  selector: 'app-root',
  // templateUrl: './app.component.html',
  standalone: true,
  imports: [
    HomeComponent,
    RouterLink,
    RouterOutlet,
    HttpClientModule,
    FormsModule,
    HttpClientModule,
    HighSeasMaterialModule,
    MatPaginatorModule,
    MatNativeDateModule,
  ],
  providers: [{ provide: MAT_FORM_FIELD_DEFAULT_OPTIONS, useValue: { appearance: 'fill' } }, MovieService, HttpClientModule,],
  styleUrls: ['./app.component.sass', '../styles.sass'],
  template: `
    <main class="earth-spirit">
      <header class="brand-name">
        <div class="container-fluid">
          <a [routerLink]="['/']">
            <img id="the-spirits" src="/assets/home-icon.jpg" height=150px width=150px alt="logo" aria-hidden="true">
          </a>
          <button mat-button [matMenuTriggerFor]="movies" class="navbar-links big-btn">Movies</button>
          <button mat-button [matMenuTriggerFor]="shows" class="navbar-links big-btn">Shows</button>

          <mat-menu #movies="matMenu">
            <button mat-menu-item [matMenuTriggerFor]="newupcomingtopratedpopularmovies">Now Playing / Popular / Top Rated / Upcoming Movies</button>
            <button mat-menu-item [matMenuTriggerFor]="discoversearchmovies">Discover & Search Movies</button>
          </mat-menu>

          <mat-menu #shows="matMenu">
            <button mat-menu-item [matMenuTriggerFor]="newupcomingtopratedpopularshows">Now Playing / Popular / Top Rated / Upcoming Shows</button>
            <button mat-menu-item [matMenuTriggerFor]="discoversearchshows">Discover & Search Shows</button>
          </mat-menu>

          <mat-menu #newupcomingtopratedpopularmovies="matMenu">
            <button mat-menu-item> <a [routerLink]="['now-playing/movies']">Now Playing Movies</a></button>
            <button mat-menu-item><a [routerLink]="['popular/movies']">Popular Movies</a></button>
            <button mat-menu-item><a [routerLink]="['top-rated/movies']">Top Rated Movies</a></button>
            <button mat-menu-item><a [routerLink]="['upcoming/movies']">Upcoming Movies</a></button>
          </mat-menu>

          <mat-menu #discoversearchmovies="matMenu">
            <button mat-menu-item><a [routerLink]="['discover/movies']">Discover Movies</a></button>
            <button mat-menu-item><a [routerLink]="['search/movies']">Search Movies</a></button>
          </mat-menu>

          <mat-menu #newupcomingtopratedpopularshows="matMenu">
            <button mat-menu-item><a [routerLink]="['airing-today/shows']">Airing Today Shows</a></button>
            <button mat-menu-item><a [routerLink]="['popular/shows']">Popular Shows</a></button>
            <button mat-menu-item><a [routerLink]="['top-rated/shows']">Top Rated Shows</a></button>
            <button mat-menu-item><a [routerLink]="['on-the-air/shows']">On The Air Shows</a></button>
          </mat-menu>

          <mat-menu #discoversearchshows="matMenu">
            <button mat-menu-item><a [routerLink]="['discover/shows']">Discover Shows</a></button>
            <button mat-menu-item><a [routerLink]="['search/shows']">Search Shows</a></button>
          </mat-menu>
        </div>
      </header>
      <section class="content the-girls">
        <router-outlet></router-outlet>
      </section>
    </main>
  `
})

export class AppComponent {
  public movies$: Observable<MovieResult[]>;

  constructor(private movieService: MovieService) {

  }

  title = 'High Seas';

  ngOnInit() {

  }
}
