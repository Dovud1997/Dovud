import { useCallback, useEffect, useState, type FormEvent } from "react";
import { api, login, wsSceneUrl, type Agent, type LogItem } from "./api";
import { AgentForm } from "./components/AgentForm";
import { PixelScene } from "./components/PixelScene";
import "./App.css";

export default function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem("token"));
  const [email, setEmail] = useState("admin@example.com");
  const [password, setPassword] = useState("admin123");
  const [agents, setAgents] = useState<Agent[]>([]);
  const [logs, setLogs] = useState<LogItem[]>([]);
  const [authError, setAuthError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    const [a, l] = await Promise.all([api.agents(), api.logs()]);
    setAgents(a);
    setLogs(l);
  }, []);

  useEffect(() => {
    if (!token) return;
    refresh().catch(() => {
      localStorage.removeItem("token");
      setToken(null);
    });
  }, [token, refresh]);

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
          setLogs((prev) => [
            {
              id: crypto.randomUUID(),
              agent_id: data.agent_id,
              level: data.status === "error" ? "error" : "info",
              message: data.status_message || data.status,
              meta: null,
              created_at: new Date().toISOString(),
            },
            ...prev,
          ].slice(0, 150));
        }
        if (data.type === "agent_event" || data.type === "job_update") {
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
  }, [token]);

  async function onLogin(e: FormEvent) {
    e.preventDefault();
    setAuthError(null);
    try {
      const t = await login(email, password);
      setToken(t);
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : "Login failed");
    }
  }

  if (!token) {
    return (
      <div className="login-shell">
        <form className="login-card" onSubmit={onLogin}>
          <p className="brand">DOVUD</p>
          <h1>Agent Ops</h1>
          <p className="lede">Единая точка управления AI-агентами каналов.</p>
          <label>
            Email
            <input value={email} onChange={(e) => setEmail(e.target.value)} />
          </label>
          <label>
            Password
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
          </label>
          {authError && <div className="banner bad">{authError}</div>}
          <button type="submit">Войти</button>
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
        <button
          className="ghost"
          onClick={() => {
            localStorage.removeItem("token");
            setToken(null);
          }}
        >
          Выйти
        </button>
      </header>

      <section className="scene-section">
        <div className="scene-wrap">
          <PixelScene agents={agents} />
        </div>
        <aside className="side-log panel">
          <header className="panel-head">
            <h2>Статусы / лог</h2>
            <p>Realtime из WebSocket + журнал действий.</p>
          </header>
          <ul className="status-list">
            {agents.map((a) => (
              <li key={a.id}>
                <span className={`dot ${a.status}`} />
                <div>
                  <strong>{a.name}</strong>
                  <em>{a.platform} · {a.status}</em>
                  <small>{a.status_message || "—"}</small>
                </div>
                <div className="mini-actions">
                  {!a.is_active && (
                    <button className="ghost" onClick={() => api.activateAgent(a.id).then(refresh)}>On</button>
                  )}
                  <button
                    className="ghost"
                    onClick={() => api.command(a.id, "test_connection").then(refresh)}
                  >
                    Ping
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

      <section className="bottom-grid">
        <AgentForm onCreated={refresh} />
        <div className="panel">
          <header className="panel-head">
            <h2>Агенты</h2>
            <p>Активные модули и быстрые команды MVP.</p>
          </header>
          <ul className="agent-table">
            {agents.map((a) => (
              <li key={a.id}>
                <div>
                  <strong>{a.name}</strong>
                  <span>{a.id.slice(0, 8)}… · AI: {a.ai_mode}</span>
                </div>
                <button
                  className="ghost"
                  onClick={() =>
                    api.command(a.id, "publish_post", { text: "MVP test post from admin" }).then(refresh)
                  }
                >
                  Publish test
                </button>
              </li>
            ))}
          </ul>
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
