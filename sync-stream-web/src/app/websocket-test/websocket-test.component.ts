// Create a new diagnostic component
// src/app/diagnostic/websocket-test.component.ts

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { environment } from '../../environments/environment';

@Component({
  selector: 'app-websocket-test',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="diagnostic-container">
      <h2>WebSocket Connection Diagnostic</h2>

      <div class="input-group">
        <label for="wsUrl">WebSocket URL:</label>
        <input type="text" id="wsUrl" [(ngModel)]="wsUrl" class="full-width">
      </div>

      <div class="button-group">
        <button (click)="testConnection()" [disabled]="isConnecting">Test Connection</button>
        <button (click)="testRESTEndpoint()" [disabled]="isChecking">Check Server</button>
      </div>

      <div class="results">
        <h3>Connection Log:</h3>
        <div class="log-container">
          <pre>{{ log }}</pre>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .diagnostic-container {
      padding: 20px;
      background-color: #2a2a2a;
      border-radius: 8px;
      max-width: 800px;
      margin: 20px auto;
    }

    h2, h3 {
      color: white;
    }

    .input-group {
      margin-bottom: 15px;
    }

    label {
      display: block;
      margin-bottom: 5px;
      color: #ddd;
    }

    .full-width {
      width: 100%;
      padding: 8px;
      background-color: #333;
      border: 1px solid #555;
      border-radius: 4px;
      color: white;
    }

    .button-group {
      display: flex;
      gap: 10px;
      margin-bottom: 15px;
    }

    button {
      padding: 10px 15px;
      background-color: #007bff;
      color: white;
      border: none;
      border-radius: 4px;
      cursor: pointer;
    }

    button:disabled {
      background-color: #555;
      cursor: not-allowed;
    }

    .results {
      margin-top: 20px;
    }

    .log-container {
      background-color: #333;
      border-radius: 4px;
      padding: 10px;
      max-height: 400px;
      overflow-y: auto;
    }

    pre {
      color: #ddd;
      white-space: pre-wrap;
      word-break: break-all;
    }
  `]
})
export class WebsocketTestComponent implements OnInit {
  wsUrl: string = '';
  log: string = '';
  isConnecting: boolean = false;
  isChecking: boolean = false;
  socket: WebSocket | null = null;

  ngOnInit(): void {
    // Get base WS URL from environment
    const baseUrl = environment.apiUrl.replace(/^http/, 'ws');
    this.wsUrl = `${baseUrl}/ws?clientId=diagnostic-test&token=test-token`;

    this.appendLog('🔧 Diagnostic tool initialized');
    this.appendLog(`🌐 Default WebSocket URL: ${this.wsUrl}`);
    this.appendLog(`💻 Browser: ${navigator.userAgent}`);
    this.checkBrowserSupport();
  }

  checkBrowserSupport(): void {
    if (!window.WebSocket) {
      this.appendLog('❌ WebSocket not supported in this browser!');
    } else {
      this.appendLog('✅ WebSocket is supported in this browser');
    }
  }

  appendLog(message: string): void {
    const timestamp = new Date().toISOString().substr(11, 8);
    this.log = `[${timestamp}] ${message}\n${this.log}`;
  }

  testConnection(): void {
    if (!this.wsUrl) {
      this.appendLog('❌ WebSocket URL is required');
      return;
    }

    this.isConnecting = true;
    this.appendLog(`🔄 Testing connection to: ${this.wsUrl}`);

    try {
      // Close any existing connection
      if (this.socket) {
        this.socket.close();
        this.socket = null;
      }

      this.socket = new WebSocket(this.wsUrl);

      // Set connection timeout
      const timeoutId = setTimeout(() => {
        this.appendLog('⏱️ Connection timeout after 10 seconds');
        if (this.socket) {
          this.socket.close();
          this.socket = null;
        }
        this.isConnecting = false;
      }, 10000);

      this.socket.onopen = () => {
        clearTimeout(timeoutId);
        this.appendLog('✅ Connection established successfully!');
        this.appendLog(`State: ${this.getReadyStateText(this.socket?.readyState)}`);

        // Send a test message
        if (this.socket) {
          this.socket.send(JSON.stringify({ type: 'ping', payload: { timestamp: new Date().toISOString() } }));
          this.appendLog('📤 Sent test ping message');
        }

        this.isConnecting = false;
      };

      this.socket.onmessage = (event) => {
        this.appendLog(`📩 Message received: ${event.data}`);
      };

      this.socket.onerror = (error) => {
        clearTimeout(timeoutId);
        this.appendLog(`❌ Connection error: ${JSON.stringify(error)}`);
        this.appendLog('Note: Browser security restricts error details. Check console for more info.');
        console.error('WebSocket error:', error);
        this.isConnecting = false;
      };

      this.socket.onclose = (event) => {
        clearTimeout(timeoutId);
        if (event.wasClean) {
          this.appendLog(`🔌 Connection closed cleanly, code=${event.code}, reason=${event.reason || 'None'}`);
        } else {
          this.appendLog(`⚠️ Connection died, code=${event.code}, reason=${event.reason || 'None'}`);

          if (event.code === 1006) {
            this.appendLog('❓ Code 1006 indicates abnormal closure - common causes:');
            this.appendLog('  - CORS policy blocking the connection');
            this.appendLog('  - Server not running or unreachable');
            this.appendLog('  - Authentication failure');
            this.appendLog('  - Firewall or proxy blocking WebSockets');
          }
        }
        this.isConnecting = false;
      };
    } catch (error) {
      this.appendLog(`❌ Error creating WebSocket: ${error}`);
      this.isConnecting = false;
    }
  }

  testRESTEndpoint(): void {
    this.isChecking = true;
    const apiUrl = environment.apiUrl;

    this.appendLog(`🔄 Testing REST API endpoint: ${apiUrl}/health`);

    fetch(`${apiUrl}/health`)
      .then(response => {
        if (response.ok) {
          this.appendLog(`✅ Server is online! Status: ${response.status}`);
          return response.json();
        } else {
          this.appendLog(`❌ Server returned error status: ${response.status}`);
          throw new Error(`HTTP error ${response.status}`);
        }
      })
      .then(data => {
        this.appendLog(`📋 Server response: ${JSON.stringify(data)}`);
      })
      .catch(error => {
        this.appendLog(`❌ Failed to reach server: ${error.message}`);
        this.appendLog('  - Check if server is running');
        this.appendLog('  - Verify API URL in environment.ts');
        this.appendLog('  - Check for CORS issues');
      })
      .finally(() => {
        this.isChecking = false;
      });
  }

  getReadyStateText(state: number | undefined): string {
    if (state === undefined) return 'UNKNOWN';
    switch (state) {
      case WebSocket.CONNECTING: return 'CONNECTING (0)';
      case WebSocket.OPEN: return 'OPEN (1)';
      case WebSocket.CLOSING: return 'CLOSING (2)';
      case WebSocket.CLOSED: return 'CLOSED (3)';
      default: return `UNKNOWN (${state})`;
    }
  }
}
