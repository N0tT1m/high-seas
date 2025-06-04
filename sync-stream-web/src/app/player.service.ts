// Updated PlayerService to support direct URLs
import { Injectable, OnDestroy } from '@angular/core';
import { BehaviorSubject, Subject, Observable, interval, Subscription, of } from 'rxjs';
import { takeUntil, filter } from 'rxjs/operators';
import { MediaSessionService } from './media-session.service';
import { SyncService } from './sync.service';
import { PlexService } from './plex.service';
import { PlayerState } from './media.model';
import { environment } from '../environments/environment';

/**
 * Handles video playback and synchronization with the media session
 */
@Injectable({
  providedIn: 'root'
})
export class PlayerService implements OnDestroy {
  private videoElement: HTMLVideoElement | null = null;
  private progressInterval: Subscription | null = null;
  private updateRate = environment.progressUpdateRate || 1000; // Update progress every second
  private syncInterval = environment.syncInterval || 10000; // Sync with server every 10 seconds
  private lastSyncTime = 0;
  private destroy$ = new Subject<void>();

  private stateSubject = new BehaviorSubject<PlayerState>({
    mediaKey: null,
    isPlaying: false,
    isPaused: false,
    isStopped: true,
    isBuffering: false,
    isError: false,
    position: 0,
    duration: 0,
    volume: 1,
    muted: false,
    title: '',
    subtitle: '',
    streamUrl: null,
    error: null
  });

  public state$ = this.stateSubject.asObservable();

  constructor(
    private mediaSessionService: MediaSessionService,
    private syncService: SyncService,
    private plexService: PlexService
  ) {
    // Listen for remote control events
    this.setupRemoteControlListeners();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.stopProgressTracking();
  }

  /**
   * Initialize the player with a video element
   */
  initialize(videoElement: HTMLVideoElement): void {
    this.videoElement = videoElement;
    this.setupVideoListeners();
    console.log('PlayerService initialized with video element');
  }

  /**
   * Setup video element event listeners
   */
  private setupVideoListeners(): void {
    if (!this.videoElement) {
      console.error('No video element to set up listeners for');
      return;
    }

    console.log('Setting up video listeners');

    this.videoElement.addEventListener('play', () => this.handlePlay());
    this.videoElement.addEventListener('pause', () => this.handlePause());
    this.videoElement.addEventListener('ended', () => this.handleEnded());
    this.videoElement.addEventListener('timeupdate', () => this.handleTimeUpdate());
    this.videoElement.addEventListener('durationchange', () => this.handleDurationChange());
    this.videoElement.addEventListener('waiting', () => this.handleBuffering(true));
    this.videoElement.addEventListener('canplay', () => this.handleBuffering(false));
    this.videoElement.addEventListener('error', (e) => this.handleError(e));
    this.videoElement.addEventListener('volumechange', () => this.handleVolumeChange());
  }

  /**
   * Setup listeners for remote control events
   */
  private setupRemoteControlListeners(): void {
    // Play events from other clients
    this.mediaSessionService.remotePlay$
      .pipe(takeUntil(this.destroy$))
      .subscribe(event => {
        if (this.stateSubject.value.mediaKey === event.mediaKey && this.videoElement) {
          this.videoElement.currentTime = event.position / 1000; // Convert to seconds
          this.videoElement.play();
        }
      });

    // Pause events from other clients
    this.mediaSessionService.remotePause$
      .pipe(takeUntil(this.destroy$))
      .subscribe(event => {
        if (this.stateSubject.value.mediaKey === event.mediaKey && this.videoElement) {
          this.videoElement.currentTime = event.position / 1000; // Convert to seconds
          this.videoElement.pause();
        }
      });

    // Seek events from other clients
    this.mediaSessionService.remoteSeek$
      .pipe(takeUntil(this.destroy$))
      .subscribe(event => {
        if (this.stateSubject.value.mediaKey === event.mediaKey && this.videoElement) {
          this.videoElement.currentTime = event.position / 1000; // Convert to seconds
        }
      });

    // Stop events from other clients
    this.mediaSessionService.remoteStop$
      .pipe(takeUntil(this.destroy$))
      .subscribe(event => {
        if (this.stateSubject.value.mediaKey === event.mediaKey) {
          this.stop();
        }
      });
  }

