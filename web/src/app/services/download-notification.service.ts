import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

export interface DownloadNotification {
  id: string;
  type: 'success' | 'error' | 'info' | 'loading';
  title: string;
  message: string;
  timestamp: Date;
  duration?: number; // Auto-hide after milliseconds (0 means no auto-hide)
}

@Injectable({
  providedIn: 'root'
})
export class DownloadNotificationService {
  private notifications$ = new BehaviorSubject<DownloadNotification[]>([]);
  private activeDownloads$ = new BehaviorSubject<Set<string>>(new Set());

  constructor() {}

  getNotifications(): Observable<DownloadNotification[]> {
    return this.notifications$.asObservable();
  }

  getActiveDownloads(): Observable<Set<string>> {
    return this.activeDownloads$.asObservable();
  }

  isDownloadActive(tmdbId: number, contentType: 'movie' | 'show' | 'anime'): boolean {
    const key = `${contentType}-${tmdbId}`;
    return this.activeDownloads$.value.has(key);
  }

  startDownload(tmdbId: number, contentType: 'movie' | 'show' | 'anime', title: string): string {
    const key = `${contentType}-${tmdbId}`;
    const activeDownloads = new Set(this.activeDownloads$.value);
    activeDownloads.add(key);
    this.activeDownloads$.next(activeDownloads);

    return this.showNotification({
      type: 'loading',
      title: 'Download Started',
      message: `Searching for "${title}"...`,
      duration: 0 // Don't auto-hide loading notifications
    });
  }

  completeDownload(tmdbId: number, contentType: 'movie' | 'show' | 'anime', title: string): string {
    const key = `${contentType}-${tmdbId}`;
    const activeDownloads = new Set(this.activeDownloads$.value);
    activeDownloads.delete(key);
    this.activeDownloads$.next(activeDownloads);

    return this.showNotification({
      type: 'success',
      title: 'Download Request Submitted',
      message: `"${title}" has been added to the download queue.`,
      duration: 5000
    });
  }

  failDownload(tmdbId: number, contentType: 'movie' | 'show' | 'anime', title: string, error?: string): string {
    const key = `${contentType}-${tmdbId}`;
    const activeDownloads = new Set(this.activeDownloads$.value);
    activeDownloads.delete(key);
    this.activeDownloads$.next(activeDownloads);

    return this.showNotification({
      type: 'error',
      title: 'Download Failed',
      message: error || `Failed to download "${title}". Please try again later.`,
      duration: 8000
    });
  }

  showInfo(title: string, message: string): string {
    return this.showNotification({
      type: 'info',
      title,
      message,
      duration: 4000
    });
  }

  private showNotification(notification: Omit<DownloadNotification, 'id' | 'timestamp'>): string {
    const id = this.generateId();
    const fullNotification: DownloadNotification = {
      ...notification,
      id,
      timestamp: new Date()
    };

    const current = this.notifications$.value;
    this.notifications$.next([fullNotification, ...current]);

    // Auto-hide if duration is specified
    if (notification.duration && notification.duration > 0) {
      setTimeout(() => {
        this.hideNotification(id);
      }, notification.duration);
    }

    return id;
  }

  hideNotification(id: string): void {
    const current = this.notifications$.value;
    const filtered = current.filter(n => n.id !== id);
    this.notifications$.next(filtered);
  }

  clearAll(): void {
    this.notifications$.next([]);
  }

  private generateId(): string {
    return Math.random().toString(36).substring(2) + Date.now().toString(36);
  }
}