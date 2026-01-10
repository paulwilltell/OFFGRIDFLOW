import Link from 'next/link';

export default function HomePage() {
  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
      color: 'white',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '2rem',
      textAlign: 'center'
    }}>
      <h1 style={{ fontSize: '3rem', marginBottom: '1rem', fontWeight: 'bold' }}>
        OffGridFlow
      </h1>
      <p style={{ fontSize: '1.5rem', color: '#94a3b8', marginBottom: '2rem', maxWidth: '600px' }}>
        Enterprise Carbon Emissions Tracking and CSRD Compliance Platform
      </p>
      <p style={{ fontSize: '1.1rem', color: '#64748b', marginBottom: '3rem', maxWidth: '500px' }}>
        Track Scope 1, 2 and 3 emissions. Automate CSRD, SEC, and CBAM reporting. 
        10x cheaper than Watershed.
      </p>
      <div style={{ display: 'flex', gap: '1rem' }}>
        <Link href="/login" style={{
          background: '#22c55e',
          color: 'white',
          padding: '1rem 2rem',
          borderRadius: '8px',
          fontWeight: 'bold',
          textDecoration: 'none'
        }}>
          Get Started
        </Link>
        <Link href="/pricing" style={{
          background: 'transparent',
          border: '2px solid #22c55e',
          color: '#22c55e',
          padding: '1rem 2rem',
          borderRadius: '8px',
          fontWeight: 'bold',
          textDecoration: 'none'
        }}>
          View Pricing
        </Link>
      </div>
      <p style={{ marginTop: '4rem', color: '#475569', fontSize: '0.9rem' }}>
        Trusted by sustainability teams across the EU
      </p>
    </div>
  );
}