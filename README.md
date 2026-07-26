# DOVUD — платформа AI-агентов

Единая multi-tenant система управления AI-агентами для каналов (Telegram, Instagram, YouTube).  
Агенты — плагины; конфиги через админку; управление через API / Telegram-бот; realtime пиксельная сцена.

## Документация
- [Стек](docs/STACK.md) · [Архитектура](docs/ARCHITECTURE.md) · [Схема данных](docs/DATA_SCHEMA.md)

## Возможности
- Multi-tenant организации + регистрация
- Плагины: Telegram / Instagram / YouTube
- Очередь задач (in-process или Redis/ARQ)
- Автоответы: шаблоны или LLM (OpenAI / Anthropic)
- Уведомления в Telegram / webhook
- Админка: форма агента, AI-настройки, пиксельная сцена, лог

## Быстрый старт

### Backend
```bash
cd backend
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
uvicorn app.main:app --reload --port 8000
```

- API: http://127.0.0.1:8000/docs  
- Логин: `admin@example.com` / `admin123`  
- Demo-токены плагинов: `demo:local`

### Admin
```bash
cd admin
npm install
npm run dev
```

### Control-бот
```bash
export CONTROL_BOT_TOKEN=...
cd backend && python -m bot.control_bot
```
Команда `/notify` привязывает чат к уведомлениям организации.

### Тесты
```bash
cd backend && pytest -q
```

## Добавить платформу
Создайте `backend/plugins/<platform>/plugin.py` с классом `Plugin(BaseAgentPlugin)` — registry подхватит без изменений ядра.
