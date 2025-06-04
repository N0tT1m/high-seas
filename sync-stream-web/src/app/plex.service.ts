// src/app/plex.service.ts
import { Injectable } from '@angular/core';
import {HttpClient, HttpErrorResponse, HttpHeaders} from '@angular/common/http';
import { Observable, BehaviorSubject, of } from 'rxjs';
import { map, tap, switchMap, catchError } from 'rxjs/operators';
import { environment } from '../environments/environment';
import { MediaItem, Library } from './media.model';
import { MediaSessionService } from './media-session.service';
import { AuthService } from './auth.service';

@Injectable({
  providedIn: 'root'
})
export class PlexService {
  private serversSubject = new BehaviorSubject<any[]>([]);
  public servers$ = this.serversSubject.asObservable();

  private selectedServerSubject = new BehaviorSubject<any | null>(null);
  public selectedServer$ = this.selectedServerSubject.asObservable();

  private librariesSubject = new BehaviorSubject<Library[]>([]);
  public libraries$ = this.librariesSubject.asObservable();

  private continueWatchingSubject = new BehaviorSubject<MediaItem[]>([]);
  public continueWatching$ = this.continueWatchingSubject.asObservable();

  private recentlyAddedSubject = new BehaviorSubject<MediaItem[]>([]);
  public recentlyAdded$ = this.recentlyAddedSubject.asObservable();

  // Declare the plexServerUrl property
  private plexServerUrl: string = '';

  constructor(
    private http: HttpClient,
    private mediaSessionService: MediaSessionService,
    private authService: AuthService
  ) {
    // Initialize with environment plexUrl as default
    this.plexServerUrl = environment.plexUrl || '';

    // Get servers to update the URL if available
    this.getServers().subscribe(servers => {
      if (servers.length > 0) {
        this.plexServerUrl = servers[0].url || environment.plexUrl || '';
        console.log('Plex server URL set to:', this.plexServerUrl);
      }
    });
  }

