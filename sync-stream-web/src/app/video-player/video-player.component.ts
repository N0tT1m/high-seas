// Enhanced video-player.component.ts for Plex streaming
import { Component, OnInit, AfterViewInit, OnDestroy, ViewChild, ElementRef, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { PlayerService } from '../player.service';
import { MediaSessionService } from '../media-session.service';
import { MediaItem } from '../media.model';
import { MatIconModule } from '@angular/material/icon';
import { AuthService } from '../auth.service';
import {formatPlexStreamUrl, isValidStreamUrl, createDirectStreamUrl} from '../utils/plex-stream.utils';

// Define HLS interface for TypeScript
interface HlsEvent {
  type: string;
  payload: any;
  fatal: boolean;
}

interface HlsInstance {
  loadSource(url: string): void;
  attachMedia(element: HTMLVideoElement): void;
  on(event: string, callback: (event: Event, data: HlsEvent) => void): void;
  destroy(): void;
  recoverMediaError(): void;
}

interface HlsConstructor {
  new (config?: any): HlsInstance;
  isSupported(): boolean;
  Events: { [key: string]: string };
  ErrorTypes: { [key: string]: string };
}

// Declare global HLS availability
declare global {
  interface Window {
    Hls?: HlsConstructor;
  }
}

// In video-player.component.ts - Add this method before the component definition

/**
 * Helper function to check if HLS.js is available
 * @returns true if HLS.js is available, false otherwise
 */
function isHlsJsAvailable(): boolean {
  return typeof window !== 'undefined' &&
    typeof window.Hls !== 'undefined' &&
    window.Hls.isSupported();
}

/**
 * Helper function to check if the browser has native HLS support
 * @returns true if browser has native HLS support, false otherwise
 */
function hasNativeHlsSupport(): boolean {
  if (typeof document === 'undefined') return false;

  // Create a test video element
  const video = document.createElement('video');

  // Check if the video element can play HLS
  return video.canPlayType('application/vnd.apple.mpegurl') !== '' ||
    video.canPlayType('application/x-mpegURL') !== '';
}


@Component({
  selector: 'app-video-player',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  providers: [],
  templateUrl: './video-player.component.html',
  styles: [`
    .video-player-container {
      position: relative;
      width: 100%;
      height: 100%;
      background-color: #000;
      overflow: hidden;
    }

    /* Add your other styles here */
  `]
})
export class VideoPlayerComponent implements OnInit, AfterViewInit, OnDestroy {
  @ViewChild('videoElement') videoElement!: ElementRef<HTMLVideoElement>;
  @Input() autoplay = false;

  // Internal state
  playing = false;
  currentTime = 0;
  duration = 0;
  volume = 1;
  muted = false;
  fullscreen = false;
  showControls = true;
  loading = false;
  error = false;
  errorMessage = '';
  progress = 0;
  buffered = 0;
  hideControlsTimer: any;

  // Track HLS.js instance if needed
  private hls: HlsInstance | null = null;

  // UI state
  showVolumeSlider = false;

  private destroy$ = new Subject<void>();

  constructor(
    private playerService: PlayerService,
    private mediaSessionService: MediaSessionService,
    private authService: AuthService
  ) {}

  ngOnInit(): void {
    // Subscribe to player state changes
    this.playerService.state$
      .pipe(takeUntil(this.destroy$))
      .subscribe(state => {
        this.playing = state.isPlaying;
        this.currentTime = state.position;
        this.duration = state.duration;
        this.volume = state.volume;
        this.muted = state.muted;
        this.loading = state.isBuffering;
        this.error = state.isError;

        if (state.error) {
          this.errorMessage = state.error;
        }

        // Update progress percentage
        if (this.duration > 0) {
          this.progress = (this.currentTime / this.duration) * 100;
        } else {
          this.progress = 0;
        }
      });
  }

  ngAfterViewInit(): void {
    // Initialize the player service with our video element
    if (this.videoElement && this.videoElement.nativeElement) {
      this.playerService.initialize(this.videoElement.nativeElement);
      console.log('Video player initialized with video element');

      // Set up event listeners for debugging
      this.videoElement.nativeElement.addEventListener('error', (e) => {
        console.error('Video element error:', e);
        console.error('Video error details:', this.videoElement.nativeElement.error);
      });

      this.videoElement.nativeElement.addEventListener('loadstart', () => {
        console.log('Video loadstart event');
      });

      this.videoElement.nativeElement.addEventListener('loadedmetadata', () => {
        console.log('Video loadedmetadata event');
      });

      this.videoElement.nativeElement.addEventListener('canplay', () => {
        console.log('Video canplay event');
      });
    } else {
      console.error('Video element not available in VideoPlayerComponent');
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.clearHideControlsTimer();

    // Clean up HLS.js if it was used
    if (this.hls) {
      this.hls.destroy();
    }
  }

  // Update the loadMedia method in video-player.component.ts to improve stream detection

  loadMedia(media: MediaItem): void {
    console.log('Loading media in VideoPlayerComponent:', media);

    if (!media) {
      console.error('No media provided to loadMedia');
      this.errorMessage = 'No media to play';
      this.error = true;
      return;
    }

    if (!media.streamUrl) {
      console.error('Media has no streamUrl:', media);
      this.errorMessage = 'No stream URL available for this media';
      this.error = true;
      return;
    }

    // Get auth token for formatting the URL
    const token = this.authService.getToken();

    // Format the stream URL to ensure it works in browsers
    const formattedUrl = this.formatPlexStreamUrl(media.streamUrl, token!);
    console.log('Using formatted stream URL:', formattedUrl);

    if (!this.isValidStreamUrl(formattedUrl)) {
      console.warn('Stream URL may not be valid for direct browser playback:', formattedUrl);
    }

    // ALWAYS use HLS.js for Plex streams to avoid format errors
    // This is crucial for universal/transcode URLs from Plex
    if (formattedUrl.includes('/video/:/transcode/universal/start')) {
      this.setupHlsStream(formattedUrl, media);
    }
    // Use HLS.js for .m3u8 files as well
    else if (formattedUrl.includes('.m3u8')) {
      this.setupHlsStream(formattedUrl, media);
    }
    // Only use standard playback for direct file URLs
    else {
      this.loadStandardStream(formattedUrl, media);
    }
  }

  // Then update the setupHlsStream method to handle fallbacks better:

  private setupHlsStream(streamUrl: string, media: MediaItem): void {
    console.log('Setting up HLS stream playback');

    // First check if HLS.js is available
    if (isHlsJsAvailable()) {
      console.log('Using HLS.js for playback');

      try {
        // Clean up any existing HLS instance
        if (this.hls) {
          this.hls.destroy();
        }

        // Create new HLS instance
        if (window.Hls) {  // Add null check here
          const Hls = window.Hls;
          this.hls = new Hls({
            enableWorker: true,
            lowLatencyMode: false,
            backBufferLength: 90,
            // Add additional HLS config for better compatibility
            maxBufferLength: 30,
            maxMaxBufferLength: 600
          });

          this.hls.loadSource(streamUrl);
          this.hls.attachMedia(this.videoElement.nativeElement);

          // Use string indexing for Events
          this.hls.on(Hls.Events['MANIFEST_PARSED'], (_event: Event, _data: HlsEvent) => {
            console.log('HLS manifest parsed, can play now');

            // Load media into player service for tracking
            this.playerService.loadMedia(
              media.key,
              media.title,
              media.type === 'show' ? 'TV Show' : media.type === 'movie' ? 'Movie' : media.type,
              streamUrl
            ).subscribe({
              next: (success) => {
                console.log('Media loaded successfully in player service:', success);
                if (success && this.autoplay) {
                  this.play();
                }
              },
              error: (error) => {
                console.error('Error loading media in player service:', error);
                this.errorMessage = 'Failed to load media: ' + (error.message || 'Unknown error');
                this.error = true;
              }
            });
          });

          // Use string indexing for Events and ErrorTypes
          this.hls.on(Hls.Events['ERROR'], (_event: Event, data: HlsEvent) => {
            console.error('HLS error:', data);
            if (data.fatal) {
              if (Hls.ErrorTypes) {  // Add null check here
                switch(data.type) {
                  case Hls.ErrorTypes['NETWORK_ERROR']:
                    console.error('HLS fatal network error');
                    this.errorMessage = 'Network error while loading video. Please check your internet connection.';
                    this.error = true;
                    break;
                  case Hls.ErrorTypes['MEDIA_ERROR']:
                    console.error('HLS media error - trying to recover');
                    this.hls?.recoverMediaError();
                    break;
                  default:
                    console.error('HLS fatal error, cannot recover');
                    this.errorMessage = 'Cannot play this video format. Try a different media file.';
                    this.error = true;
                    this.hls?.destroy();
                    break;
                }
              } else {
                console.error('HLS.ErrorTypes is undefined');
                this.errorMessage = 'Video playback error';
                this.error = true;
              }
            }
          });
        } else {
          console.error('HLS.js is not available in window object');
          this.tryFallbackPlayback(streamUrl, media);
        }
      } catch (error) {
        console.error('Error initializing HLS.js:', error);
        this.tryFallbackPlayback(streamUrl, media);
      }
    } else {
      this.tryFallbackPlayback(streamUrl, media);
    }
  }

  /**
   * Try fallback playback methods when HLS.js isn't available
   */
  private tryFallbackPlayback(streamUrl: string, media: MediaItem): void {
    // Check for native HLS support (Safari)
    if (hasNativeHlsSupport()) {
      console.log('Using native HLS support');
      this.loadStandardStream(streamUrl, media);
    } else {
      // For Plex streams, we can try to change the protocol to HTTP
      if (streamUrl.includes('/video/:/transcode/universal/start') && streamUrl.includes('protocol=hls')) {
        console.log('Falling back to HTTP streaming');
        const httpUrl = streamUrl.replace('protocol=hls', 'protocol=http');
        this.loadStandardStream(httpUrl, media);
      } else {
        console.error('HLS is not supported in this browser and no fallback is available');
        this.errorMessage = 'Your browser does not support this video format. Please try using a modern browser like Chrome or Safari.';
        this.error = true;
      }
    }
  }

  /**
   * Load a standard video stream (non-HLS)
   */
  private loadStandardStream(streamUrl: string, media: MediaItem): void {
    console.log('Using standard video playback for URL:', streamUrl);

    // Load media into player service
    this.playerService.loadMedia(
      media.key,
      media.title,
      media.type === 'show' ? 'TV Show' : media.type === 'movie' ? 'Movie' : media.type,
      streamUrl
    ).subscribe({
      next: (success) => {
        console.log('Media loaded successfully:', success);
        if (success && this.autoplay) {
          this.play();
        }
      },
      error: (error) => {
        console.error('Error loading media:', error);
        this.errorMessage = 'Failed to load media: ' + (error.message || 'Unknown error');
        this.error = true;
      }
    });
  }

  /**
   * Play/pause toggle
   */
  togglePlayPause(): void {
    this.playerService.togglePlayPause();
  }

  /**
   * Play the video
   */
  play(): void {
    this.playerService.play();
  }

  /**
   * Pause the video
   */
  pause(): void {
    this.playerService.pause();
  }

  /**
   * Stop the video
   */
  stop(): void {
    this.playerService.stop();

    // Clean up HLS if it was used
    if (this.hls) {
      this.hls.destroy();
      this.hls = null;
    }
  }

  /**
   * Toggle volume mute
   */
  toggleMute(): void {
    this.playerService.toggleMute();
  }

  /**
   * Seek to a specific position
   */
  seek(event: MouseEvent): void {
    const progressBar = event.currentTarget as HTMLElement;
    const rect = progressBar.getBoundingClientRect();
    const percentage = (event.clientX - rect.left) / rect.width;
    const position = percentage * this.duration;

    this.playerService.seek(position);
  }

  /**
   * Set volume
   */
  setVolume(event: Event): void {
    const input = event.target as HTMLInputElement;
    const volume = parseFloat(input.value);
    this.playerService.setVolume(volume);
  }

  /**
   * Skip forward
   */
  skipForward(): void {
    this.playerService.seekForward(30); // 30 seconds
  }

  /**
   * Skip backward
   */
  skipBackward(): void {
    this.playerService.seekBackward(10); // 10 seconds
  }

  /**
   * Toggle fullscreen
   */
  toggleFullscreen(): void {
    const playerElement = this.videoElement?.nativeElement?.parentElement;

    if (playerElement) {
      if (!document.fullscreenElement) {
        // Enter fullscreen
        playerElement.requestFullscreen?.() ||
        (playerElement as any).webkitRequestFullscreen?.() ||
        (playerElement as any).msRequestFullscreen?.();
        this.fullscreen = true;
      } else {
        // Exit fullscreen
        document.exitFullscreen?.() ||
        (document as any).webkitExitFullscreen?.() ||
        (document as any).msExitFullscreen?.();
        this.fullscreen = false;
      }
    }
  }

  /**
   * Show controls when mouse moves over player
   */
  showPlayerControls(): void {
    this.showControls = true;
    this.clearHideControlsTimer();

    // Hide controls after 3 seconds of inactivity
    this.hideControlsTimer = setTimeout(() => {
      if (this.playing) {
        this.showControls = false;
      }
    }, 3000);
  }

  /**
   * Clear the hide controls timer
   */
  clearHideControlsTimer(): void {
    if (this.hideControlsTimer) {
      clearTimeout(this.hideControlsTimer);
      this.hideControlsTimer = null;
    }
  }

  /**
   * Format time in seconds to MM:SS or HH:MM:SS format
   */
  formatTime(timeInMs: number): string {
    const timeInSeconds = Math.floor(timeInMs / 1000);
    const hours = Math.floor(timeInSeconds / 3600);
    const minutes = Math.floor((timeInSeconds % 3600) / 60);
    const seconds = timeInSeconds % 60;

    if (hours > 0) {
      return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    } else {
      return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
  }

  /**
   * Toggle volume slider visibility
   */
  toggleVolumeSlider(): void {
    this.showVolumeSlider = !this.showVolumeSlider;
  }

  /**
   * Format a Plex stream URL for browser compatibility
   */
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

      // Add protocol parameter to ensure HLS or HTTP progressive streaming
      if (!finalUrl.includes('protocol=')) {
        finalUrl += '&protocol=http';
      }

      // Force transcoding for maximum compatibility
      finalUrl += '&fastSeek=1&session=plex-web-player';
    }

    return finalUrl;
  }

  /**
   * Check if a stream URL is valid for browser playback
   */
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
}
