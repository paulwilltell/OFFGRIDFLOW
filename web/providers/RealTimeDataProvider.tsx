'use client';

import React, { createContext, useContext, useEffect, useRef, useCallback, useMemo, useState } from 'react';
import { EmissionData } from '@/stores/carbonStore';

type UpdateCallback = (data: Partial<EmissionData>) => void;
type UnsubscribeFunction = () => void;
type ConnectionStatus = 'disabled' | 'connecting' | 'connected' | 'offline';

interface RealTimeContextValue {
  subscribe: (tenantId: string, callback: UpdateCallback) => UnsubscribeFunction;
  status: ConnectionStatus;
  isConnected: boolean;
}

const RealTimeContext = createContext<RealTimeContextValue | null>(null);

// -----------------------------------------------------------------------------
// Subscription manager
//
// Design principles (Million Fold Precision):
//   1. NEVER fabricate data. If real-time data is unavailable, the UI stays on
//      the last known real values from the REST API.
//   2. NEVER crash the host app. Every WebSocket interaction is wrapped in
//      try/catch and failures are logged but non-fatal.
//   3. NEVER touch auth state. WebSocket errors do not clear tokens or force
//      redirects — they are a non-critical enhancement layer.
//   4. Be honest. Status is one of: disabled, connecting, connected, offline.
// -----------------------------------------------------------------------------

class SubscriptionManager {
  private subscribers: Map<string, Set<UpdateCallback>> = new Map();
  private statusListeners: Set<(status: ConnectionStatus) => void> = new Set();
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private readonly maxReconnectAttempts = 3;
  private readonly reconnectDelay = 2000;
  private status: ConnectionStatus = 'disabled';
  private disabled = false;

  private setStatus(status: ConnectionStatus) {
    if (this.status === status) return;
    this.status = status;
    this.statusListeners.forEach((listener) => {
      try { listener(status); } catch { /* listener errors never propagate */ }
    });
  }

  getStatus(): ConnectionStatus {
    return this.status;
  }

  onStatusChange(listener: (status: ConnectionStatus) => void): () => void {
    this.statusListeners.add(listener);
    listener(this.status);
    return () => { this.statusListeners.delete(listener); };
  }

