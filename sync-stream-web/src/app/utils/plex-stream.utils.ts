// src/app/core/utils/plex-stream.utils.ts

/**
 * Utility functions for handling Plex streams
 */

/**
 * Formats a Plex streaming URL to be directly playable in a browser
 *
 * @param streamUrl The original Plex stream URL
 * @param token Authentication token
 * @returns A properly formatted streaming URL
 */
// Update the formatPlexStreamUrl method in video-player.component.ts

export function formatPlexStreamUrl(streamUrl: string, token: string): string {
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

    // IMPORTANT: Set protocol to HLS instead of HTTP for better browser compatibility
    if (!finalUrl.includes('protocol=')) {
      finalUrl += '&protocol=hls';
    } else if (finalUrl.includes('protocol=http')) {
      // Replace http protocol with hls if it's already set
      finalUrl = finalUrl.replace('protocol=http', 'protocol=hls');
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

    // Force transcoding for maximum compatibility
    finalUrl += '&fastSeek=1&session=plex-web-player';
  }

  return finalUrl;
}

/**
 * Checks if a stream URL is properly formatted for browser playback
 *
 * @param url The stream URL to check
 * @returns true if the URL appears to be properly formatted
 */
export function isValidStreamUrl(url: string): boolean {
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

/**
 * Creates a direct stream URL for a Plex media item
 *
 * @param serverUrl Plex server base URL
 * @param mediaId Media ID
 * @param token Authentication token
 * @returns A direct stream URL
 */
export function createDirectStreamUrl(serverUrl: string, mediaId: string, token: string): string {
  // Clean up the mediaId
  let id = mediaId;
  if (id.includes('library/metadata/')) {
    const matches = id.match(/library\/metadata\/(\d+)/);
    if (matches && matches[1]) {
      id = matches[1];
    }
  }

  // Ensure the server URL doesn't have a trailing slash
  const baseUrl = serverUrl.endsWith('/') ? serverUrl.slice(0, -1) : serverUrl;

  // Create the streaming URL
  return formatPlexStreamUrl(
    `${baseUrl}/video/:/transcode/universal/start?path=/library/metadata/${id}`,
    token
  );
}
