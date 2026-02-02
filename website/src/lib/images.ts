export const IMAGE_SIZES = [200, 400, 600, 800, 1200, 1600] as const;
export type ImageSize = (typeof IMAGE_SIZES)[number];

export function getSizedImageUrl(url: string | undefined | null, size: ImageSize): string {
  if (!url) return '';
  if (url.startsWith('http') || url.startsWith('blob:') || url.startsWith('data:')) return url;
  const baseUrl = url.includes('?') ? url.split('?')[0] : url;
  return `${baseUrl}?size=${size}`;
}

export function getImageSrcSet(url: string | undefined | null): string {
  if (!url || url.startsWith('http') || url.startsWith('blob:') || url.startsWith('data:')) return '';
  const baseUrl = url.includes('?') ? url.split('?')[0] : url;
  return IMAGE_SIZES.map((size) => `${baseUrl}?size=${size} ${size}w`).join(', ');
}

export function getAvatarUrl(url: string | undefined | null): string {
  if (!url) return '';
  if (url.startsWith('http') || url.startsWith('blob:') || url.startsWith('data:')) return url;
  return url.includes('?') ? url.split('?')[0] : url;
}
