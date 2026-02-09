'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';

const personas = [
  {
    id: 'ciso',
    label: 'Security Leaders',
    headline: 'Harden risk posture with provable controls and audit-ready evidence.',
    bullets: [
      'SOC 2, GDPR, and ISO-ready audit trails with immutable change tracking',
      'Fine-grained RBAC and tenant isolation across global subsidiaries',
      'Data residency guardrails for multi-region compliance',
    ],
    kpi: { label: 'Audit prep time', value: '↓ 62%' },
  },
  {
    id: 'it',
    label: 'IT Directors',
    headline: 'Unify emissions data pipelines without ripping and replacing tooling.',
    bullets: [
      'Native integrations for ERP, cloud, and procurement systems',
      'API-first ingestion with schema validation and data lineage',
      'Scales to millions of records per month with predictable latency',
    ],
    kpi: { label: 'Integration time', value: '↓ 45%' },
  },
  {
    id: 'ops',
    label: 'Operations VPs',
    headline: 'Turn compliance into measurable operational savings.',
    bullets: [
      'Automated supplier onboarding and emissions validation',
      'Scenario modeling for cost and carbon reduction',
      'Executive-ready dashboards with regional rollups',
    ],
    kpi: { label: 'Reporting cycle', value: '↓ 70%' },
  },
];

const demoModules = [
  {
    id: 'security',
    label: 'Security & Compliance',
    outcome: 'Reduce audit effort by 40% with continuous evidence capture.',
    narrative:
      'Every data change is logged, signed, and stored in a compliance-ready timeline. Automated gap detection flags missing evidence before audits.',
    screens: ['Policy control map', 'Audit evidence timeline', 'Vendor risk review'],
    dataPoints: ['2,481 control checks', '97% evidence coverage', '0 critical gaps'],
    impact: [
      'SOC 2 evidence pack generated in minutes',
      'Role-based approvals across global teams',
      'Continuous monitoring with exception alerts',
    ],
  },
  {
    id: 'reporting',
    label: 'Reporting & Analytics',
    outcome: 'Deliver board-ready climate disclosures with confidence.',
    narrative:
      'Automated reporting templates align with CSRD, SEC, and SB 253. Interactive dashboards connect emissions drivers to financial impact.',
    screens: ['Executive scorecard', 'Scope 1-3 rollup', 'Disclosure builder'],
    dataPoints: ['14 regions', '3,912 facilities', '84% reduction roadmap'],
    impact: [
      'Exportable CSRD and SEC packages',
      'Scenario modeling tied to capex planning',
      'Granular drill-down for audit assurance',
    ],
  },
  {
    id: 'integration',
    label: 'Integration Workflow',
    outcome: 'Connect data sources in days, not months.',
    narrative:
      'Ingestion pipelines validate data quality in real time, with lineage tracking and automated reconciliation against ERP baselines.',
    screens: ['Connector health view', 'Data lineage graph', 'Reconciliation checks'],
    dataPoints: ['12 connectors live', '99.4% data quality', '24,681 records/day'],
    impact: [
      'Pre-built connectors for ERP and cloud',
      'Automated anomaly detection',
      'Versioned data lineage for traceability',
    ],
  },
];

const deepDiveModules = [
  {
    title: 'Security & Compliance',
    items: [
      'SOC 2, GDPR, HIPAA-ready control mapping',
      'SSO, SCIM provisioning, and advanced RBAC',
      'Immutable audit logs with exportable evidence packs',
    ],
  },
  {
    title: 'Scalability & Performance',
    items: [
      'Regional data residency enforcement',
      'High-volume ingestion with backpressure controls',
      'Real-time dashboards optimized for enterprise scale',
    ],
  },
  {
    title: 'Integration Ecosystem',
    items: [
      'ERP + procurement integrations (SAP, Oracle, Workday)',
      'Cloud telemetry ingestion (AWS, Azure, GCP)',
      'Open API for custom partner pipelines',
    ],
  },
  {
    title: 'Administration & Governance',
    items: [
      'Centralized user and tenant management',
      'Policy engines for approval workflows',
      'Regional governance with delegated controls',
    ],
  },
];

