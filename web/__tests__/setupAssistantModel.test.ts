import {
  buildAssistantSteps,
  getAssistantProgress,
  getNextAssistantStep,
  type AssistantProgressInput,
} from '@/app/(app)/components/setupAssistantModel';

function makeProgress(overrides: Partial<AssistantProgressInput> = {}): AssistantProgressInput {
  return {
    profileReviewed: false,
    dashboardTourCompleted: false,
    hasUploadedData: false,
    hasConnectedDataSource: false,
    hasGeneratedReport: false,
    ...overrides,
  };
}

describe('setupAssistantModel', () => {
  it('marks the first incomplete step as current', () => {
    const steps = buildAssistantSteps(
      makeProgress({
        profileReviewed: true,
        dashboardTourCompleted: true,
      }),
    );

    expect(steps.find((step) => step.id === 'profile')?.status).toBe('complete');
    expect(steps.find((step) => step.id === 'tour')?.status).toBe('complete');
    expect(steps.find((step) => step.id === 'upload')?.status).toBe('current');
    expect(steps.find((step) => step.id === 'connect')?.status).toBe('pending');
  });

  it('calculates overall and required completion percentages', () => {
    const steps = buildAssistantSteps(
      makeProgress({
        profileReviewed: true,
        dashboardTourCompleted: true,
        hasUploadedData: true,
      }),
    );

    expect(getAssistantProgress(steps)).toEqual({
      total: 5,
      completeCount: 3,
      requiredTotal: 4,
      requiredCompleteCount: 3,
      percentComplete: 60,
      requiredPercentComplete: 75,
    });
  });

  it('returns the next incomplete step', () => {
    const steps = buildAssistantSteps(
      makeProgress({
        profileReviewed: true,
        dashboardTourCompleted: true,
        hasUploadedData: true,
        hasConnectedDataSource: true,
      }),
    );

    expect(getNextAssistantStep(steps)?.id).toBe('report');
  });
});
