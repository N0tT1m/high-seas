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
  templateUrl: './app.component.html',
})
export class AppComponent {
  public movies$: Observable<MovieResult[]>;
  public moviesDropdownOpen = false;
  public searchMoviesDropdownOpen = false;
  public showsDropdownOpen = false;
  public searchShowsDropdownOpen = false;

  constructor(private movieService: MovieService) {}

  title = 'High Seas';

  ngOnInit() {}

  toggleDropdown(event: Event, dropdown: string) {
    event.preventDefault();
    event.stopPropagation();

    switch (dropdown) {
      case 'movies':
        this.moviesDropdownOpen = !this.moviesDropdownOpen;
        break;
      case 'search-movies':
        this.searchMoviesDropdownOpen = !this.searchMoviesDropdownOpen;
        break;
      case 'shows':
        this.showsDropdownOpen = !this.showsDropdownOpen;
        break;
      case 'search-shows':
        this.searchShowsDropdownOpen = !this.searchShowsDropdownOpen;
        break;
    }
  }
}
