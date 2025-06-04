// src/app/features/home/home.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { PlexService } from '../plex.service';
import { MediaItem } from '../media.model';
import { MediaCardComponent } from '../media-card/media-card.component';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterLink, MediaCardComponent],
  template: `
    <div class="home-container">
      <section class="section">
        <h2 class="section-title">Continue Watching</h2>
        <div class="media-grid" *ngIf="continueWatching.length > 0; else emptyContinue">
          <app-media-card
            *ngFor="let item of continueWatching"
            [item]="item"
            (click)="navigateToMedia(item)">
          </app-media-card>
        </div>
        <ng-template #emptyContinue>
          <div class="empty-state">
            <p>No items to continue watching</p>
          </div>
        </ng-template>
      </section>

      <section class="section">
        <h2 class="section-title">Recently Added</h2>
        <div class="media-grid" *ngIf="recentlyAdded.length > 0; else emptyRecent">
          <app-media-card
            *ngFor="let item of recentlyAdded"
            [item]="item"
            (click)="navigateToMedia(item)">
          </app-media-card>
        </div>
        <ng-template #emptyRecent>
          <div class="empty-state">
            <p>No recently added content</p>
          </div>
        </ng-template>
      </section>
    </div>
  `,
  styles: [`
    .home-container {
      padding: 1rem;
    }

    .section {
      margin-bottom: 2rem;
    }

    .section-title {
      font-size: 1.5rem;
      font-weight: 500;
      margin-bottom: 1rem;
      color: white;
    }

    .media-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 1.5rem;
    }

    .empty-state {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 150px;
      background-color: rgba(255, 255, 255, 0.05);
      border-radius: 8px;
      color: #888;
    }
  `]
})
export class HomeComponent implements OnInit {
  continueWatching: MediaItem[] = [];
  recentlyAdded: MediaItem[] = [];
  isLoading = true;
  error = '';

  constructor(
    private plexService: PlexService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadData();
  }

  // Update the loadData method to handle null responses
  loadData(): void {
    this.isLoading = true;
    this.error = '';

    // Get Continue Watching data
    this.plexService.getContinueWatching().subscribe({
      next: (items) => {
        this.continueWatching = items || []; // Handle null response
        this.isLoading = false;
      },
      error: (error) => {
        console.error('Error loading continue watching data:', error);
        this.error = 'Error loading continue watching data';
        this.continueWatching = []; // Initialize to empty array on error
        this.isLoading = false;
      }
    });

    // Get Recently Added data
    this.plexService.getRecentlyAdded().subscribe({
      next: (items) => {
        this.recentlyAdded = items || []; // Handle null response
      },
      error: (error) => {
        console.error('Error loading recently added data:', error);
        this.error = this.error ? this.error + ' and recently added data' : 'Error loading recently added data';
        this.recentlyAdded = []; // Initialize to empty array on error
      }
    });
  }

  navigateToMedia(item: MediaItem): void {
    this.router.navigate(['/media', item.key.split('/').pop()]);
  }
}
