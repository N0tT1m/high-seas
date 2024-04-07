import { Component } from '@angular/core';

@Component({
  selector: 'app-top-rated-movies',
  standalone: true,
<<<<<<< HEAD
  imports: [],
  templateUrl: './top-rated-movies.component.html',
  styleUrl: './top-rated-movies.component.sass'
=======
  imports: [CommonModule, GalleryModule, RouterModule],
  providers: [MovieService],
  template: `
  <div class="container">
    <div class="gallery-wrapper">
      <gallery class="gallery" id="moviesGallery"></gallery>
    </div>
    <div class="movie-title-wrapper">
      <h3 class="title" *ngFor="let movie of movieTitles">
        <a [routerLink]="['/top-rated/movies/details', movie['id'], movie['page']]">{{ movie['title'] }}</a>
      </h3>
    </div>
  </div>
  `,
  styleUrls: ['./top-rated-movies.component.sass', '../../styles.sass']
>>>>>>> 94055854302073ac7133a05d7c939df5e7945950
})
export class TopRatedMoviesComponent {

}
