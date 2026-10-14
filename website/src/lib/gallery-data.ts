import { getEvents, getEventsByEventIdMedia } from '@lib/api';
import { createPublicResource } from '@lib/hooks/createPublicResource';

export const galleryEvents = createPublicResource(async () => {
  const { data, error } = await getEvents();
  if (error) throw error;
  return data ?? [];
});
export const galleryMedia = createPublicResource(async (eventId: string) => {
  const { data, error } = await getEventsByEventIdMedia({ path: { eventId } });
  if (error) throw error;
  return data ?? [];
});
