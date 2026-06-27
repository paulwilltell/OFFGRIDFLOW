'use client';

import {
  useState,
  useRef,
  useEffect,
  useCallback,
  useMemo,
  type ButtonHTMLAttributes,
  type HTMLAttributes,
  type KeyboardEvent,
  type ReactNode,
} from "react";

/**
 * USPS Advocate - 2026 UI refresh
 * - Zero external UI deps (CSS variables + modern layout primitives)
 * - Better behaviors: theming, offline resilience, local history, keyboard UX, message actions, toasts
 */

const DEFAULT_API_URL = "http://localhost:3001/api/v1";
const API_URL =
  (typeof window !== "undefined" && (window as any).__USPS_ADVOCATE_API_URL__) ||
  (typeof process !== "undefined" && process.env && process.env.NEXT_PUBLIC_USPS_ADVOCATE_API_URL) ||
  DEFAULT_API_URL;

type ToastKind = "success" | "danger" | "neutral";
type BadgeTone = ToastKind | "info";

type Toast = {
  id: string;
  kind: ToastKind;
  title: string;
  detail?: string;
};

type MessageData = {
  title?: string;
  content?: string[];
  sources?: string[];
  followUp?: string[];
};

type Message = {
  role: "user" | "assistant";
  ts: number;
  text?: string;
  data?: MessageData;
};

type EmployeeType = "career" | "noncareer";
type AssistantPayload = { text?: string; data?: MessageData };

type AdvocateArticle = {
  title?: string;
  content?: string;
  sources?: string[];
};

type AdvocateResponse = {
  sessionId?: string;
  articles?: AdvocateArticle[];
  quickActions?: Array<{ label: string }>;
};

// --- API Client -------------------------------------------------------------
const apiCall = async (path: string, options: RequestInit = {}) => {
  try {
    const headers = {
      "Content-Type": "application/json",
      ...((options.headers as Record<string, string>) ?? {}),
    };
    const res = await fetch(`${API_URL}${path}`, {
      ...options,
      headers,
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`API Error ${res.status}: ${text}`);
    }
    return res.json();
  } catch (err) {
    console.error("API call failed:", err);
    throw err;
  }
};

// --- Static Fallback Responses (when backend unavailable) -------------------
const FALLBACK_RESPONSES: Record<string, MessageData> = {
  general: {
    title: "USPS Advocate",
    content: [
      "**Tell me what's happening** and I'll give you practical, step-by-step guidance.",
      "> This tool is not a lawyer. It's a structured assistant for common USPS employee situations.",
      "",
      "**Fast starts:**",
      "- Investigation / Weingarten rights",
      "- Discipline and proposed removal",
      "- Grievances and steward requests",
      "- FMLA and medical leave basics",
      "- Harassment / EEO / retaliation",
    ],
    sources: ["USPS ELM", "Joint Contract Administration Manual (JCAM)"],
    followUp: [
      "I'm being investigated",
      "I received an adverse action",
      "I need FMLA help",
      "I'm being harassed",
      "Who represents me?",
    ],
  },
  investigation: {
    title: "Investigation and Weingarten rights",
    content: [
      "**If this is an investigatory meeting** and you believe it could lead to discipline, you can request a steward.",
      "Say: 'If this discussion could lead to discipline, I request union representation.'",
      "Ask what the meeting is about and which policies are at issue.",
      "Take notes: date, time, who was present, and key questions.",
      "Do not guess. If you don't know, say you need to check records.",
    ],
    sources: ["NLRB Weingarten Rights", "USPS ELM", "JCAM"],
    followUp: [
      "How do I ask for a steward?",
      "Can they deny representation?",
      "What should I write down?",
      "What happens after the meeting?",
    ],
  },
  discipline: {
    title: "Discipline, LOW, and suspensions",
    content: [
      "Ask for a copy of the discipline notice and the evidence relied on.",
      "Contact your steward quickly; deadlines can be short.",
      "Write a timeline and list witnesses or supporting records.",
      "Identify mitigation: length of service, prior record, intent, training, policy clarity.",
      "Keep communication factual and professional.",
    ],
    sources: ["USPS ELM", "JCAM"],
    followUp: [
      "I received a letter of warning",
      "How long do I have to grieve?",
      "What mitigation matters?",
      "Help me draft a timeline",
    ],
  },
  removal: {
    title: "Proposed removal response",
    content: [
      "A Notice of Proposed Removal includes charges, evidence, and a reply deadline.",
      "You can submit a written and/or oral reply; ask your steward for help.",
      "Request the full investigative file and any statements.",
      "Prepare mitigation facts and propose a lesser discipline where appropriate.",
      "Track all dates and keep copies of everything submitted.",
    ],
    sources: ["USPS ELM", "JCAM"],
    followUp: [
      "How do I respond to a removal?",
      "What should my reply include?",
      "Can I ask for more time?",
      "What are common mitigation factors?",
    ],
  },
  grievance: {
    title: "Grievances and steward requests",
    content: [
      "Work with your steward to identify contract or ELM provisions involved.",
      "Document the issue, impact, and requested remedy.",
      "Gather supporting evidence and witness statements.",
      "File within contractual time limits and get a timestamp or receipt.",
      "Stay consistent, concise, and focused on the remedy.",
    ],
    sources: ["JCAM", "USPS ELM"],
    followUp: [
      "How do I start a grievance?",
      "What is a reasonable remedy?",
      "Can I get a steward at Step 1?",
      "Help me outline the facts",
    ],
  },
  fmla: {
    title: "FMLA and medical leave basics",
    content: [
      "If you have a serious health condition, request FMLA forms promptly.",
      "Submit medical certification by the deadline and keep copies.",
      "Communicate restrictions and return-to-work updates.",
      "If denied, request the reason in writing and discuss next steps with your union.",
      "Track absences and paperwork dates.",
    ],
    sources: ["DOL FMLA", "USPS ELM"],
    followUp: [
      "What counts as a serious health condition?",
      "How long do I have to return forms?",
      "What if my supervisor denies leave?",
      "What should I document?",
    ],
  },
  harassment: {
    title: "Harassment, EEO, and retaliation",
    content: [
      "Document incidents with dates, locations, witnesses, and exact words or actions.",
      "Report through management and/or EEO channels and ask about timelines.",
      "If retaliation is occurring, document it separately.",
      "Request interim measures if safety or work environment is affected.",
      "Keep copies of all communications.",
    ],
    sources: ["EEOC", "USPS ELM"],
    followUp: [
      "How do I start an EEO complaint?",
      "What if my supervisor is involved?",
      "How should I document incidents?",
      "What are interim measures?",
    ],
  },
  representation: {
    title: "Representation and union help",
    content: [
      "USPS unions include NALC (city carriers), APWU (clerks/maintenance), NRLCA (rural carriers), and Mail Handlers.",
      "Ask for a steward for investigatory meetings or discipline discussions.",
      "You can request representation even as a non-career employee.",
      "Keep a record of who you contacted and when.",
      "If you need help identifying your local, share your craft and station.",
    ],
    sources: ["USPS ELM", "JCAM"],
    followUp: [
      "Which union covers my craft?",
      "How do I reach my local?",
      "Can a steward attend any meeting?",
      "What if no steward is available?",
    ],
  },
};

