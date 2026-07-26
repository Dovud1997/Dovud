# Схема данных

## ER (логическая)

```
User *──* Organization (через Membership)
Organization 1──* Agent
Organization 1──* NotificationTarget
Agent 1──* AgentSecret (encrypted)
Agent 1──* ReplyTemplate
Agent 1──* StyleExample
Agent 1──* AgentEvent
Agent 1──* AgentLog
Agent 1──* CommandJob
```

## Multi-tenant
Все агенты и события scoped по `org_id`.  
Клиенты передают `X-Org-Id` (админка / control-бот).  
Роли membership: `owner` | `admin` | `member`.

## Ключевые таблицы

### organizations / memberships / users
Тенанты и доступ пользователей.

### agents
`org_id`, `platform`, `status`, `ai_mode` (`off|template|llm`), `llm_provider` (`openai|anthropic`),
`system_prompt`, координаты сцены (`zone`, `pos_x`, `pos_y`), `is_active`.

### agent_secrets
Шифрованные credentials (`bot_token`, `access_token`, …) — Fernet.

### reply_templates / style_examples
Шаблоны автоответов и few-shot примеры стиля.

### agent_events / agent_logs / command_jobs
События платформ, журнал, асинхронные команды.

### notification_targets
Куда слать уведомления (`telegram` chat_id или `webhook` URL).

## Статусы на пиксельной сцене
- `online` — idle
- `busy` — публикует / отвечает
- `error` — alert
- `offline` / `draft` — серый idle
- `connecting` — переходный
