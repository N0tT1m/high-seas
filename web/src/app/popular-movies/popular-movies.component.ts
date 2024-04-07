import { Component } from '@angular/core';

@Component({
  selector: 'app-popular-movies',
  standalone: true,
<<<<<<< HEAD
  imports: [],
  templateUrl: './popular-movies.component.html',
  styleUrl: './popular-movies.component.sass'
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
        <a [routerLink]="['/popular/movies/details', movie['id'], movie['page']]">{{ movie['title'] }}</a>
      </h3>
    </div>
  </div>
  `,
  styleUrls: ['./popular-movies.component.sass', '../../styles.sass']
>>>>>>> 94055854302073ac7133a05d7c939df5e7945950
})
export class PopularMoviesComponent {

}
