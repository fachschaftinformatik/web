import type { SessionUser } from '../lib/types/session';

declare module 'next-auth' {
  interface Session {
    user?: SessionUser;
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    user?: SessionUser;
  }
}

export {};
