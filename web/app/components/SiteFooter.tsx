import Link from 'next/link';

/**
 * Site-wide footer mounted on every public marketing page.
 *
 * Six-column topical map reinforces the internal link graph:
 *  - Product, Frameworks, By Role, Trust, Resources, Brand
 *  - Legal row (Privacy, Terms, Do Not Sell) at the bottom
 *
 * Maintained as a single source of truth. Update here and every marketing
 * page automatically gains the new link.
 */
export function SiteFooter() {
  return (
    <footer className="border-t border-gray-800/50 bg-dark-900 px-6 py-12 text-gray-100">
      <div className="mx-auto max-w-5xl">
        <div className="grid gap-8 sm:grid-cols-2 lg:grid-cols-6">
          <div className="lg:col-span-1">
            <div className="text-sm font-semibold text-white">OffGridFlow</div>
            <div className="mt-1 text-xs text-gray-600">Carbon Accounting</div>
          </div>
          <div>
            <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Product</div>
            <div className="space-y-2 text-sm text-gray-400">
              <Link href="/demo" className="block hover:text-white">How It Works</Link>
              <Link href="/pricing" className="block hover:text-white">Pricing</Link>
              <Link href="/carbon-accounting-software" className="block hover:text-white">Carbon Accounting</Link>
              <Link href="/scope-1-2-3-reporting-software" className="block hover:text-white">Scope 1, 2, 3</Link>
              <Link href="/audit-ready-carbon-accounting" className="block hover:text-white">Audit-Ready</Link>
              <Link href="/case-study" className="block hover:text-white">Case Study</Link>
            </div>
          </div>
          <div>
            <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Compliance Frameworks</div>
            <div className="space-y-2 text-sm text-gray-400">
              <Link href="/sb-253-reporting-software" className="block hover:text-white">California SB 253</Link>
              <Link href="/csrd-reporting-software" className="block hover:text-white">CSRD / ESRS E1</Link>
              <Link href="/ifrs-s2-reporting-software" className="block hover:text-white">IFRS S2</Link>
              <Link href="/cbam-reporting-software" className="block hover:text-white">EU CBAM</Link>
              <Link href="/csrd-vs-ifrs-s2-carbon-reporting" className="block hover:text-white">CSRD vs IFRS S2</Link>
            </div>
          </div>
          <div>
            <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">By Role</div>
            <div className="space-y-2 text-sm text-gray-400">
              <Link href="/for-cfos" className="block hover:text-white">For CFOs</Link>
              <Link href="/for-sustainability-managers" className="block hover:text-white">Sustainability Managers</Link>
              <Link href="/for-procurement" className="block hover:text-white">Procurement</Link>
              <Link href="/carbon-accounting-software-for-finance-teams" className="block hover:text-white">Finance Teams</Link>
              <Link href="/scope-3-supplier-emissions-software" className="block hover:text-white">Scope 3 Supplier</Link>
            </div>
          </div>
          <div>
            <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Trust</div>
            <div className="space-y-2 text-sm text-gray-400">
              <Link href="/trust" className="block hover:text-white">Trust Center</Link>
              <Link href="/methodology" className="block hover:text-white">Methodology</Link>
              <Link href="/architecture" className="block hover:text-white">Architecture</Link>
              <Link href="/evidence" className="block hover:text-white">Evidence Pack</Link>
              <Link href="/security" className="block hover:text-white">Security</Link>
              <Link href="/operations" className="block hover:text-white">Operations</Link>
              <Link href="/status" className="block hover:text-white">Status</Link>
              <Link href="/how-we-operate" className="block hover:text-white">How We Operate</Link>
            </div>
          </div>
          <div>
            <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Resources</div>
            <div className="space-y-2 text-sm text-gray-400">
              <Link href="/sb-253-readiness-checklist" className="block hover:text-white">SB 253 Checklist</Link>
              <Link href="/scope-2-factor-library" className="block hover:text-white">Scope 2 Factors</Link>
              <Link href="/carbon-reporting-template" className="block hover:text-white">CSV Template</Link>
              <Link href="/aws-carbon-data" className="block hover:text-white">AWS Integration</Link>
              <Link href="/sap-carbon-reporting" className="block hover:text-white">SAP Integration</Link>
              <Link href="/csv-emissions-import" className="block hover:text-white">CSV Import</Link>
              <Link href="/about" className="block hover:text-white">About</Link>
              <a href="mailto:contact@off-grid-flow.com" className="block hover:text-white">Contact</a>
            </div>
          </div>
        </div>
        <div className="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2 border-t border-gray-800/50 pt-6 text-xs text-gray-600">
          <Link href="/privacy" className="hover:text-white">Privacy Policy</Link>
          <Link href="/terms" className="hover:text-white">Terms of Service</Link>
          <Link href="/privacy#do-not-sell" className="hover:text-white">Do Not Sell or Share</Link>
          <span>&copy; {new Date().getFullYear()} OffGridFlow LLC</span>
        </div>
      </div>
    </footer>
  );
}

export default SiteFooter;
