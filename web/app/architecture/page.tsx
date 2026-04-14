import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Architecture | OffGridFlow',
  description: 'OffGridFlow data model, entity relationships, and audit traceability architecture.',
};

const entities = [
  {
    name: 'Tenant (Organization)',
    desc: 'Multi-tenant root. Every record is scoped to a tenant. One tenant cannot see or affect another.',
    fields: ['id (UUID)', 'name', 'slug', 'plan (audit_prep | compliance_pro | enterprise)', 'is_active', 'created_at'],
  },
  {
    name: 'User',
    desc: 'Authenticated account within a tenant. Roles control access.',
    fields: ['id (UUID)', 'tenant_id (FK → Tenant)', 'email', 'password_hash (bcrypt)', 'role (admin | user | viewer)', 'mfa_enabled', 'last_login_at'],
  },
  {
    name: 'Activity',
    desc: 'A single emissions data point — the atomic unit of carbon accounting.',
    fields: ['id (UUID)', 'tenant_id (FK)', 'source (csv | aws | azure | gcp | sap)', 'category (electricity | natural_gas | diesel | ...)', 'scope (1 | 2 | 3)', 'quantity', 'unit (kWh | L | km | ...)', 'activity_date', 'location', 'meter_id'],
  },
  {
    name: 'Emission Factor',
    desc: 'Conversion factor from activity quantity to kg CO2e. Sourced from EPA, IEA, DEFRA, IPCC.',
    fields: ['id (UUID)', 'scope', 'region', 'source', 'category', 'unit', 'value_kg_co2e_per_unit', 'data_source (EPA eGRID | IEA | DEFRA | IPCC)', 'year', 'valid_from', 'valid_to', 'uncertainty_%', 'version'],
  },
  {
    name: 'Calculation Ledger',
    desc: 'Immutable record of every emissions calculation. Cannot be modified after creation.',
    fields: ['id (UUID)', 'tenant_id (FK)', 'activity_id (FK → Activity)', 'scope', 'quantity', 'unit', 'factor_id', 'factor_value (frozen copy)', 'factor_source', 'factor_region', 'method', 'formula (human-readable)', 'result_kg_co2e', 'result_tonnes_co2e', 'calculated_by (FK → User)', 'calculated_at', 'is_locked', 'version', 'supersedes_id'],
  },
  {
    name: 'Factor Snapshot',
    desc: 'A locked set of emission factors for a reporting period. Ensures reproducibility.',
    fields: ['id (UUID)', 'tenant_id (FK)', 'reporting_period_start', 'reporting_period_end', 'snapshot_name', 'factors (JSONB — frozen copies)', 'factor_count', 'status (draft | locked | superseded)', 'locked_by (FK → User)', 'locked_at'],
  },
  {
    name: 'Compliance Report',
    desc: 'Generated compliance output for a specific framework and year.',
    fields: ['id (UUID)', 'tenant_id (FK)', 'report_type (csrd | sec | california | cbam | ifrs_s2)', 'report_year', 'status (draft | in_review | approved | submitted)', 'scope1_emissions', 'scope2_emissions', 'scope3_emissions', 'total_emissions_co2e', 'generated_by (FK → User)', 'approved_by (FK → User)', 'approved_at'],
  },
  {
    name: 'Approval Workflow',
    desc: 'Tracks review and approval state for reportable entities.',
    fields: ['id (UUID)', 'tenant_id (FK)', 'entity_type (report | inventory)', 'entity_id (FK)', 'status (draft | submitted | reviewed | approved | rejected)', 'prepared_by', 'reviewed_by', 'approved_by', 'rejection_reason', 'timestamps for each state'],
  },
  {
    name: 'Audit Log',
    desc: 'Immutable event trail. Every significant action is recorded with actor attribution.',
    fields: ['id (UUID)', 'tenant_id (FK)', 'user_id (FK)', 'event_type', 'action (create | update | delete | approve | reject)', 'resource_type', 'resource_id', 'changes (JSONB before/after)', 'ip_address', 'user_agent', 'created_at'],
  },
  {
    name: 'Change Log',
    desc: 'Field-level modification history for data governance and forensic review.',
    fields: ['id (UUID)', 'tenant_id (FK)', 'entity_type', 'entity_id', 'action', 'field_name', 'old_value', 'new_value', 'changed_by (FK → User)', 'changed_at'],
  },
];

const traceabilityChain = [
  { label: 'Activity', desc: 'Raw data point (e.g., 10,000 kWh electricity in US-WEST for January 2026)' },
  { label: 'Emission Factor', desc: 'EPA eGRID 2023 US-WEST: 0.298 kg CO2e/kWh' },
  { label: 'Calculation Ledger', desc: '10,000 × 0.298 = 2,980 kg CO2e = 2.98 tonnes CO2e. Formula, factor ID, source, region, user, timestamp recorded.' },
  { label: 'Factor Snapshot', desc: 'Factor 0.298 locked to reporting period Jan-Dec 2026. Immutable.' },
  { label: 'Compliance Report', desc: '2.98 tCO2e included in Scope 2 total for SEC Climate Disclosure 2026.' },
  { label: 'Approval Workflow', desc: 'Report prepared by operator, reviewed by manager, approved by CFO. Each step timestamped.' },
  { label: 'Export + Checksum', desc: 'PDF exported with SHA256 checksum. Reconciliation verifies export matches current data.' },
];

