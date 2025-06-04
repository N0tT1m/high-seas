// src/app/features/search/search.component.ts
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { PlexService } from '../plex.service';
import { MediaItem } from '../media.model';
import { MediaCardComponent } from '../media-card/media-card.component';

@Component({
  selector: 'app-search',
  standalone: true,
  imports: [CommonModule, FormsModule, MediaCardComponent],
  template: `
    <div class="search-container">
      <div class="search-header">
        <h1>Search</h1>
        <div class="search-form">
          <input
            type="text"
            [(ngModel)]="searchQuery"
            (keyup.enter)="performSearch()"
            placeholder="Search movies, TV shows, music..."
            class="search-input"
          >
          <button (click)="performSearch()" class="search-button">
            <span class="material-icons">search</span>
          </button>
        </div>
      </div>

      <div class="search-results" *ngIf="hasSearched">
        <div class="loading-indicator" *ngIf="isLoading">
          <div class="spinner"></div>
        </div>

        <div class="error-message" *ngIf="error">
          {{ error }}
        </div>

        <div class="results-count" *ngIf="searchResults.length > 0">
          Found {{ searchResults.length }} results for "{{ lastSearchQuery }}"
        </div>

        <div class="no-results" *ngIf="!isLoading && !error && searchResults.length === 0">
          No results found for "{{ lastSearchQuery }}"
        </div>

        <div class="results-grid" *ngIf="searchResults.length > 0">
          <app-media-card
            *ngFor="let item of searchResults"
            [item]="item"
            (click)="navigateToMedia(item)">
          </app-media-card>
        </div>
      </div>

      <div class="empty-search" *ngIf="!hasSearched">
        <div class="search-prompt">
          <span class="material-icons">search</span>
          <p>Enter a search term to find content</p>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .search-container {
      padding: 1rem;
    }

    .search-header {
      margin-bottom: 2rem;
    }

    h1 {
      font-size: 2rem;
      font-weight: 500;
      margin-bottom: 1rem;
      color: white;
    }

    .search-form {
      display: flex;
      width: 100%;
      max-width: 600px;
    }

    .search-input {
      flex: 1;
      padding: 0.75rem 1rem;
      border: none;
      border-radius: 4px 0 0 4px;
      background-color: #2a2a2a;
      color: white;
      font-size: 1rem;
    }

    .search-input:focus {
      outline: none;
      box-shadow: 0 0 0 2px #ff7b00;
    }

    .search-button {
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 0.75rem 1rem;
      border: none;
      border-radius: 0 4px 4px 0;
      background-color: #ff7b00;
      color: white;
      cursor: pointer;
      transition: background-color 0.2s ease;
    }

    .search-button:hover {
      background-color: #e06e00;
    }

    .loading-indicator {
      display: flex;
      justify-content: center;
      padding: 2rem;
    }

    .spinner {
      width: 40px;
      height: 40px;
      border: 4px solid rgba(255, 255, 255, 0.3);
      border-radius: 50%;
      border-top-color: #ff7b00;
      animation: spin 1s ease-in-out infinite;
    }

    @keyframes spin {
      to { transform: rotate(360deg); }
    }

    .error-message {
      padding: 1rem;
      background-color: rgba(255, 0, 0, 0.2);
      border-radius: 4px;
      color: #ff6b6b;
      margin-bottom: 1rem;
    }

    .results-count {
      font-size: 1rem;
      color: #bbb;
      margin-bottom: 1.5rem;
    }

    .no-results {
      padding: 2rem;
      text-align: center;
      color: #888;
      background-color: rgba(255, 255, 255, 0.05);
      border-radius: 8px;
    }

    .results-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 1.5rem;
    }

    .empty-search {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 300px;
    }

    .search-prompt {
      text-align: center;
      color: #888;
    }

    .search-prompt .material-icons {
      font-size: 48px;
      margin-bottom: 1rem;
    }
  `]
})
export class SearchComponent {
  searchQuery = '';
  lastSearchQuery = '';
  searchResults: MediaItem[] = [];
  isLoading = false;
  error = '';
  hasSearched = false;

  constructor(
    private plexService: PlexService,
    private router: Router
  ) {}

  performSearch(): void {
    if (!this.searchQuery.trim()) {
      return;
    }

    this.isLoading = true;
    this.error = '';
    this.hasSearched = true;
    this.lastSearchQuery = this.searchQuery;

    this.plexService.searchMedia(this.searchQuery).subscribe(
      results => {
        this.searchResults = results;
        this.isLoading = false;
      },
      error => {
        this.error = 'Error performing search. Please try again.';
        this.isLoading = false;
        this.searchResults = [];
      }
    );
  }

  navigateToMedia(item: MediaItem): void {
    this.router.navigate(['/media', item.key.split('/').pop()]);
  }
}
