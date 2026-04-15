import { attachCSRFHeader } from '@/lib/csrf';
import { ApiRequestError, TENANT_ID_KEY, api } from '@/lib/api';
import type {
  AbatementDashboardData,
  AbatementFramework,
  EvaluateAbatementResponse,
  SelfCertifyResponse,
} from '@/lib/abatement/types';

type EvaluatePayload = {
  complianceCheckId: string;
  completed: boolean;
  justification: string;
  files: File[];
};

type SelfCertifyPayload = {
  actionItemId?: string;
  complianceCheckId?: string;
  selfCertified: boolean;
};

function getTenantHeader(headers: Headers) {
  if (typeof window === 'undefined') {
    return;
  }

  const tenantId = localStorage.getItem(TENANT_ID_KEY);
  if (tenantId) {
    headers.set('X-Tenant-ID', tenantId);
  }
}

async function parseJsonResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let payload: { error?: { code?: string; message?: string; detail?: string; fields?: Record<string, string> } } = {};
    try {
      payload = (await response.json()) as typeof payload;
    } catch {
      // ignore non-json error bodies
    }
    throw new ApiRequestError(response.status, {
      code: payload.error?.code ?? 'unknown_error',
      message: payload.error?.message ?? `Request failed with status ${response.status}`,
      detail: payload.error?.detail,
      fields: payload.error?.fields,
      status: response.status,
    });
  }

  return (await response.json()) as T;
}

export async function fetchAbatementDashboard(framework: AbatementFramework): Promise<AbatementDashboardData> {
  return api.get<AbatementDashboardData>(`/api/abatement/${framework}`);
}

export async function evaluateAbatementRisk(
  framework: AbatementFramework,
  payload: EvaluatePayload,
): Promise<EvaluateAbatementResponse> {
  const headers = new Headers();
  getTenantHeader(headers);
  await attachCSRFHeader(headers);

  const formData = new FormData();
  formData.set('compliance_check_id', payload.complianceCheckId);
  formData.set('completed', String(payload.completed));
  formData.set('justification', payload.justification);
  payload.files.forEach((file) => formData.append('evidence', file));

  const response = await fetch(`/api/abatement/${framework}/evaluate`, {
    method: 'POST',
    headers,
    credentials: 'include',
    body: formData,
  });

  return parseJsonResponse<EvaluateAbatementResponse>(response);
}

export async function selfCertifyAbatementRisk(
  framework: AbatementFramework,
  payload: SelfCertifyPayload,
): Promise<SelfCertifyResponse> {
  return api.post<SelfCertifyResponse>(`/api/abatement/${framework}/self-certify`, {
    action_item_id: payload.actionItemId,
    compliance_check_id: payload.complianceCheckId,
    self_certified: payload.selfCertified,
  });
}

export async function downloadAbatementReport(framework: AbatementFramework): Promise<void> {
  const headers = new Headers();
  getTenantHeader(headers);

  const response = await fetch(`/api/abatement/${framework}/report`, {
    method: 'GET',
    headers,
    credentials: 'include',
  });

  if (!response.ok) {
    await parseJsonResponse(response);
    return;
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  const contentDisposition = response.headers.get('Content-Disposition');
  const fileNameMatch = contentDisposition?.match(/filename="(.+)"/);
  anchor.href = url;
  anchor.download = fileNameMatch?.[1] ?? `${framework}-risk-abatement-report.pdf`;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  window.URL.revokeObjectURL(url);
}
