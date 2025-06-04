// src/app/core/models/user.media.model.ts
export interface User {
  id: string;
  username: string;
  token: string;
  tokenExpirationDate?: string;
}
// src/app/features/libraries/media.model.ts

export interface Library {
  key: string;
  title: string;
  type: string;
  agent?: string;
}

export interface MediaItem {
  key: string;
  title: string;
  type: string;
  year?: number;
  duration?: number;
  thumbnail?: string;
  summary?: string;
  librarySectionTitle?: string;
  librarySectionID?: string;
  rating?: number;
  viewOffset?: number;  // Position where playback was last stopped (milliseconds)
  viewCount?: number;   // Number of times this item has been played
  state?: string;       // 'playing', 'paused', 'stopped'
  art?: string;         // Background art URL for the media
  poster?: string;      // Alternative poster image URL
  guid?: string;        // Global unique identifier
  ratingKey?: string;   // Rating key used by Plex
  studio?: string;      // Studio that produced the content
  tagline?: string;     // Tagline or short description
  streamUrl?: string; // Add this property for direct streaming
  contentRating?: string; // Content rating (e.g., PG-13, TV-MA)
  originallyAvailableAt?: string; // Original release date
  addedAt?: number;     // Timestamp when the item was added to the library
  updatedAt?: number;   // Timestamp when the item was last updated
  genres?: string[];    // Array of genre names
  directors?: string[]; // Array of director names
  writers?: string[];   // Array of writer names
  actors?: string[];    // Array of actor names
  media?: any[];        // Media versions (various quality levels)
  parentKey?: string;   // Key of the parent item (for episodes/seasons)
  grandparentKey?: string; // Key of the grandparent item (for episodes)
  grandparentTitle?: string; // Title of the grandparent (for episodes)
  // Additional fields for TV shows
  seasonCount?: number;
  // Additional fields for episodes
  episodeNumber?: number;
}

export interface Episode extends MediaItem {
  parentTitle: string;  // Series name
  parentKey: string;    // Series key
  seasonNumber: number;
  episodeNumber: number;
}

export interface Season {
  key: string;
  title: string;
  seasonNumber: number;
  parentTitle: string;  // Series name
  thumbnail?: string;
  episodes?: Episode[];
}

export interface MediaSession {
  mediaKey: string;
  position: number;
  duration: number;
  state: 'playing' | 'paused' | 'stopped';
  metadata?: Record<string, string>;
  lastUpdated: Date;
  clientId?: string;    // ID of the client that last updated the session
  lastClient?: string;  // Name or ID of the last client that played this item
}

export interface StreamInfo {
  streamUrl: string;
  subtitleUrl?: string;
  audioUrl?: string;
}

export interface PlayerState {
  mediaKey: string | null;
  isPlaying: boolean;
  isPaused: boolean;
  isStopped: boolean;
  isBuffering: boolean;
  isError: boolean;
  isPendingUserInteraction?: boolean; // Add this new flag
  position: number;
  duration: number;
  volume: number;
  muted: boolean;
  title: string;
  subtitle: string;
  streamUrl: string | null;
  error: string | null;
}
