import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Movie, MovieResult, Genre } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { SearchMovieListComponent } from '../search-movie-list/search-movie-list.component';
import { GalleryModule, Gallery, GalleryRef, ImageItem } from 'ng-gallery';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-popular',
  standalone: true,
  imports: [CommonModule, GalleryModule, RouterModule],
  providers: [MovieService],
  template: `
  <div class="container">
    <div class="gallery-wrapper">
      <gallery class="gallery" id="moviesGallery"></gallery>
    </div>
    <div class="movie-title-wrapper">
      <h3 class="title" *ngFor="let movie of movieTitles">
        <a [routerLink]="['/popular/movies/details', movie['id'], movie['page']]">{{ movie['title'] }}</a>
      </h3>
    </div>
  </div>
  `,
  styleUrls: ['./popular-movies-gallery.component.sass']
})
export class PopularGalleryMoviesComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  public fetchedMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0 }]
  public filteredMoviesList: Movie[] = [];
  public allMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0 }]
  public galleryMoviesRef: GalleryRef;
  public moviesLength: number;
  public totalMovies: number;
  public releaseYear: string[] = [""];
  public endYear: string[] = [""];
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public pages: number[] = [0];
  public movieTitles = [{}];

  public movieService: MovieService = inject(MovieService)

  constructor(private gallery: Gallery) {
    this.movieService.getInitialPopularPage().subscribe((resp) => {
      this.moviesLength = resp['results'].length;
      this.totalMovies = resp['total_pages'];
      console.log(resp['total_results']);
    })
  }

  ngOnInit() {
    // Get the galleryRef by id
    this.galleryMoviesRef = this.gallery.ref('moviesGallery');

    let page = 1;

    this.movieService.getPopularMovies(page).subscribe((resp) => {
      resp['results'].forEach((movie) => {
        let page = resp['page'];
        let isAdult = movie['adult'];
        let backdropPath = movie['backdrop_path'];
        let genreIds = movie['genre_ids'];
        let id = movie['id'];
        let releaseDate = movie['release_date'];
        let video = movie['video'];
        let title = movie['title'];
        let originalLanguage = movie['original_language'];
        let originalTitle = movie['original_title'];
        let overview = movie['overview'];
        let popularity = movie['popularity'];
        let posterPath = this.baseUrl + movie['poster_path'];
        let voteAverage = movie['vote_average'];
        let voteCount = movie['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];

        let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video}]

        this.allMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });

        this.pages.push(page)
        console.log('LENB', this.allMovies);
      })

      this.allMovies.splice(0, 1);
    })

    page++;

    this.movieService.getPopularMovies(page).subscribe((resp) => {
      resp['results'].forEach((movie) => {
        let page = resp['page'];
        let isAdult = movie['adult'];
        let backdropPath = movie['backdrop_path'];
        let genreIds = movie['genre_ids'];
        let id = movie['id'];
        let releaseDate = movie['release_date'];
        let video = movie['video'];
        let title = movie['title'];
        let originalLanguage = movie['original_language'];
        let originalTitle = movie['original_title'];
        let overview = movie['overview'];
        let popularity = movie['popularity'];
        let posterPath = this.baseUrl + movie['poster_path'];
        let voteAverage = movie['vote_average'];
        let voteCount = movie['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];
        let in_plex = resp['in_plex'];

        let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video}]

        this.allMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });

        this.pages.push(page)
        console.log('LENB', this.allMovies);
      })
    })

    page++;

    this.movieService.getPopularMovies(page).subscribe((resp) => {
      resp['results'].forEach((movie) => {
        let page = resp['page'];
        let isAdult = movie['adult'];
        let backdropPath = movie['backdrop_path'];
        let genreIds = movie['genre_ids'];
        let id = movie['id'];
        let releaseDate = movie['release_date'];
        let video = movie['video'];
        let title = movie['title'];
        let originalLanguage = movie['original_language'];
        let originalTitle = movie['original_title'];
        let overview = movie['overview'];
        let popularity = movie['popularity'];
        let posterPath = this.baseUrl + movie['poster_path'];
        let voteAverage = movie['vote_average'];
        let voteCount = movie['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];
        let in_plex = resp['in_plex'];

        let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video}]

        this.allMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });

        this.pages.push(page)
        console.log('LENB', this.allMovies);
      })

      this.allMovies.splice(0, 1);

      for (var p = 0; p < this.allMovies.length; p++) {
        for (var j = 0; j < this.allMovies[p].results.length; j++) {
          if (this.movieTitles.includes(this.allMovies[p].results[j].title)) {
            continue
          } else {
            this.galleryMoviesRef.addImage({ src: this.allMovies[p].results[j].poster_path, thumb: this.allMovies[p].results[j].poster_path })
          }
        }
      }

      for (var p = 0; p < this.allMovies.length; p++) {
        for (var j = 0; j < this.allMovies[p].results.length; j++) {
          if (this.movieTitles.includes(this.allMovies[p].results[j].title)) {
            continue
          } else {
            this.movieTitles.push({ 'title': this.allMovies[p].results[j].title, 'id': this.allMovies[p].results[j].id, 'page': this.allMovies[p].page })
          }
        }
      }
    })

    this.galleryMoviesRef.play()
  }
}
