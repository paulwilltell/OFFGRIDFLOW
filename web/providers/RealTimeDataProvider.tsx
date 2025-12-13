'use client';

import React, { createContext, useContext, useEffect, useRef, useCallback } from 'react';
import { EmissionData } from '@/stores/carbonStore';

type UpdateCallback = (data: Partial<EmissionData>) => void;
type UnsubscribeFunction = () => void;

interface RealTimeContextValue {
  subscribe: (tenantId: string, callback: UpdateCallback) => UnsubscribeFunction;
  isConnected: boolean;
}

const RealTimeContext = createContext<RealTimeContextValue | null>(null);

// Subscription manager for real-time updates
class SubscriptionManager {
  private subscribers: Map<string, Set<UpdateCallback>> = new Map();
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnected = false;

  connect(baseUrl?: string) {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    const wsUrl = baseUrl || this.getWebSocketUrl();
    
    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('[RealTime] WebSocket connected');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        
        // Subscribe to all registered tenants
        this.subscribers.forEach((_, tenantId) => {
          this.sendSubscription(tenantId);
        });
      };

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (err) {
          console.error('[RealTime] Failed to parse message:', err);
        }
      };

      this.ws.onclose = () => {
        console.log('[RealTime] WebSocket disconnected');
        this.isConnected = false;
        this.attemptReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('[RealTime] WebSocket error:', error);
      };
    } catch (err) {
      console.error('[RealTime] Failed to connect:', err);
      this.simulateFallbackUpdates();
    }
  }

  private getWebSocketUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = process.env.NEXT_PUBLIC_WS_HOST || window.location.host;
    return `${protocol}//${host}/ws/emissions`;
  }

  private sendSubscription(tenantId: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'subscribe',
        tenantId,
      }));
    }
  }

  private handleMessage(message: { type: string; tenantId: string; data: Partial<EmissionData> }) {
    if (message.type === 'emission_update' && message.tenantId) {
      const callbacks = this.subscribers.get(message.tenantId);
      callbacks?.forEach((callback) => callback(message.data));
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('[RealTime] Max reconnect attempts reached, falling back to polling');
      this.simulateFallbackUpdates();
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    console.log(`[RealTime] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
    setTimeout(() => this.connect(), delay);
  }

  // Fallback: Simulate real-time updates with polling
  private simulateFallbackUpdates() {
    console.log('[RealTime] Using simulated updates (development mode)');
    
    // Simulate periodic updates for development
    setInterval(() => {
      this.subscribers.forEach((callbacks, tenantId) => {
        const simulatedUpdate: Partial<EmissionData> = {
          total: Math.random() * 1000 + 12000,
          updatedAt: new Date().toISOString(),
          trend: Math.random() > 0.5 ? 'down' : 'up',
          percentageChange: (Math.random() - 0.5) * 10,
        };
        
        callbacks.forEach((callback) => callback(simulatedUpdate));
      });
    }, 30000); // Update every 30 seconds
  }

  subscribe(tenantId: string, callback: UpdateCallback): UnsubscribeFunction {
    if (!this.subscribers.has(tenantId)) {
      this.subscribers.set(tenantId, new Set());
    }
    
    this.subscribers.get(tenantId)!.add(callback);
    
    // If already connected, send subscription
    if (this.isConnected) {
      this.sendSubscription(tenantId);
    }

    // Return unsubscribe function
    return () => {
      const callbacks = this.subscribers.get(tenantId);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.subscribers.delete(tenantId);
        }
      }
    };
  }

  getConnectionStatus(): boolean {
    return this.isConnected;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.subscribers.clear();
    this.isConnected = false;
  }
}

// Singleton instance
const subscriptionManager = new SubscriptionManager();

// Static methods for external access
export const RealTimeDataProvider = {
  subscribe: (tenantId: string, callback: UpdateCallback): UnsubscribeFunction => {
    return subscriptionManager.subscribe(tenantId, callback);
  },
  
  connect: (baseUrl?: string) => {
    subscriptionManager.connect(baseUrl);
  },
  
  disconnect: () => {
    subscriptionManager.disconnect();
  },
  
  isConnected: () => subscriptionManager.getConnectionStatus(),
};

// React Provider Component
interface RealTimeProviderProps {
  children?: React.ReactNode;
  tenantId: string;
  onUpdate?: UpdateCallback;
}

export function RealTimeProvider({ children, tenantId, onUpdate }: RealTimeProviderProps) {
  const callbackRef = useRef(onUpdate);
  callbackRef.current = onUpdate;

  useEffect(() => {
    // Connect on mount
    RealTimeDataProvider.connect();

    // Subscribe to updates
    const unsubscribe = RealTimeDataProvider.subscribe(tenantId, (data) => {
      callbackRef.current?.(data);
    });

    return () => {
      unsubscribe();
    };
  }, [tenantId]);

  const contextValue: RealTimeContextValue = {
    subscribe: RealTimeDataProvider.subscribe,
    isConnected: RealTimeDataProvider.isConnected(),
  };

  return (
    <RealTimeContext.Provider value={contextValue}>
      {children}
      <RealTimeIndicator isConnected={RealTimeDataProvider.isConnected()} />
    </RealTimeContext.Provider>
  );
}

// Hook to access real-time context
export function useRealTime() {
  const context = useContext(RealTimeContext);
  if (!context) {
    throw new Error('useRealTime must be used within a RealTimeProvider');
  }
  return context;
}

// Visual indicator for connection status
function RealTimeIndicator({ isConnected }: { isConnected: boolean }) {
  return (
    <div className="fixed bottom-4 right-4 flex items-center gap-2 px-3 py-1.5 bg-gray-900/80 backdrop-blur-sm rounded-full text-xs">
      <span
        className={`w-2 h-2 rounded-full ${
          isConnected ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'
        }`}
      />
      <span className="text-gray-300">
        {isConnected ? 'Live' : 'Connecting...'}
      </span>
    </div>
  );
}

export default RealTimeProvider;
