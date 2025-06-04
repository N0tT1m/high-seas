import { Component, inject, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Observable, Subscription } from 'rxjs';
import { DownloadNotificationService, DownloadNotification } from '../../services/download-notification.service';

@Component({
  selector: 'app-download-notifications',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="notifications-container" *ngIf="(notifications$ | async)?.length">
      <div 
        *ngFor="let notification of notifications$ | async; trackBy: trackByNotificationId"
        class="notification"
        [ngClass]="'notification-' + notification.type"
        (click)="dismissNotification(notification.id)"
      >
        <div class="notification-icon">
          <span *ngIf="notification.type === 'loading'">⏳</span>
          <span *ngIf="notification.type === 'success'">✅</span>
          <span *ngIf="notification.type === 'error'">❌</span>
          <span *ngIf="notification.type === 'info'">ℹ️</span>
        </div>
        
        <div class="notification-content">
          <div class="notification-title">{{ notification.title }}</div>
          <div class="notification-message">{{ notification.message }}</div>
          <div class="notification-time">{{ formatTime(notification.timestamp) }}</div>
        </div>
        
        <div class="notification-close" *ngIf="notification.type !== 'loading'">
          <span>×</span>
        </div>
        
        <div class="notification-progress" *ngIf="notification.type === 'loading'"></div>
      </div>
    </div>
  `,
  styleUrls: ['./download-notifications.component.sass']
})
export class DownloadNotificationsComponent implements OnInit, OnDestroy {
  private notificationService = inject(DownloadNotificationService);
  
  notifications$: Observable<DownloadNotification[]>;
  private subscription?: Subscription;

  ngOnInit() {
    this.notifications$ = this.notificationService.getNotifications();
  }

  ngOnDestroy() {
    this.subscription?.unsubscribe();
  }

  trackByNotificationId(index: number, notification: DownloadNotification): string {
    return notification.id;
  }

  dismissNotification(id: string) {
    this.notificationService.hideNotification(id);
  }

  formatTime(timestamp: Date): string {
    const now = new Date();
    const diff = now.getTime() - timestamp.getTime();
    
    if (diff < 60000) { // Less than 1 minute
      return 'Just now';
    } else if (diff < 3600000) { // Less than 1 hour
      const minutes = Math.floor(diff / 60000);
      return `${minutes}m ago`;
    } else {
      const hours = Math.floor(diff / 3600000);
      return `${hours}h ago`;
    }
  }
}