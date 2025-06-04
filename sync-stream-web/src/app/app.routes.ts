// app.routes.ts
import { Routes } from '@angular/router';
import { AuthGuard } from './auth-guard.service';
import {WebsocketTestComponent} from './websocket-test/websocket-test.component';

export const routes: Routes = [
  {
    path: '',
    redirectTo: '/home',
    pathMatch: 'full'
  },
  {
    path: 'login',
    loadComponent: () => import('./login/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'signup',
    loadComponent: () => import('./signup/signup.component').then(m => m.SignupComponent)
  },
  {
    path: 'diagnostics',
    component: WebsocketTestComponent
  },
  {
    path: 'home',
    loadComponent: () => import('./home/home.component').then(m => m.HomeComponent),
    canActivate: [AuthGuard]
  },
  {
    path: 'libraries',
    loadComponent: () => import('./libraries-list/libraries-list.component').then(m => m.LibrariesListComponent),
    canActivate: [AuthGuard]
  },
  {
    path: 'library/:id',
    loadComponent: () => import('./library-detail/library-detail.component').then(m => m.LibraryDetailComponent),
    canActivate: [AuthGuard]
  },
  {
    path: 'media/:id',
    loadComponent: () => import('./media-detail/media-detail.component').then(m => m.MediaDetailComponent),
    canActivate: [AuthGuard]
  },
  {
    path: 'player/:id',
    loadComponent: () => import('./media-player/media-player.component').then(m => m.MediaPlayerComponent),
    canActivate: [AuthGuard]
  },
  {
    path: 'search',
    loadComponent: () => import('./search/search.component').then(m => m.SearchComponent),
    canActivate: [AuthGuard]
  },
  {
    path: 'settings',
    loadComponent: () => import('./settings/settings.component').then(m => m.SettingsComponent),
    canActivate: [AuthGuard]
  },
  {
    path: '**',
    redirectTo: '/home'
  }
];
