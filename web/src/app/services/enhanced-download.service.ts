import { Injectable, inject } from '@angular/core';
import { Observable, catchError, map, tap, finalize, throwError } from 'rxjs';
import { MovieService } from '../movies.service';
import { TvShowService } from '../tv-service.service';
import { AnimeService } from '../anime.service';
import { DownloadNotificationService } from './download-notification.service';

export interface DownloadRequest {
  title: string;
  tmdbId: number;
  quality: string;
  description: string;
  contentType: 'movie' | 'show' | 'anime';
  originalLanguage?: string;
  seasons?: number[]; // For TV shows
  year?: string; // For anime movies
}

export interface DownloadResult {
  success: boolean;
  message: string;
  tmdbId: number;
  title: string;
}

@Injectable({
  providedIn: 'root'
})
export class EnhancedDownloadService {
  private movieService = inject(MovieService);
  private tvShowService = inject(TvShowService);
  private animeService = inject(AnimeService);
  private notificationService = inject(DownloadNotificationService);

  constructor() {}

  /**
   * Universal download method that handles movies, TV shows, and anime
   */
  downloadContent(request: DownloadRequest): Observable<DownloadResult> {
    // Validate request
    if (!request.title || !request.tmdbId || !request.quality) {
      return throwError(() => new Error('Invalid download request: missing required fields'));
    }

    // Check if download is already in progress
    if (this.notificationService.isDownloadActive(request.tmdbId, request.contentType)) {
      return throwError(() => new Error('Download already in progress for this content'));
    }

    // Start loading notification
    this.notificationService.startDownload(request.tmdbId, request.contentType, request.title);

    // Route to appropriate download method
    let downloadObservable: Observable<any>;

    switch (request.contentType) {
      case 'movie':
        downloadObservable = this.downloadMovie(request);
        break;
      case 'show':
        downloadObservable = this.downloadTvShow(request);
        break;
      case 'anime':
        downloadObservable = this.downloadAnime(request);
        break;
      default:
        return throwError(() => new Error('Invalid content type'));
    }

    return downloadObservable.pipe(
      map((response) => ({
        success: true,
        message: 'Download request submitted successfully',
        tmdbId: request.tmdbId,
        title: request.title
      })),
      tap((result) => {
        if (result.success) {
          this.notificationService.completeDownload(request.tmdbId, request.contentType, request.title);
        }
      }),
      catchError((error) => {
        console.error(`[Enhanced Download] Error downloading ${request.contentType}:`, error);
        
        let errorMessage = 'An unknown error occurred';
        if (error?.error?.message) {
          errorMessage = error.error.message;
        } else if (error?.message) {
          errorMessage = error.message;
        } else if (typeof error === 'string') {
          errorMessage = error;
        }

        this.notificationService.failDownload(
          request.tmdbId, 
          request.contentType, 
          request.title, 
          errorMessage
        );

        return throwError(() => ({
          success: false,
          message: errorMessage,
          tmdbId: request.tmdbId,
          title: request.title
        }));
      }),
      finalize(() => {
        // Ensure loading state is cleared even if something goes wrong
        // The notification service handles this, but this is a safety net
      })
    );
  }

  private downloadMovie(request: DownloadRequest): Observable<any> {
    const isAnime = request.originalLanguage === 'ja';
    
    if (isAnime) {
      return this.movieService.makeAnimeMovieDownloadRequest(
        request.title,
        request.title, // name
        request.year || '',
        request.quality,
        request.tmdbId,
        request.description
      );
    } else {
      return this.movieService.makeMovieDownloadRequest(
        request.title,
        request.quality,
        request.tmdbId,
        request.description
      );
    }
  }

  private downloadTvShow(request: DownloadRequest): Observable<any> {
    if (!request.seasons || request.seasons.length === 0) {
      return throwError(() => new Error('Season information is required for TV show downloads'));
    }

    const isAnime = request.originalLanguage === 'ja';
    
    if (isAnime) {
      return this.tvShowService.makeAnimeShowDownloadRequest(
        request.title,
        request.seasons,
        request.quality,
        request.tmdbId,
        request.description
      );
    } else {
      return this.tvShowService.makeTvShowDownloadRequest(
        request.title,
        request.seasons,
        request.quality,
        request.tmdbId,
        request.description
      );
    }
  }

  private downloadAnime(request: DownloadRequest): Observable<any> {
    // For anime, we use the same logic as movie/show but force anime endpoints
    if (request.seasons && request.seasons.length > 0) {
      // Anime TV show
      return this.tvShowService.makeAnimeShowDownloadRequest(
        request.title,
        request.seasons,
        request.quality,
        request.tmdbId,
        request.description
      );
    } else {
      // Anime movie
      return this.movieService.makeAnimeMovieDownloadRequest(
        request.title,
        request.title,
        request.year || '',
        request.quality,
        request.tmdbId,
        request.description
      );
    }
  }

  /**
   * Check if a download is currently active
   */
  isDownloadActive(tmdbId: number, contentType: 'movie' | 'show' | 'anime'): boolean {
    return this.notificationService.isDownloadActive(tmdbId, contentType);
  }

  /**
   * Get active downloads observable
   */
  getActiveDownloads() {
    return this.notificationService.getActiveDownloads();
  }

  /**
   * Get notifications observable
   */
  getNotifications() {
    return this.notificationService.getNotifications();
  }
}