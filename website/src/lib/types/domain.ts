export type Brand<TValue, TBrand extends string> = TValue & {
  readonly __brand: TBrand;
};

type EmailPattern = `${string}@${string}.${string}`;

export type UserId = Brand<string, 'UserId'>;
export type PostId = Brand<string, 'PostId'>;
export type CommentId = Brand<string, 'CommentId'>;
export type EmailAddress = Brand<EmailPattern, 'EmailAddress'>;

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const normalizeString = (value: unknown): string | null => {
  if (typeof value === 'string') {
    const normalized = value.trim();
    return normalized.length > 0 ? normalized : null;
  }

  if (typeof value === 'number') {
    return String(value);
  }

  return null;
};

const toBrandedId = <TBrand extends string>(value: unknown): Brand<string, TBrand> | null => {
  const normalized = normalizeString(value);
  if (!normalized) {
    return null;
  }

  return normalized as Brand<string, TBrand>;
};

export const toUserId = (value: unknown): UserId | null => toBrandedId<'UserId'>(value);
export const toPostId = (value: unknown): PostId | null => toBrandedId<'PostId'>(value);
export const toCommentId = (value: unknown): CommentId | null => toBrandedId<'CommentId'>(value);

export const unwrapId = <TBrand extends string>(value: Brand<string, TBrand>): string => value;

export const toEmailAddress = (value: unknown): EmailAddress | null => {
  const normalized = normalizeString(value)?.toLowerCase();
  if (!normalized || !EMAIL_PATTERN.test(normalized)) {
    return null;
  }

  return normalized as EmailAddress;
};

export const isEmailAddress = (value: unknown): value is EmailAddress => toEmailAddress(value) !== null;
