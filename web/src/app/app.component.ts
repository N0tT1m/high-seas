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

@Component({
  selector: 'app-root',
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
    <div class="background-image"></div>
    <main class="earth-spirit">
      <header class="brand-name">
        <div class="container-fluid">
          <nav class="dota-navbar">
            <div class="navbar-content">
              <a [routerLink]="['/']">
                <img id="big-titty-waifu-svg" src="/assets/home-icon.jpg" height="50" width="50" alt="logo" aria-hidden="true">
              </a>
              <div class="menu-wrapper">
                <div class="left-menu">
                  <div class="dropdown">
                    <a class="dropdown-toggle">Now Playing / Popular / Top Rated / Upcoming Movies</a>
                    <div class="dropdown-menu">
                      <a [routerLink]="['now-playing/movies']">Now Playing Movies</a>
                      <a [routerLink]="['popular/movies']">Popular Movies</a>
                      <a [routerLink]="['top-rated/movies']">Top Rated Movies</a>
                      <a [routerLink]="['upcoming/movies']">Upcoming Movies</a>
                    </div>
                  </div>
                  <div class="dropdown">
                    <a class="dropdown-toggle">Discover & Search Movies</a>
                    <div class="dropdown-menu">
                      <a [routerLink]="['discover/movies']">Discover Movies</a>
                      <a [routerLink]="['search/movies']">Search Movies</a>
                    </div>
                  </div>
                  <div class="dropdown">
                    <a class="dropdown-toggle">Airing Today / Popular / Top Rated / On The Air Shows</a>
                    <div class="dropdown-menu">
                      <a [routerLink]="['airing-today/shows']">Airing Today Shows</a>
                      <a [routerLink]="['popular/shows']">Popular Shows</a>
                      <a [routerLink]="['top-rated/shows']">Top Rated Shows</a>
                      <a [routerLink]="['on-the-air/shows']">On The Air Shows</a>
                    </div>
                  </div>
                  <div class="dropdown">
                    <a class="dropdown-toggle">Discover & Search Shows</a>
                    <div class="dropdown-menu">
                      <a [routerLink]="['discover/shows']">Discover Shows</a>
                      <a [routerLink]="['search/shows']">Search Shows</a>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </nav>
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
