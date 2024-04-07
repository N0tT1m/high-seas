import { Component } from '@angular/core';

@Component({
<<<<<<< HEAD
  selector: 'app-top-rated-tv-shows',
  standalone: true,
  imports: [],
  templateUrl: './top-rated-tv-shows.component.html',
  styleUrl: './top-rated-tv-shows.component.sass'
=======
  selector: 'app-top-rated-shows',
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
        <a [routerLink]="['/top-rated/shows/details', show['id'], show['page']]">{{ show['name'] }}</a>
      </h3>
    </div>
  </div>
  `,
  styleUrls: ['./top-rated-tv-shows.component.sass', '../../styles.sass']
>>>>>>> 94055854302073ac7133a05d7c939df5e7945950
})
export class TopRatedTvShowsComponent {

}
