import type { AbatementEvaluation } from './types';

function hasEvidence(evidenceUrls: string[]): boolean {
  return evidenceUrls.length > 0;
}

function containsAny(input: string, phrases: string[]): boolean {
  const normalized = input.toLowerCase();
  return phrases.some((phrase) => normalized.includes(phrase.toLowerCase()));
}

function hasMarketInstrumentReference(input: string): boolean {
  return [
    /\brenewable energy certificate(s)?\b/i,
    /\brec(s)?\b/i,
    /\bgreen power\b/i,
    /\bppa\b/i,
    /\bpower purchase agreement(s)?\b/i,
    /\bguarantee of origin\b/i,
    /\beac(s)?\b/i,
    /\bresidual mix\b/i,
  ].some((pattern) => pattern.test(input));
}

export function evaluateScope2MarketBased(
  justification: string,
  evidenceUrls: string[],
): AbatementEvaluation {
  const hasInstrument = hasMarketInstrumentReference(justification);
  const hasCoverage = containsAny(justification, [
    'office',
    'facility',
    'site',
    'meter',
    'account',
    'portfolio',
    'all locations',
  ]);
  const uploadedEvidence = hasEvidence(evidenceUrls);

  const criteriaChecked: string[] = [];
  if (hasInstrument) criteriaChecked.push('market-based instrument reference');
  if (hasCoverage) criteriaChecked.push('portfolio coverage detail');
  if (uploadedEvidence) criteriaChecked.push('supporting evidence');

  if (hasInstrument && hasCoverage && uploadedEvidence) {
    return {
      status: 'recommended',
      feedback:
        'Market-based instruments are referenced, the covered sites are identified, and supporting evidence was uploaded.',
      criteriaChecked,
    };
  }

  if (hasInstrument && uploadedEvidence) {
    return {
      status: 'needs_clarification',
      feedback:
        'Market-based instruments are referenced and evidence is present, but the covered sites or meters are not clearly described.',
      criteriaChecked,
    };
  }

  if (hasInstrument) {
    return {
      status: 'needs_clarification',
      feedback:
        'Market-based instruments are mentioned, but supporting evidence is missing. Add the relevant certificate or contract.',
      criteriaChecked,
    };
  }

  return {
    status: 'insufficient',
    feedback:
      'No market-based instrument is described. Add the REC, contract, tariff, or residual-mix evidence that supports the adjustment.',
    criteriaChecked,
  };
}