export default function ArchitecturePage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <main className="mx-auto max-w-5xl px-6 py-16">
        <h1 className="text-3xl font-bold text-white">Data Architecture</h1>
        <p className="mt-3 text-gray-400">
          OffGridFlow uses a single canonical data model with strict tenant isolation.
          Every record traces back to its source through an immutable audit chain.
        </p>

        {/* Traceability Chain */}
        <section className="mt-12">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">
            End-to-End Traceability Chain
          </h2>
          <p className="mt-2 text-sm text-gray-500">
            Every compliance output can be traced back through this chain to the original activity data.
          </p>
          <div className="mt-6 space-y-3">
            {traceabilityChain.map((step, i) => (
              <div key={step.label} className="flex items-start gap-4">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-600/10 text-xs font-bold text-primary-400">
                  {i + 1}
                </div>
                <div className="flex-1 rounded-lg border border-gray-800 bg-gray-800/30 px-4 py-3">
                  <span className="text-sm font-semibold text-white">{step.label}</span>
                  <span className="ml-2 text-sm text-gray-400">{step.desc}</span>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Entity Map */}
        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">
            Entity Definitions
          </h2>
          <p className="mt-2 text-sm text-gray-500">
            Core tables in the OffGridFlow data model. All tables are tenant-scoped — one tenant cannot access another&apos;s data.
          </p>
          <div className="mt-6 space-y-4">
            {entities.map((entity) => (
              <details key={entity.name} className="group rounded-xl border border-gray-800 bg-gray-800/20">
                <summary className="cursor-pointer px-5 py-4 text-sm font-semibold text-white">
                  {entity.name}
                  <span className="ml-3 text-xs font-normal text-gray-500">{entity.desc}</span>
                </summary>
                <div className="border-t border-gray-800 px-5 py-4">
                  <div className="flex flex-wrap gap-2">
                    {entity.fields.map((field) => (
                      <span key={field} className="rounded bg-gray-800 px-2.5 py-1 text-xs text-gray-300 font-mono">
                        {field}
                      </span>
                    ))}
                  </div>
                </div>
              </details>
            ))}
          </div>
        </section>

        {/* Relationships */}
        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">
            Key Relationships
          </h2>
          <div className="mt-6 overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50">
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">From</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">To</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Relationship</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Constraint</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {[
                  ['Activity', 'Tenant', 'belongs to', 'CASCADE delete'],
                  ['Activity', 'Calculation Ledger', 'calculated by', 'FK activity_id'],
                  ['Calculation Ledger', 'Factor Snapshot', 'uses factors from', 'FK factor_snapshot_id'],
                  ['Calculation Ledger', 'User', 'calculated by', 'FK calculated_by'],
                  ['Compliance Report', 'Tenant', 'belongs to', 'CASCADE delete'],
                  ['Compliance Report', 'User', 'generated / approved by', 'FK generated_by, approved_by'],
                  ['Approval Workflow', 'Compliance Report', 'governs', 'FK entity_id'],
                  ['Audit Log', 'Tenant + User', 'records actions by', 'FK tenant_id, user_id'],
                  ['Change Log', 'any entity', 'tracks field changes', 'entity_type + entity_id'],
                  ['Factor Snapshot', 'Tenant', 'belongs to', 'CASCADE delete, UNIQUE per period'],
                ].map(([from, to, rel, constraint]) => (
                  <tr key={`${from}-${to}`} className="border-b border-gray-800/30">
                    <td className="px-4 py-2 font-medium text-white">{from}</td>
                    <td className="px-4 py-2">{to}</td>
                    <td className="px-4 py-2 text-gray-400">{rel}</td>
                    <td className="px-4 py-2 font-mono text-xs text-gray-500">{constraint}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        {/* Tenant Isolation */}
        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">
            Tenant Isolation
          </h2>
          <div className="mt-4 rounded-xl border border-gray-800 bg-gray-800/20 p-6">
            <p className="text-sm text-gray-300">
              Every database query includes a tenant filter (<code className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-primary-400">WHERE tenant_id = $1</code>).
              This is enforced at the application layer through middleware that extracts the tenant ID from the authenticated JWT session.
              There is no admin endpoint that can query across tenants. Soft deletes use <code className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-primary-400">deleted_at</code> timestamps,
              and unique constraints include the <code className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-primary-400">WHERE deleted_at IS NULL</code> predicate to prevent namespace collisions.
            </p>
          </div>
        </section>

        {/* Links */}
        <div className="mt-12 flex flex-wrap gap-4 text-sm">
          <Link href="/evidence" className="text-primary-400 hover:underline">Evidence Pack</Link>
          <Link href="/methodology" className="text-primary-400 hover:underline">Methodology Library</Link>
          <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>
          <Link href="/operations" className="text-primary-400 hover:underline">Operations Proof</Link>
          <Link href="/security" className="text-primary-400 hover:underline">Security</Link>
          <Link href="/status" className="text-primary-400 hover:underline">Status</Link>
        </div>
      </main>
    </div>
  );
}