  // Helper method to get HTTP options with auth token
  private getHttpOptions() {
    const token = this.authService.getToken();
    return {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      })
    };
  }

  // Get all servers available to the user
  getServers(): Observable<any[]> {
    if (!this.authService.isAuthenticated()) {
      console.warn('Attempted to fetch servers while not authenticated');
      return of([]);
    }

    return this.http.get<any[]>(
      `${environment.apiUrl}/api/servers`,
      this.getHttpOptions()
    ).pipe(
      tap(servers => {
        this.serversSubject.next(servers);
        if (servers.length > 0 && !this.selectedServerSubject.value) {
          this.selectServer(servers[0]);
        }
      }),
      catchError(error => {
        console.error('Error fetching servers:', error);
        return of([]);
      })
    );
  }

  // Select a server to work with
  selectServer(server: any): void {
    this.selectedServerSubject.next(server);
    // When a server is selected, fetch its libraries
    this.getLibraries().subscribe();
    // Update the Plex server URL
    if (server && server.url) {
      this.plexServerUrl = server.url;
      console.log('Plex server URL updated to:', this.plexServerUrl);
    }
  }

  // Get direct stream URL for a media item
  getStreamUrl(mediaKey: string): Observable<string> {
    if (!this.authService.isAuthenticated()) {
      return of('');
    }

    // Handle Plex-style paths
    let normalizedMediaKey = mediaKey;
    if (mediaKey.startsWith('/library/metadata/') || mediaKey.startsWith('library/metadata/')) {
      // For Plex-style paths, remove leading slash and keep the path structure
      normalizedMediaKey = mediaKey.startsWith('/') ? mediaKey.substring(1) : mediaKey;
    } else if (!mediaKey.includes('library/metadata/')) {
      // If it's just an ID, format it correctly
      normalizedMediaKey = `library/metadata/${mediaKey}`;
    }

    console.log(`Getting stream URL for normalized path: ${normalizedMediaKey}`);

    return this.http.get<{ streamUrl: string }>(
      `${environment.apiUrl}/api/media/${normalizedMediaKey}/stream`,
      this.getHttpOptions()
    ).pipe(
      map(response => {
        const streamUrl = response.streamUrl;
        console.log(`Received stream URL: ${streamUrl}`);

        // Check if this is a direct file path or a Plex URL
        if (streamUrl && streamUrl.includes('/video/:/transcode/universal/start')) {
          // This is a Plex universal transcoding URL
          // Make sure it has the correct authentication token
          const token = this.authService.getToken();
          let finalUrl = streamUrl;

          // Add X-Plex-Token if it's not already there
          if (!finalUrl.includes('X-Plex-Token')) {
            finalUrl += (finalUrl.includes('?') ? '&' : '?') + `X-Plex-Token=${token}`;
          }

          // Add additional parameters to ensure the URL is directly playable in browsers
          if (!finalUrl.includes('directPlay=0')) {
            finalUrl += '&directPlay=0';
          }
          if (!finalUrl.includes('directStream=1')) {
            finalUrl += '&directStream=1';
          }
          if (!finalUrl.includes('mediaIndex=0')) {
            finalUrl += '&mediaIndex=0';
          }
          if (!finalUrl.includes('videoQuality=100')) {
            finalUrl += '&videoQuality=100';
          }
          if (!finalUrl.includes('audioBoost=100')) {
            finalUrl += '&audioBoost=100';
          }

          console.log(`Enhanced stream URL: ${finalUrl}`);
          return finalUrl;
        }

        // If it's already a direct file URL, just return it
        return streamUrl;
      }),
      catchError(error => {
        console.error(`Error fetching stream URL for ${mediaKey}:`, error);

        // Fallback: Try to construct a direct Plex URL if the API fails
        if (this.plexServerUrl) {
          const token = this.authService.getToken();
          let id = mediaKey;

          // Extract the ID from the path if needed
          if (id.includes('library/metadata/')) {
            const matches = id.match(/library\/metadata\/(\d+)/);
            if (matches && matches[1]) {
              id = matches[1];
            }
          }

          const fallbackUrl = `${this.plexServerUrl}/video/:/transcode/universal/start?path=/library/metadata/${id}&X-Plex-Token=${token}&directPlay=0&directStream=1&mediaIndex=0&videoQuality=100&audioBoost=100`;

          console.log(`Using fallback stream URL: ${fallbackUrl}`);
          return of(fallbackUrl);
        }

        return of('');
      })
    );
  }

  // Update the formatPlexStreamUrl method in video-player.component.ts

  private formatPlexStreamUrl(streamUrl: string, token: string): string {
    if (!streamUrl) {
      return '';
    }

    let finalUrl = streamUrl;

    // Check if this is a Plex transcoding URL
    if (streamUrl.includes('/video/:/transcode/universal/start')) {
      // Add X-Plex-Token if it's not already there
      if (!finalUrl.includes('X-Plex-Token')) {
        finalUrl += (finalUrl.includes('?') ? '&' : '?') + `X-Plex-Token=${token}`;
      }

      // Add required parameters for browser playback
      if (!finalUrl.includes('directPlay=0')) {
        finalUrl += '&directPlay=0';
      }
      if (!finalUrl.includes('directStream=1')) {
        finalUrl += '&directStream=1';
      }
      if (!finalUrl.includes('mediaIndex=0')) {
        finalUrl += '&mediaIndex=0';
      }

      // IMPORTANT: Set protocol to HLS instead of HTTP for better browser compatibility
      if (!finalUrl.includes('protocol=')) {
        finalUrl += '&protocol=hls';
      } else if (finalUrl.includes('protocol=http')) {
        // Replace http protocol with hls if it's already set
        finalUrl = finalUrl.replace('protocol=http', 'protocol=hls');
      }

      // Ensure proper video format for browser compatibility
      if (!finalUrl.includes('videoFormat=')) {
        finalUrl += '&videoFormat=h264';
      }
      if (!finalUrl.includes('audioFormat=')) {
        finalUrl += '&audioFormat=aac';
      }

      // Set quality parameters
      if (!finalUrl.includes('videoQuality=')) {
        finalUrl += '&videoQuality=100';
      }
      if (!finalUrl.includes('audioBoost=')) {
        finalUrl += '&audioBoost=100';
      }

      // Force transcoding for maximum compatibility
      finalUrl += '&fastSeek=1&session=plex-web-player';
    }

    return finalUrl;
  }

  private isValidStreamUrl(url: string): boolean {
    if (!url) {
      return false;
    }

    // Check if it's a Plex transcoding URL with the necessary parameters
    if (url.includes('/video/:/transcode/universal/start')) {
      return url.includes('X-Plex-Token') &&
        (url.includes('directPlay=0') || url.includes('directStream=1'));
    }

    // Check if it's a direct file URL with a common video format
    const videoExtensions = ['.mp4', '.webm', '.ogg', '.m3u8', '.mpd'];
    return videoExtensions.some(ext => url.toLowerCase().includes(ext));
  }

  // Get auth headers
  private getAuthHeaders(): HttpHeaders {
    const token = this.authService.getToken();
    return new HttpHeaders({
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    });
  }

  // Get all libraries
  getLibraries(): Observable<Library[]> {
    console.log('PlexService: Getting libraries');

    return this.http.get<any[]>(`${environment.apiUrl}/api/libraries`, this.getHttpOptions())
      .pipe(
        tap(response => console.log('Libraries response:', response)),
        map(response => {
          if (!Array.isArray(response)) {
            console.error('Unexpected response format:', response);
            return [];
          }
          return response as Library[];
        }),
        tap(libraries => {
          console.log('Processed libraries:', libraries);
          this.librariesSubject.next(libraries);
        }),
        catchError(this.handleError('getLibraries', []))
      );
  }

  // Get items from a library
  getLibraryItems(libraryKey: string): Observable<MediaItem[]> {
    console.log(`PlexService: Getting items for library ${libraryKey}`);

    return this.http.get<any[]>(
      `${environment.apiUrl}/api/library/${libraryKey}/items`,
      this.getHttpOptions()
    ).pipe(
      tap(response => console.log(`Raw library items response for ${libraryKey}:`, response)),
      map(items => {
        if (!Array.isArray(items)) {
          console.error('Unexpected response format for library items:', items);
          return [];
        }

        // Transform the items to MediaItem objects
        // The backend is now sending complete thumbnail URLs with auth tokens
        return items.map(item => ({
          ...item,
          // We no longer need to modify the thumbnail URL as the backend provides the full URL
          // Just make sure it exists or use a placeholder
          thumbnail: item.thumbnail || '/assets/images/placeholder.jpg'
        }));
      }),
      tap(items => {
        console.log(`Processed ${items.length} items for library ${libraryKey}`);
        // Log a few sample thumbnails for debugging
        if (items.length > 0) {
          console.log('Sample media items:');
          items.slice(0, 3).forEach((item, i) => {
            console.log(`Item ${i}: ${item.title} - Thumbnail: ${item.thumbnail}`);
          });
        }
      }),
      catchError(this.handleError('getLibraryItems', []))
    );
  }

  getMediaInfo(mediaKey: string): Observable<any> {
    console.log(`PlexService: Getting media info for path: ${mediaKey}`);

    return this.http.get<any>(
      `${environment.apiUrl}/api/media/${mediaKey}`,
      this.getHttpOptions()
    ).pipe(
      tap(response => console.log(`Media info response for ${mediaKey}:`, response)),
      map(item => {
        if (!item) {
          console.error('Unexpected empty response for media info');
          return null;
        }

        // Make sure thumbnail is properly set
        return {
          ...item,
          thumbnail: item.thumbnail || '/assets/images/placeholder.jpg'
        };
      }),
      catchError(error => {
        console.error(`Error getting media info for ${mediaKey}:`, error);

        // Check if error is 404
        if (error.status === 404) {
          console.error(`The endpoint /api/media/${mediaKey} was not found. Verify this endpoint exists on the backend.`);
        }

        // Check if error is CORS
        if (error.status === 0 && error.statusText === 'Unknown Error') {
          console.error('Possible CORS issue detected. Make sure the backend has proper CORS headers.');
        }

        return this.handleError('getMediaInfo', null)(error);
      })
    );
  }

  updateMediaPosition(mediaKey: string, position: number): Observable<any> {
    // Fix: Ensure mediaKey doesn't start with a slash
    const normalizedMediaKey = mediaKey.startsWith('/') ? mediaKey.substring(1) : mediaKey;

    return this.http.post(
      `${environment.apiUrl}/api/media/${normalizedMediaKey}/position`,
      { position },
      this.getHttpOptions()
    ).pipe(
      catchError(error => {
        console.error(`Error updating position for ${mediaKey}:`, error);
        return of(null);
      })
    );
  }

  // Get continue watching items
  getContinueWatching(): Observable<any[]> {
    console.log('PlexService: Getting continue watching');

    return this.http.get<any[]>(
      `${environment.apiUrl}/api/continue-watching`,
      this.getHttpOptions()
    ).pipe(
      tap(response => console.log('Continue watching response:', response)),
      map(items => {
        if (!Array.isArray(items)) {
          console.error('Unexpected response format for continue watching:', items);
          return [];
        }

        // Transform items
        return items.map(item => ({
          ...item,
          thumbnail: item.thumbnail || '/assets/images/placeholder.jpg'
        }));
      }),
      tap(items => {
        console.log(`Processed ${items.length} continue watching items`);
        this.continueWatchingSubject.next(items);
      }),
      catchError(this.handleError('getContinueWatching', []))
    );
  }

  // Get recently added items
  getRecentlyAdded(): Observable<any[]> {
    console.log('PlexService: Getting recently added');

    return this.http.get<any[]>(
      `${environment.apiUrl}/api/recently-added`,
      this.getHttpOptions()
    ).pipe(
      tap(response => console.log('Recently added response:', response)),
      map(items => {
        if (!Array.isArray(items)) {
          console.error('Unexpected response format for recently added:', items);
          return [];
        }

        // Transform items
        return items.map(item => ({
          ...item,
          thumbnail: item.thumbnail || '/assets/images/placeholder.jpg'
        }));
      }),
      tap(items => {
        console.log(`Processed ${items.length} recently added items`);
        this.recentlyAddedSubject.next(items);
      }),
      catchError(this.handleError('getRecentlyAdded', []))
    );
  }

  // Error handler
  private handleError<T>(operation = 'operation', result?: T) {
    return (error: HttpErrorResponse): Observable<T> => {
      console.error(`${operation} failed: ${error.message}`);
      console.error('Error details:', error);

      // Check for specific error types
      if (error.status === 401) {
        console.error('Authentication error - redirecting to login');
        // You might want to redirect to login here
      }

      // Return a default result or rethrow based on your needs
      return of(result as T);
    };
  }

  // Search for media across libraries
  searchMedia(query: string): Observable<MediaItem[]> {
    if (!this.authService.isAuthenticated()) {
      return of([]);
    }

    // In a real implementation, you'd have a proper search endpoint
    return this.http.get<MediaItem[]>(
      `${environment.apiUrl}/api/search?query=${encodeURIComponent(query)}`,
      this.getHttpOptions()
    ).pipe(
      tap(response => console.log(`Search response for "${query}":`, response)),
      map(items => {
        if (!Array.isArray(items)) {
          console.error('Unexpected response format for search:', items);
          return [];
        }

        // Transform items
        return items.map(item => ({
          ...item,
          thumbnail: item.thumbnail || '/assets/images/placeholder.jpg'
        }));
      }),
      catchError(error => {
        console.error(`Error searching for "${query}":`, error);

        // Fallback to filtering local data if API search fails
        return this.libraries$.pipe(
          switchMap(libraries => {
            if (libraries.length === 0) {
              return of([]);
            }

            // Just search the first library for this example
            return this.getLibraryItems(libraries[0].key).pipe(
              map(items => items.filter(item =>
                item.title.toLowerCase().includes(query.toLowerCase())
              ))
            );
          })
        );
      })
    );
  }

  // Refresh all data
  refreshData(): void {
    if (!this.authService.isAuthenticated()) {
      console.warn('Attempted to refresh data while not authenticated');
      return;
    }

    this.getServers().subscribe();
    this.getContinueWatching().subscribe();
    this.getRecentlyAdded().subscribe();
  }

  // Important: REMOVING the getPosterUrl method since we no longer need it
  // The backend now sends complete URLs with authentication tokens
  // This method was previously overriding the correct URLs with incorrect ones
}
