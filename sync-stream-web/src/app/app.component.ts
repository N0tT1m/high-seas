// src/app/app.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { AuthService } from './auth.service';
import { SyncService } from './sync.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  title = 'Plex Web Player';
  isLoggedIn = false;
  wsConnected = false;
  currentYear = new Date().getFullYear();

  constructor(
    private authService: AuthService,
    private syncService: SyncService
  ) {}

  ngOnInit(): void {
    // Subscribe to auth state
    this.authService.currentUser$.subscribe(user => {
      this.isLoggedIn = !!user;
    });

    // Subscribe to WebSocket connection status
    this.syncService.connectionStatus$.subscribe(connected => {
      this.wsConnected = connected;
    });
  }
}
