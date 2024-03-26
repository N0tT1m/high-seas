import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router'
import { GalleryModule } from 'ng-gallery';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';

@Component({
  template: `
  <head>
    <link rel="stylesheet" href="home.component.sass">
  </head>
  <body>
    
  </body>
    `,
  standalone: true,
  imports: [GalleryModule, CommonModule, RouterModule],
  providers: [MovieService, TvShowService],
  styleUrls: ['./home.component.sass', '../../styles.sass'],
})
export class HomeComponent implements OnInit {
  constructor() {
  }

  ngOnInit(): void {

  }
}
