import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { AbatementDashboard } from '@/components/abatement/AbatementDashboard';
import type { AbatementFramework } from '@/lib/abatement/types';

const supportedFrameworks: Record<AbatementFramework, string> = {
  sb253: 'California SB 253',
  csrd: 'CSRD / ESRS E1',
  sec: 'SEC Climate Disclosure',
  ifrs: 'IFRS S2',
  cbam: 'EU CBAM',
};

type PageParams = {
  framework: string;
};

export async function generateMetadata({
  params,
}: {
  params: Promise<PageParams>;
}): Promise<Metadata> {
  const { framework } = await params;
  if (!(framework in supportedFrameworks)) {
    return {};
  }

  const label = supportedFrameworks[framework as AbatementFramework];
  return {
    title: `${label} Risk Abatement | OffGridFlow`,
    description:
      'Assess remediation justifications, attach evidence, self-certify when needed, and generate a draft risk-abatement workplan.',
  };
}

export default async function AbatementPage({
  params,
}: {
  params: Promise<PageParams>;
}) {
  const { framework } = await params;

  if (!(framework in supportedFrameworks)) {
    notFound();
  }

  return <AbatementDashboard framework={framework as AbatementFramework} />;
}
