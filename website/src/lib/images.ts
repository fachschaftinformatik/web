export const IMAGE_SIZES = [200, 400, 600, 800, 1200, 1600] as const;
export type ImageSize = (typeof IMAGE_SIZES)[number];

export function toImageSize(size: number | string | undefined | null, fallback: ImageSize = 400): ImageSize {
  const parsed = typeof size === 'number' ? size : Number.parseInt(String(size), 10);

  if (Number.isNaN(parsed)) {
    return fallback;
  }

  const exactMatch = IMAGE_SIZES.find((candidate) => candidate === parsed);
  if (exactMatch) {
    return exactMatch;
  }

  let nearest: ImageSize = IMAGE_SIZES[0];
  for (const candidate of IMAGE_SIZES) {
    if (Math.abs(candidate - parsed) < Math.abs(nearest - parsed)) {
      nearest = candidate;
    }
  }

  return nearest;
}

/**
 * Generates a URL for a specific image size.
 * Handles avatars, event covers, and media previews.
 */
export function getSizedImageUrl(url: string | undefined | null, size: ImageSize | number | string): string {
  if (!url) return '';
  
  // If it's already a full URL or blob URL, return as is
  if (url.startsWith('http') || url.startsWith('blob:') || url.startsWith('data:')) {
    return url;
  }

  const baseUrl = url.includes('?') ? url.split('?')[0] : url;
  return `${baseUrl}?size=${toImageSize(size)}`;
}

/**
 * Generates a srcset string for an image URL.
 */
export function getImageSrcSet(url: string | undefined | null): string {
  if (!url || url.startsWith('http') || url.startsWith('blob:') || url.startsWith('data:')) {
    return '';
  }

  const baseUrl = url.includes('?') ? url.split('?')[0] : url;
  return IMAGE_SIZES.map((size) => `${baseUrl}?size=${size} ${size}w`).join(', ');
}

/**
 * Helper specifically for user avatars. 
 * Avatars are served as SVGs (or original uploaded format) without previews.
 */
export function getAvatarUrl(url: string | undefined | null): string {
  if (!url) return '';
  if (url.startsWith('http') || url.startsWith('blob:') || url.startsWith('data:')) {
    return url;
  }
  return url.includes('?') ? url.split('?')[0] : url;
}
