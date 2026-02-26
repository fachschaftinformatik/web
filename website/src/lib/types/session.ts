import type { DtoUserResponse } from '@lib/api';
import type { ThemePreference } from '@lib/theme';
import type { EmailAddress, UserId } from './domain';

export type SessionUser = Omit<DtoUserResponse, 'email' | 'id' | 'theme'> & {
  email: EmailAddress;
  emailVerified: Date | null;
  id: UserId;
  theme?: ThemePreference;
};
