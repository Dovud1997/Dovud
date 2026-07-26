import { useEffect, useMemo, useState, type FormEvent } from "react";
import { api, type Platform } from "../api";

type Props = {
  onCreated: () => void;
};

export function AgentForm({ onCreated }: Props) {
  const [platforms, setPlatforms] = useState<Platform[]>([]);
  const [platform, setPlatform] = useState("telegram");
  const [name, setName] = useState("");
  const [credentials, setCredentials] = useState<Record<string, string>>({});
  const [aiMode, setAiMode] = useState("off");
  const [llmProvider, setLlmProvider] = useState("openai");
  const [systemPrompt, setSystemPrompt] = useState("");
  const [testMsg, setTestMsg] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.platforms().then((list) => {
      setPlatforms(list);
      if (list[0]) setPlatform(list[0].platform);
    }).catch((e: Error) => setError(e.message));
  }, []);

  const selected = useMemo(
    () => platforms.find((p) => p.platform === platform),
    [platforms, platform],
  );

  async function onTest() {
    if (!selected) return;
    setBusy(true);
    setTestMsg(null);
    setError(null);
    try {
      const res = await api.testConnection(platform, credentials);
      setTestMsg(res.ok ? `OK: ${res.message}` : `FAIL: ${res.message}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Test failed");
    } finally {
      setBusy(false);
    }
  }

  async function submit(activate: boolean) {
    if (!selected) return;
    setBusy(true);
    setError(null);
    try {
      await api.createAgent({
        name: name || selected.title,
        platform,
        credentials,
        ai_mode: aiMode,
        llm_provider: llmProvider,
        system_prompt: systemPrompt || null,
        activate,
      });
      setName("");
      setCredentials({});
      setSystemPrompt("");
      setTestMsg(null);
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      className="panel form"
      onSubmit={(e: FormEvent) => {
        e.preventDefault();
        void submit(false);
      }}
    >
      <header className="panel-head">
        <h2>Новый агент</h2>
        <p>Платформа → данные → тест → активация. Для локальных тестов: `demo:local`.</p>
      </header>

      <label>
        Платформа
        <select
          value={platform}
          onChange={(e) => {
            setPlatform(e.target.value);
            setCredentials({});
          }}
        >
          {platforms.map((p) => (
            <option key={p.platform} value={p.platform}>{p.title}</option>
          ))}
        </select>
      </label>

      {selected && <p className="hint">{selected.description}</p>}

      <label>
        Имя агента
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="TG Main" />
      </label>

      {selected?.fields.map((field) => (
        <label key={field.key}>
          {field.label}
          <input
            type={field.type === "password" ? "password" : "text"}
            value={credentials[field.key] || ""}
            required={field.required}
            placeholder={field.placeholder}
            onChange={(e) => setCredentials((c) => ({ ...c, [field.key]: e.target.value }))}
          />
          {field.help && <span className="hint">{field.help}</span>}
        </label>
      ))}

      <div className="two-col">
        <label>
          Режим автоответа
          <select value={aiMode} onChange={(e) => setAiMode(e.target.value)}>
            <option value="off">Выкл</option>
            <option value="template">Шаблон</option>
            <option value="llm">LLM (стиль)</option>
          </select>
        </label>
        <label>
          LLM-провайдер
          <select value={llmProvider} onChange={(e) => setLlmProvider(e.target.value)}>
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
          </select>
        </label>
      </div>

      <label>
        AI-инструкция (промпт)
        <textarea
          value={systemPrompt}
          onChange={(e) => setSystemPrompt(e.target.value)}
          placeholder="Отвечай коротко, дружелюбно, как владелец канала…"
          rows={3}
        />
      </label>

      {testMsg && <div className={`banner ${testMsg.startsWith("OK") ? "ok" : "bad"}`}>{testMsg}</div>}
      {error && <div className="banner bad">{error}</div>}

      <div className="row-actions">
        <button type="button" className="ghost" disabled={busy} onClick={onTest}>Тест соединения</button>
        <button type="submit" className="ghost" disabled={busy}>Сохранить черновик</button>
        <button type="button" disabled={busy} onClick={() => void submit(true)}>Активировать</button>
      </div>
    </form>
  );
}
