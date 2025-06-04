// src/app/features/media/media-player/media-player.component.ts
import {Component, OnInit, OnDestroy, ViewChild} from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { Subscription } from 'rxjs';
import { PlexService } from '../plex.service';
import { VideoPlayerComponent } from '../video-player/video-player.component';
import { MediaItem } from '../media.model';

@Component({
  selector: 'app-media-player',
  standalone: true,
  imports: [CommonModule, VideoPlayerComponent],
  template: `
    <div class="player-container">
      <div class="loading-overlay" *ngIf="isLoading">
        <div class="spinner"></div>
      </div>

      <div class="error-message" *ngIf="error">
        {{ error }}
        <button (click)="goBack()" class="back-button">Back</button>
      </div>

      <app-video-player
        #videoPlayer
        [autoplay]="true"
        *ngIf="!isLoading && !error && mediaInfo">
      </app-video-player>

      <div class="back-control">
        <button (click)="goBack()" class="back-button">
          <span class="material-icons">arrow_back</span>
          Back
        </button>
      </div>
    </div>
  `,
  styles: [`
    .player-container {
      position: relative;
      width: 100%;
      height: calc(100vh - 64px);
      background-color: black;
    }

    .loading-overlay {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      display: flex;
      justify-content: center;
      align-items: center;
      background-color: rgba(0, 0, 0, 0.8);
      z-index: 10;
    }

    .spinner {
      width: 50px;
      height: 50px;
      border: 5px solid rgba(255, 255, 255, 0.3);
      border-radius: 50%;
      border-top-color: #ff7b00;
      animation: spin 1s ease-in-out infinite;
    }

    @keyframes spin {
      to { transform: rotate(360deg); }
    }

    .error-message {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      padding: 2rem;
      background-color: rgba(40, 40, 40, 0.9);
      border-radius: 8px;
      color: white;
      text-align: center;
      z-index: 10;
    }

    .back-control {
      position: absolute;
      top: 20px;
      left: 20px;
      z-index: 5;
    }

    .back-button {
      display: flex;
      align-items: center;
      background-color: rgba(0, 0, 0, 0.7);
      color: white;
      border: none;
      border-radius: 4px;
      padding: 0.5rem 1rem;
      cursor: pointer;
      transition: background-color 0.2s ease;
    }

    .back-button:hover {
      background-color: rgba(0, 0, 0, 0.9);
    }

    .back-button .material-icons {
      margin-right: 0.5rem;
    }
  `]
})
export class MediaPlayerComponent implements OnInit, OnDestroy {
  mediaId: string = '';
  mediaInfo: MediaItem | null = null;
  isLoading = true;
  error = '';
  private subscriptions: Subscription[] = [];

  // Add ViewChild to reference the video player component
  @ViewChild('videoPlayer') videoPlayerComponent!: VideoPlayerComponent;

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

  loadMediaInfo(): void {
    this.isLoading = true;
    this.error = '';

    // Remove leading slash if present
    let mediaPath = this.mediaId;
    if (!mediaPath.includes('library/metadata/')) {
      // If it's just an ID, format it correctly
      mediaPath = `library/metadata/${mediaPath}`;
    } else if (mediaPath.startsWith('/')) {
      // If it's a full path with leading slash, remove the slash
      mediaPath = mediaPath.substring(1);
    }

    console.log(`Requesting media info for path: ${mediaPath}`);

    this.plexService.getMediaInfo(mediaPath).subscribe({
      next: (mediaInfo) => {
        console.log('Media info received:', mediaInfo);
        this.mediaInfo = mediaInfo;

        // Get the stream URL for this media item
        this.plexService.getStreamUrl(mediaPath).subscribe({
          next: (streamUrl) => {
            console.log('Stream URL received:', streamUrl);

            if (!streamUrl) {
              this.error = 'Unable to get stream URL. Please try again later.';
              this.isLoading = false;
              return;
            }

            // Update mediaInfo with the stream URL
            if (this.mediaInfo) {
              this.mediaInfo.streamUrl = streamUrl;
            }

            this.isLoading = false;

            // Wait for Angular to render the video player
            setTimeout(() => {
              // Try to access the video player component via ViewChild
              if (this.videoPlayerComponent) {
                console.log('Loading media via ViewChild reference');
                this.videoPlayerComponent.loadMedia(this.mediaInfo!);
              } else {
                console.error('Video player component not available via ViewChild');
                this.error = 'Error initializing video player. Please refresh the page.';
              }
            }, 100);
          },
          error: (error) => {
            console.error('Error getting stream URL:', error);
            this.error = 'Error getting stream URL. Please try again later.';
            this.isLoading = false;
          }
        });
      },
      error: (error) => {
        console.error('Error loading media:', error);
        this.error = 'Error loading media. Please try again later.';
        this.isLoading = false;
      }
    });
  }

  goBack(): void {
    this.router.navigate(['/media']);
  }
}
