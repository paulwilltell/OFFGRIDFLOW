'use client';

import { useState } from 'react';

interface WalletConnectProps {
  onConnect: (connected: boolean) => void;
}

export default function WalletConnect({ onConnect }: WalletConnectProps) {
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [address, setAddress] = useState<string | null>(null);

  const connectWallet = async () => {
    setConnecting(true);

    try {
      if (typeof window !== 'undefined' && (window as any).ethereum) {
        const accounts = await (window as any).ethereum.request({ 
          method: 'eth_requestAccounts' 
        });
        
        if (accounts && accounts.length > 0) {
          const addr = accounts[0];
          setAddress(addr);
          setConnected(true);
          onConnect(true);
        }
      } else {
        alert('MetaMask not detected. Please install MetaMask browser extension.');
      }
    } catch (error: any) {
      console.error('Wallet connection error:', error);
      alert(error.message || 'Failed to connect wallet');
    } finally {
      setConnecting(false);
    }
  };

  const disconnectWallet = () => {
    setConnected(false);
    setAddress(null);
    onConnect(false);
  };

  if (connected && address) {
    return (
      <div
        style={{
          padding: '0.5rem 1rem',
          background: '#10b981',
          color: '#fff',
          borderRadius: '8px',
          fontSize: '0.85rem',
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
        }}
      >
        <span>🔗</span>
        <span style={{ fontWeight: 600 }}>
          {address.slice(0, 6)}...{address.slice(-4)}
        </span>
        <button
          onClick={disconnectWallet}
          style={{
            background: 'transparent',
            border: 'none',
            color: '#fff',
            cursor: 'pointer',
            padding: '0.25rem',
            fontSize: '1rem',
          }}
          title="Disconnect wallet"
        >
          ✕
        </button>
      </div>
    );
  }

  return (
    <button
      onClick={connectWallet}
      disabled={connecting}
      style={{
        padding: '0.5rem 1rem',
        background: connecting ? '#1d2940' : '#8aa9ff',
        color: connecting ? '#888' : '#0a0f1e',
        border: 'none',
        borderRadius: '8px',
        fontSize: '0.85rem',
        fontWeight: 600,
        cursor: connecting ? 'not-allowed' : 'pointer',
        display: 'flex',
        alignItems: 'center',
        gap: '0.5rem',
      }}
    >
      <span>🔐</span>
      {connecting ? 'Connecting...' : 'Connect Wallet'}
    </button>
  );
}