function matchTopicLocal(text: string): keyof typeof FALLBACK_RESPONSES {
  const t = text.toLowerCase();
  if (/weingarten|investigat|ii[0-9]|pdi|fact[- ]finding|pre[- ]discipline/.test(t)) return "investigation";
  if (/proposed removal|notice of removal|removal|termination|discharge/.test(t)) return "removal";
  if (/low|letter of warning|suspension|discipline|adverse action/.test(t)) return "discipline";
  if (/fmla|medical|doctor|injur|sick|leave|restriction|light duty/.test(t)) return "fmla";
  if (/harass|eeo|retaliat|hostile|discrimin|bully/.test(t)) return "harassment";
  if (/grievance|step\s*1|step\s*2|contract|article/.test(t)) return "grievance";
  if (/representation|union rep|steward|union|nalc|apwu|nrlca|mail handlers/.test(t)) return "representation";
  return "general";
}

// --- UI Utilities ------------------------------------------------------------
const cx = (...parts: Array<string | false | null | undefined>) => parts.filter(Boolean).join(" ");

function useLocalStorageState<T>(key: string, initialValue: T) {
  const [value, setValue] = useState<T>(() => {
    try {
      const raw = localStorage.getItem(key);
      if (raw !== null) return JSON.parse(raw);
      return initialValue;
    } catch {
      return initialValue;
    }
  });

  useEffect(() => {
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch {
      // ignore
    }
  }, [key, value]);

  return [value, setValue] as const;
}

function useReducedMotion() {
  const [reduced, setReduced] = useState(false);
  useEffect(() => {
    if (typeof window === "undefined") return;
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    const set = () => setReduced(Boolean(mq.matches));
    set();
    mq.addEventListener?.("change", set);
    return () => mq.removeEventListener?.("change", set);
  }, []);
  return reduced;
}

function ToastHost({ toasts, onDismiss }: { toasts: Toast[]; onDismiss: (id: string) => void }) {
  return (
    <div className="ta-toastHost" aria-live="polite" aria-relevant="additions removals">
      {toasts.map((t) => (
        <div key={t.id} className={cx("ta-toast", `ta-toast--${t.kind}`)}>
          <div className="ta-toast__title">{t.title}</div>
          {t.detail ? <div className="ta-toast__detail">{t.detail}</div> : null}
          <button className="ta-iconBtn" onClick={() => onDismiss(t.id)} aria-label="Dismiss notification">
            x
          </button>
        </div>
      ))}
    </div>
  );
}

function Badge({ tone = "neutral", children }: { tone?: BadgeTone; children: ReactNode }) {
  return <span className={cx("ta-badge", `ta-badge--${tone}`)}>{children}</span>;
}

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "ghost";
  size?: "md";
  className?: string;
};

function Button({ variant = "primary", size = "md", className, ...props }: ButtonProps) {
  return <button className={cx("ta-btn", `ta-btn--${variant}`, `ta-btn--${size}`, className)} {...props} />;
}

type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  className?: string;
};

function IconButton({ className, ...props }: IconButtonProps) {
  return <button className={cx("ta-iconBtn", className)} {...props} />;
}

type CardProps = HTMLAttributes<HTMLDivElement> & {
  className?: string;
};

function Card({ className, ...props }: CardProps) {
  return <div className={cx("ta-card", className)} {...props} />;
}

function Divider() {
  return <div className="ta-divider" role="separator" />;
}

function SkeletonLine() {
  return <div className="ta-skelLine" />;
}

