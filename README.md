# DOVUD — платформа AI-агентов

Единая система управления AI-агентами для каналов коммуникации (Telegram, Instagram, YouTube…).  
Агенты подключаются как **плагины**, конфигурируются через админ-панель, управляются через API / Telegram control-бота, визуализируются на пиксельной сцене в реальном времени.

## Документация
- [Стек технологий](docs/STACK.md)
- [Архитектура](docs/ARCHITECTURE.md)
- [Схема данных](docs/DATA_SCHEMA.md)

## MVP (этот релиз)
- Backend FastAPI: registry плагинов, шифрование секретов, очередь задач, REST + WebSocket
- Плагин **Telegram** (`connect` / `execute_action` / `get_status`)
- Админ-панель: логин, форма добавления агента (тест соединения → активация), пиксельная сцена + лог статусов
- Control Telegram-бот (stub/worker): команды и медиа → `POST /api/commands`
- LLM-адаптер (OpenAI-compatible) + шаблоны автоответов

## Быстрый старт

### Backend
```bash
cd backend
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
uvicorn app.main:app --reload --port 8000
```

По умолчанию:
- API: http://127.0.0.1:8000
- Docs: http://127.0.0.1:8000/docs
- Логин: `admin@example.com` / `admin123`

### Admin
```bash
cd admin
npm install
npm run dev
```
Откройте http://127.0.0.1:5173

### Control-бот (опционально)
```bash
export CONTROL_BOT_TOKEN=...
export PLATFORM_API_BASE=http://127.0.0.1:8000/api
cd backend && python -m bot.control_bot
```

## Как добавить новый платформенный агент
1. Создайте `backend/plugins/<platform>/plugin.py` с классом `Plugin(BaseAgentPlugin)`
2. Опишите `manifest` (поля формы, actions, zone)
3. Перезапустите API — registry подхватит плагин **без изменений ядра**

## Структура
```
backend/app/          # ядро API, модели, очередь, LLM
backend/plugins/      # платформенные агенты
backend/bot/          # control Telegram-бот
admin/                # веб-панель + pixel scene
mobile/               # заготовка под этап 2
docs/                 # архитектура и схема
```
