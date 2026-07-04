'use client';

import { Suspense, useState, FormEvent } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { ApiRequestError, api } from '@/lib/api';
import { useSession } from '@/lib/session';

interface RegisterResponse {
  user: {
    id: string;
    email: string;
    name: string;
    first_name?: string;
    last_name?: string;
    email_verified: boolean;
  };
  tenant: {
    id: string;
    name: string;
  };
  requires_verification?: boolean;
  verification_token?: string;
}

function RegisterPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login } = useSession();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [companyName, setCompanyName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [verificationSent, setVerificationSent] = useState(false);
  const [verificationToken, setVerificationToken] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Please enter your name');
      return;
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    // Mirror the backend password policy exactly so users get instant, accurate
    // feedback instead of a server-side rejection: 8+ chars, upper, lower,
    // digit, and one special character.
    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }
    if (!/[A-Z]/.test(password)) {
      setError('Password must contain at least one uppercase letter');
      return;
    }
    if (!/[a-z]/.test(password)) {
      setError('Password must contain at least one lowercase letter');
      return;
    }
    if (!/[0-9]/.test(password)) {
      setError('Password must contain at least one number');
      return;
    }
    if (!/[^A-Za-z0-9]/.test(password)) {
      setError('Password must contain at least one special character (e.g. ! @ # $)');
      return;
    }

    setLoading(true);

    try {
      const response = await api.post<RegisterResponse>('/api/auth/register', {
        name: name.trim(),
        email,
        password,
        company_name: companyName || undefined,
        terms_accepted: true,
        terms_accepted_at: new Date().toISOString(),
      });

      if (response.requires_verification) {
        setVerificationSent(true);
        if (process.env.NODE_ENV === 'development' && response.verification_token) {
          setVerificationToken(response.verification_token);
        }
      } else {
        await login({ email, password });
        router.push('/emissions');
      }
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.fields && Object.keys(err.fields).length > 0) {
          setError(Object.values(err.fields).join(' '));
        } else {
          setError(err.message);
        }
      } else {
        setError('An unexpected error occurred. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  if (verificationSent) {
    return (
      <div className="flex min-h-screen items-center justify-center" style={{ background: '#f7f8f6', fontFamily: "'Schibsted Grotesk', system-ui, sans-serif" }}>
        <div className="w-full max-w-md space-y-6 text-center">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full" style={{ background: '#e8f0ea' }}>
            <span className="text-2xl">✉️</span>
          </div>
          <h1 className="text-2xl font-bold" style={{ color: '#16201b' }}>Check your email</h1>
          <p className="text-[15px]" style={{ color: '#6a7a71' }}>
            We sent a verification link to <strong style={{ color: '#16201b' }}>{email}</strong>. Click it to activate your workspace.
          </p>
          {process.env.NODE_ENV === 'development' && verificationToken && (
            <div className="rounded-xl border p-4" style={{ background: '#fbf8ee', borderColor: '#efe6cc' }}>
              <p className="mb-2 text-sm font-semibold">Dev mode</p>
              <Link href={`/verify-email?token=${verificationToken}`} className="text-sm font-semibold" style={{ color: '#2f6b50' }}>
                Verify now →
              </Link>
            </div>
          )}
          <Link href="/login" className="block text-[13.5px] font-semibold" style={{ color: '#2f6b50' }}>
            Back to sign in
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen" style={{ fontFamily: "'Schibsted Grotesk', system-ui, sans-serif", color: '#16201b' }}>
      {/* Brand panel */}
      <div className="hidden w-[486px] flex-col justify-between p-[52px_48px] lg:flex" style={{ background: '#1d3b2e', color: '#eaf2ec' }}>
        <div className="flex items-center gap-[11px] text-[18px] font-bold tracking-[-0.01em]">
          <span className="flex h-[26px] w-[26px] items-center justify-center rounded-[7px] text-[15px]" style={{ background: '#5fbf8e', color: '#10271d' }}>◇</span>
          OffGridFlow
        </div>
        <div>
          <div className="text-[38px] font-bold leading-[1.12] tracking-[-0.02em]">
            Carbon accounting<br />without the busywork.
          </div>
          <p className="mt-[22px] max-w-[340px] text-[15.5px] leading-[1.6]" style={{ color: '#a9c6b6' }}>
            Upload your activity data. We calculate Scope&nbsp;1, 2 and 3 and hand you an audit-ready report. That&apos;s the whole product.
          </p>
        </div>
        <div className="flex flex-col gap-[18px]">
          {['Upload your data', 'Review your footprint', 'Download the report'].map((label, i) => (
            <div key={i} className="flex items-center gap-[14px]">
              <span className="flex h-[30px] w-[30px] items-center justify-center rounded-full border text-[13px]" style={{ borderColor: '#3f6b54', fontFamily: "'IBM Plex Mono', monospace", color: '#7fce9f' }}>
                {i + 1}
              </span>
              <span className="text-[15px]" style={{ color: '#cfe2d6' }}>{label}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Form */}
      <div className="flex flex-1 flex-col items-center justify-center p-8 lg:p-16" style={{ background: '#fff' }}>
        <form onSubmit={handleSubmit} className="w-full max-w-[420px]">
          <h1 className="mb-2 text-[27px] font-bold tracking-[-0.02em]">Create your workspace</h1>
          <p className="mb-[34px] text-[15px]" style={{ color: '#6a7a71' }}>Free to set up. You only pay when you export a report.</p>

          {error && (
            <div className="mb-4 rounded-lg border p-3 text-[13.5px]" style={{ background: '#fef2f2', borderColor: '#fecaca', color: '#991b1b' }}>
              {error}
            </div>
          )}

          <label className="mb-[7px] block text-[12.5px] font-semibold" style={{ color: '#5b6b62' }}>Your name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Jane Doe"
            required
            autoComplete="name"
            className="mb-[18px] block h-[46px] w-full rounded-[9px] border px-[14px] text-[14.5px] outline-none focus:border-[#2f6b50]"
            style={{ borderColor: '#e2e7e3', color: '#16201b' }}
          />

          <label className="mb-[7px] block text-[12.5px] font-semibold" style={{ color: '#5b6b62' }}>Work email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            required
            autoComplete="email"
            className="mb-[18px] block h-[46px] w-full rounded-[9px] border px-[14px] text-[14.5px] outline-none focus:border-[#2f6b50]"
            style={{ borderColor: '#e2e7e3', color: '#16201b' }}
          />

          <label className="mb-[7px] block text-[12.5px] font-semibold" style={{ color: '#5b6b62' }}>Company</label>
          <input
            type="text"
            value={companyName}
            onChange={(e) => setCompanyName(e.target.value)}
            placeholder="Acme Manufacturing Inc."
            className="mb-[18px] block h-[46px] w-full rounded-[9px] border px-[14px] text-[14.5px] outline-none focus:border-[#2f6b50]"
            style={{ borderColor: '#e2e7e3', color: '#16201b' }}
          />

          <label className="mb-[7px] block text-[12.5px] font-semibold" style={{ color: '#5b6b62' }}>Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="8+ chars · upper, lower, number, symbol"
            required
            className="mb-[18px] block h-[46px] w-full rounded-[9px] border px-[14px] text-[14.5px] outline-none focus:border-[#2f6b50]"
            style={{ borderColor: '#e2e7e3', color: '#16201b' }}
          />

          <label className="mb-[7px] block text-[12.5px] font-semibold" style={{ color: '#5b6b62' }}>Confirm password</label>
          <input
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Confirm your password"
            required
            className="mb-[30px] block h-[46px] w-full rounded-[9px] border px-[14px] text-[14.5px] outline-none focus:border-[#2f6b50]"
            style={{ borderColor: '#e2e7e3', color: '#16201b' }}
          />

          <button
            type="submit"
            disabled={loading}
            className="flex h-[48px] w-full items-center justify-center gap-2 rounded-[9px] text-[15px] font-semibold text-white disabled:opacity-60"
            style={{ background: '#1d3b2e' }}
          >
            {loading ? 'Creating...' : 'Create workspace'} {!loading && <span className="text-[17px]">→</span>}
          </button>

          <p className="mt-[18px] text-center text-[13.5px]" style={{ color: '#8a978f' }}>
            Already have one?{' '}
            <Link href="/login" className="font-semibold" style={{ color: '#2f6b50' }}>Sign in</Link>
          </p>
        </form>
      </div>
    </div>
  );
}

export default function RegisterPage() {
  return (
    <Suspense>
      <RegisterPageContent />
    </Suspense>
  );
}
