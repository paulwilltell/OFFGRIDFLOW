import Link from 'next/link';

export interface RelatedPageLink {
  href: string;
  label: string;
  description?: string;
}

export function RelatedPages({ items }: { items: RelatedPageLink[] }) {
  if (!items.length) return null;

  return (
    <section className="mt-12 rounded-2xl border border-gray-800 bg-gray-800/20 p-6">
      <h2 className="text-lg font-semibold text-white">Related Pages</h2>
      <p className="mt-2 text-sm text-gray-400">
        Continue the research path with adjacent framework, role, and implementation pages.
      </p>
      <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {items.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="rounded-xl border border-gray-800/70 bg-gray-900/40 p-4 transition hover:border-primary-600/40"
          >
            <div className="text-sm font-medium text-white">{item.label}</div>
            {item.description && <div className="mt-1 text-xs text-gray-500">{item.description}</div>}
          </Link>
        ))}
      </div>
    </section>
  );
}

export default RelatedPages;
