// src/app/features/libraries/media-card/media-card.component.ts
import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MediaItem } from '../media.model';

@Component({
  selector: 'app-media-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="media-card" [ngClass]="{'has-progress': hasProgress}">
      <div class="thumbnail-container">
        <img
          [src]="item.thumbnail"
          [alt]="item.title"
          class="thumbnail"
          (error)="onImageError($event)"
        >
        <div class="media-type-badge" *ngIf="item.type">
          {{ getMediaTypeLabel() }}
        </div>
        <div class="duration" *ngIf="item.duration && item.duration > 0">
          {{ formatDuration(item.duration) }}
        </div>
        <div class="progress-bar" *ngIf="hasProgress">
          <div class="progress" [style.width.%]="getProgressPercentage()"></div>
        </div>
      </div>
      <div class="media-info">
        <h3 class="title" [title]="item.title">{{ item.title }}</h3>
        <div class="metadata">
          <span class="year" *ngIf="item.year">{{ item.year }}</span>
          <span class="rating" *ngIf="item.rating">{{ formatRating(item.rating) }}</span>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .media-card {
      background-color: #1f1f1f;
      border-radius: 8px;
      overflow: hidden;
      transition: transform 0.2s ease, box-shadow 0.2s ease;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
      cursor: pointer;
      position: relative;
    }

    .media-card:hover {
      transform: translateY(-5px);
      box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
    }

    .thumbnail-container {
      position: relative;
      width: 100%;
      height: 0;
      padding-top: 150%; /* 2:3 aspect ratio for movies/shows */
      overflow: hidden;
      background-color: #111;
    }

    .thumbnail {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      object-fit: cover;
      transition: transform 0.3s ease;
    }

    .media-card:hover .thumbnail {
      transform: scale(1.05);
    }

    .media-type-badge {
      position: absolute;
      top: 10px;
      left: 10px;
      background-color: rgba(255, 123, 0, 0.8);
      color: white;
      font-size: 0.7rem;
      padding: 0.2em 0.6em;
      border-radius: 4px;
      font-weight: 500;
      text-transform: uppercase;
    }

    .duration {
      position: absolute;
      bottom: 10px;
      right: 10px;
      background-color: rgba(0, 0, 0, 0.7);
      color: white;
      font-size: 0.8rem;
      padding: 0.2em 0.6em;
      border-radius: 4px;
    }

    .progress-bar {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 4px;
      background-color: rgba(255, 255, 255, 0.2);
    }

    .progress {
      height: 100%;
      background-color: #ff7b00;
    }

    .media-info {
      padding: 0.8rem;
    }

    .title {
      margin: 0 0 0.3rem 0;
      font-size: 1rem;
      color: white;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .metadata {
      display: flex;
      justify-content: space-between;
      font-size: 0.8rem;
      color: #bbb;
    }

    .year {
      color: #999;
    }

    .rating {
      color: #ff7b00;
    }
  `]
})
export class MediaCardComponent {
  @Input() item!: MediaItem;
  thumbnailError = false;

  get hasProgress(): boolean {
    return !!this.item.viewOffset && this.item.viewOffset > 0 && !!this.item.duration && this.item.duration > 0;
  }

  onImageError(event: Event): void {
    console.error('Image failed to load:', this.item.thumbnail);
    // Replace with placeholder
    (event.target as HTMLImageElement).src = '/assets/images/placeholder.jpg';
  }

  getMediaTypeLabel(): string {
    const type = this.item.type?.toLowerCase() || '';
    switch (type) {
      case 'movie': return 'Movie';
      case 'show': return 'Series';
      case 'episode': return 'Episode';
      case 'season': return 'Season';
      case 'artist': return 'Artist';
      case 'album': return 'Album';
      case 'track': return 'Track';
      default: return type || '';
    }
  }

  formatDuration(milliseconds: number): string {
    const totalSeconds = Math.floor(milliseconds / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  }

  formatRating(rating: number): string {
    return rating.toFixed(1);
  }

  getProgressPercentage(): number {
    if (!this.item.viewOffset || !this.item.duration || this.item.duration <= 0) {
      return 0;
    }
    const percentage = (this.item.viewOffset / this.item.duration) * 100;
    return Math.min(percentage, 100); // Cap at 100%
  }
}
