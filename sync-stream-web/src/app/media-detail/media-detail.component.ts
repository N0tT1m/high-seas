// src/app/features/libraries/media-detail/media-detail.component.ts
import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Subscription } from 'rxjs';
import { PlexService } from '../plex.service';
import { MediaItem } from '../media.model';

@Component({
  selector: 'app-media-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <div class="media-detail-container">
      <div class="backdrop" *ngIf="getBackdropUrl()" [style.background-image]="'url(' + getBackdropUrl() + ')'"></div>

      <div class="content-wrapper">
        <div class="header">
          <button class="back-button" (click)="goBack()">
            <span class="material-icons">arrow_back</span>
          </button>
        </div>

        <div class="loading-container" *ngIf="isLoading">
          <div class="spinner"></div>
          <p>Loading media details...</p>
        </div>

        <div class="error-message" *ngIf="error">
          {{ error }}
          <button class="retry-button" (click)="loadMediaInfo()">Retry</button>
        </div>

        <div class="media-content" *ngIf="!isLoading && !error && mediaInfo">
          <div class="poster-container">
            <img [src]="getPosterUrl()" [alt]="mediaInfo.title" class="poster" (error)="onImageError($event, 'poster')">
          </div>

          <div class="details">
            <h1 class="title">{{ mediaInfo.title }}</h1>
            <div class="metadata">
              <span class="year" *ngIf="mediaInfo.year">{{ mediaInfo.year }}</span>
              <span class="dot" *ngIf="mediaInfo.year && mediaInfo.duration">•</span>
              <span class="duration" *ngIf="mediaInfo.duration">{{ formatDuration(mediaInfo.duration) }}</span>
              <span class="dot" *ngIf="(mediaInfo.year || mediaInfo.duration) && mediaInfo.contentRating">•</span>
              <span class="content-rating" *ngIf="mediaInfo.contentRating">{{ mediaInfo.contentRating }}</span>
            </div>

            <div class="tagline" *ngIf="mediaInfo.tagline">{{ mediaInfo.tagline }}</div>

            <div class="summary" *ngIf="mediaInfo.summary">{{ mediaInfo.summary }}</div>

            <div class="genres" *ngIf="mediaInfo.genres && mediaInfo.genres.length > 0">
              <span class="genre" *ngFor="let genre of mediaInfo.genres">{{ genre }}</span>
            </div>

            <div class="action-buttons">
              <button class="play-button" (click)="playMedia()">
                <span class="material-icons">play_arrow</span>
                Play
              </button>
            </div>

            <div class="section cast" *ngIf="mediaInfo.actors && mediaInfo.actors.length > 0">
              <h3>Cast</h3>
              <div class="cast-list">
                <div class="cast-member" *ngFor="let actor of mediaInfo.actors.slice(0, 6)">
                  {{ actor }}
                </div>
              </div>
            </div>

            <div class="section crew" *ngIf="mediaInfo.directors && mediaInfo.directors.length > 0">
              <h3>Director<span *ngIf="mediaInfo.directors.length > 1">s</span></h3>
              <div class="crew-list">
                <div class="crew-member" *ngFor="let director of mediaInfo.directors">
                  {{ director }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .media-detail-container {
      position: relative;
      min-height: 100vh;
      color: white;
    }

    .backdrop {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-size: cover;
      background-position: center;
      filter: blur(8px);
      opacity: 0.3;
      z-index: -1;
    }

    .content-wrapper {
      position: relative;
      z-index: 1;
      padding: 2rem;
    }

    .header {
      display: flex;
      align-items: center;
      margin-bottom: 2rem;
    }

    .back-button {
      background: none;
      border: none;
      color: #ff7b00;
      cursor: pointer;
      padding: 0.5rem;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      min-height: 400px;
      color: #bbb;
    }

    .spinner {
      width: 50px;
      height: 50px;
      border: 4px solid rgba(255, 123, 0, 0.3);
      border-radius: 50%;
      border-top-color: #ff7b00;
      animation: spin 1s ease-in-out infinite;
      margin-bottom: 1.5rem;
    }

    @keyframes spin {
      to { transform: rotate(360deg); }
    }

    .error-message {
      padding: 1.5rem;
      background-color: rgba(255, 0, 0, 0.2);
      border-radius: 8px;
      color: #ff6b6b;
      margin-bottom: 2rem;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .retry-button {
      background-color: rgba(255, 123, 0, 0.8);
      color: white;
      border: none;
      padding: 0.5rem 1rem;
      border-radius: 4px;
      cursor: pointer;
      transition: background-color 0.2s;
    }

    .media-content {
      display: flex;
      gap: 2rem;
      max-width: 1200px;
      margin: 0 auto;
    }

    .poster-container {
      flex: 0 0 300px;
    }

    .poster {
      width: 100%;
      border-radius: 8px;
      box-shadow: 0 5px 15px rgba(0, 0, 0, 0.5);
    }

    .details {
      flex: 1;
    }

    .title {
      font-size: 2.5rem;
      margin: 0 0 0.5rem 0;
      color: white;
    }

    .metadata {
      display: flex;
      align-items: center;
      color: #bbb;
      margin-bottom: 1rem;
    }

    .dot {
      margin: 0 0.5rem;
    }

    .tagline {
      font-style: italic;
      color: #ddd;
      margin-bottom: 1rem;
    }

    .summary {
      line-height: 1.6;
      margin-bottom: 1.5rem;
      color: #bbb;
    }

    .genres {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
      margin-bottom: 1.5rem;
    }

    .genre {
      background-color: rgba(255, 123, 0, 0.2);
      color: #ff7b00;
      padding: 0.3rem 0.8rem;
      border-radius: 16px;
      font-size: 0.9rem;
    }

    .action-buttons {
      margin-bottom: 2rem;
    }

    .play-button {
      background-color: #ff7b00;
      color: white;
      border: none;
      padding: 0.8rem 1.5rem;
      border-radius: 4px;
      font-size: 1rem;
      display: flex;
      align-items: center;
      gap: 0.5rem;
      cursor: pointer;
      transition: background-color 0.2s;
    }

    .play-button:hover {
      background-color: #ff8c24;
    }

    .section {
      margin-bottom: 1.5rem;
    }

    .section h3 {
      color: white;
      font-size: 1.2rem;
      margin-bottom: 0.8rem;
    }

    .cast-list, .crew-list {
      display: flex;
      flex-wrap: wrap;
      gap: 1rem;
    }

    .cast-member, .crew-member {
      background-color: rgba(255, 255, 255, 0.1);
      padding: 0.5rem 1rem;
      border-radius: 4px;
      font-size: 0.9rem;
    }

    @media (max-width: 768px) {
      .media-content {
        flex-direction: column;
      }

      .poster-container {
        flex: 0 0 auto;
        max-width: 200px;
        margin: 0 auto 2rem;
      }
    }
  `]
})
export class MediaDetailComponent implements OnInit, OnDestroy {
  mediaId: string = '';
  mediaInfo: MediaItem | null = null;
  isLoading = true;
  error = '';
  private subscriptions: Subscription[] = [];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private plexService: PlexService
  ) {}

  ngOnInit(): void {
    const paramsSub = this.route.params.subscribe(params => {
      this.mediaId = params['id'];
      this.loadMediaInfo();
    });

    this.subscriptions.push(paramsSub);
  }

  ngOnDestroy(): void {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }

  // In media-detail.component.ts, modify the loadMediaInfo method:

  loadMediaInfo(): void {
    this.isLoading = true;
    this.error = '';

    // Just pass the mediaId, not the full path
    this.plexService.getMediaInfo(this.mediaId).subscribe(
      mediaInfo => {
        this.mediaInfo = mediaInfo;
        this.isLoading = false;
        console.log('Loaded media info:', mediaInfo);
      },
      error => {
        this.error = 'Error loading media details. Please try again later.';
        this.isLoading = false;
        console.error('Error loading media details:', error);
      }
    );
  }

  getPosterUrl(): string {
    if (!this.mediaInfo || !this.mediaInfo.thumbnail) {
      return 'assets/images/placeholder.jpg';
    }
    // Use the thumbnail directly - it should already have the full URL with auth token
    return this.mediaInfo.thumbnail;
  }

  getBackdropUrl(): string {
    if (!this.mediaInfo || !this.mediaInfo.art) {
      return '';
    }
    // Use the art URL directly - it should already have the full URL with auth token
    return this.mediaInfo.art;
  }

  onImageError(event: Event, type: string): void {
    console.error(`Failed to load ${type} image:`, event);
    (event.target as HTMLImageElement).src = 'assets/images/placeholder.jpg';
  }

  formatTime(ms: number): string {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    } else {
      return `${minutes}m ${seconds % 60}s`;
    }
  }

  formatDuration(ms: number): string {
    const minutes = Math.floor(ms / 60000);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    } else {
      return `${minutes}m`;
    }
  }

  playMedia(): void {
    this.router.navigate(['/player', this.mediaId]);
  }

  goBack(): void {
    this.router.navigate(['/home']);
  }
}
