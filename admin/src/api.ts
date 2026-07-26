const API_BASE = import.meta.env.VITE_API_BASE || "http://127.0.0.1:8000/api";

export type PlatformField = {
  key: string;
  label: string;
  type: string;
  required: boolean;
  secret: boolean;
  help: string;
  placeholder: string;
};

export type Platform = {
  platform: string;
  title: string;
  description: string;
  zone: string;
  fields: PlatformField[];
  actions: string[];
};

export type Agent = {
  id: string;
  name: string;
  platform: string;
  status: string;
  status_message: string | null;
  ai_mode: string;
  system_prompt: string | null;
  zone: string;
  pos_x: number;
  pos_y: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  has_secrets: boolean;
};

export type LogItem = {
  id: string;
  agent_id: string | null;
  level: string;
  message: string;
  meta: Record<string, unknown> | null;
  created_at: string;
};

function authHeaders(): HeadersInit {
  const token = localStorage.getItem("token");
  return token ? { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } : { "Content-Type": "application/json" };
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { ...authHeaders(), ...(init?.headers || {}) },
  });
  if (!res.ok) {
    const detail = await res.text();
    throw new Error(detail || res.statusText);
  }
  return res.json() as Promise<T>;
}

export async function login(email: string, password: string): Promise<string> {
  const data = await request<{ access_token: string }>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  localStorage.setItem("token", data.access_token);
  return data.access_token;
}

export const api = {
  platforms: () => request<Platform[]>("/platforms"),
  agents: () => request<Agent[]>("/agents"),
  createAgent: (body: Record<string, unknown>) =>
    request<Agent>("/agents", { method: "POST", body: JSON.stringify(body) }),
  activateAgent: (id: string) => request<Agent>(`/agents/${id}/activate`, { method: "POST" }),
  testConnection: (platform: string, credentials: Record<string, string>) =>
    request<{ ok: boolean; message: string }>("/agents/test-connection", {
      method: "POST",
      body: JSON.stringify({ platform, credentials }),
    }),
  logs: () => request<LogItem[]>("/logs"),
  scene: () => request<Agent[]>("/scene"),
  command: (agent_id: string, action: string, payload: Record<string, unknown> = {}) =>
    request("/commands", { method: "POST", body: JSON.stringify({ agent_id, action, payload }) }),
};

export function wsSceneUrl(): string {
  const token = localStorage.getItem("token") || "";
  const base = API_BASE.replace(/^http/, "ws");
  return `${base}/ws/scene?token=${encodeURIComponent(token)}`;
}
