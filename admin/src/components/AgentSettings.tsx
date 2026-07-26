import { useEffect, useState } from "react";
import { api, type Agent, type StyleExample, type Template } from "../api";

type Props = {
  agent: Agent | null;
  onChanged: () => void;
};

export function AgentSettings({ agent, onChanged }: Props) {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [examples, setExamples] = useState<StyleExample[]>([]);
  const [tplName, setTplName] = useState("Default");
  const [tplBody, setTplBody] = useState("Спасибо за сообщение! Мы скоро ответим.");
  const [tplTrigger, setTplTrigger] = useState("");
  const [exUser, setExUser] = useState("Сколько стоит?");
  const [exReply, setExReply] = useState("Пиши в личку — скину актуальный прайс 🙂");
  const [previewIn, setPreviewIn] = useState("Привет!");
  const [previewOut, setPreviewOut] = useState<string | null>(null);
  const [aiMode, setAiMode] = useState("off");
  const [llmProvider, setLlmProvider] = useState("openai");
  const [prompt, setPrompt] = useState("");
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!agent) return;
    setAiMode(agent.ai_mode);
    setLlmProvider(agent.llm_provider || "openai");
    setPrompt(agent.system_prompt || "");
    void Promise.all([api.templates(agent.id), api.styleExamples(agent.id)]).then(([t, e]) => {
      setTemplates(t);
      setExamples(e);
    });
  }, [agent]);

  if (!agent) {
    return (
      <div className="panel">
        <header className="panel-head">
          <h2>Настройки агента</h2>
          <p>Выберите агента в списке справа.</p>
        </header>
      </div>
    );
  }

  return (
    <div className="panel">
      <header className="panel-head">
        <h2>{agent.name}</h2>
        <p>Шаблоны, few-shot примеры и превью автоответа.</p>
      </header>

      <div className="two-col">
        <label>
          Режим
          <select value={aiMode} onChange={(e) => setAiMode(e.target.value)}>
            <option value="off">Выкл</option>
            <option value="template">Шаблон</option>
            <option value="llm">LLM</option>
          </select>
        </label>
        <label>
          Провайдер
          <select value={llmProvider} onChange={(e) => setLlmProvider(e.target.value)}>
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
          </select>
        </label>
      </div>
      <label>
        Промпт
        <textarea rows={2} value={prompt} onChange={(e) => setPrompt(e.target.value)} />
      </label>
      <button
        type="button"
        className="ghost"
        onClick={async () => {
          await api.updateAgent(agent.id, {
            ai_mode: aiMode,
            llm_provider: llmProvider,
            system_prompt: prompt,
          });
          setMsg("Сохранено");
          onChanged();
        }}
      >
        Сохранить AI-настройки
      </button>

      <h3>Шаблоны</h3>
      <ul className="compact-list">
        {templates.map((t) => (
          <li key={t.id}>
            <div>
              <strong>{t.name}</strong>
              <span>{t.trigger_pattern || "default"} · {t.body.slice(0, 60)}</span>
            </div>
            <button className="ghost" onClick={() => api.deleteTemplate(agent.id, t.id).then(() => api.templates(agent.id).then(setTemplates))}>×</button>
          </li>
        ))}
      </ul>
      <div className="stack-form">
        <input value={tplName} onChange={(e) => setTplName(e.target.value)} placeholder="Имя" />
        <input value={tplTrigger} onChange={(e) => setTplTrigger(e.target.value)} placeholder="trigger regex (опц.)" />
        <textarea value={tplBody} onChange={(e) => setTplBody(e.target.value)} rows={2} />
        <button
          type="button"
          className="ghost"
          onClick={async () => {
            await api.createTemplate(agent.id, {
              name: tplName,
              body: tplBody,
              trigger_pattern: tplTrigger || null,
              is_default: !tplTrigger,
            });
            setTemplates(await api.templates(agent.id));
          }}
        >
          Добавить шаблон
        </button>
      </div>

      <h3>Примеры стиля</h3>
      <ul className="compact-list">
        {examples.map((ex) => (
          <li key={ex.id}>
            <div>
              <strong>{ex.user_message}</strong>
              <span>→ {ex.assistant_reply}</span>
            </div>
            <button className="ghost" onClick={() => api.deleteStyleExample(agent.id, ex.id).then(() => api.styleExamples(agent.id).then(setExamples))}>×</button>
          </li>
        ))}
      </ul>
      <div className="stack-form">
        <input value={exUser} onChange={(e) => setExUser(e.target.value)} placeholder="Сообщение пользователя" />
        <input value={exReply} onChange={(e) => setExReply(e.target.value)} placeholder="Ответ в вашем стиле" />
        <button
          type="button"
          className="ghost"
          onClick={async () => {
            await api.createStyleExample(agent.id, { user_message: exUser, assistant_reply: exReply });
            setExamples(await api.styleExamples(agent.id));
          }}
        >
          Добавить пример
        </button>
      </div>

      <h3>Превью ответа</h3>
      <div className="stack-form">
        <input value={previewIn} onChange={(e) => setPreviewIn(e.target.value)} />
        <button
          type="button"
          onClick={async () => {
            const res = await api.previewReply(agent.id, previewIn);
            setPreviewOut(res.reply ?? "(нет ответа — режим off или нет шаблонов)");
          }}
        >
          Сгенерировать
        </button>
        {previewOut && <div className="banner ok">{previewOut}</div>}
      </div>
      {msg && <div className="banner ok">{msg}</div>}
    </div>
  );
}
