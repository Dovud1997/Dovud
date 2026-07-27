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

export type Org = { id: string; name: string; slug: string; role: string };

export type Agent = {
  id: string;
  org_id: string;
  name: string;
  platform: string;
  status: string;
  status_message: string | null;
  ai_mode: string;
  llm_provider: string;
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

export type Template = {
  id: string;
  agent_id: string;
  name: string;
  body: string;
  trigger_pattern: string | null;
  is_default: boolean;
  created_at: string;
};

export type StyleExample = {
  id: string;
  agent_id: string;
  user_message: string;
  assistant_reply: string;
  created_at: string;
};

function orgId(): string | null {
  return localStorage.getItem("org_id");
}

function authHeaders(): HeadersInit {
  const token = localStorage.getItem("token");
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  const oid = orgId();
  if (oid) headers["X-Org-Id"] = oid;
  return headers;
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
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export type LoginResult = {
  access_token: string;
  user_id: string;
  email: string;
  orgs: Org[];
};

function persistAuth(data: LoginResult) {
  localStorage.setItem("token", data.access_token);
  if (data.orgs[0]) localStorage.setItem("org_id", data.orgs[0].id);
  localStorage.setItem("orgs", JSON.stringify(data.orgs));
}

export async function login(email: string, password: string): Promise<LoginResult> {
  const data = await request<LoginResult>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  persistAuth(data);
  return data;
}

export async function register(payload: {
  email: string;
  password: string;
  display_name?: string;
  org_name: string;
}): Promise<LoginResult> {
  const data = await request<LoginResult>("/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  persistAuth(data);
  return data;
}

export const api = {
  platforms: () => request<Platform[]>("/platforms"),
  orgs: () => request<Org[]>("/me/orgs"),
  createOrg: (name: string) =>
    request<Org>("/orgs", { method: "POST", body: JSON.stringify({ name }) }),
  agents: () => request<Agent[]>("/agents"),
  createAgent: (body: Record<string, unknown>) =>
    request<Agent>("/agents", { method: "POST", body: JSON.stringify(body) }),
  updateAgent: (id: string, body: Record<string, unknown>) =>
    request<Agent>(`/agents/${id}`, { method: "PATCH", body: JSON.stringify(body) }),
  activateAgent: (id: string) => request<Agent>(`/agents/${id}/activate`, { method: "POST" }),
  testConnection: (platform: string, credentials: Record<string, string>) =>
    request<{ ok: boolean; message: string }>("/agents/test-connection", {
      method: "POST",
      body: JSON.stringify({ platform, credentials }),
    }),
  logs: () => request<LogItem[]>("/logs"),
  events: () => request<LogItem[]>("/events").catch(() => []),
  scene: () => request<Agent[]>("/scene"),
  command: (
    agent_id: string,
    action: string,
    payload: Record<string, unknown> = {},
  ) =>
    request("/commands", { method: "POST", body: JSON.stringify({ agent_id, action, payload }) }),
  uploadMedia: async (file: File) => {
    const token = localStorage.getItem("token");
    const oid = orgId();
    const headers: Record<string, string> = {};
    if (token) headers.Authorization = `Bearer ${token}`;
    if (oid) headers["X-Org-Id"] = oid;
    const body = new FormData();
    body.append("file", file);
    const res = await fetch(`${API_BASE}/media/upload`, { method: "POST", headers, body });
    if (!res.ok) throw new Error(await res.text());
    return res.json() as Promise<{
      id: string;
      filename: string;
      content_type: string;
      public_url: string;
      media_kind: string;
    }>;
  },
  simulateEvent: (agent_id: string, type: string, payload: Record<string, unknown> = {}) =>
    request(`/agents/${agent_id}/simulate-event`, {
      method: "POST",
      body: JSON.stringify({ type, payload, auto_reply: true }),
    }),
  templates: (agent_id: string) => request<Template[]>(`/agents/${agent_id}/templates`),
  createTemplate: (agent_id: string, body: Record<string, unknown>) =>
    request<Template>(`/agents/${agent_id}/templates`, { method: "POST", body: JSON.stringify(body) }),
  deleteTemplate: (agent_id: string, template_id: string) =>
    request(`/agents/${agent_id}/templates/${template_id}`, { method: "DELETE" }),
  styleExamples: (agent_id: string) => request<StyleExample[]>(`/agents/${agent_id}/style-examples`),
  createStyleExample: (agent_id: string, body: Record<string, unknown>) =>
    request<StyleExample>(`/agents/${agent_id}/style-examples`, {
      method: "POST",
      body: JSON.stringify(body),
    }),
  deleteStyleExample: (agent_id: string, example_id: string) =>
    request(`/agents/${agent_id}/style-examples/${example_id}`, { method: "DELETE" }),
  previewReply: (agent_id: string, message: string) =>
    request<{ reply: string | null; mode: string }>(`/agents/${agent_id}/auto-reply/preview`, {
      method: "POST",
      body: JSON.stringify({ message }),
    }),
  notificationTargets: () => request<Array<{ id: string; channel: string; address: string }>>("/notifications/targets"),
  createNotificationTarget: (channel: string, address: string) =>
    request("/notifications/targets", {
      method: "POST",
      body: JSON.stringify({ channel, address, is_active: true }),
    }),
  setTelegramWebhook: (agent_id: string) =>
    request<{ ok: boolean; url: string }>(`/agents/${agent_id}/telegram/set-webhook`, { method: "POST" }),
  listenerStatus: () =>
    request<{ enabled: boolean; running: boolean; active_pollers: string[] }>("/telegram/listener-status"),
};

export function setOrgId(id: string) {
  localStorage.setItem("org_id", id);
}

export function wsSceneUrl(): string {
  const token = localStorage.getItem("token") || "";
  const oid = orgId() || "";
  const base = API_BASE.replace(/^http/, "ws");
  return `${base}/ws/scene?token=${encodeURIComponent(token)}&org_id=${encodeURIComponent(oid)}`;
}
