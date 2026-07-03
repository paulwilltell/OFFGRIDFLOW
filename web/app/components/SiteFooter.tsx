import Link from 'next/link';

/**
 * Minimal site-wide footer. Only links to pages that exist:
 * Pricing, Trust, Security, and the legal + auth routes.
 */
export function SiteFooter() {
  return (
    <footer className="border-t border-gray-800/50 bg-dark-900 px-6 py-10 text-gray-100">
      <div className="mx-auto max-w-5xl">
        <div className="flex flex-wrap items-center justify-between gap-6">
          <div>
            <div className="text-sm font-semibold text-white">OffGridFlow</div>
            <div className="mt-1 text-xs text-gray-600">Carbon accounting without the busywork.</div>
          </div>
          <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-gray-400">
            <Link href="/pricing" className="hover:text-white">Pricing</Link>
            <Link href="/trust" className="hover:text-white">Trust</Link>
            <Link href="/security" className="hover:text-white">Security</Link>
            <Link href="/login" className="hover:text-white">Sign in</Link>
            <a href="mailto:contact@off-grid-flow.com" className="hover:text-white">Contact</a>
          </div>
        </div>
        <div className="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2 border-t border-gray-800/50 pt-6 text-xs text-gray-600">
          <Link href="/privacy" className="hover:text-white">Privacy Policy</Link>
          <Link href="/terms" className="hover:text-white">Terms of Service</Link>
          <span>&copy; {new Date().getFullYear()} OffGridFlow LLC</span>
        </div>
      </div>
    </footer>
  );
}

export default SiteFooter;
