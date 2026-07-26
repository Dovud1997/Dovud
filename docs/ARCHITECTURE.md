# Архитектура платформы AI-агентов

## Принципы
1. **Ядро не знает о платформах** — Instagram/Telegram/YouTube подключаются как плагины.
2. **Единый контракт агента** — `connect`, `execute_action`, `listen_events`, `get_status`, `disconnect`.
3. **Конфиги через админку** — токены и инструкции в БД (секреты зашифрованы), не в коде.
4. **Одна точка управления** — REST + WebSocket; control Telegram-бот и будущий mobile — клиенты этого API.
5. **Realtime-сцена** — статусы агентов пушатся по WebSocket в пиксельный дашборд.

## Высокоуровневая схема

```
┌─────────────┐   ┌──────────────┐   ┌────────────────┐
│ Admin Web   │   │ Telegram     │   │ Mobile (later) │
│ (pixel UI)  │   │ Control Bot  │   │                │
└──────┬──────┘   └──────┬───────┘   └───────┬────────┘
       │ REST/WS         │ REST              │ REST/WS
       └─────────────────┼───────────────────┘
                         ▼
              ┌────────────────────┐
              │   API Gateway      │
              │   (FastAPI)        │
              └─────────┬──────────┘
         ┌──────────────┼──────────────┐
         ▼              ▼              ▼
   ┌──────────┐  ┌────────────┐  ┌─────────────┐
   │ Config / │  │ Task Queue │  │ Event Hub   │
   │ Secrets  │  │ (ARQ/IP)   │  │ (WS + DB)   │
   └──────────┘  └─────┬──────┘  └─────────────┘
                       ▼
              ┌────────────────────┐
              │  Agent Registry    │
              │  + Plugin Loader   │
              └─────────┬──────────┘
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
   ┌─────────┐    ┌──────────┐    ┌──────────┐
   │Telegram │    │Instagram │    │ YouTube  │
   │ Plugin  │    │ (future) │    │ (future) │
   └─────────┘    └──────────┘    └──────────┘
```

## Плагинная модель
Новая платформа = пакет в `backend/plugins/<name>/` с:
- `manifest.json` — id, название, поля формы подключения
- `plugin.py` — класс, реализующий `BaseAgentPlugin`
- авторегистрация через `AgentRegistry.discover()`

Ядро **не меняется** при добавлении Instagram/YouTube — только новый плагин + запись в registry.

## Контракт агента

```python
class BaseAgentPlugin(Protocol):
    platform: str

    async def connect(self, credentials: dict) -> ConnectResult: ...
    async def execute_action(self, action: str, payload: dict) -> ActionResult: ...
    async def listen_events(self) -> AsyncIterator[AgentEvent]: ...
    async def get_status(self) -> AgentStatus: ...
    async def disconnect(self) -> None: ...
```

## Потоки данных

### Команда из control-бота
1. Пользователь шлёт фото + «поставь в историю Instagram».
2. Бот парсит intent → `POST /api/commands`.
3. API кладёт задачу в очередь.
4. Worker берёт задачу → `registry.get(agent).execute_action(...)`.
5. Результат → событие в Event Hub → уведомление в бот + WS в админку.

### Событие с платформы
1. Плагин получает webhook/poll (лайк, сообщение).
2. Публикует `AgentEvent` в Event Hub.
3. Если включён автоответ — очередь задачи `auto_reply` (template или LLM).
4. Клиенты получают уведомление.

## Структура репозитория

```
/
├── docs/                 # Стек, архитектура, схема данных
├── backend/
│   ├── app/
│   │   ├── api/          # REST + WebSocket
│   │   ├── agents/       # Base + Registry (ядро)
│   │   ├── core/         # config, db, security
│   │   ├── models/       # SQLAlchemy
│   │   ├── services/     # business logic
│   │   ├── workers/      # queue + job runners
│   │   └── llm/          # LLM adapter
│   ├── plugins/          # Платформенные агенты (плагины)
│   │   └── telegram/
│   └── bot/              # Control Telegram-бот
├── admin/                # Vite React админ-панель
├── mobile/               # Placeholder
└── docker-compose.yml    # Прод-скелет
```

## Допущения MVP
- Single-tenant: один владелец (логин/пароль из env).
- SQLite по умолчанию; PostgreSQL через `DATABASE_URL`.
- In-process очередь без Redis.
- Один рабочий плагин: **Telegram**.
- Пиксельная сцена: Canvas 2D, зоны платформ, спрайты со статусом.
- Mobile — только заготовка папки.