  /**
   * Load a media item
   * @param mediaKey The unique identifier for the media
   * @param title Media title
   * @param subtitle Optional subtitle (e.g., TV Show)
   * @param streamUrl The direct URL to stream the media
   */
  loadMedia(mediaKey: string, title: string, subtitle: string = '', streamUrl: string = ''): Observable<boolean> {
    console.log(`Loading media: ${title} (${mediaKey})`);
    console.log(`Stream URL: ${streamUrl}`);

    // Set the current session first
    this.mediaSessionService.setCurrentSession(mediaKey);

    return new Observable<boolean>(observer => {
      if (!this.videoElement) {
        console.error('Video element not initialized');
        this.updateState({
          isError: true,
          error: 'Video element not initialized'
        });
        observer.next(false);
        observer.complete();
        return;
      }

      // Use the provided stream URL directly
      if (streamUrl) {
        this.updateState({
          mediaKey,
          title,
          subtitle,
          streamUrl,
          isBuffering: true,
          isError: false,
          error: null
        });

        console.log(`Setting video source to ${streamUrl}`);
        this.videoElement.src = streamUrl;
        this.videoElement.load();

        // Get the saved position
        this.mediaSessionService.getSession(mediaKey).subscribe({
          next: session => {
            if (session && session.position > 0) {
              // Resume from the saved position
              this.videoElement!.currentTime = session.position / 1000; // Convert to seconds
              console.log(`Resuming playback from ${session.position}ms`);
            }
            observer.next(true);
            observer.complete();
          },
          error: error => {
            console.error('Error getting session:', error);
            observer.next(true);
            observer.complete();
          }
        });
      } else {
        // Fallback to using PlexService if no direct URL is provided
        console.log('No direct stream URL provided, fetching from PlexService');
        this.plexService.getStreamUrl(mediaKey).subscribe({
          next: fetchedStreamUrl => {
            if (!fetchedStreamUrl) {
              console.error('Failed to get stream URL from PlexService');
              this.updateState({
                isError: true,
                error: 'Failed to get stream URL'
              });
              observer.next(false);
              observer.complete();
              return;
            }

            this.updateState({
              mediaKey,
              title,
              subtitle,
              streamUrl: fetchedStreamUrl,
              isBuffering: true,
              isError: false,
              error: null
            });

            console.log(`Setting video source to ${fetchedStreamUrl}`);
            this.videoElement!.src = fetchedStreamUrl;
            this.videoElement!.load();

            // Get the saved position
            this.mediaSessionService.getSession(mediaKey).subscribe({
              next: session => {
                if (session && session.position > 0) {
                  // Resume from the saved position
                  this.videoElement!.currentTime = session.position / 1000; // Convert to seconds
                  console.log(`Resuming playback from ${session.position}ms`);
                }
                observer.next(true);
                observer.complete();
              },
              error: error => {
                console.error('Error getting session:', error);
                observer.next(true);
                observer.complete();
              }
            });
          },
          error: error => {
            console.error('Error getting stream URL:', error);
            this.updateState({
              isError: true,
              error: 'Failed to get stream URL: ' + (error.message || 'Unknown error')
            });
            observer.next(false);
            observer.complete();
          }
        });
      }
    });
  }

  /**
   * Play the loaded media
   */
  play(): void {
    if (!this.videoElement) {
      console.error('Cannot play: video element not initialized');
      return;
    }

    if (!this.stateSubject.value.streamUrl) {
      console.error('Cannot play: no stream URL set');
      this.updateState({
        isError: true,
        error: 'No media loaded'
      });
      return;
    }

    // Wrap the play call in a user interaction check
    const playPromise = this.videoElement.play();

    if (playPromise !== undefined) {
      playPromise.catch(error => {
        if (error.name === 'NotAllowedError') {
          console.log('Autoplay prevented due to lack of user interaction - waiting for user interaction');
          // Store that we want to play as soon as possible
          this.updateState({
            isPendingUserInteraction: true
          });

          // Add a one-time event listener for user interaction
          const userInteractionHandler = () => {
            this.play();
            // Remove all the event listeners after first interaction
            ['click', 'touchend', 'keydown'].forEach(event => {
              document.removeEventListener(event, userInteractionHandler);
            });
          };

          // Add listeners for common user interactions
          ['click', 'touchend', 'keydown'].forEach(event => {
            document.addEventListener(event, userInteractionHandler, { once: true });
          });
        } else {
          console.error('Error playing video:', error);
          this.updateState({
            isError: true,
            error: 'Failed to play video: ' + error.message
          });
        }
      });
    }
  }

