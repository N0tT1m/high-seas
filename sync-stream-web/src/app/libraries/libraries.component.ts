// src/app/features/libraries/libraries-list/libraries-list.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { PlexService } from '../plex.service';
import { Library } from '../media.model';

@Component({
  selector: 'app-libraries-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <div class="libraries-container">
      <h1>Libraries</h1>

      <div class="loading-overlay" *ngIf="isLoading">
        <div class="spinner"></div>
      </div>

      <div class="error-message" *ngIf="error">
        {{ error }}
      </div>

      <div class="libraries-grid" *ngIf="!isLoading && !error">
        <div class="library-card" *ngFor="let library of libraries" (click)="openLibrary(library)">
          <div class="library-icon" [ngClass]="getLibraryIconClass(library.type)"></div>
          <div class="library-info">
            <h3>{{ library.title }}</h3>
            <span class="library-type">{{ getLibraryTypeLabel(library.type) }}</span>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .libraries-container {
      padding: 1rem;
      position: relative;
      min-height: 300px;
    }

    h1 {
      font-size: 2rem;
      font-weight: 500;
      margin-bottom: 2rem;
      color: white;
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
      background-color: rgba(0, 0, 0, 0.7);
      z-index: 10;
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

    .libraries-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
      gap: 1.5rem;
    }

    .library-card {
      background-color: #2a2a2a;
      border-radius: 8px;
      overflow: hidden;
      cursor: pointer;
      transition: transform 0.2s ease, box-shadow 0.2s ease;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
      display: flex;
      padding: 1.5rem;
    }

    .library-card:hover {
      transform: translateY(-5px);
      box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
    }

    .library-icon {
      width: 50px;
      height: 50px;
      border-radius: 50%;
      background-color: #1f1f1f;
      display: flex;
      justify-content: center;
      align-items: center;
      margin-right: 1rem;
      position: relative;
    }

    .library-icon::before {
      font-family: 'Material Icons';
      font-size: 24px;
      color: #ff7b00;
    }

    .library-icon.movie::before {
      content: 'movie';
    }

    .library-icon.show::before {
      content: 'tv';
    }

    .library-icon.music::before {
      content: 'music_note';
    }

    .library-icon.photo::before {
      content: 'photo';
    }

    .library-info {
      display: flex;
      flex-direction: column;
    }

    .library-info h3 {
      margin: 0 0 0.5rem 0;
      font-size: 1.2rem;
      color: white;
    }

    .library-type {
      font-size: 0.8rem;
      color: #bbb;
      text-transform: uppercase;
    }
  `]
})
export class LibrariesListComponent implements OnInit {
  libraries: Library[] = [];
  isLoading = true;
  error = '';

  constructor(
    private plexService: PlexService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadLibraries();
  }

  loadLibraries(): void {
    this.isLoading = true;
    this.plexService.getLibraries().subscribe(
      libraries => {
        this.libraries = libraries;
        this.isLoading = false;
      },
      error => {
        this.error = 'Error loading libraries. Please try again later.';
        this.isLoading = false;
      }
    );
  }

  getLibraryIconClass(type: string): string {
    switch (type.toLowerCase()) {
      case 'movie': return 'movie';
      case 'show': return 'show';
      case 'artist': return 'music';
      case 'photo': return 'photo';
      default: return 'movie';
    }
  }

  getLibraryTypeLabel(type: string): string {
    switch (type.toLowerCase()) {
      case 'movie': return 'Movies';
      case 'show': return 'TV Shows';
      case 'artist': return 'Music';
      case 'photo': return 'Photos';
      default: return type;
    }
  }

  openLibrary(library: Library): void {
    this.router.navigate(['/library', library.key]);
  }
}
