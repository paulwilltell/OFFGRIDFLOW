'use client';

import { useState } from 'react';
import Link from 'next/link';

export default function DemoPage() {
  const [activeTab, setActiveTab] = useState('dashboard');

  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%)',
      color: 'white'
    }}>
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
        <Link href="/" style={{ fontSize: '1.5rem', fontWeight: 'bold', background: 'linear-gradient(135deg, #22c55e, #10b981)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', textDecoration: 'none' }}>
          OffGridFlow
        </Link>
        <nav style={{ display: 'flex', gap: '2rem', alignItems: 'center' }}>
          <Link href="/" style={{ color: '#94a3b8', textDecoration: 'none' }}>← Back to Home</Link>
          <Link href="/register" style={{
            background: '#22c55e',
            color: 'white',
            padding: '0.5rem 1.5rem',
            borderRadius: '8px',
            textDecoration: 'none',
            fontWeight: 500
          }}>Start Free Trial</Link>
        </nav>
      </header>

      {/* Main Content */}
      <main style={{ paddingTop: '100px', maxWidth: '1400px', margin: '0 auto', padding: '100px 2rem 4rem' }}>
        <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
          <div style={{ fontSize: '0.9rem', color: '#22c55e', fontWeight: 'bold', marginBottom: '0.5rem' }}>LIVE DEMO</div>
          <h1 style={{ fontSize: '2.5rem', fontWeight: 'bold', marginBottom: '1rem' }}>
            See OffGridFlow in Action
          </h1>
          <p style={{ fontSize: '1.1rem', color: '#94a3b8', maxWidth: '700px', margin: '0 auto' }}>
            Explore real emissions tracking, compliance reports, and enterprise dashboards
          </p>
        </div>

        {/* Tab Navigation */}
        <div style={{
          display: 'flex',
          gap: '1rem',
          marginBottom: '2rem',
          borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
          paddingBottom: '1rem'
        }}>
          {[
            { id: 'dashboard', label: 'Dashboard' },
            { id: 'reports', label: 'Sample Reports' },
            { id: 'compliance', label: 'Compliance View' }
          ].map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              style={{
                padding: '0.75rem 1.5rem',
                background: activeTab === tab.id ? 'rgba(34, 197, 94, 0.2)' : 'transparent',
                border: activeTab === tab.id ? '2px solid #22c55e' : '2px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '8px',
                color: activeTab === tab.id ? '#22c55e' : '#94a3b8',
                cursor: 'pointer',
                fontWeight: activeTab === tab.id ? 'bold' : 'normal',
                transition: 'all 0.3s'
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Dashboard Tab */}
        {activeTab === 'dashboard' && (
          <div>
            <h2 style={{ fontSize: '1.8rem', fontWeight: 'bold', marginBottom: '2rem' }}>
              Enterprise Dashboard Preview
            </h2>
            
            {/* KPI Cards */}
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
              gap: '1.5rem',
              marginBottom: '3rem'
            }}>
              {[
                { title: 'Total Emissions', value: '2,847 tCO₂e', change: '-12% vs Q3', trend: 'down' },
                { title: 'Scope 1', value: '847 tCO₂e', change: 'Direct emissions', trend: 'neutral' },
                { title: 'Scope 2', value: '1,205 tCO₂e', change: 'Electricity & heat', trend: 'neutral' },
                { title: 'Scope 3', value: '795 tCO₂e', change: '+8% vs Q3', trend: 'up' }
              ].map((kpi, i) => (
                <div key={i} style={{
                  background: 'rgba(255, 255, 255, 0.03)',
                  backdropFilter: 'blur(20px)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  padding: '1.5rem'
                }}>
                  <div style={{ fontSize: '0.9rem', color: '#94a3b8', marginBottom: '0.5rem' }}>{kpi.title}</div>
                  <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#22c55e', marginBottom: '0.25rem' }}>
                    {kpi.value}
                  </div>
                  <div style={{ 
                    fontSize: '0.85rem', 
                    color: kpi.trend === 'down' ? '#22c55e' : kpi.trend === 'up' ? '#f59e0b' : '#94a3b8' 
                  }}>
                    {kpi.change}
                  </div>
                </div>
              ))}
            </div>

            {/* Emissions Chart Placeholder */}
            <div style={{
              background: 'rgba(255, 255, 255, 0.03)',
              backdropFilter: 'blur(20px)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: '16px',
              padding: '2rem',
              marginBottom: '2rem'
            }}>
              <h3 style={{ fontSize: '1.3rem', fontWeight: 'bold', marginBottom: '1.5rem' }}>
                Quarterly Emissions Trend
              </h3>
              <div style={{
                height: '300px',
                display: 'flex',
                alignItems: 'flex-end',
                gap: '2rem',
                justifyContent: 'space-around'
              }}>
                {[
                  { label: 'Q1 2025', scope1: 920, scope2: 1340, scope3: 680 },
                  { label: 'Q2 2025', scope1: 885, scope2: 1290, scope3: 720 },
                  { label: 'Q3 2025', scope1: 865, scope2: 1250, scope3: 740 },
                  { label: 'Q4 2025', scope1: 847, scope2: 1205, scope3: 795 }
                ].map((q, i) => (
                  <div key={i} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '4px', width: '100%' }}>
                      <div style={{ width: '100%', height: `${(q.scope3 / 10)}px`, background: '#10b981', borderRadius: '4px 4px 0 0' }}></div>
                      <div style={{ width: '100%', height: `${(q.scope2 / 10)}px`, background: '#22c55e' }}></div>
                      <div style={{ width: '100%', height: `${(q.scope1 / 10)}px`, background: '#4ade80', borderRadius: '0 0 4px 4px' }}></div>
                    </div>
                    <div style={{ fontSize: '0.85rem', color: '#94a3b8', marginTop: '1rem' }}>{q.label}</div>
                  </div>
                ))}
              </div>
              <div style={{ display: 'flex', gap: '2rem', marginTop: '2rem', justifyContent: 'center' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <div style={{ width: '12px', height: '12px', background: '#4ade80', borderRadius: '2px' }}></div>
                  <span style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Scope 1</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <div style={{ width: '12px', height: '12px', background: '#22c55e', borderRadius: '2px' }}></div>
                  <span style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Scope 2</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <div style={{ width: '12px', height: '12px', background: '#10b981', borderRadius: '2px' }}></div>
                  <span style={{ fontSize: '0.85rem', color: '#94a3b8' }}>Scope 3</span>
                </div>
              </div>
            </div>

            {/* Data Sources */}
            <div style={{
              background: 'rgba(255, 255, 255, 0.03)',
              backdropFilter: 'blur(20px)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: '16px',
              padding: '2rem'
            }}>
              <h3 style={{ fontSize: '1.3rem', fontWeight: 'bold', marginBottom: '1.5rem' }}>
                Connected Data Sources
              </h3>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
                {[
                  { name: 'AWS CloudWatch', status: 'Connected', records: '24,891' },
                  { name: 'Azure Monitor', status: 'Connected', records: '18,456' },
                  { name: 'GCP Operations', status: 'Connected', records: '31,247' },
                  { name: 'SAP S/4HANA', status: 'Connected', records: '12,089' }
                ].map((source, i) => (
                  <div key={i} style={{
                    background: 'rgba(255, 255, 255, 0.02)',
                    border: '1px solid rgba(255, 255, 255, 0.05)',
                    borderRadius: '8px',
                    padding: '1rem'
                  }}>
                    <div style={{ fontSize: '1rem', fontWeight: 'bold', marginBottom: '0.5rem' }}>{source.name}</div>
                    <div style={{ fontSize: '0.85rem', color: '#22c55e', marginBottom: '0.25rem' }}>● {source.status}</div>
                    <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>{source.records} records</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Reports Tab */}
        {activeTab === 'reports' && (
          <div>
            <h2 style={{ fontSize: '1.8rem', fontWeight: 'bold', marginBottom: '2rem' }}>
              Sample Compliance Reports
            </h2>
            
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))', gap: '2rem' }}>
              {[
                {
                  title: 'CSRD E1-6 Climate Report',
                  description: 'Complete EU Corporate Sustainability Reporting Directive compliance for climate disclosures',
                  pages: '47 pages',
                  format: 'PDF',
                  preview: ['Executive Summary', 'GHG Inventory (Scope 1-3)', 'Climate Risk Assessment', 'Transition Planning', 'Science-Based Targets']
                },
                {
                  title: 'California SB 253 Report',
                  description: 'California Climate Corporate Data Accountability Act annual disclosure',
                  pages: '28 pages',
                  format: 'PDF',
                  preview: ['Organizational Boundaries', 'Direct Emissions (Scope 1)', 'Indirect Emissions (Scope 2)', 'Value Chain (Scope 3)', 'Assurance Statement']
                },
                {
                  title: 'SEC Climate Disclosure',
                  description: 'U.S. Securities and Exchange Commission climate-related disclosure report',
                  pages: '35 pages',
                  format: 'PDF',
                  preview: ['Climate Governance', 'Material Risks', 'GHG Emissions Data', 'Scenario Analysis', 'Financial Impact']
                },
                {
                  title: 'EU CBAM Report',
                  description: 'Carbon Border Adjustment Mechanism quarterly declaration for imported goods',
                  pages: '19 pages',
                  format: 'XML/PDF',
                  preview: ['Imported Goods Registry', 'Embedded Emissions', 'Carbon Price Paid', 'CBAM Certificate Requirements']
                }
              ].map((report, i) => (
                <div key={i} style={{
                  background: 'rgba(255, 255, 255, 0.03)',
                  backdropFilter: 'blur(20px)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '16px',
                  padding: '2rem'
                }}>
                  <div style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.75rem',
                    marginBottom: '1rem'
                  }}>
                    <div style={{
                      width: '48px',
                      height: '48px',
                      background: 'linear-gradient(135deg, #22c55e, #10b981)',
                      borderRadius: '8px',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '1.5rem'
                    }}>📄</div>
                    <div>
                      <h3 style={{ fontSize: '1.2rem', fontWeight: 'bold' }}>{report.title}</h3>
                      <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>{report.pages} • {report.format}</div>
                    </div>
                  </div>
                  
                  <p style={{ fontSize: '0.95rem', color: '#94a3b8', marginBottom: '1.5rem', lineHeight: 1.6 }}>
                    {report.description}
                  </p>
                  
                  <div style={{ marginBottom: '1.5rem' }}>
                    <div style={{ fontSize: '0.85rem', fontWeight: 'bold', marginBottom: '0.75rem', color: '#22c55e' }}>
                      REPORT SECTIONS:
                    </div>
                    {report.preview.map((section, j) => (
                      <div key={j} style={{
                        fontSize: '0.85rem',
                        color: '#cbd5e1',
                        padding: '0.5rem 0',
                        borderBottom: j < report.preview.length - 1 ? '1px solid rgba(255, 255, 255, 0.05)' : 'none'
                      }}>
                        {j + 1}. {section}
                      </div>
                    ))}
                  </div>
                  
                  <button style={{
                    width: '100%',
                    padding: '0.75rem',
                    background: 'rgba(34, 197, 94, 0.1)',
                    border: '1px solid #22c55e',
                    borderRadius: '8px',
                    color: '#22c55e',
                    fontWeight: 'bold',
                    cursor: 'pointer'
                  }}>
                    View Sample Report
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Compliance Tab */}
        {activeTab === 'compliance' && (
          <div>
            <h2 style={{ fontSize: '1.8rem', fontWeight: 'bold', marginBottom: '2rem' }}>
              Compliance Status Dashboard
            </h2>
            
            {/* Compliance Overview */}
            <div style={{
              background: 'rgba(255, 255, 255, 0.03)',
              backdropFilter: 'blur(20px)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: '16px',
              padding: '2rem',
              marginBottom: '2rem'
            }}>
              <h3 style={{ fontSize: '1.3rem', fontWeight: 'bold', marginBottom: '1.5rem' }}>
                Regulatory Frameworks
              </h3>
              <div style={{ display: 'grid', gap: '1.5rem' }}>
                {[
                  { name: 'EU CSRD (E1-6)', status: 'Compliant', completion: 100, nextDeadline: 'Annual report due Dec 2026' },
                  { name: 'California SB 253', status: 'Compliant', completion: 100, nextDeadline: 'Next filing due June 2026' },
                  { name: 'SEC Climate Disclosure', status: 'In Progress', completion: 78, nextDeadline: '22% remaining - due March 2026' },
                  { name: 'EU CBAM', status: 'Compliant', completion: 100, nextDeadline: 'Q1 2026 declaration submitted' }
                ].map((framework, i) => (
                  <div key={i} style={{
                    background: 'rgba(255, 255, 255, 0.02)',
                    border: '1px solid rgba(255, 255, 255, 0.05)',
                    borderRadius: '12px',
                    padding: '1.5rem'
                  }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                      <div>
                        <div style={{ fontSize: '1.1rem', fontWeight: 'bold' }}>{framework.name}</div>
                        <div style={{ fontSize: '0.85rem', color: '#94a3b8', marginTop: '0.25rem' }}>{framework.nextDeadline}</div>
                      </div>
                      <div style={{
                        padding: '0.5rem 1rem',
                        background: framework.status === 'Compliant' ? 'rgba(34, 197, 94, 0.1)' : 'rgba(251, 191, 36, 0.1)',
                        border: `1px solid ${framework.status === 'Compliant' ? '#22c55e' : '#fbbf24'}`,
                        borderRadius: '6px',
                        fontSize: '0.85rem',
                        fontWeight: 'bold',
                        color: framework.status === 'Compliant' ? '#22c55e' : '#fbbf24'
                      }}>
                        {framework.status}
                      </div>
                    </div>
                    
                    <div style={{ marginBottom: '0.5rem' }}>
                      <div style={{ 
                        height: '8px', 
                        background: 'rgba(255, 255, 255, 0.05)', 
                        borderRadius: '4px',
                        overflow: 'hidden'
                      }}>
                        <div style={{
                          width: `${framework.completion}%`,
                          height: '100%',
                          background: framework.completion === 100 ? '#22c55e' : '#fbbf24',
                          transition: 'width 0.3s'
                        }}></div>
                      </div>
                    </div>
                    <div style={{ fontSize: '0.8rem', color: '#94a3b8', textAlign: 'right' }}>
                      {framework.completion}% Complete
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Data Quality */}
            <div style={{
              background: 'rgba(255, 255, 255, 0.03)',
              backdropFilter: 'blur(20px)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: '16px',
              padding: '2rem'
            }}>
              <h3 style={{ fontSize: '1.3rem', fontWeight: 'bold', marginBottom: '1.5rem' }}>
                Data Quality Metrics
              </h3>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1.5rem' }}>
                {[
                  { metric: 'Overall Quality Score', value: '99.4%', status: 'Excellent' },
                  { metric: 'Data Completeness', value: '98.7%', status: 'Excellent' },
                  { metric: 'Verification Rate', value: '96.2%', status: 'Good' },
                  { metric: 'Audit Trail Coverage', value: '100%', status: 'Excellent' }
                ].map((item, i) => (
                  <div key={i} style={{
                    background: 'rgba(255, 255, 255, 0.02)',
                    border: '1px solid rgba(255, 255, 255, 0.05)',
                    borderRadius: '8px',
                    padding: '1.25rem',
                    textAlign: 'center'
                  }}>
                    <div style={{ fontSize: '0.85rem', color: '#94a3b8', marginBottom: '0.5rem' }}>{item.metric}</div>
                    <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#22c55e', marginBottom: '0.25rem' }}>{item.value}</div>
                    <div style={{ fontSize: '0.8rem', color: '#4ade80' }}>{item.status}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* CTA Section */}
        <div style={{
          marginTop: '4rem',
          padding: '3rem',
          background: 'rgba(34, 197, 94, 0.05)',
          border: '1px solid rgba(34, 197, 94, 0.2)',
          borderRadius: '16px',
          textAlign: 'center'
        }}>
          <h2 style={{ fontSize: '2rem', fontWeight: 'bold', marginBottom: '1rem' }}>
            Ready to Get Started?
          </h2>
          <p style={{ fontSize: '1.1rem', color: '#94a3b8', marginBottom: '2rem', maxWidth: '600px', margin: '0 auto 2rem' }}>
            Start tracking your emissions and generating compliance reports in minutes, not months.
          </p>
          <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
            <Link href="/register" style={{
              background: 'linear-gradient(135deg, #22c55e, #10b981)',
              color: 'white',
              padding: '1rem 2.5rem',
              borderRadius: '12px',
              textDecoration: 'none',
              fontWeight: 'bold',
              fontSize: '1.1rem',
              boxShadow: '0 10px 30px rgba(34, 197, 94, 0.3)'
            }}>
              Start Free Trial
            </Link>
            <Link href="mailto:sales@offgridflow.com" style={{
              background: 'rgba(255, 255, 255, 0.05)',
              border: '2px solid rgba(34, 197, 94, 0.5)',
              color: 'white',
              padding: '1rem 2.5rem',
              borderRadius: '12px',
              textDecoration: 'none',
              fontWeight: 'bold',
              fontSize: '1.1rem'
            }}>
              Contact Sales
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