const resources = [
  'Enterprise datasheet (PDF)',
  'ROI calculator + cost savings model',
  'Implementation & onboarding guide',
  'API documentation overview',
  'Full case studies library',
  'Security & compliance whitepaper',
];

const trustLogos = [
  'GlobalEnergy',
  'Atlas Logistics',
  'Northwind Manufacturing',
  'Helios Financial',
  'Pinnacle Utilities',
];

const analystBadges = ['Gartner MQ', 'Forrester Wave', 'IDC MarketScape'];
const complianceBadges = ['SOC 2 Type II', 'ISO 27001', 'GDPR Ready', 'HIPAA Ready'];
const pressBadges = ['Forbes', 'WSJ', 'TechCrunch'];

export default function DemoPage() {
  const [activeModule, setActiveModule] = useState(demoModules[0].id);
  const [activePersona, setActivePersona] = useState(personas[0].id);

  const selectedModule = useMemo(
    () => demoModules.find((module) => module.id === activeModule) ?? demoModules[0],
    [activeModule]
  );
  const selectedPersona = useMemo(
    () => personas.find((persona) => persona.id === activePersona) ?? personas[0],
    [activePersona]
  );

  return (
    <div
      style={{
        minHeight: '100vh',
        background:
          'radial-gradient(circle at top left, rgba(56, 189, 248, 0.18), transparent 45%), radial-gradient(circle at 70% 20%, rgba(74, 222, 128, 0.18), transparent 50%), linear-gradient(135deg, #0b1222 0%, #0f172a 45%, #111827 100%)',
        color: '#e2e8f0',
      }}
    >
      <header
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          padding: '1rem 2.5rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          background: 'rgba(6, 10, 22, 0.75)',
          backdropFilter: 'blur(10px)',
          borderBottom: '1px solid rgba(148, 163, 184, 0.15)',
          zIndex: 1000,
        }}
      >
        <Link
          href="/"
          style={{
            fontSize: '1.4rem',
            fontWeight: 700,
            letterSpacing: '0.02em',
            background: 'linear-gradient(135deg, #38bdf8, #22c55e)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            textDecoration: 'none',
          }}
        >
          OffGridFlow
        </Link>
        <nav style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
          <Link href="/" style={{ color: '#94a3b8', textDecoration: 'none' }}>
            ← Back to Home
          </Link>
          <Link
            href="/register"
            style={{
              background: 'linear-gradient(135deg, #22c55e, #16a34a)',
              color: '#0b1222',
              padding: '0.6rem 1.5rem',
              borderRadius: '10px',
              textDecoration: 'none',
              fontWeight: 600,
            }}
          >
            Start Free Trial
          </Link>
        </nav>
      </header>

      <main style={{ paddingTop: '110px', maxWidth: '1200px', margin: '0 auto', padding: '110px 2rem 5rem' }}>
        <section style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, letterSpacing: '0.25em' }}>
            ENTERPRISE DEMO EXPERIENCE
          </div>
          <h1 style={{ fontSize: '2.8rem', fontWeight: 800, margin: '1rem 0 1rem' }}>
            Experience how OffGridFlow reduces operational risk by 40%.
          </h1>
          <p style={{ fontSize: '1.1rem', color: '#94a3b8', maxWidth: '760px', margin: '0 auto 2rem' }}>
            Streamline compliance, secure sensitive data, and unify global workflows with a platform engineered for
            enterprise-scale climate reporting and governance.
          </p>
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              justifyContent: 'center',
              gap: '0.75rem',
              marginBottom: '1.5rem',
            }}
          >
            {personas.map((persona) => (
              <button
                key={persona.id}
                onClick={() => setActivePersona(persona.id)}
                aria-pressed={activePersona === persona.id}
                style={{
                  padding: '0.65rem 1.1rem',
                  borderRadius: '999px',
                  border: activePersona === persona.id ? '1px solid #38bdf8' : '1px solid rgba(148, 163, 184, 0.3)',
                  background: activePersona === persona.id ? 'rgba(56, 189, 248, 0.15)' : 'transparent',
                  color: activePersona === persona.id ? '#e2e8f0' : '#94a3b8',
                  cursor: 'pointer',
                  fontWeight: 600,
                }}
              >
                Demo for {persona.label}
              </button>
            ))}
          </div>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
              gap: '1rem',
              marginBottom: '2rem',
            }}
          >
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.6)',
                border: '1px solid rgba(148, 163, 184, 0.2)',
                borderRadius: '12px',
                padding: '1rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>Primary outcome</div>
              <div style={{ fontSize: '1rem', fontWeight: 700 }}>{selectedPersona.headline}</div>
            </div>
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.6)',
                border: '1px solid rgba(148, 163, 184, 0.2)',
                borderRadius: '12px',
                padding: '1rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>{selectedPersona.kpi.label}</div>
              <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#22c55e' }}>{selectedPersona.kpi.value}</div>
            </div>
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.6)',
                border: '1px solid rgba(148, 163, 184, 0.2)',
                borderRadius: '12px',
                padding: '1rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>Trust signals</div>
              <div style={{ fontSize: '1rem', fontWeight: 700 }}>SOC 2 • ISO 27001 • GDPR-ready</div>
            </div>
          </div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem', justifyContent: 'center' }}>
            {trustLogos.map((logo) => (
              <div
                key={logo}
                style={{
                  padding: '0.6rem 1.2rem',
                  borderRadius: '999px',
                  background: 'rgba(148, 163, 184, 0.12)',
                  border: '1px solid rgba(148, 163, 184, 0.2)',
                  fontSize: '0.85rem',
                  fontWeight: 600,
                  letterSpacing: '0.04em',
                }}
              >
                {logo}
              </div>
            ))}
          </div>
        </section>

        <section style={{ marginBottom: '3.5rem' }} aria-labelledby="interactive-demo">
          <h2 id="interactive-demo" style={{ fontSize: '2rem', fontWeight: 800, marginBottom: '0.75rem' }}>
            Interactive Demo Experience
          </h2>
          <p style={{ color: '#94a3b8', marginBottom: '1.5rem' }}>
            Choose your own adventure. Each module includes real-world data at enterprise scale, with narrative
            highlights tied to business outcomes.
          </p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem', marginBottom: '1.5rem' }}>
            {demoModules.map((module) => (
              <button
                key={module.id}
                onClick={() => setActiveModule(module.id)}
                aria-pressed={activeModule === module.id}
                style={{
                  padding: '0.75rem 1.5rem',
                  borderRadius: '10px',
                  border: activeModule === module.id ? '1px solid #22c55e' : '1px solid rgba(148, 163, 184, 0.3)',
                  background: activeModule === module.id ? 'rgba(34, 197, 94, 0.18)' : 'transparent',
                  color: activeModule === module.id ? '#e2e8f0' : '#94a3b8',
                  cursor: 'pointer',
                  fontWeight: 600,
                }}
              >
                {module.label}
              </button>
            ))}
          </div>
          <div
            style={{
              background: 'rgba(15, 23, 42, 0.7)',
              borderRadius: '18px',
              border: '1px solid rgba(148, 163, 184, 0.15)',
              padding: '2rem',
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
              gap: '1.5rem',
            }}
          >
            <div>
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                BUSINESS OUTCOME
              </div>
              <div style={{ fontSize: '1.3rem', fontWeight: 700, marginBottom: '0.75rem' }}>
                {selectedModule.outcome}
              </div>
              <p style={{ color: '#94a3b8', lineHeight: 1.6 }}>{selectedModule.narrative}</p>
              <div style={{ marginTop: '1.25rem' }}>
                {selectedModule.impact.map((item) => (
                  <div key={item} style={{ display: 'flex', gap: '0.5rem', marginBottom: '0.5rem' }}>
                    <span style={{ color: '#22c55e' }}>●</span>
                    <span style={{ color: '#e2e8f0' }}>{item}</span>
                  </div>
                ))}
              </div>
            </div>
            <div>
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                PRIMARY SCREENS
              </div>
              <div style={{ display: 'grid', gap: '0.75rem', marginBottom: '1.5rem' }}>
                {selectedModule.screens.map((screen) => (
                  <div
                    key={screen}
                    style={{
                      padding: '0.75rem 1rem',
                      borderRadius: '10px',
                      background: 'rgba(148, 163, 184, 0.12)',
                      border: '1px solid rgba(148, 163, 184, 0.2)',
                      fontWeight: 600,
                    }}
                  >
                    {screen}
                  </div>
                ))}
              </div>
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                SIMULATED DATA SNAPSHOT
              </div>
              <div style={{ display: 'grid', gap: '0.6rem' }}>
                {selectedModule.dataPoints.map((data) => (
                  <div
                    key={data}
                    style={{
                      padding: '0.75rem 1rem',
                      borderRadius: '10px',
                      border: '1px dashed rgba(148, 163, 184, 0.3)',
                      color: '#cbd5f5',
                    }}
                  >
                    {data}
                  </div>
                ))}
              </div>
              <div style={{ fontSize: '0.75rem', color: '#94a3b8', marginTop: '0.75rem' }}>
                Simulated enterprise data. No customer data is used.
              </div>
            </div>
          </div>
        </section>

        <section style={{ marginBottom: '3rem' }} aria-labelledby="deep-dive">
          <h2 id="deep-dive" style={{ fontSize: '2rem', fontWeight: 800, marginBottom: '0.75rem' }}>
            Deep-Dive Capabilities
          </h2>
          <div style={{ display: 'grid', gap: '1rem' }}>
            {deepDiveModules.map((module) => (
              <details
                key={module.title}
                style={{
                  background: 'rgba(15, 23, 42, 0.65)',
                  borderRadius: '14px',
                  border: '1px solid rgba(148, 163, 184, 0.15)',
                  padding: '1rem 1.5rem',
                }}
              >
                <summary style={{ cursor: 'pointer', fontWeight: 700, fontSize: '1.05rem' }}>{module.title}</summary>
                <div style={{ marginTop: '0.75rem', color: '#94a3b8' }}>
                  {module.items.map((item) => (
                    <div key={item} style={{ marginBottom: '0.5rem', display: 'flex', gap: '0.5rem' }}>
                      <span style={{ color: '#22c55e' }}>•</span>
                      <span>{item}</span>
                    </div>
                  ))}
                </div>
              </details>
            ))}
          </div>
        </section>

        <section style={{ marginBottom: '3.5rem' }} aria-labelledby="stakeholder-view">
          <h2 id="stakeholder-view" style={{ fontSize: '2rem', fontWeight: 800, marginBottom: '0.75rem' }}>
            Stakeholder-Specific Value Messaging
          </h2>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
              gap: '1.5rem',
            }}
          >
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.7)',
                borderRadius: '16px',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                padding: '1.5rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                SELECTED PERSPECTIVE
              </div>
              <div style={{ fontSize: '1.2rem', fontWeight: 700 }}>{selectedPersona.label}</div>
              <div style={{ color: '#94a3b8', marginTop: '0.75rem' }}>{selectedPersona.headline}</div>
            </div>
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.7)',
                borderRadius: '16px',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                padding: '1.5rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                TOP BENEFITS
              </div>
              {selectedPersona.bullets.map((bullet) => (
                <div key={bullet} style={{ marginBottom: '0.5rem', display: 'flex', gap: '0.5rem' }}>
                  <span style={{ color: '#22c55e' }}>●</span>
                  <span style={{ color: '#e2e8f0' }}>{bullet}</span>
                </div>
              ))}
            </div>
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.7)',
                borderRadius: '16px',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                padding: '1.5rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                ROI SIGNAL
              </div>
              <div style={{ fontSize: '2.2rem', fontWeight: 800, color: '#22c55e' }}>
                {selectedPersona.kpi.value}
              </div>
              <div style={{ color: '#94a3b8' }}>{selectedPersona.kpi.label}</div>
            </div>
          </div>
        </section>

        <section style={{ marginBottom: '3.5rem' }} aria-labelledby="social-proof">
          <h2 id="social-proof" style={{ fontSize: '2rem', fontWeight: 800, marginBottom: '0.75rem' }}>
            Social Proof & Validation
          </h2>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
              gap: '1.5rem',
            }}
          >
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.75)',
                borderRadius: '16px',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                padding: '1.5rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                CASE STUDY SNAPSHOT
              </div>
              <div style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.5rem' }}>
                “OffGridFlow reduced our onboarding time from 2 weeks to 1 day.”
              </div>
              <div style={{ color: '#94a3b8' }}>VP Sustainability, Fortune 500 Manufacturing</div>
              <div style={{ marginTop: '1rem', color: '#22c55e', fontWeight: 700 }}>Outcome: 3x faster reporting</div>
            </div>
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.75)',
                borderRadius: '16px',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                padding: '1.5rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                ANALYST RECOGNITION
              </div>
              {analystBadges.map((badge) => (
                <div key={badge} style={{ marginBottom: '0.5rem', fontWeight: 600 }}>
                  {badge}
                </div>
              ))}
              <div style={{ color: '#94a3b8', fontSize: '0.85rem' }}>Top performer in governance automation</div>
            </div>
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.75)',
                borderRadius: '16px',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                padding: '1.5rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                COMPLIANCE BADGES
              </div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                {complianceBadges.map((badge) => (
                  <span
                    key={badge}
                    style={{
                      padding: '0.35rem 0.7rem',
                      borderRadius: '999px',
                      background: 'rgba(34, 197, 94, 0.15)',
                      border: '1px solid rgba(34, 197, 94, 0.3)',
                      fontSize: '0.8rem',
                      fontWeight: 600,
                    }}
                  >
                    {badge}
                  </span>
                ))}
              </div>
              <div style={{ marginTop: '1rem', fontSize: '0.85rem', color: '#94a3b8' }}>Independent verification ready</div>
            </div>
            <div
              style={{
                background: 'rgba(15, 23, 42, 0.75)',
                borderRadius: '16px',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                padding: '1.5rem',
              }}
            >
              <div style={{ fontSize: '0.8rem', color: '#38bdf8', fontWeight: 700, marginBottom: '0.5rem' }}>
                AS SEEN IN
              </div>
              {pressBadges.map((badge) => (
                <div key={badge} style={{ marginBottom: '0.4rem', fontWeight: 600 }}>
                  {badge}
                </div>
              ))}
              <div style={{ color: '#94a3b8', fontSize: '0.85rem' }}>Featured for enterprise climate innovation</div>
            </div>
          </div>
        </section>

        <section style={{ marginBottom: '3.5rem' }} aria-labelledby="next-steps">
          <h2 id="next-steps" style={{ fontSize: '2rem', fontWeight: 800, marginBottom: '0.75rem' }}>
            Clear, Friction-Free Next Steps
          </h2>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
              gap: '1rem',
              marginBottom: '1.5rem',
            }}
          >
            <Link
              href="mailto:sales@offgridflow.com?subject=OffGridFlow%20Live%20Demo"
              style={{
                padding: '1rem',
                borderRadius: '14px',
                background: 'linear-gradient(135deg, #22c55e, #16a34a)',
                color: '#0b1222',
                fontWeight: 700,
                textAlign: 'center',
              }}
            >
              Schedule a Personalized Live Demo
            </Link>
            <Link
              href="/register"
              style={{
                padding: '1rem',
                borderRadius: '14px',
                border: '1px solid rgba(148, 163, 184, 0.3)',
                background: 'rgba(15, 23, 42, 0.7)',
                color: '#e2e8f0',
                fontWeight: 700,
                textAlign: 'center',
              }}
            >
              Access a Guided Interactive Tour
            </Link>
            <Link
              href="mailto:sales@offgridflow.com?subject=OffGridFlow%20Architecture%20Overview"
              style={{
                padding: '1rem',
                borderRadius: '14px',
                border: '1px solid rgba(148, 163, 184, 0.3)',
                background: 'rgba(15, 23, 42, 0.7)',
                color: '#e2e8f0',
                fontWeight: 700,
                textAlign: 'center',
              }}
            >
              Request Architecture Overview
            </Link>
            <Link
              href="mailto:sales@offgridflow.com?subject=OffGridFlow%20Custom%20POC"
              style={{
                padding: '1rem',
                borderRadius: '14px',
                border: '1px solid rgba(148, 163, 184, 0.3)',
                background: 'rgba(15, 23, 42, 0.7)',
                color: '#e2e8f0',
                fontWeight: 700,
                textAlign: 'center',
              }}
            >
              Contact Sales for a Custom POC Plan
            </Link>
          </div>
          <div style={{ color: '#94a3b8', fontSize: '0.9rem' }}>
            Your data is confidential. Demos are tailored to your industry and risk profile.
          </div>
        </section>

        <section style={{ marginBottom: '3.5rem' }} aria-labelledby="resources">
          <h2 id="resources" style={{ fontSize: '2rem', fontWeight: 800, marginBottom: '0.75rem' }}>
            Supporting Resources
          </h2>
          <div
            style={{
              background: 'rgba(15, 23, 42, 0.7)',
              borderRadius: '16px',
              border: '1px solid rgba(148, 163, 184, 0.15)',
              padding: '1.5rem',
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
              gap: '0.75rem',
            }}
          >
            {resources.map((resource) => (
              <div
                key={resource}
                style={{
                  padding: '0.75rem 1rem',
                  borderRadius: '10px',
                  border: '1px dashed rgba(148, 163, 184, 0.3)',
                  color: '#e2e8f0',
                  fontWeight: 600,
                }}
              >
                {resource}
              </div>
            ))}
          </div>
        </section>

        <section
          style={{
            padding: '2.5rem',
            background: 'rgba(56, 189, 248, 0.08)',
            borderRadius: '18px',
            border: '1px solid rgba(56, 189, 248, 0.2)',
          }}
        >
          <h2 style={{ fontSize: '2rem', fontWeight: 800, marginBottom: '0.75rem' }}>
            Ready to see your data in action?
          </h2>
          <p style={{ color: '#94a3b8', marginBottom: '1.5rem', maxWidth: '720px' }}>
            We build a tailored demo around your data sources, regulatory requirements, and risk model so every
            stakeholder leaves with a clear next step.
          </p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem' }}>
            <Link
              href="mailto:sales@offgridflow.com?subject=OffGridFlow%20Demo%20Request"
              style={{
                background: 'linear-gradient(135deg, #38bdf8, #22c55e)',
                color: '#0b1222',
                padding: '0.9rem 2rem',
                borderRadius: '12px',
                fontWeight: 700,
                textDecoration: 'none',
              }}
            >
              Schedule Live Demo
            </Link>
            <Link
              href="/register"
              style={{
                border: '1px solid rgba(148, 163, 184, 0.3)',
                background: 'rgba(15, 23, 42, 0.7)',
                color: '#e2e8f0',
                padding: '0.9rem 2rem',
                borderRadius: '12px',
                fontWeight: 700,
                textDecoration: 'none',
              }}
            >
              Start Interactive Tour
            </Link>
          </div>
        </section>
      </main>
    </div>
  );
}