  /**
   * Pause the media
   */
  pause(): void {
    if (this.videoElement) {
      this.videoElement.pause();
    }
  }

  /**
   * Toggle play/pause
   */
  togglePlayPause(): void {
    if (this.stateSubject.value.isPlaying) {
      this.pause();
    } else {
      this.play();
    }
  }

  /**
   * Stop playback
   */
  stop(): void {
    if (this.videoElement) {
      this.videoElement.pause();
      this.videoElement.currentTime = 0;
      this.videoElement.src = '';
    }

    this.stopProgressTracking();

    // Update state
    this.updateState({
      isPlaying: false,
      isPaused: false,
      isStopped: true,
      position: 0,
      streamUrl: null
    });

    // Update session
    if (this.stateSubject.value.mediaKey) {
      this.syncService.sendStopEvent(
        this.stateSubject.value.mediaKey,
        this.stateSubject.value.position
      );

      this.mediaSessionService.updateSession({
        mediaKey: this.stateSubject.value.mediaKey,
        position: this.stateSubject.value.position,
        state: 'stopped',
        clientId: this.syncService.getClientId()
      }).subscribe();
    }

    this.mediaSessionService.clearCurrentSession();
  }

  /**
   * Seek to a specific position (in milliseconds)
   */
  seek(position: number): void {
    if (this.videoElement) {
      this.videoElement.currentTime = position / 1000; // Convert to seconds

      // Update the media session immediately
      const mediaKey = this.stateSubject.value.mediaKey;
      if (mediaKey) {
        this.mediaSessionService.updateLocalSession({
          mediaKey,
          position
        });

        // Send position update to other clients
        this.syncService.updatePosition(mediaKey, position);
      }
    }
  }

  /**
   * Seek forward by a specific amount (in seconds)
   */
  seekForward(seconds: number = 10): void {
    if (this.videoElement) {
      const newTime = this.videoElement.currentTime + seconds;
      this.videoElement.currentTime = Math.min(newTime, this.videoElement.duration);
    }
  }

  /**
   * Seek backward by a specific amount (in seconds)
   */
  seekBackward(seconds: number = 10): void {
    if (this.videoElement) {
      const newTime = this.videoElement.currentTime - seconds;
      this.videoElement.currentTime = Math.max(newTime, 0);
    }
  }

  /**
   * Set volume (0-1)
   */
  setVolume(volume: number): void {
    if (this.videoElement) {
      this.videoElement.volume = Math.max(0, Math.min(1, volume));
    }
  }

  /**
   * Toggle mute
   */
  toggleMute(): void {
    if (this.videoElement) {
      this.videoElement.muted = !this.videoElement.muted;
    }
  }

  /**
   * Handle play event
   */
  private handlePlay(): void {
    this.updateState({
      isPlaying: true,
      isPaused: false,
      isStopped: false
    });

    this.startProgressTracking();

    // Update session and send play event
    if (this.stateSubject.value.mediaKey) {
      this.syncService.sendPlayEvent(
        this.stateSubject.value.mediaKey,
        this.stateSubject.value.position
      );

      this.mediaSessionService.updateSession({
        mediaKey: this.stateSubject.value.mediaKey,
        position: this.stateSubject.value.position,
        state: 'playing',
        clientId: this.syncService.getClientId()
      }).subscribe();
    }
  }

