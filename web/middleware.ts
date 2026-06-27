import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const protectedPaths = [
  '/audit',
  '/blockchain',
  '/compliance',
  '/dashboard',
  '/emissions',
  '/onboarding',
  '/settings',
  '/workflow',
];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const isProtected = protectedPaths.some(
    (p) => pathname === p || pathname.startsWith(p + '/')
  );

  if (!isProtected) {
    return NextResponse.next();
  }

  const token =
    request.cookies.get('offgrid_session')?.value ||
    request.cookies.get('offgridflow_session')?.value;

  if (token) {
    const parts = token.split('.');
    if (parts.length === 3) {
      try {
        const payload = JSON.parse(atob(parts[1]));
        if (payload.exp && payload.exp * 1000 > Date.now()) {
          return NextResponse.next();
        }
      } catch {
        // malformed token — fall through to redirect
      }
    }
  }

  const loginUrl = new URL('/login', request.url);
  loginUrl.searchParams.set('returnTo', pathname);
  return NextResponse.redirect(loginUrl);
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/emissions/:path*',
    '/compliance/:path*',
    '/audit/:path*',
    '/onboarding/:path*',
    '/settings/:path*',
    '/workflow/:path*',
    '/blockchain/:path*',
  ],
};
