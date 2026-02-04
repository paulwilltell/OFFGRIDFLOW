
'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

type Entitlements = {
  live_dashboard: boolean;
  export_reports: boolean;
  level: 'self' | 'concierge' | 'enterprise';
};

type Session = {
  user: {
    email: string;
    company: string;
    id: string;
  };
  entitlements: Entitlements;
  createdAt: string;
};

type GlobePoint = {
  lat: number;
  lng: number;
  level: 'critical' | 'warning' | 'info' | 'ok';
  intensity: number;
};

type GlobeData = {
  mode: 'DEMO' | 'LIVE';
  metrics: {
    today_kg: number;
    month_kg: number;
    scope2_pct: number;
    readiness_pct: number;
  };
  points: GlobePoint[];
};

type ToastTone = 'ok' | 'warn' | 'err';

type Toast = {
  id: number;
  title: string;
  message: string;
  tone: ToastTone;
};

type AuditEvent = {
  id: number;
  text: string;
  tone: string;
};

const MockBackend = {
  session: null as Session | null,

  async fetchSession() {
    await this.simulateLatency(300);
    return this.session;
  },

  async login(email: string, company: string) {
    await this.simulateLatency(800);
    const entitlementLevel = email.includes('enterprise')
      ? 'enterprise'
      : email.includes('concierge')
        ? 'concierge'
        : 'self';

    this.session = {
      user: { email, company, id: `usr_${Math.random().toString(36).slice(2, 9)}` },
      entitlements: {
        live_dashboard: true,
        export_reports: entitlementLevel !== 'self',
        level: entitlementLevel,
      },
      createdAt: new Date().toISOString(),
    };
    return this.session;
  },

  async logout() {
    await this.simulateLatency(200);
    this.session = null;
  },

  async getGlobeData(): Promise<GlobeData> {
    await this.simulateLatency(400);
    const isEntitled = this.session?.entitlements?.live_dashboard;

    if (!isEntitled) {
      return {
        mode: 'DEMO',
        metrics: {
          today_kg: 842.7 + (Math.random() * 100 - 50),
          month_kg: 24392 + Math.floor(Math.random() * 1000 - 500),
          scope2_pct: 68,
          readiness_pct: 94,
        },
        points: this.generateDemoPoints(150),
      };
    }

    return {
      mode: 'LIVE',
      metrics: {
        today_kg: 1247.3 + (Math.random() * 200 - 100),
        month_kg: 38471 + Math.floor(Math.random() * 2000 - 1000),
        scope2_pct: 72,
        readiness_pct: 98,
      },
      points: this.generateLivePoints(200),
    };
  },

  async exportReport() {
    await this.simulateLatency(1500);
    if (!this.session) throw new Error('UNAUTHORIZED');
    if (!this.session.entitlements.export_reports) throw new Error('PAYMENT_REQUIRED');

    return {
      downloadUrl: `#strategy-pack-${Date.now()}`,
      expiresIn: '15m',
      reportId: `pkg_${Math.random().toString(36).slice(2, 9)}`,
    };
  },

  generateDemoPoints(count: number): GlobePoint[] {
    return Array.from({ length: count }, () => ({
      lat: Math.random() * 160 - 80,
      lng: Math.random() * 360 - 180,
      level: Math.random() < 0.02 ? 'critical' : Math.random() < 0.08 ? 'warning' : Math.random() < 0.35 ? 'info' : 'ok',
      intensity: Math.random(),
    }));
  },

  generateLivePoints(count: number): GlobePoint[] {
    const facilities = [
      { lat: 37.7749, lng: -122.4194, name: 'SF HQ' },
      { lat: 34.0522, lng: -118.2437, name: 'LA Facility' },
      { lat: 40.7128, lng: -74.006, name: 'NY Office' },
      { lat: 51.5074, lng: -0.1278, name: 'London' },
      { lat: 35.6762, lng: 139.6503, name: 'Tokyo' },
    ];

    const points: GlobePoint[] = [];
    facilities.forEach((fac) => {
      points.push({
        lat: fac.lat + (Math.random() - 0.5) * 2,
        lng: fac.lng + (Math.random() - 0.5) * 2,
        level: Math.random() > 0.7 ? 'warning' : 'ok',
        intensity: 0.5 + Math.random() * 0.5,
      });
    });

    while (points.length < count) {
      points.push({
        lat: Math.random() * 160 - 80,
        lng: Math.random() * 360 - 180,
        level: Math.random() < 0.05 ? 'critical' : 'ok',
        intensity: Math.random() * 0.5,
      });
    }

    return points;
  },

  simulateLatency(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  },
};

