# DOVUD — платформа AI-агентов

Единая multi-tenant система управления AI-агентами для каналов (Telegram, Instagram, YouTube).  
Агенты — плагины; конфиги через админку; управление через API / Telegram-бот; realtime пиксельная сцена.

## Документация
- [Стек](docs/STACK.md) · [Архитектура](docs/ARCHITECTURE.md) · [Схема данных](docs/DATA_SCHEMA.md)

## Возможности
- Multi-tenant организации + регистрация
- Плагины: Telegram / Instagram / YouTube
- **Instagram Content Publishing** — posts / Stories / Reels через Graph API
- Медиа-хостинг (`POST /api/media/upload` → `/media/{file}`) для публичных URL
- Control-бот: фото/видео → загрузка на платформу → publish на Instagram
- Очередь задач (in-process или Redis/ARQ)
- Автоответы: шаблоны или LLM (OpenAI / Anthropic)
- Уведомления в Telegram / webhook
- Админка: форма агента, AI-настройки, пиксельная сцена, лог, Media publish

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
Положите **новый** токен только в локальный `backend/.env` (не в чат и не в git):
```env
CONTROL_BOT_TOKEN=...
```
```bash
./scripts/control-bot.sh
# или: cd backend && python -m bot.control_bot
```
Команда `/notify` привязывает чат к уведомлениям организации.

### Telegram inbound
- **Long-polling** включается автоматически для активных Telegram-агентов (не demo).
- **Webhook** (если есть публичный URL): кнопка Webhook в админке или  
  `POST /api/agents/{id}/telegram/set-webhook`  
  URL: `{PUBLIC_BASE_URL}/api/webhooks/telegram/{agent_id}`

### Instagram publish
1. Создайте Instagram-агента (access token + `ig_user_id`; для UI: `demo:local`).
2. Укажите **публичный** `PUBLIC_BASE_URL` (Meta должна достучаться до `/media/...`).
3. Из control-бота: фото + «поставь в историю Instagram» — бот скачает файл, зальёт на платформу и поставит `publish_story`.
4. Из админки: кнопка **Media** у агента — URL или выбор файла.

### Тесты
```bash
cd backend && pytest -q
```

## Добавить платформу
Создайте `backend/plugins/<platform>/plugin.py` с классом `Plugin(BaseAgentPlugin)` — registry подхватит без изменений ядра.
