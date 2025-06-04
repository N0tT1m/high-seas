// src/app/features/libraries/library-detail/library-detail.component.ts
import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { Subscription } from 'rxjs';
import { PlexService } from '../plex.service';
import { MediaItem } from '../media.model';
import { MediaCardComponent } from '../media-card/media-card.component';

@Component({
  selector: 'app-library-detail',
  standalone: true,
  imports: [CommonModule, MediaCardComponent],
  template: `
    <div class="library-detail-container">
      <div class="header">
        <button class="back-button" (click)="goBack()">
          <span class="material-icons">arrow_back</span>
        </button>
        <h1>{{ libraryName }}</h1>
      </div>

      <div class="loading-container" *ngIf="isLoading">
        <div class="spinner"></div>
        <p>Loading content...</p>
      </div>

      <div class="error-message" *ngIf="error">
        {{ error }}
        <button class="retry-button" (click)="loadLibraryContent()">Retry</button>
      </div>

      <div class="content-grid" *ngIf="!isLoading && !error && items.length > 0">
        <app-media-card
          *ngFor="let item of items"
          [item]="item"
          (click)="navigateToMedia(item)">
        </app-media-card>
      </div>

      <div class="empty-state" *ngIf="!isLoading && !error && items.length === 0">
        <span class="material-icons">movie_filter</span>
        <p>No items found in this library</p>
        <p class="hint">This library appears to be empty</p>
      </div>

      <div class="debug-panel" *ngIf="isDebugMode">
        <div class="debug-info">
          <h3>Debug Information</h3>
          <p>Library Key: {{ libraryKey }}</p>
          <p>Items Loaded: {{ items.length }}</p>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .library-detail-container {
      padding: 1.5rem;
      min-height: 600px;
      position: relative;
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
      margin-right: 1rem;
      padding: 0.5rem;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    h1 {
      font-size: 2rem;
      font-weight: 500;
      color: white;
      margin: 0;
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

    .retry-button:hover {
      background-color: rgba(255, 123, 0, 1);
    }

    .content-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
      gap: 2rem 1.5rem;
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      background-color: rgba(255, 255, 255, 0.05);
      border-radius: 8px;
      padding: 4rem 2rem;
      color: #888;
      text-align: center;
    }

    .empty-state .material-icons {
      font-size: 4rem;
      color: #666;
      margin-bottom: 1.5rem;
    }

    .empty-state p {
      margin: 0.5rem 0;
      font-size: 1.2rem;
    }

    .empty-state .hint {
      font-size: 0.9rem;
      color: #666;
      margin-top: 0.5rem;
    }

    .debug-panel {
      margin-top: 3rem;
      padding: 1rem;
      background-color: rgba(0, 0, 0, 0.3);
      border-radius: 8px;
    }

    .debug-info {
      font-family: monospace;
      color: #bbb;
    }

    .debug-info h3 {
      margin-top: 0;
      color: #ff7b00;
    }
  `]
})
export class LibraryDetailComponent implements OnInit, OnDestroy {
  libraryKey = '';
  libraryName = '';
  items: MediaItem[] = [];
  isLoading = true;
  error = '';
  isDebugMode = false; // Set to true to show debug information
  private subscriptions: Subscription[] = [];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private plexService: PlexService
  ) {}

  ngOnInit(): void {
    // Enable debug mode by checking for URL parameter or environment
    const debugMode = localStorage.getItem('plexSyncDebug');
    if (debugMode === 'true') {
      this.isDebugMode = true;
    }

    const paramsSub = this.route.params.subscribe(params => {
      this.libraryKey = params['id'];
      console.log(`Initializing library detail for library key: ${this.libraryKey}`);
      this.loadLibraryContent();
    });

    this.subscriptions.push(paramsSub);
  }

  ngOnDestroy(): void {
    this.subscriptions.forEach(sub => sub.unsubscribe());
    console.log('Library detail component destroyed, subscriptions cleaned up');
  }

  loadLibraryContent(): void {
    this.isLoading = true;
    this.error = '';

    console.log(`Loading content for library: ${this.libraryKey}`);

    // Set default library name based on key while loading
    switch (this.libraryKey) {
      case '1':
        this.libraryName = 'Movies';
        break;
      case '2':
        this.libraryName = 'TV Shows';
        break;
      default:
        this.libraryName = `Library ${this.libraryKey}`;
    }

    const loadSub = this.plexService.getLibraryItems(this.libraryKey).subscribe(
      items => {
        console.log(`Loaded ${items.length} items for library ${this.libraryKey}`);
        this.items = items;
        this.isLoading = false;

        // Update library name from items if available
        if (items.length > 0 && items[0].librarySectionTitle) {
          this.libraryName = items[0].librarySectionTitle;
        }
      },
      error => {
        console.error('Error loading library content:', error);
        this.error = 'Error loading library content. Please try again later.';
        this.isLoading = false;
      }
    );

    this.subscriptions.push(loadSub);
  }

  navigateToMedia(item: MediaItem): void {
    console.log('Navigating to media item:', item);

    // Extract numeric ID from key (e.g., "/library/metadata/101" -> "101")
    const matches = item.key.match(/\/library\/metadata\/(\d+)/);
    const mediaId = matches ? matches[1] : item.key.split('/').pop();

    this.router.navigate(['/media', mediaId]);
  }

  goBack(): void {
    this.router.navigate(['/libraries']);
  }
}
