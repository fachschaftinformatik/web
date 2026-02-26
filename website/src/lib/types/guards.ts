import { z } from 'zod';
import { zDtoSearchResult, zDtoUserResponse } from '@lib/api/zod.gen';
import type { DtoSearchResult, DtoUserResponse } from '@lib/api';
import type { ThemePreference } from '@lib/theme';
import { toEmailAddress, toUserId } from './domain';
import type { SessionUser } from './session';

const zSearchResultArray = z.array(zDtoSearchResult);

export const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

export const isThemePreference = (value: unknown): value is ThemePreference =>
  value === 'light' || value === 'dark' || value === 'system';

export const toThemePreference = (value: unknown): ThemePreference =>
  isThemePreference(value) ? value : 'system';

export const toApiUser = (value: unknown): DtoUserResponse | null => {
  const parsed = zDtoUserResponse.safeParse(value);
  return parsed.success ? parsed.data : null;
};

export const toSessionUser = (value: unknown): SessionUser | null => {
  const apiUser = toApiUser(value);
  if (!apiUser) {
    return null;
  }

  const id = toUserId(apiUser.id);
  const email = toEmailAddress(apiUser.email);
  if (!id || !email) {
    return null;
  }

  return {
    ...apiUser,
    id,
    email,
    emailVerified: null,
    theme: apiUser.theme ? toThemePreference(apiUser.theme) : undefined,
  };
};

export const isSessionUser = (value: unknown): value is SessionUser => toSessionUser(value) !== null;

export const toSearchResults = (value: unknown): DtoSearchResult[] => {
  const parsed = zSearchResultArray.safeParse(value);
  return parsed.success ? parsed.data : [];
};

export const toErrorMessage = (error: unknown): string | null => {
  if (typeof error === 'string') {
    return error;
  }

  if (!isRecord(error)) {
    return null;
  }

  const message = error.message;
  if (typeof message === 'string' && message.length > 0) {
    return message;
  }

  const apiError = error.error;
  if (typeof apiError === 'string' && apiError.length > 0) {
    return apiError;
  }

  const statusText = error.statusText;
  if (typeof statusText === 'string' && statusText.length > 0) {
    return statusText;
  }

  return null;
};
