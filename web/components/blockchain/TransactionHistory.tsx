'use client';

import { useEffect, useState } from 'react';
import { api, ApiRequestError } from '../../lib/api';

interface Transaction {
  id: string;
  type: 'mint' | 'buy' | 'sell' | 'transfer';
  timestamp: string;
  amount: number;
  price?: number;
  from?: string;
  to?: string;
  txHash: string;
  status: 'pending' | 'confirmed' | 'failed';
}

interface TransactionHistoryProps {
  walletConnected: boolean;
}

export default function TransactionHistory({ walletConnected }: TransactionHistoryProps) {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!walletConnected) return;

    const fetchTransactions = async () => {
      setLoading(true);
      setError(null);

      try {
        const data = await api.get<Transaction[]>('/api/blockchain/transactions');
        setTransactions(data);
      } catch (err) {
        if (err instanceof ApiRequestError) {
          if (err.status === 404) {
            setError('Transaction history not available');
          } else {
            setError(err.message);
          }
        } else {
          setError('Failed to load transaction history');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchTransactions();
  }, [walletConnected]);

  if (!walletConnected) {
    return null;
  }

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'mint': return '⚡';
      case 'buy': return '🛒';
      case 'sell': return '💰';
      case 'transfer': return '📤';
      default: return '📝';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'confirmed': return '#10b981';
      case 'pending': return '#eab308';
      case 'failed': return '#ef4444';
      default: return '#888';
    }
  };

  return (
    <div style={{ padding: '1.5rem', background: '#0f172a', borderRadius: '12px', border: '1px solid #1d2940' }}>
      <h3 style={{ fontSize: '1.1rem', marginBottom: '1rem', color: '#8aa9ff' }}>Transaction History</h3>

      {loading ? (
        <div style={{ color: '#888', textAlign: 'center', padding: '2rem 0' }}>Loading transactions...</div>
      ) : error ? (
        <div style={{ color: '#fca5a5', textAlign: 'center', padding: '2rem 0' }}>
          <div style={{ marginBottom: '0.5rem' }}>⚠️</div>
          <div>{error}</div>
        </div>
      ) : transactions.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem 0' }}>
          <div style={{ fontSize: '3rem', marginBottom: '0.5rem' }}>📋</div>
          <p style={{ color: '#888' }}>No transactions yet</p>
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '600px' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid #1d2940' }}>
                <th style={{ padding: '0.75rem', textAlign: 'left', fontSize: '0.85rem', color: '#888' }}>Type</th>
                <th style={{ padding: '0.75rem', textAlign: 'left', fontSize: '0.85rem', color: '#888' }}>Date</th>
                <th style={{ padding: '0.75rem', textAlign: 'right', fontSize: '0.85rem', color: '#888' }}>Amount</th>
                <th style={{ padding: '0.75rem', textAlign: 'right', fontSize: '0.85rem', color: '#888' }}>Price</th>
                <th style={{ padding: '0.75rem', textAlign: 'left', fontSize: '0.85rem', color: '#888' }}>Status</th>
                <th style={{ padding: '0.75rem', textAlign: 'left', fontSize: '0.85rem', color: '#888' }}>Tx Hash</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((tx) => (
                <tr key={tx.id} style={{ borderTop: '1px solid #1d2940' }}>
                  <td style={{ padding: '0.75rem' }}>
                    <span style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <span>{getTypeIcon(tx.type)}</span>
                      <span style={{ textTransform: 'capitalize' }}>{tx.type}</span>
                    </span>
                  </td>
                  <td style={{ padding: '0.75rem', fontSize: '0.85rem', color: '#888' }}>
                    {new Date(tx.timestamp).toLocaleDateString()}
                  </td>
                  <td style={{ padding: '0.75rem', textAlign: 'right', fontWeight: 600 }}>
                    {tx.amount}
                  </td>
                  <td style={{ padding: '0.75rem', textAlign: 'right', color: '#888' }}>
                    {tx.price ? `$${tx.price}` : '—'}
                  </td>
                  <td style={{ padding: '0.75rem' }}>
                    <span
                      style={{
                        padding: '0.25rem 0.5rem',
                        background: `${getStatusColor(tx.status)}20`,
                        color: getStatusColor(tx.status),
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        textTransform: 'capitalize',
                      }}
                    >
                      {tx.status}
                    </span>
                  </td>
                  <td style={{ padding: '0.75rem', fontSize: '0.85rem', fontFamily: 'monospace' }}>
                    <a
                      href={`https://etherscan.io/tx/${tx.txHash}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{ color: '#8aa9ff', textDecoration: 'none' }}
                    >
                      {tx.txHash.slice(0, 6)}...{tx.txHash.slice(-4)}
                    </a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
