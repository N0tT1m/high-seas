import { Component } from '@angular/core';
import {RouterLink} from "@angular/router";
import {DropdownComponent} from "../app-dropdown/app-dropdown.component";

@Component({
  selector: 'app-navbar',
  templateUrl: './navbar.component.html',
  standalone: true,
  imports: [
    RouterLink,
    DropdownComponent
  ],
  styleUrls: ['./navbar.component.css']
})
export class NavbarComponent {
  moviesDropdownOpen = false;
  searchMoviesDropdownOpen = false;
  showsDropdownOpen = false;
  searchShowsDropdownOpen = false;

  toggleDropdown(event: Event, dropdownName: string) {
    event.stopPropagation();
    switch (dropdownName) {
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

  moviesLinks = [
    { label: 'Now Playing Movies', route: 'now-playing/movies' },
    { label: 'Popular Movies', route: 'popular/movies' },
    { label: 'Top Rated Movies', route: 'top-rated/movies' },
    { label: 'Upcoming Movies', route: 'upcoming/movies' },
    { label: 'Now Playing Movies Gallery', route: 'now-playing/movies/gallery' },
    { label: 'Popular Movies Gallery', route: 'popular/movies/gallery' },
    { label: 'Top Rated Movies Gallery', route: 'top-rated/movies/gallery' },
    { label: 'Upcoming Movies Gallery', route: 'upcoming/movies/gallery' }
  ];

  searchMoviesLinks = [
    { label: 'Discover Movies', route: 'discover/movies' },
    { label: 'Search Movies', route: 'search/movies' }
  ];

  showsLinks = [
    { label: 'Airing Today Shows', route: 'airing-today/shows' },
    { label: 'Popular Shows', route: 'popular/shows' },
    { label: 'Top Rated Shows', route: 'top-rated/shows' },
    { label: 'On The Air Shows', route: 'on-the-air/shows' },
    { label: 'Airing Today Shows Gallery', route: 'airing-today/shows/gallery' },
    { label: 'Popular Shows Gallery', route: 'popular/shows/gallery' },
    { label: 'Top Rated Shows Gallery', route: 'top-rated/shows/gallery' },
    { label: 'On The Air Shows Gallery', route: 'on-the-air/shows/gallery' }
  ];

  searchShowsLinks = [
    { label: 'Discover Shows', route: 'discover/shows' },
    { label: 'Search Shows', route: 'search/shows' }
  ];
}
