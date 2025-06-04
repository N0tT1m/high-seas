// src/app/core/services/auth.service.ts
import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { Router } from '@angular/router';
import { environment } from '../environments/environment';
import { User } from './media.model';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();
  private tokenExpirationTimer: any;

  // Define consistent headers for all requests
  private httpOptions = {
    headers: new HttpHeaders({
      'Content-Type': 'application/json'
    })
  };

  constructor(private http: HttpClient, private router: Router) {
    this.loadStoredUser();
  }

  signup(username: string, password: string): Observable<User> {
    return this.http.post<any>(
      `${environment.apiUrl}/api/auth/register`,
      { username, password },
      this.httpOptions
    ).pipe(
      map(response => {
        // Convert response to User model
        const user: User = {
          id: response.userId,
          username: username,
          token: response.token,
          tokenExpirationDate: new Date(new Date().getTime() + 24 * 60 * 60 * 1000).toISOString() // Set expiration to 24 hours from now
        };

        this.handleAuthentication(user);
        return user;
      }),
      catchError(error => {
        console.error('Registration error:', error);
        throw error;
      })
    );
  }

  login(username: string, password: string): Observable<User> {
    return this.http.post<any>(
      `${environment.apiUrl}/api/auth/login`,
      { username, password },
      this.httpOptions
    ).pipe(
      map(response => {
        // Convert response to User model
        const user: User = {
          id: response.userId,
          username: username,
          token: response.token,
          tokenExpirationDate: response.expirationDate || new Date(new Date().getTime() + 24 * 60 * 60 * 1000).toISOString()
        };

        this.handleAuthentication(user);
        return user;
      }),
      catchError(error => {
        console.error('Login error:', error);
        throw error;
      })
    );
  }

  logout(): void {
    localStorage.removeItem('userData');
    this.currentUserSubject.next(null);
    this.router.navigate(['/login']);

    if (this.tokenExpirationTimer) {
      clearTimeout(this.tokenExpirationTimer);
      this.tokenExpirationTimer = null;
    }
  }

  private loadStoredUser(): void {
    const userData = localStorage.getItem('userData');
    if (!userData) {
      return;
    }

    const user: User = JSON.parse(userData);
    if (!user.token) {
      return;
    }

    // Check token expiration
    if (user.tokenExpirationDate) {
      const expirationDate = new Date(user.tokenExpirationDate);
      if (expirationDate <= new Date()) {
        this.logout();
        return;
      }

      // Set auto-logout timer
      this.autoLogout(expirationDate.getTime() - new Date().getTime());
    }

    this.currentUserSubject.next(user);
  }

  private autoLogout(expirationDuration: number): void {
    this.tokenExpirationTimer = setTimeout(() => {
      this.logout();
    }, expirationDuration);
  }

  private handleAuthentication(user: User): void {
    // Save user to local storage
    localStorage.setItem('userData', JSON.stringify(user));
    this.currentUserSubject.next(user);

    // Set up auto-logout if expiration date is provided
    if (user.tokenExpirationDate) {
      const expirationDate = new Date(user.tokenExpirationDate);
      const expirationDuration = expirationDate.getTime() - new Date().getTime();
      this.autoLogout(expirationDuration);
    }
  }

  getToken(): string | null {
    const currentUser = this.currentUserSubject.value;
    return currentUser ? currentUser.token : null;
  }

  isAuthenticated(): boolean {
    return !!this.currentUserSubject.value;
  }

  getCurrentUserId(): string | null {
    const currentUser = this.currentUserSubject.value;
    return currentUser ? currentUser.id : null;
  }
}
