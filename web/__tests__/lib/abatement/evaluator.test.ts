import { evaluateScope2MarketBased } from '@/lib/abatement/evaluator';

describe('evaluateScope2MarketBased', () => {
  it('returns recommended when the justification references RECs and evidence is present', () => {
    const result = evaluateScope2MarketBased(
      'Uploaded renewable energy certificates covering all facility meters for the reporting year.',
      ['/api/abatement/sb253/evidence/123'],
    );

    expect(result.status).toBe('recommended');
    expect(result.criteriaChecked).toContain('market-based instrument reference');
    expect(result.criteriaChecked).toContain('supporting evidence');
  });

  it('returns needs clarification when market-based instruments are described but evidence is missing', () => {
    const result = evaluateScope2MarketBased('Uploaded RECs for all three offices.', []);

    expect(result.status).toBe('needs_clarification');
  });

  it('returns insufficient when no market-based instrument is described', () => {
    const result = evaluateScope2MarketBased(
      'Corrected the electricity bills and re-imported the usage values.',
      ['/api/abatement/sb253/evidence/123'],
    );

    expect(result.status).toBe('insufficient');
  });
});
