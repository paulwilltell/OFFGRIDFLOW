/**
 * @fileoverview Versioned methodology reference for all emissions calculations
 * and compliance reports. Every calculated value and every exported report
 * carries this version stamp so prior outputs can be reproduced exactly.
 *
 * This is central to Gatekeeper Panel 1B: Reproducibility by Version and
 * Traceability to Source.
 */

export interface MethodologyVersion {
  /** Version identifier following semver (major.minor.patch). */
  version: string;
  /** ISO 8601 date this version went live. */
  effectiveDate: string;
  /** Human-readable summary of what this version covers. */
  summary: string;
  /** Standards this version aligns with. */
  standards: ReadonlyArray<string>;
  /** Factor source bundles included in this version. */
  factorSources: ReadonlyArray<{
    name: string;
    vintage: string;
    coverage: string;
  }>;
  /** Notable changes from the previous version, if any. */
  changesSinceLastVersion?: ReadonlyArray<string>;
}

/**
 * The currently active methodology version. Update this constant when
 * factor sources, calculation methods, or GHG Protocol guidance changes.
 *
 * Every calculation recorded in the calculation_ledger table embeds this
 * version. When auditors ask "why this number", the version + the
 * snapshot together reproduce the calculation deterministically.
 */
export const CURRENT_METHODOLOGY: MethodologyVersion = {
  version: '2026.1.0',
  effectiveDate: '2026-04-13',
  summary:
    'Scope 1, 2, and 3 calculations aligned with the GHG Protocol Corporate Standard. Scope 2 supports location-based and market-based methods. Scope 3 covers all 15 categories with activity-based, spend-based, and supplier-specific tiers.',
  standards: [
    'GHG Protocol Corporate Accounting and Reporting Standard',
    'GHG Protocol Scope 2 Guidance',
    'GHG Protocol Corporate Value Chain (Scope 3) Standard',
    'IPCC AR6 Global Warming Potential (100-year)',
  ],
  factorSources: [
    { name: 'EPA eGRID', vintage: '2023', coverage: '27 US subregions' },
    { name: 'IEA Emission Factors', vintage: '2023', coverage: '55+ countries' },
    { name: 'UK DEFRA Conversion Factors', vintage: '2024', coverage: 'Global Scope 1/2/3 activities' },
    { name: 'IPCC AR6 GWP-100', vintage: '2021', coverage: 'Refrigerants and fugitive gases' },
    { name: 'GHG Protocol Scope 3 Guidance', vintage: '2013+supplements', coverage: '15 categories' },
  ],
  changesSinceLastVersion: [
    'Initial published version (2026.1.0).',
    'Factor snapshots introduced for period-locked reproducibility.',
    'Export reconciliation with SHA-256 checksums enabled.',
  ],
};

/**
 * Returns a short label suitable for embedding in dashboards and exports.
 * Example: "Methodology v2026.1.0"
 */
export function methodologyLabel(): string {
  return `Methodology v${CURRENT_METHODOLOGY.version}`;
}

/**
 * Returns a machine-readable stamp to attach to API payloads and exports.
 */
export function methodologyStamp() {
  return {
    methodology_version: CURRENT_METHODOLOGY.version,
    methodology_effective_date: CURRENT_METHODOLOGY.effectiveDate,
    methodology_standards: [...CURRENT_METHODOLOGY.standards],
  } as const;
}
