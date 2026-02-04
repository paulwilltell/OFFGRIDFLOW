'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';

export default function HomePage() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [stats, setStats] = useState({
    emissions: '127.8M',
    organizations: '2,847',
    reports: '24,912',
    dataQuality: '99.4%'
  });

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Set canvas size
    const updateCanvasSize = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };
    updateCanvasSize();
    window.addEventListener('resize', updateCanvasSize);

    // Globe animation
    let frame = 0;
    const nodes = [
      { x: 0.3, y: 0.3, label: 'AWS' },
      { x: 0.7, y: 0.3, label: 'Azure' },
      { x: 0.5, y: 0.6, label: 'GCP' },
      { x: 0.4, y: 0.7, label: 'SAP' }
    ];

    const animate = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      frame += 0.02;

      // Draw connections
      ctx.strokeStyle = 'rgba(34, 197, 94, 0.2)';
      ctx.lineWidth = 1;
      for (let i = 0; i < nodes.length; i++) {
        for (let j = i + 1; j < nodes.length; j++) {
          ctx.beginPath();
          ctx.moveTo(nodes[i].x * canvas.width, nodes[i].y * canvas.height);
          ctx.lineTo(nodes[j].x * canvas.width, nodes[j].y * canvas.height);
          ctx.stroke();
        }
      }

      // Draw nodes with pulse
      nodes.forEach((node, i) => {
        const pulse = Math.sin(frame + i) * 3 + 8;
        const x = node.x * canvas.width;
        const y = node.y * canvas.height;

        // Pulse glow
        ctx.beginPath();
        ctx.arc(x, y, pulse + 5, 0, Math.PI * 2);
        const gradient = ctx.createRadialGradient(x, y, 0, x, y, pulse + 5);
        gradient.addColorStop(0, 'rgba(34, 197, 94, 0.3)');
        gradient.addColorStop(1, 'rgba(34, 197, 94, 0)');
        ctx.fillStyle = gradient;
        ctx.fill();

        // Node circle
        ctx.beginPath();
        ctx.arc(x, y, pulse, 0, Math.PI * 2);
        ctx.fillStyle = '#22c55e';
        ctx.fill();

        // Label
        ctx.fillStyle = '#fff';
        ctx.font = '14px system-ui';
        ctx.textAlign = 'center';
        ctx.fillText(node.label, x, y - 15);
      });

      requestAnimationFrame(animate);
    };
    animate();

    return () => {
      window.removeEventListener('resize', updateCanvasSize);
    };
  }, []);

  return (
    <div style={{
      position: 'relative',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%)',
      color: 'white',
      overflow: 'hidden'
    }}>
      {/* Animated background canvas */}
      <canvas
        ref={canvasRef}
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          opacity: 0.4
        }}
      />

      {/* Header */}
      <header style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        padding: '1rem 2rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        background: 'rgba(15, 23, 42, 0.8)',
        backdropFilter: 'blur(10px)',
        borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
        zIndex: 1000
      }}>
        <div style={{ fontSize: '1.5rem', fontWeight: 'bold', background: 'linear-gradient(135deg, #22c55e, #10b981)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
          OffGridFlow
        </div>
        <nav style={{ display: 'flex', gap: '2rem', alignItems: 'center' }}>
          <Link href="/login" style={{ color: '#94a3b8', textDecoration: 'none' }}>Login</Link>
          <Link href="/register" style={{
            background: '#22c55e',
            color: 'white',
            padding: '0.5rem 1.5rem',
            borderRadius: '8px',
            textDecoration: 'none',
            fontWeight: 500
          }}>Get Started</Link>
        </nav>
      </header>

      {/* Hero Section */}
      <main style={{ position: 'relative', zIndex: 1, paddingTop: '100px' }}>
        <section style={{
          maxWidth: '1200px',
          margin: '0 auto',
          padding: '4rem 2rem',
          textAlign: 'center'
        }}>
          <h1 style={{
            fontSize: 'clamp(2.5rem, 6vw, 4rem)',
            fontWeight: 'bold',
            marginBottom: '1.5rem',
            background: 'linear-gradient(135deg, #fff, #94a3b8)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            lineHeight: 1.2
          }}>
            Enterprise Carbon Emissions Tracking
          </h1>
          <p style={{
            fontSize: 'clamp(1.1rem, 2vw, 1.5rem)',
            color: '#94a3b8',
            marginBottom: '3rem',
            maxWidth: '800px',
            margin: '0 auto 3rem'
          }}>
            Automated CSRD, SEC, and CBAM reporting. Track Scope 1, 2, and 3 emissions with enterprise-grade accuracy.
          </p>

          {/* CTA Buttons */}
          <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', marginBottom: '4rem' }}>
            <Link href="/register" style={{
              background: 'linear-gradient(135deg, #22c55e, #10b981)',
              color: 'white',
              padding: '1rem 2.5rem',
              borderRadius: '12px',
              textDecoration: 'none',
              fontWeight: 'bold',
              fontSize: '1.1rem',
              boxShadow: '0 10px 30px rgba(34, 197, 94, 0.3)',
              transition: 'transform 0.2s',
              display: 'inline-block'
            }}>
              Start Free Trial
            </Link>
            <Link href="/demo" style={{
              background: 'rgba(255, 255, 255, 0.05)',
              backdropFilter: 'blur(10px)',
              border: '2px solid rgba(34, 197, 94, 0.5)',
              color: 'white',
              padding: '1rem 2.5rem',
              borderRadius: '12px',
              textDecoration: 'none',
              fontWeight: 'bold',
              fontSize: '1.1rem'
            }}>
              View Demo
            </Link>
          </div>

          {/* Live Stats */}
          <div style={{
            background: 'rgba(255, 255, 255, 0.03)',
            backdropFilter: 'blur(20px)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '20px',
            padding: '2rem',
            marginBottom: '4rem',
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3)'
          }}>
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: '2rem'
            }}>
              <div>
                <div style={{ fontSize: '2.5rem', fontWeight: 'bold', color: '#22c55e', marginBottom: '0.5rem' }}>
                  {stats.emissions}
                </div>
                <div style={{ color: '#94a3b8', fontSize: '0.9rem' }}>tCO₂e Tracked</div>
              </div>
              <div>
                <div style={{ fontSize: '2.5rem', fontWeight: 'bold', color: '#22c55e', marginBottom: '0.5rem' }}>
                  {stats.organizations}
                </div>
                <div style={{ color: '#94a3b8', fontSize: '0.9rem' }}>Organizations</div>
              </div>
              <div>
                <div style={{ fontSize: '2.5rem', fontWeight: 'bold', color: '#22c55e', marginBottom: '0.5rem' }}>
                  {stats.reports}
                </div>
                <div style={{ color: '#94a3b8', fontSize: '0.9rem' }}>Reports Generated</div>
              </div>
              <div>
                <div style={{ fontSize: '2.5rem', fontWeight: 'bold', color: '#22c55e', marginBottom: '0.5rem' }}>
                  {stats.dataQuality}
                </div>
                <div style={{ color: '#94a3b8', fontSize: '0.9rem' }}>Data Quality</div>
              </div>
            </div>
          </div>

          {/* Feature Cards */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
            gap: '2rem',
            marginTop: '4rem'
          }}>
            {[
              { title: 'Multi-Cloud Integration', desc: 'Connect AWS, Azure, GCP, and SAP automatically', icon: '☁️' },
              { title: 'Real-Time Tracking', desc: 'Monitor emissions across all scopes in real-time', icon: '📊' },
              { title: 'CSRD Compliance', desc: 'Automated E1-6 reporting for EU regulations', icon: '✓' },
              { title: '10x Cost Savings', desc: 'Enterprise features at fraction of Big 4 cost', icon: '💰' }
            ].map((feature, i) => (
              <div key={i} style={{
                background: 'rgba(255, 255, 255, 0.03)',
                backdropFilter: 'blur(20px)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '16px',
                padding: '2rem',
                textAlign: 'left',
                transition: 'all 0.3s'
              }}>
                <div style={{ fontSize: '2.5rem', marginBottom: '1rem' }}>{feature.icon}</div>
                <h3 style={{ fontSize: '1.3rem', fontWeight: 'bold', marginBottom: '0.5rem' }}>
                  {feature.title}
                </h3>
                <p style={{ color: '#94a3b8', fontSize: '0.95rem' }}>
                  {feature.desc}
                </p>
              </div>
            ))}
          </div>
        </section>

        {/* Pricing Teaser */}
        <section style={{
          background: 'rgba(34, 197, 94, 0.05)',
          borderTop: '1px solid rgba(34, 197, 94, 0.2)',
          padding: '4rem 2rem',
          textAlign: 'center',
          marginTop: '4rem'
        }}>
          <h2 style={{ fontSize: '2rem', fontWeight: 'bold', marginBottom: '1rem' }}>
            Transparent Pricing
          </h2>
          <p style={{ color: '#94a3b8', fontSize: '1.1rem', marginBottom: '2rem' }}>
            Starting at $2,500/year. No hidden fees. Cancel anytime.
          </p>
          <Link href="/pricing" style={{
            color: '#22c55e',
            textDecoration: 'none',
            fontWeight: 'bold',
            fontSize: '1.1rem'
          }}>
            View Full Pricing →
          </Link>
        </section>
      </main>

      {/* Footer */}
      <footer style={{
        position: 'relative',
        zIndex: 1,
        borderTop: '1px solid rgba(255, 255, 255, 0.1)',
        padding: '2rem',
        textAlign: 'center',
        color: '#64748b',
        fontSize: '0.9rem'
      }}>
        <p>© 2026 OffGridFlow LLC. All rights reserved.</p>
      </footer>
    </div>
  );
}