  /**
   * Handle pause event
   */
  private handlePause(): void {
    // Don't mark as paused if we're at the end (this means it's actually stopped)
    if (this.videoElement &&
      this.videoElement.currentTime >= this.videoElement.duration - 0.5) {
      this.handleEnded();
      return;
    }

    this.updateState({
      isPlaying: false,
      isPaused: true,
      isStopped: false
    });

    this.stopProgressTracking();

    // Update session and send pause event
    if (this.stateSubject.value.mediaKey) {
      this.syncService.sendPauseEvent(
        this.stateSubject.value.mediaKey,
        this.stateSubject.value.position
      );

      this.mediaSessionService.updateSession({
        mediaKey: this.stateSubject.value.mediaKey,
        position: this.stateSubject.value.position,
        state: 'paused',
        clientId: this.syncService.getClientId()
      }).subscribe();
    }
  }

  /**
   * Handle video end
   */
  private handleEnded(): void {
    this.updateState({
      isPlaying: false,
      isPaused: false,
      isStopped: true,
      position: this.stateSubject.value.duration
    });

    this.stopProgressTracking();

    // Update session and send stop event
    if (this.stateSubject.value.mediaKey) {
      this.syncService.sendStopEvent(
        this.stateSubject.value.mediaKey,
        this.stateSubject.value.duration
      );

      this.mediaSessionService.updateSession({
        mediaKey: this.stateSubject.value.mediaKey,
        position: this.stateSubject.value.duration,
        state: 'stopped',
        clientId: this.syncService.getClientId()
      }).subscribe();
    }
  }

  /**
   * Handle time update
   */
  private handleTimeUpdate(): void {
    if (this.videoElement) {
      this.updateState({
        position: Math.floor(this.videoElement.currentTime * 1000) // Convert to milliseconds
      });
    }
  }

  /**
   * Handle duration change
   */
  private handleDurationChange(): void {
    if (this.videoElement) {
      this.updateState({
        duration: Math.floor(this.videoElement.duration * 1000) // Convert to milliseconds
      });
    }
  }

  /**
   * Handle buffering
   */
  private handleBuffering(isBuffering: boolean): void {
    this.updateState({ isBuffering });
  }

  /**
   * Handle video error
   */
  private handleError(event: Event): void {
    let errorMessage = 'Unknown video error';
    if (this.videoElement && this.videoElement.error) {
      switch (this.videoElement.error.code) {
        case MediaError.MEDIA_ERR_ABORTED:
          errorMessage = 'You aborted the video playback.';
          break;
        case MediaError.MEDIA_ERR_NETWORK:
          errorMessage = 'A network error caused the video download to fail.';
          break;
        case MediaError.MEDIA_ERR_DECODE:
          errorMessage = 'The video playback was aborted due to a corruption problem.';
          break;
        case MediaError.MEDIA_ERR_SRC_NOT_SUPPORTED:
          errorMessage = 'The video format is not supported.';
          break;
      }
    }

    this.updateState({
      isPlaying: false,
      isPaused: false,
      isStopped: true,
      isBuffering: false,
      isError: true,
      error: errorMessage
    });

    this.stopProgressTracking();
  }

  /**
   * Handle volume change
   */
  private handleVolumeChange(): void {
    if (this.videoElement) {
      this.updateState({
        volume: this.videoElement.volume,
        muted: this.videoElement.muted
      });
    }
  }

  /**
   * Start tracking playback progress
   */
  private startProgressTracking(): void {
    this.stopProgressTracking();

    this.progressInterval = interval(this.updateRate)
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        const now = Date.now();
        const mediaKey = this.stateSubject.value.mediaKey;

        if (mediaKey && this.stateSubject.value.isPlaying) {
          // Update local session
          this.mediaSessionService.updateLocalSession({
            mediaKey,
            position: this.stateSubject.value.position,
            duration: this.stateSubject.value.duration,
            state: 'playing'
          });

          // Sync with server less frequently
          if (now - this.lastSyncTime >= this.syncInterval) {
            this.syncService.updatePosition(mediaKey, this.stateSubject.value.position);
            this.lastSyncTime = now;
          }
        }
      });
  }

  /**
   * Stop tracking playback progress
   */
  private stopProgressTracking(): void {
    if (this.progressInterval) {
      this.progressInterval.unsubscribe();
      this.progressInterval = null;
    }
  }

  /**
   * Update player state
   */
  private updateState(stateUpdate: Partial<PlayerState>): void {
    this.stateSubject.next({
      ...this.stateSubject.value,
      ...stateUpdate
    });
  }
}
