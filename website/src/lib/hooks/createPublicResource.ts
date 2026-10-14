import { useCallback, useEffect, useSyncExternalStore } from 'react';

// Client-memory cache for public data only. Revalidation retains the last successful
// result, so revisiting a page does not replace loaded content with skeletons.
export function createPublicResource<T>(load: (key: string) => Promise<T>) {
  type Snapshot = { data?: T; error?: unknown };
  type Entry = { snapshot?: Snapshot; updatedAt: number; pending?: Promise<void>; listeners: Set<() => void> };
  const entries = new Map<string, Entry>();
  const entryFor = (key: string) => {
    let entry = entries.get(key);
    if (!entry) {
      entry = { updatedAt: 0, listeners: new Set() };
      entries.set(key, entry);
    }
    return entry;
  };
  const refresh = (key: string) => {
    const entry = entryFor(key);
    if (entry.pending) return entry.pending;
    entry.pending = load(key).then(data => {
      entry.snapshot = { data };
      entry.updatedAt = Date.now();
    }).catch(error => {
      entry.snapshot = { data: entry.snapshot?.data, error };
    }).finally(() => {
      entry.pending = undefined;
      entry.listeners.forEach(listener => listener());
    });
    return entry.pending;
  };
  function useResource(key: string | null) {
    const subscribe = useCallback((listener: () => void) => {
      if (key === null) return () => {};
      const entry = entryFor(key);
      entry.listeners.add(listener);
      return () => { entry.listeners.delete(listener); };
    }, [key]);
    const snapshot = useCallback(() => key === null ? undefined : entryFor(key).snapshot, [key]);
    const data = useSyncExternalStore(subscribe, snapshot, () => undefined);
    useEffect(() => {
      if (key !== null && Date.now() - entryFor(key).updatedAt > 30_000) void refresh(key);
    }, [key]);
    return data;
  }
  return { useResource, refresh };
}
