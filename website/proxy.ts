import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // We can't easily check auth state in middleware without a session cookie
  // that we can verify. Since the current app uses an API call to check auth,
  // we might want to keep the client-side checks for now or implement a more robust
  // session management if we want to use middleware effectively.
  
  // For now, let's just let the client-side components handle it as they did before.
  // The Sidebar and components already handle missing users.
  
  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};
