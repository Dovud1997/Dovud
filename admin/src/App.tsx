import { useCallback, useEffect, useState, type FormEvent } from "react";
import {
  api,
  login,
  register,
  setOrgId,
  wsSceneUrl,
  type Agent,
  type LogItem,
  type Org,
} from "./api";
import { AgentForm } from "./components/AgentForm";
import { AgentSettings } from "./components/AgentSettings";
import { PixelScene } from "./components/PixelScene";
import "./App.css";

export default function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem("token"));
  const [orgs, setOrgs] = useState<Org[]>(() => {
    try {
      return JSON.parse(localStorage.getItem("orgs") || "[]") as Org[];
    } catch {
      return [];
    }
  });
  const [currentOrg, setCurrentOrg] = useState(localStorage.getItem("org_id") || "");
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("admin@example.com");
  const [password, setPassword] = useState("admin123");
  const [orgName, setOrgName] = useState("My Studio");
  const [agents, setAgents] = useState<Agent[]>([]);
  const [logs, setLogs] = useState<LogItem[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);
  const [notifyChat, setNotifyChat] = useState("");

  const refresh = useCallback(async () => {
    const [a, l] = await Promise.all([api.agents(), api.logs()]);
    setAgents(a);
    setLogs(l);
    if (selectedId && !a.find((x) => x.id === selectedId)) setSelectedId(a[0]?.id ?? null);
    if (!selectedId && a[0]) setSelectedId(a[0].id);
  }, [selectedId]);

  useEffect(() => {
    if (!token) return;
    refresh().catch(() => {
      localStorage.removeItem("token");
      setToken(null);
    });
  }, [token, currentOrg, refresh]);

  useEffect(() => {
    if (!token) return;
    let ws: WebSocket | null = null;
    let closed = false;
    let retry: number | undefined;

    const connect = () => {
      ws = new WebSocket(wsSceneUrl());
      ws.onmessage = (ev) => {
        const data = JSON.parse(ev.data);
        if (data.type === "snapshot" && Array.isArray(data.agents)) {
          setAgents((prev) => mergeScene(prev, data.agents));
        }
        if (data.type === "agent_status") {
          setAgents((prev) =>
            prev.map((a) =>
              a.id === data.agent_id
                ? {
                    ...a,
                    status: data.status,
                    status_message: data.status_message,
                    pos_x: data.pos_x ?? a.pos_x,
                    pos_y: data.pos_y ?? a.pos_y,
                    zone: data.zone ?? a.zone,
                  }
                : a,
            ),
          );
        }
        if (data.type === "agent_event" || data.type === "job_update" || data.type === "notification") {
          api.logs().then(setLogs).catch(() => undefined);
        }
      };
      ws.onclose = () => {
        if (!closed) retry = window.setTimeout(connect, 2000);
      };
    };
    connect();
    return () => {
      closed = true;
      if (retry) window.clearTimeout(retry);
      ws?.close();
    };
  }, [token, currentOrg]);

  async function onAuth(e: FormEvent) {
    e.preventDefault();
    setAuthError(null);
    try {
      const data =
        mode === "login"
          ? await login(email, password)
          : await register({ email, password, org_name: orgName });
      setToken(data.access_token);
      setOrgs(data.orgs);
      setCurrentOrg(data.orgs[0]?.id || "");
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : "Auth failed");
    }
  }

  const selected = agents.find((a) => a.id === selectedId) || null;

  if (!token) {
    return (
      <div className="login-shell">
        <form className="login-card" onSubmit={onAuth}>
          <p className="brand">DOVUD</p>
          <h1>Agent Ops</h1>
          <p className="lede">Multi-tenant платформа AI-агентов каналов.</p>
          <div className="row-actions">
            <button type="button" className={mode === "login" ? "" : "ghost"} onClick={() => setMode("login")}>Вход</button>
            <button type="button" className={mode === "register" ? "" : "ghost"} onClick={() => setMode("register")}>Регистрация</button>
          </div>
          <label>
            Email
            <input value={email} onChange={(e) => setEmail(e.target.value)} />
          </label>
          <label>
            Password
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
          </label>
          {mode === "register" && (
            <label>
              Организация
              <input value={orgName} onChange={(e) => setOrgName(e.target.value)} />
            </label>
          )}
          {authError && <div className="banner bad">{authError}</div>}
          <button type="submit">{mode === "login" ? "Войти" : "Создать аккаунт"}</button>
        </form>
      </div>
    );
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="brand">DOVUD</p>
          <h1>Agent World</h1>
        </div>
        <div className="top-controls">
          <label className="inline">
            Org
            <select
              value={currentOrg}
              onChange={(e) => {
                setOrgId(e.target.value);
                setCurrentOrg(e.target.value);
              }}
            >
              {orgs.map((o) => (
                <option key={o.id} value={o.id}>{o.name}</option>
              ))}
            </select>
          </label>
          <button
            className="ghost"
            onClick={async () => {
              const name = prompt("Название новой организации");
              if (!name) return;
              const org = await api.createOrg(name);
              const list = await api.orgs();
              setOrgs(list);
              localStorage.setItem("orgs", JSON.stringify(list));
              setOrgId(org.id);
              setCurrentOrg(org.id);
            }}
          >
            + Org
          </button>
          <button
            className="ghost"
            onClick={() => {
              localStorage.clear();
              setToken(null);
            }}
          >
            Выйти
          </button>
        </div>
      </header>

      <section className="scene-section">
        <div className="scene-wrap">
          <PixelScene agents={agents} />
        </div>
        <aside className="side-log panel">
          <header className="panel-head">
            <h2>Статусы / лог</h2>
            <p>Realtime WebSocket + журнал.</p>
          </header>
          <ul className="status-list">
            {agents.map((a) => (
              <li key={a.id} className={selectedId === a.id ? "active" : ""}>
                <span className={`dot ${a.status}`} />
                <div onClick={() => setSelectedId(a.id)} style={{ cursor: "pointer" }}>
                  <strong>{a.name}</strong>
                  <em>{a.platform} · {a.status}</em>
                  <small>{a.status_message || "—"}</small>
                </div>
                <div className="mini-actions">
                  {!a.is_active && (
                    <button className="ghost" onClick={() => api.activateAgent(a.id).then(refresh)}>On</button>
                  )}
                  <button className="ghost" onClick={() => api.command(a.id, "test_connection").then(refresh)}>Ping</button>
                  <button
                    className="ghost"
                    onClick={() =>
                      api.simulateEvent(a.id, "message", { text: "Привет! Есть вопрос по прайсу" }).then(refresh)
                    }
                  >
                    Event
                  </button>
                </div>
              </li>
            ))}
            {agents.length === 0 && <li className="empty">Агентов пока нет</li>}
          </ul>
          <ul className="log-list">
            {logs.map((log) => (
              <li key={log.id} className={log.level}>
                <time>{new Date(log.created_at).toLocaleTimeString()}</time>
                <span>{log.message}</span>
              </li>
            ))}
          </ul>
        </aside>
      </section>

      <section className="bottom-grid three">
        <AgentForm onCreated={refresh} />
        <AgentSettings agent={selected} onChanged={refresh} />
        <div className="panel">
          <header className="panel-head">
            <h2>Агенты и уведомления</h2>
            <p>Быстрые команды и Telegram notify chat_id.</p>
          </header>
          <ul className="agent-table">
            {agents.map((a) => (
              <li key={a.id}>
                <div onClick={() => setSelectedId(a.id)} style={{ cursor: "pointer" }}>
                  <strong>{a.name}</strong>
                  <span>{a.platform} · AI: {a.ai_mode}/{a.llm_provider}</span>
                </div>
                <button
                  className="ghost"
                  onClick={() =>
                    api.command(a.id, "publish_post", { text: "MVP test post from admin" }).then(refresh)
                  }
                >
                  Publish
                </button>
              </li>
            ))}
          </ul>
          <div className="stack-form" style={{ marginTop: "1rem" }}>
            <input
              value={notifyChat}
              onChange={(e) => setNotifyChat(e.target.value)}
              placeholder="Telegram chat_id для уведомлений"
            />
            <button
              type="button"
              className="ghost"
              onClick={async () => {
                if (!notifyChat.trim()) return;
                await api.createNotificationTarget("telegram", notifyChat.trim());
                setNotifyChat("");
                alert("Notification target сохранён");
              }}
            >
              Сохранить notify target
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}

function mergeScene(prev: Agent[], scene: Agent[]): Agent[] {
  if (prev.length === 0) return scene as Agent[];
  const map = new Map(scene.map((a) => [a.id, a]));
  return prev.map((a) => {
    const s = map.get(a.id);
    return s
      ? {
          ...a,
          status: s.status,
          status_message: s.status_message,
          pos_x: s.pos_x,
          pos_y: s.pos_y,
          zone: s.zone,
          is_active: s.is_active,
        }
      : a;
  });
}
