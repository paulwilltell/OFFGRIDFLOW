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

  const token =
    request.cookies.get('offgrid_session')?.value ||
    request.cookies.get('offgridflow_session')?.value;

  if (token) {
    return NextResponse.next();
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
    '/settings/:path*',
    '/workflow/:path*',
    '/blockchain/:path*',
  ],
};