  connect(baseUrl?: string) {
    if (typeof window === 'undefined') return;
    if (this.disabled) return;
    if (this.ws?.readyState === WebSocket.OPEN) return;
    if (this.ws?.readyState === WebSocket.CONNECTING) return;

    const wsUrl = baseUrl || this.getWebSocketUrl();
    if (!wsUrl) {
      this.setStatus('disabled');
      return;
    }

    this.setStatus('connecting');

    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        this.reconnectAttempts = 0;
        this.setStatus('connected');
        try {
          this.subscribers.forEach((_, tenantId) => this.sendSubscription(tenantId));
        } catch { /* ignore */ }
      };

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.handleMessage(message);
        } catch {
          // Silently drop malformed messages. Never throw from ws handlers.
        }
      };

      this.ws.onclose = () => {
        this.setStatus('offline');
        this.attemptReconnect();
      };

      this.ws.onerror = () => {
        // WebSocket error — log to console for devs but do not surface to UI.
        // Browser prints its own error to the console; no additional log needed.
      };
    } catch {
      // Constructor can throw on invalid URL or mixed content. Silently disable.
      this.disabled = true;
      this.setStatus('disabled');
    }
  }

  private getWebSocketUrl(): string | null {
    try {
      const override = process.env.NEXT_PUBLIC_WS_URL;
      if (override) return override;

      // If no explicit WS URL is configured, don't try to connect.
      // (Railway deployment uses a separate API host — we only enable WS
      // when NEXT_PUBLIC_WS_URL is explicitly set.)
      const wsHost = process.env.NEXT_PUBLIC_WS_HOST;
      if (!wsHost) return null;

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${wsHost}/ws/emissions`;
    } catch {
      return null;
    }
  }

  private sendSubscription(tenantId: string) {
    if (this.ws?.readyState !== WebSocket.OPEN) return;
    try {
      this.ws.send(JSON.stringify({ type: 'subscribe', tenantId }));
    } catch {
      // Send can fail if socket is closing. Non-fatal.
    }
  }

  private handleMessage(message: { type?: string; tenantId?: string; data?: Partial<EmissionData> }) {
    if (message?.type !== 'emission_update' || !message.tenantId || !message.data) return;
    const callbacks = this.subscribers.get(message.tenantId);
    if (!callbacks) return;
    callbacks.forEach((callback) => {
      try { callback(message.data!); } catch { /* subscriber errors never propagate */ }
    });
  }

  private attemptReconnect() {
    if (this.disabled) return;
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      // Give up silently. The dashboard continues to work with REST-only data.
      this.disabled = true;
      this.setStatus('disabled');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  subscribe(tenantId: string, callback: UpdateCallback): UnsubscribeFunction {
    if (!this.subscribers.has(tenantId)) {
      this.subscribers.set(tenantId, new Set());
    }
    this.subscribers.get(tenantId)!.add(callback);

    if (this.status === 'connected') {
      this.sendSubscription(tenantId);
    }

    return () => {
      const callbacks = this.subscribers.get(tenantId);
      if (!callbacks) return;
      callbacks.delete(callback);
      if (callbacks.size === 0) {
        this.subscribers.delete(tenantId);
      }
    };
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      try { this.ws.close(); } catch { /* ignore */ }
      this.ws = null;
    }
    this.subscribers.clear();
    this.setStatus('disabled');
  }
}

// Singleton instance (browser-scoped)
const subscriptionManager = new SubscriptionManager();

// Static accessors for non-React code
export const RealTimeDataProvider = {
  subscribe: (tenantId: string, callback: UpdateCallback): UnsubscribeFunction =>
    subscriptionManager.subscribe(tenantId, callback),
  connect: (baseUrl?: string) => subscriptionManager.connect(baseUrl),
  disconnect: () => subscriptionManager.disconnect(),
  getStatus: (): ConnectionStatus => subscriptionManager.getStatus(),
};

// -----------------------------------------------------------------------------
// React Provider
// -----------------------------------------------------------------------------

interface RealTimeProviderProps {
  children?: React.ReactNode;
  tenantId: string;
  onUpdate?: UpdateCallback;
  baseUrl?: string;
}

export function RealTimeProvider({ children, tenantId, onUpdate, baseUrl }: RealTimeProviderProps) {
  const callbackRef = useRef(onUpdate);
  callbackRef.current = onUpdate;

  const [status, setStatus] = useState<ConnectionStatus>(subscriptionManager.getStatus());

  // Track connection status as actual React state so the UI re-renders.
  useEffect(() => {
    const unsubscribeStatus = subscriptionManager.onStatusChange(setStatus);
    return unsubscribeStatus;
  }, []);

  useEffect(() => {
    try {
      subscriptionManager.connect(baseUrl);
    } catch {
      // Connection setup failures never crash the parent tree.
    }

    const unsubscribe = subscriptionManager.subscribe(tenantId, (data) => {
      try { callbackRef.current?.(data); } catch { /* ignore downstream errors */ }
    });

    return () => {
      try { unsubscribe(); } catch { /* ignore */ }
    };
  }, [tenantId, baseUrl]);

  const contextValue = useMemo<RealTimeContextValue>(
    () => ({
      subscribe: RealTimeDataProvider.subscribe,
      status,
      isConnected: status === 'connected',
    }),
    [status],
  );

  return (
    <RealTimeErrorBoundary>
      <RealTimeContext.Provider value={contextValue}>
        {children}
        <RealTimeIndicator status={status} />
      </RealTimeContext.Provider>
    </RealTimeErrorBoundary>
  );
}

export function useRealTime(): RealTimeContextValue {
  const context = useContext(RealTimeContext);
  if (!context) {
    // Return a no-op context when used outside a provider. Never throw —
    // a missing provider must not crash a dashboard that works fine without
    // real-time updates.
    return {
      subscribe: () => () => {},
      status: 'disabled',
      isConnected: false,
    };
  }
  return context;
}

// -----------------------------------------------------------------------------
// Error boundary — isolates any WebSocket / subscription error from the host
// React tree so a real-time failure never takes down the dashboard.
// -----------------------------------------------------------------------------

interface ErrorBoundaryProps { children?: React.ReactNode }
interface ErrorBoundaryState { hasError: boolean }

class RealTimeErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch() {
    // Real-time layer failed — disable it so future renders are quiet.
    try { subscriptionManager.disconnect(); } catch { /* ignore */ }
  }

  render() {
    if (this.state.hasError) {
      // Render children without the real-time layer. The dashboard continues
      // to function with REST-only data.
      return <>{this.props.children}</>;
    }
    return <>{this.props.children}</>;
  }
}

// -----------------------------------------------------------------------------
// Status indicator
//
// Only shown when real-time is actively connected. A frozen "Connecting..."
// pill in the corner of every page is worse than no indicator at all.
// -----------------------------------------------------------------------------

function RealTimeIndicator({ status }: { status: ConnectionStatus }) {
  if (status !== 'connected') return null;
  return (
    <div className="fixed bottom-4 right-4 flex items-center gap-2 px-3 py-1.5 bg-gray-900/80 backdrop-blur-sm rounded-full text-xs pointer-events-none">
      <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
      <span className="text-gray-300">Live</span>
    </div>
  );
}

export default RealTimeProvider;
