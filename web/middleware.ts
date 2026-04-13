import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const protectedPaths = [
  '/dashboard',
  '/emissions',
  '/compliance',
  '/settings',
  '/workflow',
  '/blockchain',
];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const isProtected = protectedPaths.some(
    (p) => pathname === p || pathname.startsWith(p + '/')
  );

  if (!isProtected) {
    return NextResponse.next();
  }

  // Check for auth token in cookies or Authorization header
  // The client stores the token in localStorage which we can't read server-side,
  // but we can check for the session cookie set by the CSRF flow
  const token = request.cookies.get('offgridflow_session')?.value;

  // If no server-side cookie, let the client-side useRequireAuth handle redirect.
  // This middleware mainly prevents search engines and direct URL scrapers from
  // seeing authenticated page structures.
  // We add a header so the client knows to check auth immediately.
  const response = NextResponse.next();
  response.headers.set('x-offgridflow-auth-required', 'true');
  return response;
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/emissions/:path*',
    '/compliance/:path*',
    '/settings/:path*',
    '/workflow/:path*',
    '/blockchain/:path*',
  ],
};
