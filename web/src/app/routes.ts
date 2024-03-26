import { Routes } from '@angular/router';
import { HomeComponent } from './home/home.component';
import { DiscoverMovieDetailsComponent } from './discover-movie-details/discover-movie-details.component';
import { TvShowDetailsComponent } from './discover-tv-show-details/discover-tv-show-details.component';
import { DiscoverShowsComponent } from './discover-shows/discover-shows.component';
import { DiscoverMoviesComponent } from './discover-movies/discover-movies.component';
import { SearchMoviesComponent } from './search-movies/search-movies.component';
import { SearchShowsComponent } from './search-shows/search-shows.component';
import { SearchMovieDetailsComponent } from './search-movie-details/search-movie-details.component';
import { SearchTvShowDetailsComponent } from './search-shows-details/search-show-details.component';
import { NowPlayingMoviesComponent } from './now-playing-movies/now-playing-movies.component';
import { PopularMoviesComponent } from './popular-movies/popular-movies.component';
import { TopRatedMoviesComponent } from './top-rated-movies/top-rated-movies.component';
import { UpcomingMoviesComponent } from './upcoming-movies/upcoming-movies.component';
import { NowPlayingMoviesDetails } from './now-playing-movies-details/now-playing-movies-details.component';
import { PopularMoviesDetailsComponent } from './popular-movies-details/popular-movies-details.component';
import { TopRatedMoviesDetailsComponent } from './top-rated-movies-details/top-rated-movies-details.component';
import { UpcomingMoviesDetailsComponent } from './upcoming-movies-details/upcoming-movies-details.component';
import { AiringTodayTvShowsComponent } from './airing-today-tv-shows/airing-today-tv-shows.component';
import { AiringTodayTvShowsDetailsComponent } from './airing-today-tv-shows-details/airing-today-tv-shows-details.component';
import { PopularTvShowsComponent } from './popular-tv-shows/popular-tv-shows.component';
import { PopularTvShowsDetailsComponent } from './popular-tv-shows-details/popular-tv-shows-details.component';
import { TopRatedTvShowsComponent } from './top-rated-tv-shows/top-rated-tv-shows.component';
import { TopRatedTvShowsDetailsComponent } from './top-rated-tv-shows-details/top-rated-tv-shows-details.component';
import { OnTheAirTvShowsComponent } from './on-the-air-tv-shows/on-the-air-tv-shows.component';
import { OnTheAirTvShowsDetailsComponent } from './on-the-air-tv-shows-details/on-the-air-tv-shows-details.component';

const routeConfig: Routes = [
  {
    path: '',
    component: HomeComponent,
    title: 'Home page'
  },
  {
    path: 'discover-movie/:id/:page/:releaseYear/:endYear/:genre',
    component: DiscoverMovieDetailsComponent,
    title: 'Discover Movie details'
  },
  {
    path: 'discover-show/:id/:page/:airDate/:genre',
    component: TvShowDetailsComponent,
    title: 'Discover Show details'
  },
  {
    path: 'discover/shows',
    component: DiscoverShowsComponent,
    title: 'Discover Shows'
  },
  {
    path: 'discover/movies',
    component: DiscoverMoviesComponent,
    title: 'Discover Movies'
  },
  {
    path: 'search-movie/:id/:page/:query',
    component: SearchMovieDetailsComponent,
    title: 'Search Movie details'
  },
  {
    path: 'search-show/:id/:page/:query',
    component: SearchTvShowDetailsComponent,
    title: 'Search Show details'
  },
  {
    path: 'search/shows',
    component: SearchShowsComponent,
    title: 'Search Shows'
  },
  {
    path: 'search/movies',
    component: SearchMoviesComponent,
    title: 'Search Movies'
  },
  {
    path: 'now-playing/movies',
    component: NowPlayingMoviesComponent,
    title: 'Now Playing Movies'
  },
  {
    path: 'popular/movies',
    component: PopularMoviesComponent,
    title: 'Popular Movies'
  },
  {
    path: 'top-rated/movies',
    component: TopRatedMoviesComponent,
    title: 'Top Rated Movies'
  },
  {
    path: 'upcoming/movies',
    component: UpcomingMoviesComponent,
    title: 'Upcoming Movies'
  },
  {
    path: 'now-playing/movies/details/:id/:page',
    component: NowPlayingMoviesDetails,
    title: 'Now Playing Movies Details'
  },
  {
    path: 'popular/movies/details/:id/:page',
    component: PopularMoviesDetailsComponent,
    title: 'Popular Movies Details'
  },
  {
    path: 'top-rated/movies/details/:id/:page',
    component: TopRatedMoviesDetailsComponent,
    title: 'Top Rated Movies Details'
  },
  {
    path: 'upcoming/movies/details/:id/:page',
    component: UpcomingMoviesDetailsComponent,
    title: 'Upcoming Movies Details'
  },
  {
    path: 'airing-today/shows',
    component: AiringTodayTvShowsComponent,
    title: 'Airing Today Tv Shows'
  },
  {
    path: 'popular/shows',
    component: PopularTvShowsComponent,
    title: 'Popular Tv Shows'
  },
  {
    path: 'top-rated/shows',
    component: TopRatedTvShowsComponent,
    title: 'Top Rated Tv Shows'
  },
  {
    path: 'on-the-air/shows',
    component: OnTheAirTvShowsComponent,
    title: 'On The Error Tv Shows'
  },
  {
    path: 'airing-today/shows/details/:id/:page',
    component: AiringTodayTvShowsDetailsComponent,
    title: 'Airing Today Tv Shows Details'
  },
  {
    path: 'popular/shows/details/:id/:page',
    component: PopularTvShowsDetailsComponent,
    title: 'Popular Tv Shows Details'
  },
  {
    path: 'top-rated/shows/details/:id/:page',
    component: TopRatedTvShowsDetailsComponent,
    title: 'Top Rated Tv Shows Details'
  },
  {
    path: 'on-the-air/shows/details/:id/:page',
    component: OnTheAirTvShowsDetailsComponent,
    title: 'Upcoming Tv Shows Details'
  },
];

export default routeConfig;


/*
Copyright Google LLC. All Rights Reserved.
Use of this source code is governed by an MIT-style license that
can be found in the LICENSE file at https://angular.io/license
*/
