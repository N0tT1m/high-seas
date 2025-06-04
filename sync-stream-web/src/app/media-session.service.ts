// src/app/media-session.service.ts
import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { BehaviorSubject, Observable, of, Subject } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';
import { environment } from '../environments/environment';
import { MediaSession } from './media.model';
import { AuthService } from './auth.service';

interface SessionResponse {
  position: number;
  duration?: number;
  state?: string;
  lastClient?: string;
}

@Injectable({
  providedIn: 'root'
})
export class MediaSessionService {
  private activeSessionsSubject = new BehaviorSubject<MediaSession[]>([]);
  public activeSessions$ = this.activeSessionsSubject.asObservable();

  private currentSessionSubject = new BehaviorSubject<MediaSession | null>(null);
  public currentSession$ = this.currentSessionSubject.asObservable();

  // Subjects for remote control events
  private remotePlaySubject = new Subject<{ mediaKey: string, position: number }>();
  public remotePlay$ = this.remotePlaySubject.asObservable();

  private remotePauseSubject = new Subject<{ mediaKey: string, position: number }>();
  public remotePause$ = this.remotePauseSubject.asObservable();

  private remoteSeekSubject = new Subject<{ mediaKey: string, position: number }>();
  public remoteSeek$ = this.remoteSeekSubject.asObservable();

  private remoteStopSubject = new Subject<{ mediaKey: string, position: number }>();
  public remoteStop$ = this.remoteStopSubject.asObservable();

  constructor(
    private http: HttpClient,
    private authService: AuthService
  ) {
    // Try to load sessions from localStorage
    this.loadSavedSessions();
  }

