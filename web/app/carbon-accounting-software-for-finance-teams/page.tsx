import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/carbon-accounting-software-for-finance-teams';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Carbon Accounting Software for Finance Teams',
  description:
    'CFO-grade carbon accounting. Ledger-style data model, immutable audit trail, dual reporting (SEC, CSRD, SB 253, IFRS S2), reconciliation checksums, and published methodology.',
  keyword: 'carbon accounting for finance',
});

const faqs = [
  {
    question: 'Why should the CFO own carbon reporting, not just sustainability?',
    answer:
      'Climate disclosures are now financial disclosures. SEC Climate, CSRD, and SB 253 require Scope 1/2 (and material Scope 3) to appear alongside audited financials. A sustainability tool that lacks ledger discipline creates restatement risk. Finance teams that own the data pipeline avoid that risk.',
  },
  {
    question: 'How does OffGridFlow compare to a general ledger?',
    answer:
      'Activities are analogous to transactions, factors to chart-of-accounts mappings, the calculation ledger to journal entries, approval workflow to close-period controls, and factor snapshots to accounting-policy version locks. If you can defend a financial close, you can defend an OffGridFlow close.',
  },
  {
    question: 'Can we use the same workflow for multiple frameworks?',
    answer:
      'Yes. Reconcile once, disclose many times. The same activity ledger produces SEC, CSRD, SB 253, CBAM, and IFRS S2 draft outputs with each framework&apos;s required packaging. You maintain one source of truth.',
  },
  {
    question: 'Does OffGridFlow integrate with our ERP?',
    answer:
      'SAP S/4HANA connector, plus CSV and API fallbacks. Activity data ingestion can be scheduled so the carbon close becomes a monthly rhythm alongside the financial close.',
  },
  {
    question: 'What about internal controls?',
    answer:
      'RBAC with admin/user/viewer roles, approval workflow with separation-of-duty enforcement, immutable audit logs, and change-log field-level tracking. All documented in the public Trust Center.',
  },
];

export default function FinanceTeamsPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'For Finance Teams', path: PATH },
      ]}
      h1="Carbon Accounting Software for Finance Teams"
      dek="Ledger-style data model. Reconciliation checksums. Versioned methodology. Separation of duties. OffGridFlow is what a CFO actually wants when sustainability lands on the 10-K."
      slug="finance-teams"
      ctaUtm="for_finance"
      faqs={faqs}
      relatedPages={[
        { href: '/for-cfos', label: 'For CFOs', description: 'See the leadership-level control model and close process.' },
        { href: '/audit-ready-carbon-accounting', label: 'Audit-ready carbon accounting', description: 'Focus on defensibility and assurance evidence.' },
        { href: '/sb-253-reporting-software', label: 'SB 253 reporting', description: 'Map finance ownership to a concrete disclosure regime.' },
      ]}
    >
      <p>
        When climate disclosures entered the 10-K (SEC Climate), the annual report (CSRD), and
        the CARB filing (SB 253), emissions data became financial data. CFOs and controllers
        inherit a new line item that their existing sustainability tools were never designed to
        defend.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What finance teams need that most tools miss</h2>
      <ul>
        <li>Immutable ledger so prior-period disclosures can be reproduced on demand</li>
        <li>Factor snapshot locks so restatements are explicit and traceable</li>
        <li>Approval workflow with separation of duties (preparer, reviewer, approver)</li>
        <li>Change log with field-level before/after and actor attribution</li>
        <li>Reconciliation checksums that prove the PDF matches the underlying data</li>
        <li>Multi-framework output from one source of truth</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">Carbon close alongside financial close</h2>
      <p>
        Run the carbon close on the same monthly rhythm as the financial close. Import
        activities, run calculations, review anomalies, submit for review, approve, lock factor
        snapshot. Next year&apos;s assurance engagement reconstructs last year&apos;s numbers
        without needing last year&apos;s spreadsheets.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What it replaces</h2>
      <p>
        Six-figure Big 4 engagements, bespoke consulting spreadsheets, ad hoc email chains with
        suppliers for Scope 3 data. OffGridFlow starts at $6,500/year and produces the same
        audit-ready draft outputs — with a published methodology auditors can reference.
      </p>

    </MoneyPageLayout>
  );
}
