import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TvShowResult, Genre, TvShow } from '../http-service/http-service.component';
import { MovieService } from '../movies.service';
import { SearchMovieListComponent } from '../search-movie-list/search-movie-list.component';
import { GalleryModule, Gallery, GalleryRef, ImageItem } from 'ng-gallery';
import { RouterModule } from '@angular/router';
import { TvShowService } from '../tv-service.service';

@Component({
  selector: 'app-now-playing',
  standalone: true,
  imports: [CommonModule, GalleryModule, RouterModule],
  providers: [TvShowService],
  template: `
  <div class="container">
    <div class="gallery-wrapper">
      <gallery class="gallery" id="showsGallery"></gallery>
    </div>
    <div class="show-names-wrapper">
      <h3 class="name" *ngFor="let show of showNames">
        <a [routerLink]="['/popular/shows/details', show['id'], show['page']]">{{ show['name'] }}</a>
      </h3>
    </div>
  </div>
  `,
  styleUrls: ['./popular-tv-shows.component.sass']
})
export class PopularGalleryTvShowsComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';

  public fetchedShows: TvShow[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public filteredShowsList: TvShow[] = [];
  public allShows: TvShow[] = [{ page: 0, results: [{ adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false }], total_pages: 0, total_result: 0 }]
  public showsLength: number;
  public totalShows: number;
  public releaseYear: string[] = [""];
  public endYear: string[] = [""];
  public genreDetails: Genre[] = [{ id: 0, name: "None" }];
  public genre: number = this.genreDetails[0]['id'];
  public pages: number[] = [0]
  public showNames = [{}];
  public galleryTvShowsRef: GalleryRef;

  public tvShowService: TvShowService = inject(TvShowService)

  constructor(private gallery: Gallery) {
    this.tvShowService.getInitialPopularPage().subscribe((resp) => {
      this.showsLength = resp['results'].length;
      this.totalShows = resp['total_pages'];
      console.log(resp['total_results']);
    })
  }

  ngOnInit() {
    // Get the galleryRef by id
    this.galleryTvShowsRef = this.gallery.ref('showsGallery');

    let page = 1;

    this.tvShowService.getPopular(page).subscribe((resp) => {
      resp['results'].forEach((show) => {
        let page = resp['page'];
        let isAdult = show['adult'];
        let backdropPath = show['backdrop_path'];
        let genreIds = show['genre_ids'];
        let id = show['id'];
        let firstAirDate = show['first_air_date'];
        let video = show['video'];
        let name = show['name'];
        let originalLanguage = show['original_language'];
        let originalName = show['original_name'];
        let overview = show['overview'];
        let popularity = show['popularity'];
        let posterPath = this.baseUrl + show['poster_path'];
        let voteAverage = show['vote_average'];
        let voteCount = show['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];

        let result: TvShowResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, name: name, first_air_date: firstAirDate, original_language: originalLanguage, original_name: originalName, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

        this.allShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });



        this.pages.push(page)
        console.log('LENB', this.allShows);
      })

      this.allShows.splice(0, 1);
    })

    page++;

    this.tvShowService.getPopular(page).subscribe((resp) => {
      resp['results'].forEach((show) => {
        let page = resp['page'];
        let isAdult = show['adult'];
        let backdropPath = show['backdrop_path'];
        let genreIds = show['genre_ids'];
        let id = show['id'];
        let firstAirDate = show['first_air_date'];
        let video = show['video'];
        let name = show['name'];
        let originalLanguage = show['original_language'];
        let originalName = show['original_name'];
        let overview = show['overview'];
        let popularity = show['popularity'];
        let posterPath = this.baseUrl + show['poster_path'];
        let voteAverage = show['vote_average'];
        let voteCount = show['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];

        let result: TvShowResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, name: name, first_air_date: firstAirDate, original_language: originalLanguage, original_name: originalName, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

        this.allShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });


        this.pages.push(page)
        console.log('LENB', this.allShows);
      })
    })

    page++;

    this.tvShowService.getPopular(page).subscribe((resp) => {
      resp['results'].forEach((show) => {
      let page = resp['page'];
        let isAdult = show['adult'];
        let backdropPath = show['backdrop_path'];
        let genreIds = show['genre_ids'];
        let id = show['id'];
        let firstAirDate = show['first_air_date'];
        let video = show['video'];
        let name = show['name'];
        let originalLanguage = show['original_language'];
        let originalName = show['original_name'];
        let overview = show['overview'];
        let popularity = show['popularity'];
        let posterPath = this.baseUrl + show['poster_path'];
        let voteAverage = show['vote_average'];
        let voteCount = show['vote_count'];
        let totalPages = resp['total_pages'];
        let totalResult = resp['total_result'];

        let result: TvShowResult[] = [{ adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, name: name, first_air_date: firstAirDate, original_language: originalLanguage, original_name: originalName, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video }]

        this.allShows.push({ page: page, results: result, total_pages: totalPages, total_result: totalResult });


        this.pages.push(page)
        console.log('LENB', this.allShows);
      })

      this.allShows.splice(0, 1);

      for (var p = 0; p < this.allShows.length; p++) {
        for (var j = 0; j < this.allShows[p].results.length; j++) {
          if (this.showNames.includes(this.allShows[p].results[j].name)) {
            continue
          } else {
            this.galleryTvShowsRef.addImage({ src: this.allShows[p].results[j].poster_path, thumb: this.allShows[p].results[j].poster_path })
          }
        }
      }

      for (var p = 0; p < this.allShows.length; p++) {
        for (var j = 0; j < this.allShows[p].results.length; j++) {
          if (this.showNames.includes(this.allShows[p].results[j].name)) {
            continue
          } else {
            this.showNames.push({ 'name': this.allShows[p].results[j].name, 'id': this.allShows[p].results[j].id, 'page': this.allShows[p].page })
          }
        }
      }
    })

    this.galleryTvShowsRef.play()
  }
}
