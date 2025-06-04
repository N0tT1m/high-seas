// src/app/core/services/sync.service.ts
import { Injectable, OnDestroy } from '@angular/core';
import { BehaviorSubject, Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { environment } from '../environments/environment'; // Updated path
import { AuthService } from './auth.service';
import { MediaSessionService } from './media-session.service'; // Updated path
import { v4 as uuidv4 } from 'uuid';

export interface WebSocketMessage {
  type: string;
  payload: any;
}

/**
 * Handles websocket communication for real-time sync between devices
 */
@Injectable({
  providedIn: 'root'
})
export class SyncService implements OnDestroy {
  private socket: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 2000; // Start with 2 seconds
  private reconnectTimer: any;
  private destroy$ = new Subject<void>();
  private clientId = '';

  // In sync.service.ts - Add this new property to the class
  private useFallbackMode = false;

  private connectionStatusSubject = new BehaviorSubject<boolean>(false);
  public connectionStatus$ = this.connectionStatusSubject.asObservable();

  constructor(
    private authService: AuthService,
    private mediaSessionService: MediaSessionService
  ) {
    // Generate a unique client ID for this browser/device
    this.clientId = localStorage.getItem('plexClientId') || uuidv4();
    localStorage.setItem('plexClientId', this.clientId);

    // Listen for auth changes to connect/disconnect
    this.authService.currentUser$
      .pipe(takeUntil(this.destroy$))
      .subscribe(user => {
        if (user) {
          this.connect();
        } else {
          this.disconnect();
        }
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.disconnect();
  }

  /**
   * Disconnect the WebSocket
   */
  private disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    this.connectionStatusSubject.next(false);
  }

  /**
   * Handle received WebSocket messages
   */
  private handleMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case 'auth_success':
        console.log('Authentication successful');
        // After successful auth, request sessions
        this.sendMessage('get_sessions', {});
        break;

      case 'auth_error':
        console.error('Authentication failed:', message.payload);
        // Force reconnect to try authentication again
        this.disconnect();
        this.scheduleReconnect();
        break;

      case 'position_update':
        if (message.payload.clientId !== this.clientId) {
          // Only update if it's from another client
          this.mediaSessionService.handleRemotePositionUpdate(
            message.payload.mediaKey,
            message.payload.position
          );
        }
        break;

      case 'play_event':
        if (message.payload.clientId !== this.clientId) {
          this.mediaSessionService.handleRemotePlayEvent(
            message.payload.mediaKey,
            message.payload.position
          );
        }
        break;

      case 'pause_event':
        if (message.payload.clientId !== this.clientId) {
          this.mediaSessionService.handleRemotePauseEvent(
            message.payload.mediaKey,
            message.payload.position
          );
        }
        break;

      case 'stop_event':
        if (message.payload.clientId !== this.clientId) {
          this.mediaSessionService.handleRemoteStopEvent(
            message.payload.mediaKey,
            message.payload.position
          );
        }
        break;

      case 'sessions':
        this.mediaSessionService.importRemoteSessions(message.payload);
        break;

      default:
        console.log('Unknown message type:', message.type);
    }
  }

  /**
   * Setup WebSocket event handlers
   */
  private setupSocketEvents(): void {
    if (!this.socket) return;

    this.socket.onopen = () => {
      console.log('WebSocket connected successfully');
      this.connectionStatusSubject.next(true);
      this.reconnectAttempts = 0;

      // The token is already in the URL, no need to send authentication again
      // Just request current sessions
      setTimeout(() => {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
          this.sendMessage('get_sessions', {});
        }
      }, 500); // Small delay to ensure server is ready
    };

    this.socket.onclose = (event) => {
      console.log('WebSocket disconnected', event);
      this.connectionStatusSubject.next(false);

      // Code 1000 is normal closure, 1001 is going away
      if (event.code !== 1000 && event.code !== 1001) {
        this.scheduleReconnect();
      }
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      // Don't close here - the onclose handler will be called next
    };

    this.socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error, event.data);
      }
    };
  }

  /**
   * Send a WebSocket message with better error handling
   */
  public sendMessage(type: string, payload: any): void {
    if (!this.socket) {
      console.warn('WebSocket not initialized, queuing message for later');
      // Store message for later sending
      setTimeout(() => {
        this.connect();
      }, 100);
      return;
    }

    if (this.socket.readyState !== WebSocket.OPEN) {
      console.warn(`WebSocket not open (state: ${this.socket.readyState}), queuing message`);
      // Could implement a message queue here if needed
      return;
    }

    const message = {
      type,
      payload
    };

    try {
      this.socket.send(JSON.stringify(message));
    } catch (error) {
      console.error('Error sending WebSocket message:', error);
      this.disconnect();
      this.scheduleReconnect();
    }
  }

  /**
   * Creates a fallback mode that doesn't rely on WebSockets
   */
  // Implement the fallback mode
  private enableFallbackMode(): void {
    console.log('⚠️ Using HTTP fallback mode instead of WebSockets');

    // Start polling for session updates
    const pollInterval = 10000; // 10 seconds
    setInterval(() => {
      if (this.authService.getToken()) {
        this.pollForSessions();
      }
    }, pollInterval);

    // Reset connection status
    this.connectionStatusSubject.next(false);
  }

  // Add polling method
  private pollForSessions(): void {
    const token = this.authService.getToken();
    fetch(`${environment.apiUrl}/api/continue-watching`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
      .then(response => response.json())
      .then(data => {
        this.mediaSessionService.importRemoteSessions(data);
      })
      .catch(error => {
        console.error('Error polling for sessions:', error);
      });
  }

  /**
   * Fallback implementation for sending messages
   */
  private sendMessageFallback(type: string, payload: any): void {
    switch (type) {
      case 'update_position':
        // Use HTTP POST instead of WebSocket
        fetch(`${environment.apiUrl}/api/media/${payload.mediaKey}/position`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.authService.getToken()}`
          },
          body: JSON.stringify({
            position: payload.position,
            clientId: this.clientId
          })
        }).catch(error => {
          console.error('Error sending position update:', error);
        });
        break;

      // Handle other message types similarly
      case 'play':
      case 'pause':
      case 'stop':
        // Similar implementations using HTTP endpoints
        break;
    }
  }

  // Update the connect method to attempt a websocket connection but also prepare for fallback
  private connect(): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      return;
    }

    const token = this.authService.getToken();
    if (!token) {
      console.error('No token available for WebSocket connection');
      return;
    }

    // Check if we should use fallback mode
    if (this.useFallbackMode) {
      this.enableFallbackMode();
      return;
    }

    try {
      // Replace http/https with ws/wss
      const wsBase = environment.apiUrl.replace(/^http/, 'ws');
      // Add token to query parameters for authentication
      const wsUrl = `${wsBase}/ws?clientId=${this.clientId}&token=${encodeURIComponent(token)}`;

      console.log('⚠️ Attempting to connect to WebSocket:', wsUrl);

      this.socket = new WebSocket(wsUrl);
      this.setupSocketEvents();

      // Set a timeout to detect stalled connections
      setTimeout(() => {
        if (this.socket && this.socket.readyState !== WebSocket.OPEN) {
          console.error('WebSocket connection timed out');
          this.socket.close();
          this.socket = null;
          this.scheduleReconnect();
        }
      }, 5000);
    } catch (error) {
      console.error('Error creating WebSocket:', error);
      this.scheduleReconnect();
    }
  }

  private establishConnection(wsUrl: string): void {
    try {
      // Set up global error handler for unhandled WebSocket errors
      window.addEventListener('error', (event) => {
        if (event.target instanceof WebSocket) {
          console.error('⚠️ Unhandled WebSocket error:', event);
        }
      }, { once: true });

      this.socket = new WebSocket(wsUrl);

      // Create a connection timeout
      const connectionTimeout = setTimeout(() => {
        if (this.socket && this.socket.readyState !== WebSocket.OPEN) {
          console.error('⏱️ WebSocket connection timed out after 10 seconds');
          this.socket.close();
          this.scheduleReconnect();
        }
      }, 10000);

      this.socket.onopen = () => {
        clearTimeout(connectionTimeout);
        console.log('✅ WebSocket connected successfully');
        this.connectionStatusSubject.next(true);
        this.reconnectAttempts = 0;

        // After connection, immediately ping to verify two-way communication
        this.sendMessage('ping', { timestamp: new Date().toISOString() });
      };

      this.socket.onclose = (event) => {
        clearTimeout(connectionTimeout);
        console.log('⚠️ WebSocket disconnected', event);
        this.connectionStatusSubject.next(false);

        // Log more detailed information about the close event
        if (event.wasClean) {
          console.log(`Clean close: Code ${event.code}, Reason: ${event.reason || 'None provided'}`);
        } else {
          console.error(`Connection died: Code ${event.code}, Reason: ${event.reason || 'None provided'}`);

          // Code 1006 specifically means abnormal closure - often firewall or security related
          if (event.code === 1006) {
            console.error('Code 1006 indicates abnormal closure - check CORS, server availability, or network issues');
          }
        }

        // Only reconnect for abnormal closures
        if (!event.wasClean) {
          this.scheduleReconnect();
        }
      };

      this.socket.onerror = (error) => {
        console.error('❌ WebSocket error:', error);
        // Error details are limited in the error event due to browser security restrictions
        console.error('WebSocket errors are often caused by CORS, authentication, or network issues');
      };

      this.socket.onmessage = (event) => {
        try {
          console.log('📩 WebSocket message received:', event.data);
          const message = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error, event.data);
        }
      };
    } catch (error) {
      console.error('❌ Error creating WebSocket:', error);
      this.scheduleReconnect();
    }
  }

  // Add this method to handle max reconnect attempts
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('🛑 Max reconnection attempts reached, switching to fallback mode');
      this.useFallbackMode = true;
      this.enableFallbackMode();
      return;
    }

    const delay = Math.min(
      this.reconnectDelay * Math.pow(1.5, this.reconnectAttempts),
      30000 // Cap at 30 seconds
    );

    console.log(`⏱️ Scheduling reconnect in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`);

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect();
    }, delay);
  }

// Function to test WebSocket connectivity
  private testWebSocketConnection(wsUrl: string): Promise<string> {
    return new Promise((resolve, reject) => {
      console.log(`Testing WebSocket connection to: ${wsUrl}`);

      const socket = new WebSocket(wsUrl);
      let timeoutId: any;

      // Set a timeout to catch stalled connections
      timeoutId = setTimeout(() => {
        socket.close();
        reject('Connection timeout after 10 seconds');
      }, 10000);

      socket.onopen = () => {
        clearTimeout(timeoutId);
        console.log('Test connection successful!');
        socket.close();
        resolve('Connection successful');
      };

      socket.onerror = (error) => {
        clearTimeout(timeoutId);
        console.error('Test connection error:', error);
        socket.close();
        reject('Connection error');
      };

      socket.onclose = (event) => {
        clearTimeout(timeoutId);
        if (event.wasClean) {
          console.log(`Test connection closed cleanly, code=${event.code}`);
        } else {
          console.error(`Test connection died, code=${event.code}`);
        }
        reject(`Connection closed: ${event.code}`);
      };
    });
  }

  /**
   * Manually initiate a reconnection
   */
  public reconnect(): void {
    this.disconnect();
    this.reconnectAttempts = 0;
    this.connect();
  }

  // Override the send methods to use HTTP when in fallback mode
  public updatePosition(mediaKey: string, position: number): void {
    if (this.useFallbackMode) {
      this.sendPositionUpdateHttp(mediaKey, position);
    } else {
      this.sendMessage('update_position', {
        mediaKey,
        position,
        clientId: this.clientId
      });
    }
  }

  // Implementation for HTTP fallback
  private sendPositionUpdateHttp(mediaKey: string, position: number): void {
    const token = this.authService.getToken();
    fetch(`${environment.apiUrl}/api/media/${mediaKey}/position`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        position: position,
        duration: 0, // You may need to get the actual duration
        state: 'playing',
        clientId: this.clientId
      })
    })
      .catch(error => {
        console.error('Error sending position update via HTTP:', error);
      });
  }


  /**
   * Send play event
   */
  public sendPlayEvent(mediaKey: string, position: number): void {
    this.sendMessage('play', {
      mediaKey,
      position,
      clientId: this.clientId
    });
  }

  /**
   * Send pause event
   */
  public sendPauseEvent(mediaKey: string, position: number): void {
    this.sendMessage('pause', {
      mediaKey,
      position,
      clientId: this.clientId
    });
  }

  /**
   * Send stop event
   */
  public sendStopEvent(mediaKey: string, position: number): void {
    this.sendMessage('stop', {
      mediaKey,
      position,
      clientId: this.clientId
    });
  }

  /**
   * Get the client ID for this device
   */
  public getClientId(): string {
    return this.clientId;
  }
}