  /**
   * Get HTTP options with auth token
   */
  private getHttpOptions() {
    const token = this.authService.getToken();
    return {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      })
    };
  }

  /**
   * Get the appropriate API endpoint for a mediaKey
   * This handles both regular mediaKeys and Plex-style paths
   */
  private getApiEndpoint(mediaKey: string): string {
    // Check if it's a Plex-style path
    if (mediaKey.startsWith('/library/metadata/') || mediaKey.startsWith('library/metadata/')) {
      // Extract the ID from the Plex path
      const normalizedPath = mediaKey.startsWith('/') ? mediaKey.substring(1) : mediaKey;
      // This will preserve the entire path structure
      return `${environment.apiUrl}/api/media/${normalizedPath}/position`;
    }

    // For other types of mediaKeys, use the regular endpoint
    // Remove any leading slash
    const normalizedMediaKey = mediaKey.startsWith('/') ? mediaKey.substring(1) : mediaKey;
    return `${environment.apiUrl}/api/media/${normalizedMediaKey}/position`;
  }

  /**
   * Update a session on the server and locally
   */
  updateSession(session: Partial<MediaSession>): Observable<any> {
    if (!session.mediaKey) {
      console.error('Cannot update session without mediaKey');
      return of(null);
    }

    // Update local copy first
    this.updateLocalSession(session);

    // Get the appropriate API endpoint
    const endpoint = this.getApiEndpoint(session.mediaKey);
    console.log(`Updating session at endpoint: ${endpoint}`);

    // Then send to server with auth headers
    return this.http.post(
      endpoint,
      {
        position: session.position,
        duration: session.duration,
        state: session.state,
        clientId: session.clientId
      },
      this.getHttpOptions()
    ).pipe(
      catchError(error => {
        console.error('Error updating session:', error);
        return of(null);
      })
    );
  }

  /**
   * Get a session from the server
   */
  getSession(mediaKey: string): Observable<SessionResponse> {
    // Get the appropriate API endpoint
    const endpoint = this.getApiEndpoint(mediaKey);
    console.log(`Getting session from endpoint: ${endpoint} with auth token`);

    return this.http.get<SessionResponse>(endpoint, this.getHttpOptions()).pipe(
      tap((session: SessionResponse) => {
        console.log(`Session response:`, session);
        // Fixed: Cast the state to a valid MediaSession state type
        const validState = this.getValidState(session.state);

        // Merge with local sessions
        this.updateLocalSession({
          mediaKey: mediaKey,
          position: session.position,
          duration: session.duration || 0,
          state: validState,
          lastClient: session.lastClient
        });
      }),
      catchError(error => {
        console.error('Error getting session:', error);
        return of({
          position: 0,
          duration: 0,
          state: 'stopped'
        } as SessionResponse);
      })
    );
  }

  /**
   * Helper method to ensure state is a valid MediaSession state
   */
  private getValidState(state?: string): 'playing' | 'paused' | 'stopped' {
    if (state === 'playing' || state === 'paused') {
      return state;
    }
    return 'stopped';
  }

  /**
   * Update a session locally (doesn't send to server)
   */
  updateLocalSession(sessionUpdate: Partial<MediaSession>): void {
    if (!sessionUpdate.mediaKey) return;

    const currentSessions = this.activeSessionsSubject.value;
    const existingSessionIndex = currentSessions.findIndex(
      s => s.mediaKey === sessionUpdate.mediaKey
    );

    // Make sure the state is a valid value
    const validState = sessionUpdate.state ?
      this.getValidState(sessionUpdate.state) :
      undefined;

    if (existingSessionIndex >= 0) {
      // Update existing session
      const updatedSessions = [...currentSessions];
      updatedSessions[existingSessionIndex] = {
        ...updatedSessions[existingSessionIndex],
        ...sessionUpdate,
        state: validState || updatedSessions[existingSessionIndex].state,
        lastUpdated: new Date()
      };
      this.activeSessionsSubject.next(updatedSessions);

      // If this is the current session, update it too
      if (this.currentSessionSubject.value?.mediaKey === sessionUpdate.mediaKey) {
        this.currentSessionSubject.next({
          ...this.currentSessionSubject.value,
          ...sessionUpdate,
          state: validState || this.currentSessionSubject.value.state,
          lastUpdated: new Date()
        });
      }
    } else {
      // Create new session
      const newSession: MediaSession = {
        mediaKey: sessionUpdate.mediaKey!,
        position: sessionUpdate.position || 0,
        duration: sessionUpdate.duration || 0,
        state: validState || 'stopped',
        clientId: sessionUpdate.clientId,
        metadata: sessionUpdate.metadata || {},
        lastUpdated: new Date()
      };

      this.activeSessionsSubject.next([...currentSessions, newSession]);
    }

    // Save to localStorage for persistence
    this.saveSessions();
  }

  /**
   * Set the current active session
   */
  setCurrentSession(mediaKey: string): void {
    const session = this.activeSessionsSubject.value.find(
      s => s.mediaKey === mediaKey
    );

    if (session) {
      this.currentSessionSubject.next(session);
    } else {
      // Create a new blank session
      this.currentSessionSubject.next({
        mediaKey: mediaKey,
        position: 0,
        duration: 0,
        state: 'stopped',
        metadata: {},
        lastUpdated: new Date()
      });
    }
  }

  /**
   * Clear the current session
   */
  clearCurrentSession(): void {
    this.currentSessionSubject.next(null);
  }

  /**
   * Handle remote position update from WebSocket
   */
  handleRemotePositionUpdate(mediaKey: string, position: number): void {
    this.updateLocalSession({
      mediaKey,
      position
    });

    this.remoteSeekSubject.next({ mediaKey, position });
  }

  /**
   * Handle remote play event from WebSocket
   */
  handleRemotePlayEvent(mediaKey: string, position: number): void {
    this.updateLocalSession({
      mediaKey,
      position,
      state: 'playing'
    });

    this.remotePlaySubject.next({ mediaKey, position });
  }

  /**
   * Handle remote pause event from WebSocket
   */
  handleRemotePauseEvent(mediaKey: string, position: number): void {
    this.updateLocalSession({
      mediaKey,
      position,
      state: 'paused'
    });

    this.remotePauseSubject.next({ mediaKey, position });
  }

  /**
   * Handle remote stop event from WebSocket
   */
  handleRemoteStopEvent(mediaKey: string, position: number): void {
    this.updateLocalSession({
      mediaKey,
      position,
      state: 'stopped'
    });

    this.remoteStopSubject.next({ mediaKey, position });
  }

  /**
   * Import a list of sessions from the server
   */
  importRemoteSessions(sessions: MediaSession[]): void {
    if (!Array.isArray(sessions)) return;

    const currentSessions = this.activeSessionsSubject.value;
    let updatedSessions = [...currentSessions];

    sessions.forEach(remoteSession => {
      const existingIndex = updatedSessions.findIndex(
        s => s.mediaKey === remoteSession.mediaKey
      );

      if (existingIndex >= 0) {
        // Update if the remote is newer
        const existingSession = updatedSessions[existingIndex];
        const existingDate = new Date(existingSession.lastUpdated);
        const remoteDate = new Date(remoteSession.lastUpdated);

        if (remoteDate > existingDate) {
          updatedSessions[existingIndex] = {
            ...existingSession,
            position: remoteSession.position,
            duration: remoteSession.duration || existingSession.duration,
            state: remoteSession.state || existingSession.state,
            lastUpdated: remoteSession.lastUpdated
          };
        }
      } else {
        // Add new session
        updatedSessions.push(remoteSession);
      }
    });

    this.activeSessionsSubject.next(updatedSessions);
    this.saveSessions();
  }

  /**
   * Save all sessions to localStorage
   */
  private saveSessions(): void {
    try {
      localStorage.setItem(
        'plex_sessions',
        JSON.stringify(this.activeSessionsSubject.value)
      );
    } catch (error) {
      console.error('Error saving sessions to localStorage:', error);
    }
  }

  /**
   * Load saved sessions from localStorage
   */
  private loadSavedSessions(): void {
    try {
      const savedSessions = localStorage.getItem('plex_sessions');
      if (savedSessions) {
        const sessions = JSON.parse(savedSessions);
        if (Array.isArray(sessions)) {
          this.activeSessionsSubject.next(sessions);
        }
      }
    } catch (error) {
      console.error('Error loading sessions from localStorage:', error);
    }
  }
}