function formatTime(ts: number) {
  try {
    return new Date(ts).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
}

function copyToClipboard(text: string) {
  if (!text) return;
  if (navigator?.clipboard?.writeText) return navigator.clipboard.writeText(text);
  const ta = document.createElement("textarea");
  ta.value = text;
  document.body.appendChild(ta);
  ta.select();
  document.execCommand("copy");
  document.body.removeChild(ta);
}

// --- Message Rendering -------------------------------------------------------
function RichText({ lines }: { lines: string[] }) {
  return (
    <div className="ta-richText">
      {lines.map((line, i) => {
        if (typeof line !== "string") return null;

        const isBq = line.startsWith("> ");
        const text = isBq ? line.replace(/^> /, "") : line;
        const parts = text.split(/\*\*(.+?)\*\*/g);

        const rendered = parts.map((part, j) =>
          j % 2 === 1 ? <strong key={j}>{part}</strong> : <span key={j}>{part}</span>
        );

        if (isBq) {
          return (
            <div key={i} className="ta-quote">
              <span>{rendered}</span>
            </div>
          );
        }
        return <p key={i}>{rendered}</p>;
      })}
    </div>
  );
}

type MessageBubbleProps = {
  msg: Message;
  onQuickAction: (text: string) => void;
  onCopy: (text: string) => void;
  isReducedMotion: boolean;
};

function MessageBubble({ msg, onQuickAction, onCopy, isReducedMotion }: MessageBubbleProps) {
  const isUser = msg.role === "user";
  const data = msg.data || {};
  const lines = data.content || (msg.text ? [msg.text] : []);

  return (
    <div className={cx("ta-msgRow", isUser ? "ta-msgRow--user" : "ta-msgRow--assistant")}>
      <div className={cx("ta-bubble", isUser ? "ta-bubble--user" : "ta-bubble--assistant")}>
        {!isUser && data.title ? (
          <div className="ta-bubble__titleRow">
            <div className="ta-bubble__title">{data.title}</div>
            <div className="ta-bubble__meta">
              <Badge tone="info">Guidance</Badge>
            </div>
          </div>
        ) : null}

        <RichText lines={lines} />

        {!isUser && data.sources?.length ? (
          <>
            <Divider />
            <div className="ta-sources">
              <div className="ta-sectionLabel">Sources</div>
              <ul>
                {data.sources.map((s, i) => (
                  <li key={i} className="ta-sourceItem">
                    <span className="ta-dot" aria-hidden="true" />
                    <span>{s}</span>
                  </li>
                ))}
              </ul>
            </div>
          </>
        ) : null}

        {!isUser && data.followUp?.length ? (
          <>
            <Divider />
            <div className="ta-quick">
              <div className="ta-sectionLabel">Suggested follow-ups</div>
              <div className="ta-chipRow">
                {data.followUp.slice(0, 8).map((q, i) => (
                  <button key={i} className="ta-chip" onClick={() => onQuickAction(q)}>
                    {q}
                  </button>
                ))}
              </div>
            </div>
          </>
        ) : null}

        <div className="ta-bubble__footer">
          <span className="ta-time">{formatTime(msg.ts)}</span>
          <div className="ta-actions">
            <IconButton onClick={() => onCopy(lines.join("\n"))} aria-label="Copy message" title="Copy">
              copy
            </IconButton>
            {!isUser ? <span className={cx("ta-pulse", isReducedMotion && "ta-pulse--off")} aria-hidden="true" /> : null}
          </div>
        </div>
      </div>
    </div>
  );
}

// --- Main Component ----------------------------------------------------------
export default function USPSAdvocate() {
  const [employeeType, setEmployeeType] = useLocalStorageState<EmployeeType>("ta.employeeType", "career");
  const [sessionId, setSessionId] = useLocalStorageState("ta.sessionId", "");
  const [messages, setMessages] = useLocalStorageState<Message[]>("ta.messages", []);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [backendOnline, setBackendOnline] = useState(true);
  const [theme, setTheme] = useLocalStorageState("ta.theme", "light");
  const [toasts, setToasts] = useState<Toast[]>([]);
  const [search, setSearch] = useState("");

  const reducedMotion = useReducedMotion();

  const listRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const pushToast = useCallback((kind: ToastKind, title: string, detail?: string) => {
    const id = `${Date.now()}-${Math.random().toString(16).slice(2)}`;
    const t = { id, kind, title, detail };
    setToasts((prev) => [t, ...prev].slice(0, 4));
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((x) => x.id !== id));
    }, 3800);
  }, []);

  const resetThread = useCallback(() => {
    setSessionId("");
    setMessages([]);
    pushToast("neutral", "Conversation cleared");
    // focus input
    requestAnimationFrame(() => inputRef.current?.focus());
  }, [pushToast, setMessages, setSessionId]);

  const scrollToBottom = useCallback(
    (behavior: ScrollBehavior = "smooth") => {
      const el = listRef.current;
      if (!el) return;
      try {
        el.scrollTo({ top: el.scrollHeight, behavior: reducedMotion ? "auto" : behavior });
      } catch {
        el.scrollTop = el.scrollHeight;
      }
    },
    [reducedMotion]
  );

  // Initialize welcome message if empty
  useEffect(() => {
    if (messages?.length) return;
    const welcome: Message = {
      role: "assistant",
      ts: Date.now(),
      text: "",
      data: {
        title: "USPS Advocate",
        content: [
          "**Tell me what's happening** and I'll give you practical, step-by-step guidance.",
          "> This tool is not a lawyer. It's a structured assistant for common USPS employee situations.",
          "",
          "**Fast starts:**",
          "- Investigation / Weingarten rights",
          "- Discipline and proposed removal",
          "- Grievances and steward requests",
          "- FMLA and medical leave basics",
          "- Harassment / EEO / retaliation",
        ],
        sources: [],
        followUp: [
          "I'm being investigated",
          "I received an adverse action",
          "I need FMLA help",
          "I'm being harassed",
          "Who represents me?",
        ],
      },
    };
    setMessages([welcome]);
  }, [messages, setMessages]);

  // Keep theme on root
  useEffect(() => {
    if (typeof document === "undefined") return;
    document.documentElement.dataset.taTheme = theme;
  }, [theme]);

  // Connectivity probe (gentle)
  useEffect(() => {
    let alive = true;
    const ping = async () => {
      try {
        await apiCall("/health");
        if (!alive) return;
        setBackendOnline(true);
      } catch {
        if (!alive) return;
        setBackendOnline(false);
      }
    };
    ping();
    const id = window.setInterval(ping, 15000);
    return () => {
      alive = false;
      window.clearInterval(id);
    };
  }, []);

  // Auto scroll on new messages/loading
  useEffect(() => {
    scrollToBottom("smooth");
  }, [messages, loading, scrollToBottom]);

  const filteredMessages = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return messages;
    return (messages || []).filter((m) => {
      const text = (m.data?.content || []).join(" ") + " " + (m.data?.title || "") + " " + (m.text || "");
      return text.toLowerCase().includes(q);
    });
  }, [messages, search]);

  const sendMessage = useCallback(
    async (text: string) => {
      const trimmed = text.trim();
      if (!trimmed || loading) return;

      // cancel any inflight request
      abortRef.current?.abort?.();
      const ac = new AbortController();
      abortRef.current = ac;

      setInput("");
      setLoading(true);

      const userMsg: Message = { role: "user", text: trimmed, ts: Date.now() };
      setMessages((prev) => [...(prev || []), userMsg]);

      const addAssistant = (payload: AssistantPayload) => {
        setMessages((prev) => [...(prev || []), { role: "assistant", ts: Date.now(), ...payload }]);
      };

      try {
        if (backendOnline) {
          const res = (await apiCall("/advocate/query", {
            method: "POST",
            signal: ac.signal,
            body: JSON.stringify({
              query: trimmed,
              employeeType,
              sessionId: sessionId || undefined,
            }),
          })) as AdvocateResponse;

          if (!sessionId && res.sessionId) setSessionId(res.sessionId);

          const article = res.articles?.[0];
          const contentLines =
            article?.content?.split("\n").map((l) => l.trim()).filter(Boolean) ||
            ["Try asking about investigations, discipline, grievances, FMLA, or harassment."];

          addAssistant({
            text: article?.content || "",
            data: {
              title: article?.title || "Guidance",
              content: contentLines,
              sources: article?.sources || [],
              followUp: res.quickActions?.map((qa) => qa.label) || [],
            },
          });
        } else {
          const topic = matchTopicLocal(trimmed);
          const resp = FALLBACK_RESPONSES[topic] || FALLBACK_RESPONSES.general;
          addAssistant({ text: "", data: resp });
        }
      } catch (err: any) {
        if (err?.name === "AbortError") {
          pushToast("neutral", "Request cancelled");
        } else {
          console.error(err);
          pushToast("danger", "Couldn't reach backend", "Showing local guidance instead.");
          setBackendOnline(false);
          const topic = matchTopicLocal(trimmed);
          const resp = FALLBACK_RESPONSES[topic] || FALLBACK_RESPONSES.general;
          addAssistant({ text: "", data: resp });
        }
      } finally {
        setLoading(false);
        requestAnimationFrame(() => inputRef.current?.focus());
      }
    },
    [backendOnline, employeeType, loading, pushToast, sessionId, setMessages, setSessionId]
  );

  const onComposerKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        sendMessage(input);
      }
      // Cmd/Ctrl+K focus search
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        const el = document.getElementById("ta-search");
        el?.focus?.();
      }
    },
    [input, sendMessage]
  );

  const onCopy = useCallback(
    (text: string) => {
      copyToClipboard(text);
      pushToast("success", "Copied");
    },
    [pushToast]
  );

  const connectionTone = backendOnline ? "success" : "danger";

  return (
    <div className="ta-root">
      <style>{GLOBAL_STYLES}</style>

      <ToastHost toasts={toasts} onDismiss={(id) => setToasts((prev) => prev.filter((t) => t.id !== id))} />

      <div className="ta-shell">
        {/* Sidebar */}
        <aside className="ta-side" aria-label="Sidebar">
          <div className="ta-brand">
            <div className="ta-logo" aria-hidden="true">
              UA
            </div>
            <div>
              <div className="ta-brand__name">USPS Advocate</div>
              <div className="ta-brand__sub">Guidance / Drafting / Next steps</div>
            </div>
          </div>

          <Card className="ta-panel">
            <div className="ta-panel__row">
              <div className="ta-panel__label">Connection</div>
              <Badge tone={connectionTone}>{backendOnline ? "Backend online" : "Local mode"}</Badge>
            </div>

            <div className="ta-panel__row">
              <div className="ta-panel__label">Employee type</div>
              <div className="ta-seg">
                <button className={cx("ta-seg__btn", employeeType === "career" && "is-active")} onClick={() => setEmployeeType("career")}>
                  Career
                </button>
                <button
                  className={cx("ta-seg__btn", employeeType === "noncareer" && "is-active")}
                  onClick={() => setEmployeeType("noncareer")}
                >
                  Non-career
                </button>
              </div>
            </div>

            <div className="ta-panel__row">
              <div className="ta-panel__label">Theme</div>
              <div className="ta-seg">
                <button className={cx("ta-seg__btn", theme === "dark" && "is-active")} onClick={() => setTheme("dark")}>
                  Dark
                </button>
                <button className={cx("ta-seg__btn", theme === "light" && "is-active")} onClick={() => setTheme("light")}>
                  Light
                </button>
              </div>
            </div>

            <Divider />

            <div className="ta-panel__label">Quick starts</div>
            <div className="ta-chipCol">
              {[
                "I'm being investigated (Weingarten)",
                "I got a LOW / suspension",
                "I received a proposed removal",
                "I need FMLA / medical leave",
                "Harassment / hostile environment",
                "How do I file a grievance?",
              ].map((q) => (
                <button key={q} className="ta-chip ta-chip--block" onClick={() => sendMessage(q)}>
                  {q}
                </button>
              ))}
            </div>

            <Divider />

            <div className="ta-sideActions">
              <Button variant="ghost" onClick={resetThread}>
                Clear
              </Button>
              <Button
                variant="ghost"
                onClick={() => {
                  setBackendOnline((v) => !v);
                  pushToast("neutral", "Mode toggled", "You can force Local mode for testing.");
                }}
              >
                Toggle Local
              </Button>
            </div>
          </Card>

          <div className="ta-footNote">
            Tip: <kbd>Enter</kbd> to send | <kbd>Shift</kbd>+<kbd>Enter</kbd> for newline | <kbd>Ctrl</kbd>/<kbd>Cmd</kbd>+<kbd>K</kbd> search
          </div>
        </aside>

        {/* Main */}
        <main className="ta-main">
          <header className="ta-topbar">
            <div className="ta-topbar__left">
              <div className="ta-title">Conversation</div>
              <div className="ta-subtitle">{backendOnline ? "Backend responses + sources" : "Offline local guidance (fallback)"}</div>
            </div>

            <div className="ta-topbar__right">
              <div className="ta-searchWrap">
                <input
                  id="ta-search"
                  className="ta-search"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder="Search this thread..."
                />
                {search ? (
                  <IconButton onClick={() => setSearch("")} aria-label="Clear search" title="Clear">
                    x
                  </IconButton>
                ) : (
                  <span className="ta-searchHint" aria-hidden="true">
                    Ctrl/Cmd+K
                  </span>
                )}
              </div>

              <Button variant="primary" onClick={() => sendMessage("Summarize the situation and next steps")}>Get next steps</Button>
            </div>
          </header>

          <div className="ta-chat" ref={listRef}>
            {filteredMessages.map((m, idx) => (
              <MessageBubble
                key={`${m.ts || idx}-${idx}`}
                msg={m}
                isReducedMotion={reducedMotion}
                onQuickAction={(q) => sendMessage(q)}
                onCopy={onCopy}
              />
            ))}

            {loading ? (
              <div className="ta-msgRow ta-msgRow--assistant">
                <div className="ta-bubble ta-bubble--assistant">
                  <div className="ta-bubble__titleRow">
                    <div className="ta-bubble__title">Thinking</div>
                    <div className="ta-bubble__meta">
                      <Badge tone="neutral">Working</Badge>
                    </div>
                  </div>
                  <div className="ta-richText">
                    <SkeletonLine />
                    <SkeletonLine />
                    <SkeletonLine />
                  </div>
                  <div className="ta-bubble__footer">
                    <span className="ta-time">...</span>
                    <div className="ta-actions">
                      <IconButton onClick={() => abortRef.current?.abort?.()} aria-label="Cancel request" title="Cancel">
                        stop
                      </IconButton>
                    </div>
                  </div>
                </div>
              </div>
            ) : null}
          </div>

          <footer className="ta-composer">
            <div className="ta-composer__box">
              <textarea
                ref={inputRef}
                className="ta-input"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={onComposerKeyDown}
                placeholder="Describe what happened (dates, who said what, any letters received)..."
                rows={1}
              />
              <div className="ta-composer__actions">
                <Button
                  variant="ghost"
                  onClick={() => {
                    onCopy(
                      (messages || [])
                        .map((m) => `[${m.role}] ${(m.data?.title || "").trim()} ${(m.data?.content || [m.text]).join(" ")}`)
                        .join("\n\n")
                    );
                  }}
                  title="Copy entire thread"
                >
                  Copy thread
                </Button>
                <Button variant="primary" disabled={loading || !input.trim()} onClick={() => sendMessage(input)}>
                  Send
                </Button>
              </div>
            </div>

            <div className="ta-composer__hint">
              Add specifics: craft/position, incident dates, manager names, what documentation exists, and what outcome you want.
            </div>
          </footer>
        </main>
      </div>
    </div>
  );
}

