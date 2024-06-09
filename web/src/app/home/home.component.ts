import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router';
import { GalleryModule } from 'ng-gallery';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';

@Component({
  selector: 'app-home',
  template: `
    <body>
    <div class="movie-poster">
      <img src="/assets/the-spirits.jpg" alt="Movie Poster" (animationend)="onAnimationEnd()">
    </div>
    <div class="star-wars-text">
      <p>
        In a land far far away a engineer had a vision for an application to find new shows and movies then obtain them in an automated way. That applcation is High Seas.
      </p>
    </div>
    <div class="application-text">
      <p>
        Welcome to High Seas.
      </p>
    </div>
    </body>
  `,
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule],
  providers: [MovieService, TvShowService],
  styleUrls: ['./home.component.sass', '../../styles.sass'],
})
export class HomeComponent implements OnInit {
  constructor() {}
  ngOnInit(): void {}
  onAnimationEnd() {
    const starWarsText = document.querySelector('.star-wars-text');
    const applicationText = document.querySelector('.application-text');
    if (starWarsText) {
      starWarsText.classList.add('hidden');
    }
  }
}
