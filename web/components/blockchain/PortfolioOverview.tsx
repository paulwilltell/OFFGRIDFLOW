'use client';

interface Portfolio {
  totalValue: number;
  totalCredits: number;
  credits: CreditHolding[];
}

interface CreditHolding {
  id: string;
  projectName: string;
  vintage: number;
  quantity: number;
  co2ePerCredit: number;
  currentPrice: number;
  totalValue: number;
}

interface PortfolioOverviewProps {
  portfolio: Portfolio | null;
  loading: boolean;
}

export default function PortfolioOverview({ portfolio, loading }: PortfolioOverviewProps) {
  if (loading) {
    return (
      <div style={{ padding: '1.5rem', background: '#0f172a', borderRadius: '12px', border: '1px solid #1d2940' }}>
        <h3 style={{ fontSize: '1.1rem', marginBottom: '1rem', color: '#8aa9ff' }}>Portfolio Overview</h3>
        <div style={{ color: '#888' }}>Loading portfolio data...</div>
      </div>
    );
  }

  if (!portfolio || portfolio.credits.length === 0) {
    return (
      <div style={{ padding: '1.5rem', background: '#0f172a', borderRadius: '12px', border: '1px solid #1d2940' }}>
        <h3 style={{ fontSize: '1.1rem', marginBottom: '1rem', color: '#8aa9ff' }}>Portfolio Overview</h3>
        <div style={{ textAlign: 'center', padding: '2rem 0' }}>
          <div style={{ fontSize: '3rem', marginBottom: '0.5rem' }}>📂</div>
          <p style={{ color: '#888', fontSize: '0.9rem' }}>No credits in your portfolio</p>
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: '1.5rem', background: '#0f172a', borderRadius: '12px', border: '1px solid #1d2940' }}>
      <h3 style={{ fontSize: '1.1rem', marginBottom: '1rem', color: '#8aa9ff' }}>Portfolio Overview</h3>
      
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1.5rem' }}>
        <div style={{ padding: '1rem', background: '#1d2940', borderRadius: '8px' }}>
          <div style={{ fontSize: '0.8rem', color: '#888', marginBottom: '0.25rem' }}>Total Value</div>
          <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>${portfolio.totalValue.toLocaleString()}</div>
        </div>
        <div style={{ padding: '1rem', background: '#1d2940', borderRadius: '8px' }}>
          <div style={{ fontSize: '0.8rem', color: '#888', marginBottom: '0.25rem' }}>Total Credits</div>
          <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>{portfolio.totalCredits.toLocaleString()}</div>
        </div>
      </div>

      <div style={{ fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.75rem', color: '#fff' }}>Holdings</div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
        {portfolio.credits.map((credit) => (
          <div
            key={credit.id}
            style={{
              padding: '0.75rem',
              background: '#1a1f36',
              borderRadius: '6px',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <div>
              <div style={{ fontSize: '0.9rem', fontWeight: 600, marginBottom: '0.25rem' }}>
                {credit.projectName}
              </div>
              <div style={{ fontSize: '0.75rem', color: '#888' }}>
                {credit.quantity} credits • {credit.vintage} vintage
              </div>
            </div>
            <div style={{ textAlign: 'right' }}>
              <div style={{ fontSize: '0.9rem', fontWeight: 600 }}>${credit.totalValue.toLocaleString()}</div>
              <div style={{ fontSize: '0.75rem', color: '#888' }}>${credit.currentPrice}/credit</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
