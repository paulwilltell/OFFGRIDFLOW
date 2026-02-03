import 'whatwg-fetch';
import '@testing-library/jest-dom';
import { TextEncoder, TextDecoder } from 'util';
import { TransformStream, ReadableStream, WritableStream } from 'stream/web';

if (!global.TextEncoder) {
  global.TextEncoder = TextEncoder;
}
if (!global.TextDecoder) {
  global.TextDecoder = TextDecoder as unknown as typeof global.TextDecoder;
}
if (!global.TransformStream) {
  global.TransformStream = TransformStream;
}
if (!global.ReadableStream) {
  global.ReadableStream = ReadableStream;
}
if (!global.WritableStream) {
  global.WritableStream = WritableStream;
}

if (!(global as typeof globalThis).BroadcastChannel) {
  class MockBroadcastChannel {
    name: string;
    onmessage: ((event: MessageEvent) => void) | null = null;

    constructor(name: string) {
      this.name = name;
    }

    postMessage = jest.fn();
    close = jest.fn();
    addEventListener = jest.fn();
    removeEventListener = jest.fn();
    dispatchEvent = jest.fn();
  }

  (global as typeof globalThis).BroadcastChannel = MockBroadcastChannel as unknown as typeof BroadcastChannel;
}

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
    back: jest.fn(),
  }),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => '/',
}));

// Mock localStorage
const localStorageStore: Record<string, string> = {};
const localStorageMock = {
  getItem: jest.fn((key: string) => (key in localStorageStore ? localStorageStore[key] : null)),
  setItem: jest.fn((key: string, value: string) => {
    localStorageStore[key] = value;
  }),
  removeItem: jest.fn((key: string) => {
    delete localStorageStore[key];
  }),
  clear: jest.fn(() => {
    Object.keys(localStorageStore).forEach((key) => delete localStorageStore[key]);
  }),
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

if (!window.matchMedia) {
  Object.defineProperty(window, 'matchMedia', {
    value: jest.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: jest.fn(), // deprecated
      removeListener: jest.fn(), // deprecated
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    })),
  });
}

// Mock URL.createObjectURL for download helpers
if (typeof URL.createObjectURL !== 'function') {
  Object.defineProperty(URL, 'createObjectURL', {
    value: jest.fn(() => 'blob:mock'),
  });
}
if (typeof URL.revokeObjectURL !== 'function') {
  Object.defineProperty(URL, 'revokeObjectURL', {
    value: jest.fn(),
  });
}

// Suppress console errors during tests
const originalError = console.error;
beforeAll(() => {
  console.error = (...args: unknown[]) => {
    if (
      typeof args[0] === 'string' &&
      args[0].includes('Warning: ReactDOM.render is no longer supported')
    ) {
      return;
    }
    originalError.call(console, ...args);
  };
});

afterAll(() => {
  console.error = originalError;
});