// --- Global Styles (single-file, tokenized) ---------------------------------
const GLOBAL_STYLES = `
@import url('https://fonts.googleapis.com/css2?family=Fraunces:wght@500;700&family=Space+Grotesk:wght@400;500;600;700&display=swap');

:root {
  --ta-font: "Space Grotesk", "Segoe UI", system-ui, -apple-system, sans-serif;
  --ta-display: "Fraunces", "Times New Roman", serif;
  --ta-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;

  --ta-radius-lg: 18px;
  --ta-radius-md: 14px;
  --ta-radius-sm: 12px;

  --ta-shadow: 0 14px 40px rgba(0, 0, 0, 0.35);
  --ta-shadow-soft: 0 10px 24px rgba(0, 0, 0, 0.22);

  --ta-dur: 180ms;
  --ta-ease: cubic-bezier(0.2, 0.8, 0.2, 1);

  --ta-gap: 14px;
}

:root[data-ta-theme="dark"] {
  --ta-bg: #070a12;
  --ta-bg2: #0b1020;
  --ta-surface: rgba(255, 255, 255, 0.06);
  --ta-surface2: rgba(255, 255, 255, 0.08);
  --ta-border: rgba(255, 255, 255, 0.1);
  --ta-text: rgba(255, 255, 255, 0.92);
  --ta-muted: rgba(255, 255, 255, 0.68);
  --ta-faint: rgba(255, 255, 255, 0.5);

  --ta-accent: #60a5fa;
  --ta-accent2: #f97316;
  --ta-good: #34d399;
  --ta-warn: #fbbf24;
  --ta-bad: #fb7185;
}

:root[data-ta-theme="light"] {
  --ta-bg: #f7f8fb;
  --ta-bg2: #ffffff;
  --ta-surface: rgba(0, 0, 0, 0.045);
  --ta-surface2: rgba(0, 0, 0, 0.06);
  --ta-border: rgba(0, 0, 0, 0.1);
  --ta-text: rgba(0, 0, 0, 0.88);
  --ta-muted: rgba(0, 0, 0, 0.62);
  --ta-faint: rgba(0, 0, 0, 0.46);

  --ta-accent: #2563eb;
  --ta-accent2: #ea580c;
  --ta-good: #059669;
  --ta-warn: #b45309;
  --ta-bad: #e11d48;
}

* {
  box-sizing: border-box;
}
html,
body {
  height: 100%;
  margin: 0;
  font-family: var(--ta-font);
  background: radial-gradient(1200px 800px at 20% 10%, rgba(96, 165, 250, 0.18), transparent 55%),
    radial-gradient(1000px 700px at 80% 30%, rgba(249, 115, 22, 0.12), transparent 55%),
    linear-gradient(180deg, var(--ta-bg), var(--ta-bg2));
  color: var(--ta-text);
}

.ta-root {
  min-height: 100vh;
}

.ta-shell {
  display: grid;
  grid-template-columns: 360px 1fr;
  min-height: 100vh;
  animation: taEnter 320ms var(--ta-ease);
}

@media (max-width: 980px) {
  .ta-shell {
    grid-template-columns: 1fr;
  }
  .ta-side {
    display: none;
  }
}

.ta-side {
  padding: 18px;
  border-right: 1px solid var(--ta-border);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.02), transparent 70%);
}

.ta-main {
  display: grid;
  grid-template-rows: auto 1fr auto;
  min-width: 0;
}

.ta-brand {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 14px 12px 18px;
}

.ta-logo {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, rgba(96, 165, 250, 0.22), rgba(249, 115, 22, 0.18));
  border: 1px solid var(--ta-border);
  box-shadow: var(--ta-shadow-soft);
  font-weight: 800;
  letter-spacing: 0.5px;
}

.ta-brand__name {
  font-weight: 800;
  letter-spacing: 0.2px;
  font-size: 15px;
  font-family: var(--ta-display);
}
.ta-brand__sub {
  font-size: 12px;
  color: var(--ta-muted);
  margin-top: 2px;
}

.ta-card {
  background: var(--ta-surface);
  border: 1px solid var(--ta-border);
  border-radius: var(--ta-radius-lg);
  box-shadow: var(--ta-shadow-soft);
}

.ta-panel {
  padding: 14px;
}

.ta-panel__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 10px;
  border-radius: 14px;
}
.ta-panel__row:hover {
  background: var(--ta-surface2);
}

.ta-panel__label {
  font-size: 12px;
  color: var(--ta-muted);
  font-weight: 700;
  letter-spacing: 0.2px;
}

.ta-seg {
  display: inline-flex;
  gap: 6px;
  padding: 4px;
  border-radius: 999px;
  border: 1px solid var(--ta-border);
  background: rgba(0, 0, 0, 0.06);
}
:root[data-ta-theme="light"] .ta-seg {
  background: rgba(255, 255, 255, 0.7);
}

.ta-seg__btn {
  border: 0;
  background: transparent;
  color: var(--ta-muted);
  padding: 8px 10px;
  border-radius: 999px;
  font-weight: 800;
  font-size: 12px;
  cursor: pointer;
  transition: transform var(--ta-dur) var(--ta-ease), background var(--ta-dur) var(--ta-ease),
    color var(--ta-dur) var(--ta-ease);
}
.ta-seg__btn.is-active {
  background: rgba(96, 165, 250, 0.2);
  color: var(--ta-text);
}
.ta-seg__btn:active {
  transform: scale(0.98);
}

.ta-divider {
  height: 1px;
  background: var(--ta-border);
  margin: 12px 8px;
}

.ta-chipCol {
  display: grid;
  gap: 8px;
  padding: 6px 8px 2px;
}
.ta-chipRow {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ta-chip {
  border: 1px solid var(--ta-border);
  background: rgba(255, 255, 255, 0.03);
  color: var(--ta-text);
  border-radius: 999px;
  padding: 10px 12px;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
  transition: transform var(--ta-dur) var(--ta-ease), background var(--ta-dur) var(--ta-ease),
    border-color var(--ta-dur) var(--ta-ease);
}
.ta-chip:hover {
  background: rgba(96, 165, 250, 0.12);
  border-color: rgba(96, 165, 250, 0.35);
}
.ta-chip:active {
  transform: scale(0.985);
}
.ta-chip--block {
  text-align: left;
  border-radius: 14px;
}

.ta-sideActions {
  display: flex;
  gap: 10px;
  justify-content: space-between;
  padding: 10px 8px 4px;
}

.ta-footNote {
  margin-top: 14px;
  padding: 10px 12px;
  font-size: 12px;
  color: var(--ta-muted);
}
kbd {
  font-family: var(--ta-mono);
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 8px;
  border: 1px solid var(--ta-border);
  background: rgba(255, 255, 255, 0.04);
}

.ta-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 16px 18px;
  border-bottom: 1px solid var(--ta-border);
  backdrop-filter: blur(10px);
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.2), rgba(0, 0, 0, 0.02));
}
:root[data-ta-theme="light"] .ta-topbar {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.85), rgba(255, 255, 255, 0.6));
}

.ta-title {
  font-weight: 900;
  font-size: 16px;
  letter-spacing: 0.2px;
  font-family: var(--ta-display);
}
.ta-subtitle {
  font-size: 12px;
  color: var(--ta-muted);
  margin-top: 2px;
}

.ta-topbar__right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ta-searchWrap {
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--ta-border);
  background: rgba(255, 255, 255, 0.03);
  padding: 8px 10px;
  border-radius: 999px;
}
.ta-search {
  border: 0;
  outline: none;
  background: transparent;
  color: var(--ta-text);
  width: 220px;
  font-weight: 700;
  font-size: 12px;
}
.ta-search::placeholder {
  color: var(--ta-faint);
}
.ta-searchHint {
  font-family: var(--ta-mono);
  font-size: 11px;
  color: var(--ta-faint);
}

.ta-chat {
  padding: 18px;
  overflow: auto;
}

.ta-msgRow {
  display: flex;
  margin-bottom: 14px;
  animation: taEnter 260ms var(--ta-ease);
}
.ta-msgRow--user {
  justify-content: flex-end;
}
.ta-msgRow--assistant {
  justify-content: flex-start;
}

.ta-bubble {
  width: min(820px, 92%);
  border-radius: var(--ta-radius-lg);
  border: 1px solid var(--ta-border);
  box-shadow: var(--ta-shadow-soft);
  overflow: hidden;
}

.ta-bubble--user {
  background: linear-gradient(135deg, rgba(96, 165, 250, 0.18), rgba(249, 115, 22, 0.12));
}
.ta-bubble--assistant {
  background: var(--ta-surface);
}

.ta-bubble__titleRow {
  padding: 12px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border-bottom: 1px solid var(--ta-border);
  background: rgba(255, 255, 255, 0.03);
}
.ta-bubble__title {
  font-weight: 900;
  font-size: 13px;
  font-family: var(--ta-display);
}
.ta-bubble__meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.ta-richText {
  padding: 12px 14px;
  font-size: 13px;
  line-height: 1.65;
  color: var(--ta-text);
}
.ta-richText p {
  margin: 8px 0;
  color: var(--ta-text);
}
.ta-richText strong {
  color: var(--ta-text);
}
.ta-quote {
  margin: 10px 0;
  padding: 10px 12px;
  border-left: 3px solid var(--ta-warn);
  background: rgba(251, 191, 36, 0.1);
  border-radius: 0 12px 12px 0;
  color: var(--ta-text);
  font-style: italic;
}

.ta-sectionLabel {
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.18px;
  color: var(--ta-muted);
  margin-bottom: 10px;
  text-transform: uppercase;
}

.ta-sources {
  padding: 12px 14px;
}
.ta-sources ul {
  list-style: none;
  margin: 0;
  padding: 0;
}
.ta-sourceItem {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 8px 0;
  border-top: 1px dashed rgba(255, 255, 255, 0.08);
}
:root[data-ta-theme="light"] .ta-sourceItem {
  border-top: 1px dashed rgba(0, 0, 0, 0.1);
}
.ta-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  margin-top: 6px;
  background: var(--ta-accent);
}

.ta-quick {
  padding: 12px 14px 14px;
}

.ta-bubble__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border-top: 1px solid var(--ta-border);
  background: rgba(255, 255, 255, 0.02);
}
.ta-time {
  font-size: 11px;
  color: var(--ta-muted);
  font-family: var(--ta-mono);
}
.ta-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.ta-btn {
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 10px 14px;
  cursor: pointer;
  font-weight: 900;
  font-size: 12px;
  transition: transform var(--ta-dur) var(--ta-ease), box-shadow var(--ta-dur) var(--ta-ease),
    background var(--ta-dur) var(--ta-ease), border-color var(--ta-dur) var(--ta-ease);
}
.ta-btn--md {
  padding: 10px 14px;
}
.ta-btn--primary {
  background: linear-gradient(135deg, var(--ta-accent), var(--ta-accent2));
  color: white;
  box-shadow: 0 10px 22px rgba(96, 165, 250, 0.2);
}
.ta-btn--primary:hover {
  box-shadow: 0 14px 28px rgba(96, 165, 250, 0.24);
}
.ta-btn--ghost {
  background: rgba(255, 255, 255, 0.03);
  color: var(--ta-text);
  border-color: var(--ta-border);
}
.ta-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.ta-btn:active {
  transform: scale(0.985);
}

.ta-iconBtn {
  border: 1px solid var(--ta-border);
  background: rgba(255, 255, 255, 0.03);
  color: var(--ta-text);
  width: 34px;
  height: 34px;
  border-radius: 999px;
  cursor: pointer;
  display: grid;
  place-items: center;
  transition: transform var(--ta-dur) var(--ta-ease), background var(--ta-dur) var(--ta-ease),
    border-color var(--ta-dur) var(--ta-ease);
}
.ta-iconBtn:hover {
  background: rgba(96, 165, 250, 0.12);
  border-color: rgba(96, 165, 250, 0.35);
}
.ta-iconBtn:active {
  transform: scale(0.98);
}

.ta-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 900;
  border: 1px solid var(--ta-border);
  background: rgba(255, 255, 255, 0.03);
}
.ta-badge--success {
  border-color: rgba(52, 211, 153, 0.35);
  background: rgba(52, 211, 153, 0.12);
}
.ta-badge--danger {
  border-color: rgba(251, 113, 133, 0.35);
  background: rgba(251, 113, 133, 0.12);
}
.ta-badge--info {
  border-color: rgba(96, 165, 250, 0.35);
  background: rgba(96, 165, 250, 0.12);
}
.ta-badge--neutral {
  border-color: var(--ta-border);
  background: rgba(255, 255, 255, 0.03);
}

.ta-composer {
  border-top: 1px solid var(--ta-border);
  padding: 14px 18px 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.02), transparent 70%);
}
.ta-composer__box {
  border: 1px solid var(--ta-border);
  border-radius: var(--ta-radius-lg);
  background: rgba(255, 255, 255, 0.03);
  padding: 10px 10px 10px;
}
.ta-input {
  width: 100%;
  resize: none;
  border: 0;
  outline: none;
  background: transparent;
  color: var(--ta-text);
  font-size: 13px;
  font-weight: 650;
  line-height: 1.55;
  padding: 10px 10px 6px;
}
.ta-input::placeholder {
  color: var(--ta-faint);
  font-weight: 600;
}
.ta-composer__actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  padding: 6px 8px 6px;
}
.ta-composer__hint {
  margin-top: 10px;
  font-size: 12px;
  color: var(--ta-muted);
}

.ta-skelLine {
  height: 10px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.06), rgba(255, 255, 255, 0.14), rgba(255, 255, 255, 0.06));
  background-size: 240% 100%;
  animation: taShimmer 1.4s var(--ta-ease) infinite;
  margin: 10px 0;
}
@keyframes taShimmer {
  0% {
    background-position: 0% 0%;
  }
  100% {
    background-position: 220% 0%;
  }
}

.ta-toastHost {
  position: fixed;
  right: 16px;
  top: 16px;
  z-index: 9999;
  display: grid;
  gap: 10px;
  width: min(380px, calc(100vw - 32px));
}
.ta-toast {
  position: relative;
  padding: 12px 12px 12px;
  border-radius: 16px;
  border: 1px solid var(--ta-border);
  background: rgba(0, 0, 0, 0.28);
  backdrop-filter: blur(12px);
  box-shadow: var(--ta-shadow);
  display: grid;
  gap: 4px;
}
:root[data-ta-theme="light"] .ta-toast {
  background: rgba(255, 255, 255, 0.85);
}
.ta-toast__title {
  font-weight: 950;
  font-size: 12px;
  font-family: var(--ta-display);
}
.ta-toast__detail {
  font-size: 12px;
  color: var(--ta-muted);
}
.ta-toast--success {
  border-color: rgba(52, 211, 153, 0.35);
}
.ta-toast--danger {
  border-color: rgba(251, 113, 133, 0.35);
}
.ta-toast--neutral {
  border-color: var(--ta-border);
}

.ta-toast .ta-iconBtn {
  position: absolute;
  right: 10px;
  top: 10px;
  width: 30px;
  height: 30px;
}

.ta-pulse {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--ta-accent);
  box-shadow: 0 0 0 0 rgba(96, 165, 250, 0.35);
  animation: taPulse 1.8s var(--ta-ease) infinite;
}
.ta-pulse--off {
  display: none;
}
@keyframes taPulse {
  0% {
    box-shadow: 0 0 0 0 rgba(96, 165, 250, 0.38);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(96, 165, 250, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(96, 165, 250, 0);
  }
}

@keyframes taEnter {
  0% {
    opacity: 0;
    transform: translateY(6px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 720px) {
  .ta-topbar {
    flex-direction: column;
    align-items: flex-start;
  }
  .ta-topbar__right {
    width: 100%;
    flex-wrap: wrap;
  }
  .ta-searchWrap {
    width: 100%;
    justify-content: space-between;
  }
  .ta-search {
    width: 100%;
  }
  .ta-composer__actions {
    flex-direction: column;
    align-items: stretch;
  }
  .ta-btn {
    width: 100%;
  }
  .ta-bubble {
    width: 100%;
  }
}
`;