export default function HomePage() {
  const [session, setSession] = useState<Session | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [modalKind, setModalKind] = useState<'discovery' | 'plans'>('discovery');
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [toasts, setToasts] = useState<Toast[]>([]);
  const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([
    { id: 1, text: '[SYSTEM] Growth transformation node initialized...', tone: 'text-emerald-700' },
  ]);
  const [auditMode, setAuditMode] = useState<'DEMO' | 'LIVE'>('DEMO');
  const [locked, setLocked] = useState(true);
  const [reduceMotion, setReduceMotion] = useState(false);
  const [form, setForm] = useState({ email: '', company: '', notes: '' });

  const lastUpdateRef = useRef<HTMLSpanElement>(null);
  const mTodayRef = useRef<HTMLDivElement>(null);
  const mMonthRef = useRef<HTMLDivElement>(null);
  const mScope2Ref = useRef<HTMLDivElement>(null);
  const mReadyRef = useRef<HTMLDivElement>(null);
  const mReadyBarRef = useRef<HTMLDivElement>(null);
  const globeTeaserRef = useRef<HTMLDivElement>(null);
  const globeLiveRef = useRef<HTMLDivElement>(null);

  const refreshIntervalRef = useRef<number | null>(null);
  const auditIntervalRef = useRef<number | null>(null);
  const threeRef = useRef<{ THREE: typeof import('three'); OrbitControls: any } | null>(null);
  const globeTeaserState = useRef<{ renderer: any; raf: number; resize: () => void } | null>(null);
  const globeLiveState = useRef<{
    renderer: any;
    scene: any;
    camera: any;
    pointsGroup: any;
    controls: any;
    raf: number;
    resize: () => void;
  } | null>(null);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const media = window.matchMedia('(prefers-reduced-motion: reduce)');
    const update = () => setReduceMotion(media.matches);
    update();
    media.addEventListener('change', update);
    return () => media.removeEventListener('change', update);
  }, []);

  const showToast = useCallback((title: string, message: string, tone: ToastTone = 'ok') => {
    const id = Date.now() + Math.floor(Math.random() * 1000);
    setToasts((prev) => [...prev, { id, title, message, tone }]);
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((toast) => toast.id !== id));
    }, 4200);
  }, []);

  const openModal = useCallback(
    (kind: 'discovery' | 'plans', plan?: string) => {
      const planSlug = plan ? plan.toLowerCase().replace(/\s+/g, '-') : '';
      setModalKind(kind);
      setSelectedPlan(plan ?? null);
      setModalOpen(true);
      setForm({
        email: plan ? `architect+${planSlug}@offgridflow.com` : '',
        company: '',
        notes: plan ? `Interested in ${plan} tier strategy.` : '',
      });
    },
    []
  );

  const closeModal = useCallback(() => {
    setModalOpen(false);
  }, []);

  useEffect(() => {
    document.body.style.overflow = modalOpen ? 'hidden' : '';
    return () => {
      document.body.style.overflow = '';
    };
  }, [modalOpen]);

  const loadThree = useCallback(async () => {
    if (threeRef.current) return threeRef.current;
    const THREE = await import('three');
    const controlsModule = await import('three/examples/jsm/controls/OrbitControls');
    threeRef.current = { THREE, OrbitControls: controlsModule.OrbitControls };
    return threeRef.current;
  }, []);

  const initGlobeTeaser = useCallback(async () => {
    const container = globeTeaserRef.current;
    if (!container || globeTeaserState.current) return;

    const { THREE } = await loadThree();
    container.innerHTML = '';

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(45, container.clientWidth / container.clientHeight, 0.1, 1000);
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    container.appendChild(renderer.domElement);

    const geometry = new THREE.IcosahedronGeometry(2, 2);
    const material = new THREE.MeshBasicMaterial({
      color: 0x00ff9d,
      wireframe: true,
      transparent: true,
      opacity: 0.2,
    });
    const globe = new THREE.Mesh(geometry, material);
    scene.add(globe);

    const innerGeo = new THREE.IcosahedronGeometry(1.9, 1);
    const innerMat = new THREE.MeshBasicMaterial({
      color: 0x00ff9d,
      wireframe: true,
      transparent: true,
      opacity: 0.05,
    });
    const inner = new THREE.Mesh(innerGeo, innerMat);
    scene.add(inner);

    camera.position.z = 5;

    let raf = 0;
    const animate = () => {
      raf = requestAnimationFrame(animate);
      globe.rotation.x += 0.001;
      globe.rotation.y += 0.002;
      inner.rotation.x -= 0.002;
      inner.rotation.y -= 0.001;
      renderer.render(scene, camera);
    };

    if (!reduceMotion) {
      animate();
    } else {
      renderer.render(scene, camera);
    }

    const resize = () => {
      if (!container.isConnected) return;
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    };

    window.addEventListener('resize', resize);
    globeTeaserState.current = { renderer, raf, resize };
  }, [loadThree, reduceMotion]);
  const initOrUpdateLiveGlobe = useCallback(
    async (points: GlobePoint[]) => {
      const container = globeLiveRef.current;
      if (!container) return;

      const { THREE, OrbitControls } = await loadThree();

      if (!globeLiveState.current) {
        container.innerHTML = '';

        const scene = new THREE.Scene();
        const camera = new THREE.PerspectiveCamera(45, container.clientWidth / container.clientHeight, 0.1, 1000);
        const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
        renderer.setSize(container.clientWidth, container.clientHeight);
        renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
        container.appendChild(renderer.domElement);

        const controls = new OrbitControls(camera, renderer.domElement);
        controls.enableDamping = true;
        controls.dampingFactor = 0.05;
        controls.autoRotate = !reduceMotion;
        controls.autoRotateSpeed = 0.8;
        controls.enableZoom = false;

        const sphereGeo = new THREE.IcosahedronGeometry(2, 3);
        const sphereMat = new THREE.MeshPhongMaterial({
          color: 0x0a0c12,
          emissive: 0x00ff9d,
          emissiveIntensity: 0.1,
          shininess: 100,
          wireframe: true,
        });
        const globe = new THREE.Mesh(sphereGeo, sphereMat);
        scene.add(globe);

        const coreGeo = new THREE.SphereGeometry(1.8, 32, 32);
        const coreMat = new THREE.MeshBasicMaterial({
          color: 0x00ff9d,
          transparent: true,
          opacity: 0.05,
        });
        const core = new THREE.Mesh(coreGeo, coreMat);
        scene.add(core);

        const pointsGroup = new THREE.Group();
        scene.add(pointsGroup);

        const ambient = new THREE.AmbientLight(0x404040, 2);
        scene.add(ambient);

        const pointLight = new THREE.PointLight(0x00ff9d, 1, 100);
        pointLight.position.set(5, 5, 5);
        scene.add(pointLight);

        camera.position.z = 5.5;

        let raf = 0;
        const animate = () => {
          raf = requestAnimationFrame(animate);
          controls.update();
          renderer.render(scene, camera);
        };

        if (!reduceMotion) {
          animate();
        } else {
          renderer.render(scene, camera);
        }

        const resize = () => {
          if (!container.isConnected) return;
          camera.aspect = container.clientWidth / container.clientHeight;
          camera.updateProjectionMatrix();
          renderer.setSize(container.clientWidth, container.clientHeight);
        };

        window.addEventListener('resize', resize);
        globeLiveState.current = { renderer, scene, camera, pointsGroup, controls, raf, resize };
      }

      const g = globeLiveState.current.pointsGroup;
      while (g.children.length) g.remove(g.children[0]);

      points.forEach((p) => {
        const color =
          p.level === 'critical'
            ? 0xff0040
            : p.level === 'warning'
              ? 0xffaa00
              : p.level === 'info'
                ? 0x00f0ff
                : 0x00ff9d;

        const size = p.intensity ? 0.04 + p.intensity * 0.05 : 0.05;
        const geometry = new THREE.SphereGeometry(size, 8, 8);
        const material = new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.9 });
        const dot = new THREE.Mesh(geometry, material);

        const phi = (90 - p.lat) * (Math.PI / 180);
        const theta = (p.lng + 180) * (Math.PI / 180);
        const radius = 2.05;

        dot.position.x = -(radius * Math.sin(phi) * Math.cos(theta));
        dot.position.y = radius * Math.cos(phi);
        dot.position.z = radius * Math.sin(phi) * Math.sin(theta);

        g.add(dot);
      });
    },
    [loadThree, reduceMotion]
  );

  const animateValue = useCallback(
    (
      ref: React.RefObject<HTMLElement>,
      value: number,
      options: { decimals?: number; suffix?: string; prefix?: string } = {}
    ) => {
      const el = ref.current;
      if (!el) return;

      const { decimals = 0, suffix = '', prefix = '' } = options;
      const start = parseFloat(el.textContent?.replace(/[^0-9.-]/g, '') || '0');
      const end = value;
      const duration = 1000;
      const startTime = performance.now();

      const update = (currentTime: number) => {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        const easeProgress = 1 - Math.pow(1 - progress, 3);
        const current = start + (end - start) * easeProgress;

        if (decimals === 0) {
          el.textContent = `${prefix}${Math.round(current).toLocaleString()}${suffix}`;
        } else {
          el.textContent = `${prefix}${current.toFixed(decimals)}${suffix}`;
        }

        if (progress < 1) requestAnimationFrame(update);
      };

      requestAnimationFrame(update);
    },
    []
  );

  const refreshDashboard = useCallback(async () => {
    if (lastUpdateRef.current) {
      lastUpdateRef.current.textContent = new Date().toLocaleTimeString('en-US', { hour12: false });
    }

    try {
      const data = await MockBackend.getGlobeData();

      animateValue(mTodayRef, data.metrics.today_kg, { decimals: 1 });
      animateValue(mMonthRef, data.metrics.month_kg, { decimals: 0 });
      animateValue(mScope2Ref, data.metrics.scope2_pct, { suffix: '%' });
      animateValue(mReadyRef, data.metrics.readiness_pct, { suffix: '%' });

      if (mReadyBarRef.current) {
        mReadyBarRef.current.style.width = `${data.metrics.readiness_pct}%`;
      }

      setAuditMode(data.mode);
      initOrUpdateLiveGlobe(data.points);

    } catch (err) {
      showToast('SYNC_ERROR', 'Signal stream interrupted. Retrying...', 'err');
    }
  }, [animateValue, initOrUpdateLiveGlobe, showToast]);

  const startAuditStream = useCallback(() => {
    if (auditIntervalRef.current) {
      clearInterval(auditIntervalRef.current);
    }

    const events = [
      { t: '[SIGNAL] ICP message calibration synced', c: 'text-emerald-500/70' },
      { t: '[PIPELINE] Discovery sprint queue updated', c: 'text-cyan-500/70' },
      { t: '[PROOF] New case study asset published', c: 'text-purple-400/70' },
      { t: '[PRICING] Value ladder iteration approved', c: 'text-emerald-500/70' },
      { t: '[OPS] Sales motion checkpoint executed', c: 'text-amber-500/70' },
      { t: '[SYSTEM] Cycle time trending downward', c: 'text-emerald-600/50' },
      { t: '[GROWTH] ARR trajectory recalculated', c: 'text-cyan-600/50' },
    ];

    auditIntervalRef.current = window.setInterval(() => {
      const now = new Date();
      const evt = events[Math.floor(Math.random() * events.length)];
      const id = Date.now() + Math.floor(Math.random() * 1000);
      setAuditEvents((prev) => {
        const next = [
          ...prev,
          {
            id,
            text: `${now.toLocaleTimeString('en-US', { hour12: false })} ${evt.t}`,
            tone: `${evt.c} border-l-2 border-emerald-500/20 pl-2 text-[9px]`,
          },
        ];
        return next.slice(-40);
      });
    }, 2200);
  }, []);

  const startRefreshInterval = useCallback(() => {
    if (refreshIntervalRef.current) {
      clearInterval(refreshIntervalRef.current);
    }

    refreshIntervalRef.current = window.setInterval(() => {
      if (document.visibilityState === 'visible') {
        refreshDashboard();
      }
    }, 30000);
  }, [refreshDashboard]);

  const bootstrap = useCallback(async () => {
    await initGlobeTeaser();

    const currentSession = await MockBackend.fetchSession();
    setSession(currentSession);
    setLocked(!currentSession?.entitlements?.live_dashboard);

    await refreshDashboard();
    startAuditStream();
    startRefreshInterval();
  }, [initGlobeTeaser, refreshDashboard, startAuditStream, startRefreshInterval]);

  useEffect(() => {
    bootstrap();

    return () => {
      if (refreshIntervalRef.current) clearInterval(refreshIntervalRef.current);
      if (auditIntervalRef.current) clearInterval(auditIntervalRef.current);

      if (globeTeaserState.current) {
        window.removeEventListener('resize', globeTeaserState.current.resize);
        cancelAnimationFrame(globeTeaserState.current.raf);
        globeTeaserState.current.renderer?.dispose?.();
        globeTeaserState.current = null;
      }

      if (globeLiveState.current) {
        window.removeEventListener('resize', globeLiveState.current.resize);
        cancelAnimationFrame(globeLiveState.current.raf);
        globeLiveState.current.renderer?.dispose?.();
        globeLiveState.current = null;
      }
    };
  }, [bootstrap]);

  const handleLogin = useCallback(
    async (email: string, company: string) => {
      try {
        const nextSession = await MockBackend.login(email, company);
        setSession(nextSession);
        setLocked(!nextSession.entitlements.live_dashboard);
        showToast('HANDSHAKE_COMPLETE', 'Discovery sprint request received. Strategy team inbound.');
        closeModal();
        await refreshDashboard();
      } catch (err) {
        showToast('TRANSMISSION_FAILED', 'Retry connection sequence.', 'err');
      }
    },
    [closeModal, refreshDashboard, showToast]
  );

  const handleLogout = useCallback(async () => {
    await MockBackend.logout();
    setSession(null);
    setLocked(true);
    showToast('DISCONNECTED', 'Session cleared. Strategic channel closed.');
    await refreshDashboard();
  }, [refreshDashboard, showToast]);

  const handleSubmit = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      if (!form.email || !form.company) {
        showToast('INPUT_ERROR', 'Provide valid identity credentials.', 'warn');
        return;
      }

      await handleLogin(form.email, form.company);
    },
    [form, handleLogin, showToast]
  );

  const handleExport = useCallback(async () => {
    showToast('GENERATING', 'Compiling strategic briefing artifact...');

    try {
      const result = await MockBackend.exportReport();
      showToast('ARTIFACT_READY', `Strategy pack ${result.reportId} compiled. Decay in ${result.expiresIn}.`);
    } catch (err) {
      if (err instanceof Error && err.message === 'UNAUTHORIZED') {
        showToast('AUTH_REQUIRED', 'Initialize session to export.', 'warn');
        openModal('discovery');
      } else if (err instanceof Error && err.message === 'PAYMENT_REQUIRED') {
        showToast('TIER_INSUFFICIENT', 'Upgrade to Pilot for export protocols.', 'warn');
      } else {
        showToast('COMPILATION_FAILED', 'Artifact generation error.', 'err');
      }
    }
  }, [openModal, showToast]);

  const scrollToId = useCallback((id: string) => {
    const target = document.getElementById(id);
    if (target) {
      target.scrollIntoView({ behavior: 'smooth' });
    }
  }, []);

  const modalCopy = useMemo(() => {
    if (modalKind === 'plans') {
      return {
        kicker: 'TIER_SELECTION',
        title: 'FOUNDERS_DISCOVERY_SPRINT',
        desc: `Founders Discovery Sprint ($4,995). We diagnose ICP, pricing, and pipeline gaps. ${
          selectedPlan ? `Requested focus: ${selectedPlan}.` : ''
        }`,
      };
    }

    return {
      kicker: 'DISCOVERY_SPRINT',
      title: 'FOUNDERS_DISCOVERY_SPRINT',
      desc: 'Founders Discovery Sprint ($4,995). We map ICP, pricing, and GTM execution into a 90-day launch plan.',
    };
  }, [modalKind, selectedPlan]);

  return (
    <div className="quantum-shell antialiased selection:bg-emerald-500/30 selection:text-emerald-200">
      <div className="scanlines" aria-hidden="true"></div>

      <nav className="fixed top-0 left-0 right-0 z-50 holo-panel border-b border-emerald-500/20">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <button
            type="button"
            className="flex items-center gap-4 group no-underline"
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
          >
            <div className="relative flex items-center justify-center h-12 w-36">
              <img
                src="/offgridflow_logo_primary.png"
                alt="OffGridFlow - Carbon Accounting"
                className="h-full w-auto object-contain filter brightness-125 group-hover:brightness-150 transition-all duration-300"
              />
              <div className="absolute -inset-2 bg-emerald-500/0 group-hover:bg-emerald-500/10 blur-md transition-all duration-500 -z-10 rounded-lg"></div>
            </div>
          </button>

          <div className="flex items-center gap-6">
            {session?.user?.email ? (
              <div className="flex items-center gap-4 px-4 py-2 holo-panel rounded-full border border-emerald-500/30">
                <div className="status-node"></div>
                <div className="flex flex-col">
                  <span className="text-[10px] text-emerald-400/60 font-mono tracking-wider">SYSTEM ONLINE</span>
                  <span className="text-xs text-emerald-100 font-mono max-w-[140px] truncate">
                    {session.user.email}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="text-xs text-emerald-600 hover:text-emerald-300 font-mono transition-colors ml-2"
                >
                  [DISCONNECT]
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-4">
                <button
                  type="button"
                  className="text-xs font-mono text-emerald-600/80 hover:text-emerald-300 transition-colors tracking-widest glitch-hover"
                  onClick={() => scrollToId('analysis')}
                >
                  ANALYSIS
                </button>
                <button
                  type="button"
                  className="text-xs font-mono text-emerald-600/80 hover:text-emerald-300 transition-colors tracking-widest glitch-hover"
                  onClick={() => scrollToId('pricing')}
                >
                  PRICING
                </button>
                <button type="button" className="btn-quantum px-4 py-2 rounded" onClick={() => scrollToId('roadmap')}>
                  ROADMAP
                </button>
                <button type="button" className="btn-primary px-6 py-2 text-xs rounded" onClick={() => openModal('discovery')}>
                  DISCOVERY SPRINT
                </button>
              </div>
            )}
          </div>
        </div>
      </nav>

      <section className="pt-32 pb-20 px-6 relative z-10">
        <div className="max-w-7xl mx-auto grid lg:grid-cols-2 gap-12 items-center">
          <div className="space-y-8 relative">
            <div className="absolute -left-4 top-0 bottom-0 w-px bg-gradient-to-b from-emerald-500/50 via-emerald-500/20 to-transparent"></div>

            <div className="inline-flex items-center gap-3 px-4 py-2 holo-panel rounded-full border border-emerald-500/30">
              <span className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse shadow-[0_0_10px_cyan]"></span>
              <span className="text-[10px] tracking-[0.3em] text-emerald-400/80 font-mono uppercase">
                Enterprise Transformation Protocol v1.0
              </span>
            </div>

            <h1 className="text-5xl md:text-7xl font-display leading-none tracking-tight">
              <span className="block text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 via-cyan-400 to-emerald-200 drop-shadow-[0_0_30px_rgba(0,255,157,0.3)]">
                FROM $2,500/YR
              </span>
              <span className="block text-emerald-100/90 mt-2">TO $409,500 ARR</span>
              <span className="block text-xl md:text-2xl font-mono font-light text-emerald-600/60 mt-4 tracking-[0.2em]">
                // THE 12-MONTH TRANSFORMATION
              </span>
            </h1>

            <p className="text-lg text-emerald-100/60 max-w-xl leading-relaxed font-light border-l-2 border-emerald-500/30 pl-6">
              The quantum leap from solo-founder startup to category-leading enterprise sales engine. A masterclass in
              positioning, pricing, and process.
            </p>

            <div className="grid sm:grid-cols-3 gap-4 pt-4">
              <div className="holo-panel p-4 rounded-lg group hover:border-emerald-400/50 transition-all cursor-default">
                <div className="text-[10px] text-emerald-600 font-mono mb-2 flex items-center gap-2">
                  <span className="w-1 h-1 bg-emerald-500 rounded-full"></span>
                  POSITIONING
                </div>
                <div className="text-sm font-display text-emerald-100">ICP TARGETING</div>
                <div className="text-[10px] text-emerald-700/60 font-mono mt-1">Messaging precision and trust</div>
              </div>
              <div className="holo-panel p-4 rounded-lg group hover:border-cyan-400/50 transition-all cursor-default">
                <div className="text-[10px] text-cyan-600 font-mono mb-2 flex items-center gap-2">
                  <span className="w-1 h-1 bg-cyan-500 rounded-full"></span>
                  PRICING
                </div>
                <div className="text-sm font-display text-emerald-100">VALUE LADDER</div>
                <div className="text-[10px] text-emerald-700/60 font-mono mt-1">Sprint to Enterprise escalation</div>
              </div>
              <div className="holo-panel p-4 rounded-lg group hover:border-purple-400/50 transition-all cursor-default">
                <div className="text-[10px] text-purple-400/80 font-mono mb-2 flex items-center gap-2">
                  <span className="w-1 h-1 bg-purple-500 rounded-full"></span>
                  PROCESS
                </div>
                <div className="text-sm font-display text-emerald-100">SYSTEMATIC GTM</div>
                <div className="text-[10px] text-emerald-700/60 font-mono mt-1">Repeatable sales motion</div>
              </div>
            </div>

            <div className="flex flex-wrap gap-4 pt-4">
              <button type="button" className="btn-primary px-8 py-4 text-sm" onClick={() => openModal('discovery')}>
                FOUNDERS DISCOVERY SPRINT
              </button>
              <button type="button" className="btn-quantum px-8 py-4 rounded" onClick={() => openModal('discovery')}>
                REQUEST STRATEGY BRIEF
              </button>
            </div>
          </div>

          <div className="space-y-6 relative">
            <div className="absolute inset-0 bg-emerald-500/5 blur-3xl rounded-full"></div>

            <div className="quantum-border holo-panel rounded-2xl p-1 relative">
              <div className="corner-accent"></div>
              <div className="corner-accent br"></div>

              <div className="relative rounded-xl overflow-hidden bg-black/40 p-6 grid-bg">
                <div className="absolute top-0 right-0 p-2">
                  <div className="flex gap-1">
                    <div className="w-1 h-1 bg-emerald-500 rounded-full animate-pulse"></div>
                    <div className="w-1 h-1 bg-emerald-500/50 rounded-full"></div>
                    <div className="w-1 h-1 bg-emerald-500/25 rounded-full"></div>
                  </div>
                </div>

                <div className="flex items-center justify-between mb-6">
                  <div>
                    <div className="text-[10px] text-emerald-600 font-mono tracking-widest mb-1">
                      PORTFOLIO_CORE_METRIC
                    </div>
                    <div className="text-xl font-display text-emerald-100">PROJECTED_IMPACT</div>
                  </div>
                  <div className="text-[10px] px-3 py-1 rounded holo-panel border border-cyan-500/30 text-cyan-400 font-mono">
                    YEAR_1_TARGET
                  </div>
                </div>

                <div className="text-center py-8">
                  <div className="text-6xl md:text-7xl font-mono font-bold text-gradient mb-2">263%</div>
                  <div className="text-lg font-display text-emerald-300 mb-4">Revenue Growth</div>
                  <div className="text-[12px] font-mono text-emerald-600/80 border-t border-emerald-500/20 pt-4 max-w-xs mx-auto">
                    <div className="flex justify-between">
                      <span>Baseline ARR:</span>
                      <span className="text-cyan-400">$150K</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Target ARR:</span>
                      <span className="text-emerald-400">$409.5K</span>
                    </div>
                  </div>
                </div>

                <div className="space-y-2 font-mono text-[10px] text-emerald-600/80 border-t border-emerald-500/20 pt-4">
                  <div className="flex justify-between">
                    <span>STRATEGY_STATUS</span>
                    <span className="text-emerald-400 flex items-center gap-2">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                      READY_FOR_EXECUTION
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div className="quantum-border holo-panel rounded-2xl p-1">
              <div className="rounded-xl bg-black/40 p-4">
                <div className="flex justify-between items-center mb-2">
                  <span className="text-[10px] font-mono text-emerald-600 tracking-widest">GROWTH_SIGNAL_MESH</span>
                  <span className="text-[10px] font-mono text-emerald-800">PREVIEW_MODE</span>
                </div>
                <div ref={globeTeaserRef} className="h-[240px] rounded-lg bg-black/60 relative overflow-hidden">
                  <div className="absolute inset-0 flex items-center justify-center text-emerald-800 font-mono text-xs">
                    INITIALIZING STRATEGY_MESH...
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="analysis" className="px-6 pb-24 relative z-10">
        <div className="max-w-7xl mx-auto">
          <div className="flex items-end justify-between gap-6 mb-12 border-b border-emerald-500/20 pb-4">
            <div>
              <div className="text-[10px] font-mono text-emerald-600 tracking-[0.3em] mb-2">CURRENT_STATE_ANALYSIS</div>
              <h2 className="text-3xl md:text-4xl font-display text-emerald-100">CRITICAL_GAPS</h2>
            </div>
            <div className="hidden md:flex items-center gap-6 text-[10px] font-mono text-emerald-700">
              <span className="flex items-center gap-2">
                <span className="w-2 h-2 bg-emerald-500 shadow-[0_0_10px_emerald]"></span> POSITIONING
              </span>
              <span className="flex items-center gap-2">
                <span className="w-2 h-2 bg-cyan-500 shadow-[0_0_10px_cyan]"></span> CONVERSION
              </span>
              <span className="flex items-center gap-2">
                <span className="w-2 h-2 bg-amber-500 shadow-[0_0_10px_amber]"></span> SOCIAL_PROOF
              </span>
            </div>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            <div className="holo-panel p-6 rounded-xl relative group cursor-default overflow-hidden">
              <div className="absolute top-0 right-0 p-4 text-6xl font-display text-emerald-500/10 group-hover:text-emerald-400/20 transition-colors">
                01
              </div>
              <div className="relative z-10">
                <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-4">CRITICAL_GAP_01</div>
                <h3 className="text-xl font-display text-emerald-100 mb-3 group-hover:text-emerald-300 transition-colors">
                  POSITIONING_SIGNAL
                </h3>
                <p className="text-sm text-emerald-100/60 leading-relaxed mb-4">
                  Messaging is feature-first and generic. ICP targeting is unclear and differentiation is weak.
                </p>
                <div className="text-[10px] font-mono text-emerald-700/80 border-t border-emerald-500/20 pt-3">
                  GRADE: D
                </div>
              </div>
              <div className="absolute bottom-0 left-0 h-1 bg-emerald-500/50 w-0 group-hover:w-full transition-all duration-700"></div>
            </div>

            <div className="holo-panel p-6 rounded-xl relative group cursor-default overflow-hidden">
              <div className="absolute top-0 right-0 p-4 text-6xl font-display text-cyan-500/10 group-hover:text-cyan-400/20 transition-colors">
                02
              </div>
              <div className="relative z-10">
                <div className="text-[10px] font-mono text-cyan-600 tracking-widest mb-4">CRITICAL_GAP_02</div>
                <h3 className="text-xl font-display text-emerald-100 mb-3 group-hover:text-cyan-300 transition-colors">
                  CONVERSION_ARCHITECTURE
                </h3>
                <p className="text-sm text-emerald-100/60 leading-relaxed mb-4">
                  Limited enterprise CTAs and no structured conversion paths. Evaluators lack guided next steps.
                </p>
                <div className="text-[10px] font-mono text-cyan-700/80 border-t border-cyan-500/20 pt-3">
                  GRADE: C-
                </div>
              </div>
              <div className="absolute bottom-0 left-0 h-1 bg-cyan-500/50 w-0 group-hover:w-full transition-all duration-700"></div>
            </div>

            <div className="holo-panel p-6 rounded-xl relative group cursor-default overflow-hidden">
              <div className="absolute top-0 right-0 p-4 text-6xl font-display text-purple-500/10 group-hover:text-purple-400/20 transition-colors">
                03
              </div>
              <div className="relative z-10">
                <div className="text-[10px] font-mono text-purple-400/80 tracking-widest mb-4">CRITICAL_GAP_03</div>
                <h3 className="text-xl font-display text-emerald-100 mb-3 group-hover:text-purple-300 transition-colors">
                  SOCIAL_PROOF_SIGNAL
                </h3>
                <p className="text-sm text-emerald-100/60 leading-relaxed mb-4">
                  No logos, testimonials, or case studies. Enterprise buyers lack validation and proof.
                </p>
                <div className="text-[10px] font-mono text-purple-700/80 border-t border-purple-500/20 pt-3">
                  GRADE: D
                </div>
              </div>
              <div className="absolute bottom-0 left-0 h-1 bg-purple-500/50 w-0 group-hover:w-full transition-all duration-700"></div>
            </div>
          </div>
        </div>
      </section>

      <section id="pricing" className="px-6 pb-24 relative z-10">
        <div className="max-w-7xl mx-auto grid lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <div className="holo-panel rounded-2xl p-1">
              <div className="rounded-xl bg-black/40 p-6">
                <div className="flex items-center justify-between mb-6 flex-wrap gap-4">
                  <div>
                    <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-1">VALUE_LADDER</div>
                    <div className="text-2xl font-display text-emerald-100">PRICING_STRATEGY</div>
                    <div className="flex items-center gap-2 mt-2 text-[10px] font-mono text-emerald-700">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                      ARR_TARGET: <span className="text-cyan-400">409,500</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <button type="button" className="btn-quantum px-4 py-2 rounded text-[10px]" onClick={handleExport}>
                      DOWNLOAD_BRIEF
                    </button>
                    <button type="button" className="btn-primary px-4 py-2 text-[10px]" onClick={() => openModal('discovery')}>
                      DISCOVERY_SPRINT
                    </button>
                  </div>
                </div>

                <div className="holo-panel p-4 rounded-xl border border-emerald-500/20 mb-6">
                  <div className="text-[10px] font-mono text-emerald-600 tracking-widest">FOUNDERS_LAUNCH_PRICING</div>
                  <div className="text-lg font-display text-emerald-100 mt-2">Phase 1: Proof and Credibility</div>
                  <div className="text-sm text-emerald-100/60 mt-2">
                    First 5-10 customers at $4,995 Discovery Sprint to collect testimonials and validation.
                  </div>
                </div>

                <div className="grid md:grid-cols-3 gap-4">
                  <div
                    className="holo-panel p-4 rounded border border-emerald-500/30 hover:border-emerald-400/50 transition-all cursor-pointer group"
                    onClick={() => openModal('plans', 'Discovery Sprint')}
                  >
                    <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-3">SPRINT</div>
                    <div className="font-display text-emerald-100 mb-1 group-hover:text-emerald-300">DISCOVERY</div>
                    <div className="text-xl font-mono text-emerald-400">$4,995</div>
                    <ul className="space-y-2 text-xs font-mono text-emerald-100/70 mt-4">
                      <li className="flex items-center gap-2"><span className="text-emerald-500">></span> ICP + messaging reset</li>
                      <li className="flex items-center gap-2"><span className="text-emerald-500">></span> Pricing ladder blueprint</li>
                      <li className="flex items-center gap-2"><span className="text-emerald-500">></span> 90-day execution plan</li>
                    </ul>
                  </div>

                  <div
                    className="holo-panel p-4 rounded border border-cyan-500/40 relative overflow-hidden cursor-pointer"
                    onClick={() => openModal('plans', 'Annual Pilot')}
                  >
                    <div className="absolute top-0 right-0 bg-cyan-500/20 text-cyan-300 text-[9px] px-2 py-1 font-mono">
                      STRATEGIC
                    </div>
                    <div className="font-display text-cyan-100 mb-1 mt-4">ANNUAL PILOT</div>
                    <div className="text-[10px] font-mono text-cyan-600">RECURRING REVENUE</div>
                    <div className="text-xl font-mono text-cyan-400 mt-2">$18K / YEAR</div>
                    <ul className="space-y-2 text-xs font-mono text-emerald-100/70 mt-4">
                      <li className="flex items-center gap-2"><span className="text-cyan-500">></span> Guided pipeline build</li>
                      <li className="flex items-center gap-2"><span className="text-cyan-500">></span> Quarterly KPI reviews</li>
                      <li className="flex items-center gap-2"><span className="text-cyan-500">></span> Conversion system tuning</li>
                    </ul>
                  </div>

                  <div
                    className="holo-panel p-4 rounded border border-purple-500/20 hover:border-purple-400/50 transition-all cursor-pointer"
                    onClick={() => openModal('plans', 'Enterprise Platform')}
                  >
                    <div className="font-display text-purple-200 mb-1">ENTERPRISE</div>
                    <div className="text-[10px] font-mono text-purple-400/80">CATEGORY LEADER</div>
                    <div className="text-xl font-mono text-purple-400 mt-2">$50K+ / YEAR</div>
                    <ul className="space-y-2 text-xs font-mono text-emerald-100/70 mt-4">
                      <li className="flex items-center gap-2"><span className="text-purple-500">></span> Dedicated sales engine</li>
                      <li className="flex items-center gap-2"><span className="text-purple-500">></span> Premium positioning</li>
                      <li className="flex items-center gap-2"><span className="text-purple-500">></span> Long-term scale support</li>
                    </ul>
                  </div>
                </div>

                <div className="holo-panel p-4 rounded border border-emerald-500/20 mt-6">
                  <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-2">ADD_ON_SERVICES</div>
                  <div className="text-sm text-emerald-100/70 font-mono">
                    CSRD Assessment ($15K) | SEC Preparation ($12K) | CBAM Setup ($8K) | Verification Coordination ($5K + fees)
                  </div>
                </div>
              </div>
            </div>

            <div className="holo-panel rounded-2xl p-1">
              <div className="rounded-xl bg-black/40 p-6 relative overflow-hidden">
                <div className="flex items-center justify-between mb-6 flex-wrap gap-4">
                  <div>
                    <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-1">GROWTH_SIGNAL_CONSOLE</div>
                    <div className="text-2xl font-display text-emerald-100">EXECUTION_READINESS</div>
                    <div className="flex items-center gap-2 mt-2 text-[10px] font-mono text-emerald-700">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                      STREAM_ACTIVE: <span ref={lastUpdateRef} className="text-cyan-400">--:--:--</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 font-mono">
                      {auditMode}
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 md:grid-cols-4 gap-3 p-4 pt-0">
                  <div className="holo-panel p-4 rounded border border-emerald-500/30">
                    <div className="text-[10px] font-mono text-emerald-700 mb-1">PIPELINE_SIGNAL</div>
                    <div className="text-2xl font-mono text-emerald-400 font-bold" ref={mTodayRef}>
                      --
                    </div>
                    <div className="text-[10px] font-mono text-emerald-600">WEEKLY_INDEX</div>
                  </div>
                  <div className="holo-panel p-4 rounded border border-cyan-500/30">
                    <div className="text-[10px] font-mono text-cyan-700 mb-1">ANNUALIZED_FLOW</div>
                    <div className="text-2xl font-mono text-cyan-400 font-bold" ref={mMonthRef}>
                      --
                    </div>
                    <div className="text-[10px] font-mono text-cyan-600">ARR_SIGNAL</div>
                  </div>
                  <div className="holo-panel p-4 rounded border border-purple-500/30">
                    <div className="text-[10px] font-mono text-purple-400/80 mb-1">CONVERSION_LIFT</div>
                    <div className="text-2xl font-mono text-purple-300 font-bold" ref={mScope2Ref}>
                      --%
                    </div>
                    <div className="text-[10px] font-mono text-purple-500/60">EFFICACY</div>
                  </div>
                  <div className="holo-panel p-4 rounded border border-amber-500/30">
                    <div className="text-[10px] font-mono text-amber-600 mb-1">READINESS</div>
                    <div className="text-2xl font-mono text-amber-400 font-bold" ref={mReadyRef}>
                      --%
                    </div>
                    <div className="h-1 mt-2 bg-black/50 rounded-full overflow-hidden">
                      <div
                        ref={mReadyBarRef}
                        className="h-full bg-amber-500 w-0 transition-all duration-1000 shadow-[0_0_10px_rgba(245,158,11,0.5)]"
                      ></div>
                    </div>
                  </div>
                </div>

                <div className="grid lg:grid-cols-3 gap-3 p-4 pt-0">
                  <div className="lg:col-span-2">
                    <div className="holo-panel rounded border border-emerald-500/20 p-1">
                      <div className="flex justify-between items-center p-2 border-b border-emerald-500/10">
                        <span className="text-[10px] font-mono text-emerald-700">MARKET_PENETRATION_MESH</span>
                        <span className="text-[10px] font-mono text-emerald-900">LIVE_FEED</span>
                      </div>
                      <div ref={globeLiveRef} className="h-[360px] bg-black/60 relative">
                        <div className="absolute inset-0 flex items-center justify-center text-emerald-900 font-mono text-xs">
                          AWAITING_STRATEGIC_ENTITLEMENT...
                        </div>
                      </div>
                    </div>
                  </div>
                  <div className="holo-panel rounded border border-emerald-500/20 p-4 flex flex-col">
                    <div className="flex justify-between items-center mb-3 border-b border-emerald-500/10 pb-2">
                      <span className="text-[10px] font-mono text-emerald-700">EXECUTION_STREAM</span>
                      <span className="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 font-mono">
                        {auditMode}
                      </span>
                    </div>
                    <div className="flex-1 overflow-y-auto font-mono text-[10px] text-emerald-600/80 space-y-1 h-[320px] pr-2">
                      {auditEvents.map((evt) => (
                        <div key={evt.id} className={evt.tone}>
                          {evt.text}
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                {locked ? (
                  <div className="quantum-lock absolute inset-0 flex items-center justify-center z-30">
                    <div className="text-center p-8 max-w-md">
                      <div className="w-16 h-16 mx-auto mb-4 relative">
                        <div
                          className="absolute inset-0 border-2 border-purple-500/30 rounded-full animate-spin"
                          style={{ animationDuration: '3s' }}
                        ></div>
                        <div
                          className="absolute inset-2 border-2 border-purple-500/60 rounded-full animate-spin"
                          style={{ animationDuration: '2s', animationDirection: 'reverse' }}
                        ></div>
                        <div className="absolute inset-0 flex items-center justify-center text-2xl"></div>
                      </div>
                      <h3 className="text-xl font-display text-purple-300 mb-2">STRATEGY_LOCK_ACTIVE</h3>
                      <p className="text-sm text-emerald-100/60 font-mono mb-6">
                        Enterprise transformation telemetry requires verified discovery sprint access.
                      </p>
                      <div className="flex gap-4 justify-center">
                        <button type="button" className="btn-primary px-6 py-3 text-xs" onClick={() => openModal('discovery')}>
                          REQUEST_SPRINT
                        </button>
                        <button type="button" className="btn-quantum px-6 py-3 rounded" onClick={() => scrollToId('pricing')}>
                          VIEW_TIERS
                        </button>
                      </div>
                    </div>
                  </div>
                ) : null}
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="holo-panel p-6 rounded-xl">
              <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-2">QUARTERLY_PATHWAY</div>
              <h3 className="text-xl font-display text-emerald-100 mb-6">ARR_GROWTH_MAP</h3>
              <div className="space-y-3 text-xs font-mono text-emerald-100/70">
                <div className="flex justify-between border-b border-emerald-500/10 pb-2">
                  <span>Q1</span>
                  <span>$52K ARR - Website optimization</span>
                </div>
                <div className="flex justify-between border-b border-emerald-500/10 pb-2">
                  <span>Q2</span>
                  <span>$130K ARR - Sales motion system</span>
                </div>
                <div className="flex justify-between border-b border-emerald-500/10 pb-2">
                  <span>Q3</span>
                  <span>$247K ARR - Authority content</span>
                </div>
                <div className="flex justify-between">
                  <span>Q4</span>
                  <span>$409.5K ARR - Category leadership</span>
                </div>
              </div>
            </div>

            <div className="holo-panel p-6 rounded-xl">
              <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-2">NEXT_STEP</div>
              <h3 className="text-xl font-display text-emerald-100 mb-4">FOUNDERS DISCOVERY SPRINT</h3>
              <p className="text-sm text-emerald-100/60 mb-6">
                $4,995 engagement to build the ICP, pricing ladder, and 90-day execution map.
              </p>
              <button type="button" className="btn-primary w-full py-3 text-xs" onClick={() => openModal('discovery')}>
                REQUEST_SPRINT
              </button>
            </div>
          </div>
        </div>
      </section>

      <section id="roadmap" className="px-6 pb-24 relative z-10">
        <div className="max-w-7xl mx-auto">
          <div className="flex items-end justify-between gap-6 mb-12 border-b border-emerald-500/20 pb-4">
            <div>
              <div className="text-[10px] font-mono text-emerald-600 tracking-[0.3em] mb-2">EXECUTION_ROADMAP</div>
              <h2 className="text-3xl md:text-4xl font-display text-emerald-100">90_DAY_IMPLEMENTATION</h2>
            </div>
            <div className="hidden md:flex items-center gap-6 text-[10px] font-mono text-emerald-700">
              <span className="flex items-center gap-2"><span className="status-node"></span> WEEKS 1-4</span>
              <span className="flex items-center gap-2"><span className="status-node"></span> WEEKS 5-8</span>
              <span className="flex items-center gap-2"><span className="status-node"></span> WEEKS 9-12</span>
            </div>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            <div className="quantum-border holo-panel p-6 rounded-xl">
              <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-3">WEEKS_1_4</div>
              <div className="text-xl font-display text-emerald-100 mb-2">Foundation and Quick Wins</div>
              <ul className="text-xs font-mono text-emerald-100/70 space-y-2">
                <li>ICP-targeted hero rewrite + CTAs</li>
                <li>Launch proof collection and testimonials</li>
                <li>Publish compliance playbook asset</li>
                <li>Build competitive comparison page</li>
              </ul>
              <div className="text-[10px] font-mono text-emerald-700/80 border-t border-emerald-500/20 pt-3 mt-4">
                Expected lift: 15-30% conversion
              </div>
            </div>

            <div className="quantum-border holo-panel p-6 rounded-xl">
              <div className="text-[10px] font-mono text-cyan-600 tracking-widest mb-3">WEEKS_5_8</div>
              <div className="text-xl font-display text-emerald-100 mb-2">Interactive Engagement</div>
              <ul className="text-xs font-mono text-emerald-100/70 space-y-2">
                <li>Animated dashboard + interactive tour</li>
                <li>ROI calculator and pricing proof</li>
                <li>First enterprise case study</li>
                <li>Sales enablement templates</li>
              </ul>
              <div className="text-[10px] font-mono text-cyan-700/80 border-t border-cyan-500/20 pt-3 mt-4">
                Expected lift: 3-5x demo completion
              </div>
            </div>

            <div className="quantum-border holo-panel p-6 rounded-xl">
              <div className="text-[10px] font-mono text-purple-400/80 tracking-widest mb-3">WEEKS_9_12</div>
              <div className="text-xl font-display text-emerald-100 mb-2">Authority Building</div>
              <ul className="text-xs font-mono text-emerald-100/70 space-y-2">
                <li>Resource hub + 5 SEO pieces</li>
                <li>Sales playbook + outreach automation</li>
                <li>Partner and analyst outreach</li>
                <li>Pipeline review cadence</li>
              </ul>
              <div className="text-[10px] font-mono text-purple-700/80 border-t border-purple-500/20 pt-3 mt-4">
                Expected lift: 20-40% sales efficiency
              </div>
            </div>
          </div>
        </div>
      </section>

      <footer className="relative z-10 border-t border-emerald-500/20 px-6 py-10 text-center text-emerald-600 text-xs font-mono">
        <div className="max-w-4xl mx-auto space-y-3">
          <div className="text-emerald-400 font-display">OFFGRIDFLOW ENTERPRISE STRATEGY SYSTEM</div>
          <div>Quantum carbon intelligence and revenue acceleration platform.</div>
          <div>© 2026 OffGridFlow LLC. All rights reserved.</div>
        </div>
      </footer>

      <div className="fixed bottom-6 right-6 z-[9999] space-y-3 pointer-events-none">
        {toasts.map((toast) => {
          const borderColors: Record<ToastTone, string> = {
            ok: 'border-emerald-500',
            warn: 'border-amber-500',
            err: 'border-red-500',
          };

          return (
            <div key={toast.id} className={`quantum-toast p-4 min-w-[300px] pointer-events-auto ${borderColors[toast.tone]}`}>
              <div className="text-emerald-400 font-display text-sm mb-1">{toast.title}</div>
              <div className="text-emerald-600/80 font-mono text-xs">{toast.message}</div>
            </div>
          );
        })}
      </div>

      {modalOpen ? (
        <div className="fixed inset-0 z-[9999]">
          <button type="button" className="absolute inset-0 bg-black/90 backdrop-blur-xl" onClick={closeModal}></button>
          <div className="absolute inset-4 md:inset-20 max-w-2xl mx-auto holo-panel rounded-2xl border border-emerald-500/30 flex flex-col">
            <div className="p-6 border-b border-emerald-500/20 flex justify-between items-center">
              <div>
                <div className="text-[10px] font-mono text-emerald-600 tracking-widest mb-1">{modalCopy.kicker}</div>
                <div className="text-2xl font-display text-emerald-100">{modalCopy.title}</div>
              </div>
              <button type="button" onClick={closeModal} className="text-emerald-600 hover:text-emerald-300 text-2xl">
                &times;
              </button>
            </div>
            <div className="p-6 flex-1 overflow-auto">
              <p className="text-sm text-emerald-100/60 font-mono mb-6">{modalCopy.desc}</p>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-1">
                  <label className="text-[10px] font-mono text-emerald-600 tracking-widest">BUSINESS_EMAIL</label>
                  <input
                    type="email"
                    required
                    className="w-full quantum-input px-4 py-3 rounded"
                    placeholder="user@corporation.units"
                    value={form.email}
                    onChange={(event) => setForm((prev) => ({ ...prev, email: event.target.value }))}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-[10px] font-mono text-emerald-600 tracking-widest">ORGANIZATION_NODE</label>
                  <input
                    type="text"
                    required
                    className="w-full quantum-input px-4 py-3 rounded"
                    placeholder="Corporate Entity"
                    value={form.company}
                    onChange={(event) => setForm((prev) => ({ ...prev, company: event.target.value }))}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-[10px] font-mono text-emerald-600 tracking-widest">STRATEGY_NOTES</label>
                  <textarea
                    className="w-full quantum-input px-4 py-3 rounded h-24 resize-none"
                    placeholder="Add strategic context..."
                    value={form.notes}
                    onChange={(event) => setForm((prev) => ({ ...prev, notes: event.target.value }))}
                  ></textarea>
                </div>
                <button type="submit" className="w-full btn-primary py-4 text-sm mt-6">
                  TRANSMIT_REQUEST
                </button>
              </form>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
