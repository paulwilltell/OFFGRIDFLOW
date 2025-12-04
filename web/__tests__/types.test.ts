import type {
  ModeResponse,
  ChatResponse,
  Scope2Summary,
  ComplianceSummary,
  FrameworkStatus,
  Scope2Emission,
  EmissionsTotals,
  CSRDComplianceResponse,
  PageInfo,
  PaginatedResponse,
} from '@/lib/types';

describe('TypeScript Types', () => {
  describe('ModeResponse', () => {
    it('should allow valid mode values', () => {
      const normalMode: ModeResponse = { mode: 'normal' };
      const offlineMode: ModeResponse = { mode: 'offline' };
      const degradedMode: ModeResponse = { mode: 'degraded' };

      expect(normalMode.mode).toBe('normal');
      expect(offlineMode.mode).toBe('offline');
      expect(degradedMode.mode).toBe('degraded');
    });
  });

  describe('ChatResponse', () => {
    it('should allow cloud source', () => {
      const response: ChatResponse = {
        output: 'Test response',
        source: 'cloud',
      };
      expect(response.source).toBe('cloud');
    });

    it('should allow local source', () => {
      const response: ChatResponse = {
        output: 'Test response',
        source: 'local',
      };
      expect(response.source).toBe('local');
    });
  });

  describe('Scope2Summary', () => {
    it('should have all required fields', () => {
      const summary: Scope2Summary = {
        scope: 'SCOPE2',
        totalKWh: 10000,
        totalEmissionsKgCO2e: 5000,
        totalEmissionsTonsCO2e: 5,
        averageEmissionFactor: 0.5,
        activityCount: 10,
        timestamp: new Date().toISOString(),
      };

      expect(summary.scope).toBe('SCOPE2');
      expect(summary.totalKWh).toBe(10000);
      expect(summary.activityCount).toBe(10);
    });

    it('should allow optional regionBreakdown', () => {
      const summary: Scope2Summary = {
        scope: 'SCOPE2',
        totalKWh: 10000,
        totalEmissionsKgCO2e: 5000,
        totalEmissionsTonsCO2e: 5,
        averageEmissionFactor: 0.5,
        activityCount: 10,
        timestamp: new Date().toISOString(),
        regionBreakdown: {
          'US-WEST': 2500,
          'US-EAST': 2500,
        },
      };

      expect(summary.regionBreakdown?.['US-WEST']).toBe(2500);
    });
  });

  describe('FrameworkStatus', () => {
    it('should support all status values', () => {
      const statuses: FrameworkStatus['status'][] = [
        'ok',
        'partial',
        'no_data',
        'not_started',
        'not_applicable',
      ];

      statuses.forEach((status) => {
        const framework: FrameworkStatus = {
          name: 'Test Framework',
          status,
        };
        expect(framework.status).toBe(status);
      });
    });

    it('should support optional scope readiness flags', () => {
      const framework: FrameworkStatus = {
        name: 'CSRD',
        status: 'partial',
        scope1Ready: true,
        scope2Ready: true,
        scope3Ready: false,
      };

      expect(framework.scope1Ready).toBe(true);
      expect(framework.scope3Ready).toBe(false);
    });
  });

  describe('ComplianceSummary', () => {
    it('should have all framework statuses', () => {
      const summary: ComplianceSummary = {
        frameworks: {
          csrd: { name: 'CSRD', status: 'ok' },
          sec: { name: 'SEC', status: 'partial' },
          cbam: { name: 'CBAM', status: 'no_data' },
          california: { name: 'California', status: 'not_started' },
        },
        totals: {
          scope1Tons: 100,
          scope2Tons: 200,
          scope3Tons: 300,
        },
        timestamp: new Date().toISOString(),
      };

      expect(summary.frameworks.csrd.status).toBe('ok');
      expect(summary.totals.scope1Tons + summary.totals.scope2Tons + summary.totals.scope3Tons).toBe(600);
    });
  });

  describe('Scope2Emission', () => {
    it('should have all required fields', () => {
      const emission: Scope2Emission = {
        id: 'emission-1',
        meterId: 'meter-1',
        location: 'Building A',
        region: 'US-WEST',
        quantityKWh: 1000,
        emissionsKgCO2e: 500,
        emissionsTonsCO2e: 0.5,
        emissionFactor: 0.5,
        methodology: 'location-based',
        dataSource: 'utility_bill',
        dataQuality: 'high',
        periodStart: '2024-01-01',
        periodEnd: '2024-01-31',
      };

      expect(emission.methodology).toBe('location-based');
    });

    it('should support market-based methodology', () => {
      const emission: Scope2Emission = {
        id: 'emission-2',
        meterId: 'meter-2',
        location: 'Building B',
        region: 'US-EAST',
        quantityKWh: 2000,
        emissionsKgCO2e: 800,
        emissionsTonsCO2e: 0.8,
        emissionFactor: 0.4,
        methodology: 'market-based',
        dataSource: 'rec',
        dataQuality: 'medium',
        periodStart: '2024-01-01',
        periodEnd: '2024-01-31',
      };

      expect(emission.methodology).toBe('market-based');
    });
  });

  describe('PageInfo', () => {
    it('should track pagination state', () => {
      const pageInfo: PageInfo = {
        page: 1,
        perPage: 20,
        total: 100,
        totalPages: 5,
        hasNext: true,
        hasPrev: false,
      };

      expect(pageInfo.totalPages).toBe(5);
      expect(pageInfo.hasNext).toBe(true);
      expect(pageInfo.hasPrev).toBe(false);
    });
  });

  describe('PaginatedResponse', () => {
    it('should wrap data with pagination info', () => {
      const response: PaginatedResponse<Scope2Emission> = {
        data: [],
        pageInfo: {
          page: 1,
          perPage: 20,
          total: 0,
          totalPages: 0,
          hasNext: false,
          hasPrev: false,
        },
      };

      expect(Array.isArray(response.data)).toBe(true);
      expect(response.pageInfo.page).toBe(1);
    });
  });

  describe('CSRDComplianceResponse', () => {
    it('should have all required CSRD fields', () => {
      const response: CSRDComplianceResponse = {
        standard: 'ESRS E1',
        orgId: 'org-123',
        year: 2024,
        totals: {
          scope1Tons: 100,
          scope2Tons: 200,
          scope3Tons: 300,
          totalTons: 600,
        },
        metrics: {},
        status: 'ok',
        timestamp: new Date().toISOString(),
      };

      expect(response.standard).toBe('ESRS E1');
      expect(response.totals.totalTons).toBe(600);
    });

    it('should support all status values', () => {
      const statuses: CSRDComplianceResponse['status'][] = ['ok', 'incomplete', 'warnings'];

      statuses.forEach((status) => {
        const response: CSRDComplianceResponse = {
          standard: 'ESRS E1',
          orgId: 'org-123',
          year: 2024,
          totals: { scope1Tons: 0, scope2Tons: 0, scope3Tons: 0, totalTons: 0 },
          metrics: {},
          status,
          timestamp: new Date().toISOString(),
        };
        expect(response.status).toBe(status);
      });
    });
  });
});
