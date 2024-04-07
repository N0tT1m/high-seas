import { Component } from '@angular/core';

@Component({
<<<<<<< HEAD
  selector: 'app-popular-tv-shows',
  standalone: true,
  imports: [],
  templateUrl: './popular-tv-shows.component.html',
  styleUrl: './popular-tv-shows.component.sass'
=======
  selector: 'app-popular-shows',
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
  styleUrls: ['./popular-tv-shows.component.sass', '../../styles.sass']
>>>>>>> 94055854302073ac7133a05d7c939df5e7945950
})
export class PopularTvShowsComponent {

}
