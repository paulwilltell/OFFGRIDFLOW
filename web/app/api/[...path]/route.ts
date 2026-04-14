import { NextRequest, NextResponse } from 'next/server';

const API_ORIGIN =
  process.env.OFFGRIDFLOW_API_ORIGIN || 'https://offgridflow-api-v2-production.up.railway.app';

const SESSION_COOKIE_NAME = 'offgrid_session';
const CSRF_COOKIE_NAME = 'csrf_token';
const SESSION_COOKIE_MAX_AGE = 7 * 24 * 60 * 60;
const CSRF_COOKIE_MAX_AGE = 24 * 60 * 60;

const FORWARDED_REQUEST_HEADERS = new Set([
  'accept',
  'content-type',
  'x-csrf-token',
  'x-org-id',
  'x-request-id',
  'x-tenant-id',
]);

const FORWARDED_RESPONSE_HEADERS = new Set([
  'cache-control',
  'content-disposition',
  'content-type',
  'expires',
  'location',
  'pragma',
]);

const AUTH_COOKIE_PATHS = new Set([
  '/auth/login',
  '/auth/register',
  '/auth/refresh',
  '/auth/verify-2fa',
]);

const AUTH_EXEMPT_PATHS = new Set([
  '/auth/csrf-token',
  '/auth/login',
  '/auth/password/forgot',
  '/auth/password/reset',
  '/auth/register',
  '/auth/verify-email',
  '/billing/webhook',
]);

function pathFromSegments(segments: string[]): string {
  return `/${segments.join('/')}`;
}

function shouldAttachBearer(path: string): boolean {
  return !AUTH_EXEMPT_PATHS.has(path);
}

function createProxyHeaders(request: NextRequest, path: string): Headers {
  const headers = new Headers();

  for (const [key, value] of request.headers.entries()) {
    const normalized = key.toLowerCase();
    if (FORWARDED_REQUEST_HEADERS.has(normalized)) {
      headers.set(key, value);
    }
  }

  const forwardedCookies = request.cookies
    .getAll()
    .map(({ name, value }) => `${name}=${value}`)
    .join('; ');

  if (forwardedCookies) {
    headers.set('cookie', forwardedCookies);
  }

  const token = request.cookies.get(SESSION_COOKIE_NAME)?.value;
  if (token && shouldAttachBearer(path) && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  return headers;
}

function applySharedResponseHeaders(upstream: Response, response: NextResponse) {
  for (const [key, value] of upstream.headers.entries()) {
    if (FORWARDED_RESPONSE_HEADERS.has(key.toLowerCase())) {
      response.headers.set(key, value);
    }
  }
}

function setSessionCookie(response: NextResponse, token: string) {
  response.cookies.set({
    name: SESSION_COOKIE_NAME,
    value: token,
    path: '/',
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    maxAge: SESSION_COOKIE_MAX_AGE,
  });
}

function setCSRFCookie(response: NextResponse, token: string) {
  response.cookies.set({
    name: CSRF_COOKIE_NAME,
    value: token,
    path: '/',
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    maxAge: CSRF_COOKIE_MAX_AGE,
  });
}

function clearAuthCookies(response: NextResponse) {
  response.cookies.set({
    name: SESSION_COOKIE_NAME,
    value: '',
    path: '/',
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    maxAge: 0,
  });
  response.cookies.set({
    name: CSRF_COOKIE_NAME,
    value: '',
    path: '/',
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    maxAge: 0,
  });
}

function parseJSONSafe(text: string): any | null {
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

async function proxyRequest(request: NextRequest, segments: string[]) {
  const path = pathFromSegments(segments);
  const targetUrl = new URL(`/api${path}`, API_ORIGIN);
  targetUrl.search = request.nextUrl.search;

  const headers = createProxyHeaders(request, path);
  const upstream = await fetch(targetUrl, {
    method: request.method,
    headers,
    body:
      request.method === 'GET' || request.method === 'HEAD'
        ? undefined
        : await request.arrayBuffer(),
    cache: 'no-store',
    redirect: 'manual',
  });

  const contentType = upstream.headers.get('content-type') || 'application/json; charset=utf-8';

  if (contentType.includes('application/json')) {
    const text = await upstream.text();
    const response = new NextResponse(text || null, { status: upstream.status });
    applySharedResponseHeaders(upstream, response);
    response.headers.set('content-type', contentType);

    const payload = parseJSONSafe(text);
    if (upstream.ok && AUTH_COOKIE_PATHS.has(path) && typeof payload?.token === 'string') {
      setSessionCookie(response, payload.token);
    }
    if (upstream.ok && path === '/auth/csrf-token' && typeof payload?.csrf_token === 'string') {
      setCSRFCookie(response, payload.csrf_token);
    }
    if (path === '/auth/logout' || (path === '/auth/me' && upstream.status === 401)) {
      clearAuthCookies(response);
    }

    return response;
  }

  const response = new NextResponse(await upstream.arrayBuffer(), { status: upstream.status });
  applySharedResponseHeaders(upstream, response);
  response.headers.set('content-type', contentType);

  if (path === '/auth/logout') {
    clearAuthCookies(response);
  }

  return response;
}

type RouteContext = {
  params: Promise<{ path: string[] }> | { path: string[] };
};

async function resolveSegments(context: RouteContext): Promise<string[]> {
  const params = await context.params;
  return params.path;
}

export async function GET(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, await resolveSegments(context));
}

export async function POST(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, await resolveSegments(context));
}

export async function PUT(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, await resolveSegments(context));
}

export async function PATCH(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, await resolveSegments(context));
}

export async function DELETE(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, await resolveSegments(context));
}

export async function OPTIONS(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, await resolveSegments(context));
}
