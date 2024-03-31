import { Component, inject, ViewChild, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Movie, MovieResult, Genre } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { SearchMovieListComponent } from '../search-movie-list/search-movie-list.component';
import { MatPaginator } from '@angular/material/paginator';
import { NgModel, FormsModule } from '@angular/forms';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';

@Component({
  selector: 'app-search-movies',
  standalone: true,
  imports: [CommonModule, SearchMovieListComponent, MatPaginatorModule, FormsModule],
  providers: [NgModel],
  // <button class="big-btn filter-button" type="button" (click)="filterResults(filter.value)">Search</button>
  template: `
  <!-- SearchMoviesComponent -->
  <div class="container">
  <section class="header-section">
    <form class="search-form" (ngSubmit)="getGenre(1)">
      <input type="text" [(ngModel)]="movieSearch" name="movieSearch" id="movieSearch" placeholder="Find Movie by Title" #filter />
      <button class="button big-btn filter-button" type="submit">Filter</button>
    </form>

    <div class="results" *ngIf="this.filteredMovieList.length != 0">
      <div class="movie-item" *ngFor="let movieItem of this.filteredMovieList; index as i;">
        <div class="movie-info">
          <app-search-movie-list
            [movieItem]="movieItem" [page]="movieItem.page" [query]="filter.value">
          </app-search-movie-list>
        </div>
      </div>
    </div>
  </section>

  <footer>
    <mat-paginator [length]=this.totalMovies
              [pageSize]=this.moviesLength
              aria-label="Select page"
              (page)="onPageChange($event)">
    </mat-paginator>
  </footer>
  </div>
  `,
  styleUrls: ['./search-movies.component.sass', '../../styles.sass']
})
export class SearchMoviesComponent {
  @ViewChild(MatPaginator) paginator: MatPaginator;

  @ViewChild('paginatorPageSize') paginatorPageSize: MatPaginator;

  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  public movieTitles = [{}];
  public fetchedMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public filteredMovieList: Movie[] = [];
  public allMovies: Movie[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, title: "", release_date: "", original_language: "", original_title: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public moviesLength: number;
  public totalMovies: number;
  public releaseYear: string[] = [""];
  public endYear: string[] = [""];
  public movieSearch: string;
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public pages: number[] = [0];
  public page: number = 1;

  public movieService: MovieService = inject(MovieService)

  constructor() {
    this.movieService.getGenres().subscribe((resp) => {
      resp['genres'].forEach((genre) => {
        var item = {id: genre.id, name: genre.name}

        this.genreDetails.push(item)
      })
    })
  }

  ngOnInit() {

  }

  // async getGenre(page: number) {
  //   while (this.fetchedMovies.length > 0) {
  //     this.fetchedMovies.pop()
  //   }

  //   // this.galleryMoviesRef.reset()

  //   while (this.movieTitles.length > 0) {
  //     this.movieTitles.pop()
  //   }

  //   this.movieService.getAllMovies(this.page, text).subscribe((resp) => {
  //     console.log(resp['results']);

  //     this.moviesLength = resp['results'].length;
  //     this.totalMovies = resp['total_results'];

  //     resp['results'].forEach((movie) => {
  //       let page = resp['page'];
  //       let isAdult = movie['adult'];
  //       let backdropPath = movie['backdrop_path'];
  //       let genreIds = movie['genre_ids'];
  //       let id = movie['id'];
  //       let releaseDate = movie['release_date'];
  //       let video = movie['video'];
  //       let title = movie['title'];
  //       let originalLanguage = movie['original_language'];
  //       let originalTitle = movie['original_title'];
  //       let overview = movie['overview'];
  //       let popularity = movie['popularity'];
  //       let posterPath = this.baseUrl + movie['poster_path'];
  //       let voteAverage = movie['vote_average'];
  //       let voteCount = movie['vote_count'];
  //       let totalPages = resp['total_pages'];
  //       let totalResult = resp['total_result'];

  //       let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

  //       this.fetchedMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });
  //     })
  //   })

  //   this.fetchedMovies.splice(0, 1);

  //   if (this.filteredMovieList.length > 0) {
  //     while (this.filteredMovieList.length > 0) {
  //       this.filteredMovieList.pop();
  //     }
  //     this.filteredMovieList = this.fetchedMovies;
  //   } else {
  //     this.filteredMovieList = this.fetchedMovies;
  //   }
  // }

  onPageChange(event?: PageEvent) {
    if (event === null) {

    } else {
      this.page = event!.pageIndex + 1;
      this.getGenre(this.page);
    }
  }

  async getGenre(page: number) {
    while (this.fetchedMovies.length > 0) {
      this.fetchedMovies.pop()
    }

    // this.galleryMoviesRef.reset()

    while (this.movieTitles.length > 0) {
      this.movieTitles.pop()
    }

    this.movieService.getAllMovies(page, this.movieSearch).subscribe((resp) => {
      console.log(resp['results']);

      this.moviesLength = resp['results'].length;
      this.totalMovies = resp['total_results'];

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

        let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

        this.fetchedMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });

        this.pages.push(page)
        this.releaseYear.push(releaseDate)
        this.endYear.push(releaseDate)
        this.genre = genreIds[0]
      })

      this.allMovies.splice(0, 1);
    })

    if (this.filteredMovieList.length > 0) {
      while (this.filteredMovieList.length > 0) {
        this.filteredMovieList.pop();
      }
      this.filteredMovieList = this.fetchedMovies;
    } else {
      this.filteredMovieList = this.fetchedMovies;
    }
    //return this.filteredMovieList = this.fetchedMovies.filter((movie) => movie.results[0]?.title.toLowerCase().includes(this.movieSearch.toLowerCase()));
  }

  filterResults(text: string) {


    if (!text) {
      return this.filteredMovieList = this.allMovies;
    }
    this.movieService.getInitialPage(text).subscribe((resp) => {

      this.moviesLength = resp['results'].length;
      this.totalMovies = resp['total_pages'];
      console.log(resp['total_results']);
      console.log(this.totalMovies);
      for (let i = 0; i < this.totalMovies; i++) {
        console.log(this.totalMovies);



        this.movieService.getAllMovies(this.page, text).subscribe((resp) => {
          resp['results'].forEach((movie) => {
            if (movie['title'] === "Meg 2: The Trench") {
              console.log(resp['page']);
            }

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

            let result: MovieResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, title: title, release_date: releaseDate, original_language: originalLanguage, original_title: originalTitle, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

            this.allMovies.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });

            this.pages.push(page)
            this.releaseYear.push(releaseDate)
            this.endYear.push(releaseDate)
            this.genre = genreIds[0]
          })

        })
        this.allMovies.splice(0, 1);
        this.releaseYear.splice(0, 1);
        this.endYear.splice(0, 1);
        this.pages.splice(0, 1);
        this.page++;
      }
    })

    return this.filteredMovieList = this.allMovies.filter((show) => show.results[0]?.title.toLowerCase().includes(text.toLowerCase()));
  }
}
