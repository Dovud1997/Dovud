"""
Control Telegram-бот — точка управления агентами.

Запуск (после настройки CONTROL_BOT_TOKEN и API):
  python -m bot.control_bot

MVP: команды /agents, /status, /cmd; приём фото+подписи как intent → API /commands.
"""

from __future__ import annotations

import asyncio
import logging
import os
import re
from typing import Any

import httpx
from aiogram import Bot, Dispatcher, F
from aiogram.filters import Command
from aiogram.types import Message

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("control_bot")

API_BASE = os.getenv("PLATFORM_API_BASE", "http://127.0.0.1:8000/api")
API_EMAIL = os.getenv("ADMIN_EMAIL", "admin@example.com")
API_PASSWORD = os.getenv("ADMIN_PASSWORD", "admin123")
BOT_TOKEN = os.getenv("CONTROL_BOT_TOKEN", "")

INTENT_MAP = [
    (re.compile(r"истори", re.I), "publish_story"),
    (re.compile(r"пост|опублик", re.I), "publish_post"),
    (re.compile(r"ответ|reply", re.I), "reply"),
    (re.compile(r"сообщ|send", re.I), "send_message"),
]


class PlatformClient:
    def __init__(self) -> None:
        self._token: str | None = None

    async def login(self) -> None:
        async with httpx.AsyncClient(timeout=20) as client:
            resp = await client.post(
                f"{API_BASE}/auth/login",
                json={"email": API_EMAIL, "password": API_PASSWORD},
            )
            resp.raise_for_status()
            self._token = resp.json()["access_token"]

    async def _headers(self) -> dict[str, str]:
        if not self._token:
            await self.login()
        return {"Authorization": f"Bearer {self._token}"}

    async def get(self, path: str) -> Any:
        async with httpx.AsyncClient(timeout=20) as client:
            resp = await client.get(f"{API_BASE}{path}", headers=await self._headers())
            if resp.status_code == 401:
                await self.login()
                resp = await client.get(f"{API_BASE}{path}", headers=await self._headers())
            resp.raise_for_status()
            return resp.json()

    async def post(self, path: str, json: dict[str, Any]) -> Any:
        async with httpx.AsyncClient(timeout=20) as client:
            resp = await client.post(f"{API_BASE}{path}", headers=await self._headers(), json=json)
            if resp.status_code == 401:
                await self.login()
                resp = await client.post(f"{API_BASE}{path}", headers=await self._headers(), json=json)
            resp.raise_for_status()
            return resp.json()


api = PlatformClient()
dp = Dispatcher()


def parse_intent(text: str) -> str:
    for pattern, action in INTENT_MAP:
        if pattern.search(text):
            return action
    return "publish_post"


@dp.message(Command("start"))
async def cmd_start(message: Message) -> None:
    await message.answer(
        "Control-бот платформы агентов.\n"
        "/agents — список\n"
        "/status — статусы\n"
        "/cmd <agent_id> <action> [text] — команда\n"
        "Или пришлите фото/видео с подписью: «поставь в историю»."
    )


@dp.message(Command("agents"))
async def cmd_agents(message: Message) -> None:
    agents = await api.get("/agents")
    if not agents:
        await message.answer("Агентов пока нет. Добавьте через админ-панель.")
        return
    lines = [f"• {a['name']} [{a['platform']}] — {a['status']} (`{a['id'][:8]}…`)" for a in agents]
    await message.answer("Агенты:\n" + "\n".join(lines))


@dp.message(Command("status"))
async def cmd_status(message: Message) -> None:
    agents = await api.get("/scene")
    if not agents:
        await message.answer("Нет агентов на сцене.")
        return
    lines = [f"{a['name']}: {a['status']} — {a.get('status_message') or '—'}" for a in agents]
    await message.answer("Статусы:\n" + "\n".join(lines))


@dp.message(Command("cmd"))
async def cmd_command(message: Message) -> None:
    parts = (message.text or "").split(maxsplit=3)
    if len(parts) < 3:
        await message.answer("Формат: /cmd <agent_id> <action> [text]")
        return
    _, agent_id, action, *rest = parts
    text = rest[0] if rest else ""
    # Resolve short id prefix
    agents = await api.get("/agents")
    match = next((a for a in agents if a["id"].startswith(agent_id) or a["id"] == agent_id), None)
    if match is None:
        await message.answer("Агент не найден")
        return
    job = await api.post(
        "/commands",
        {"agent_id": match["id"], "action": action, "payload": {"text": text}},
    )
    await message.answer(f"Команда поставлена в очередь: job={job['id'][:8]}… status={job['status']}")


@dp.message(F.photo | F.video | F.document | F.text)
async def media_or_text(message: Message) -> None:
    caption = message.caption or message.text or ""
    if caption.startswith("/"):
        return
    if not caption and not (message.photo or message.video):
        return

    agents = await api.get("/agents")
    telegram_agents = [a for a in agents if a["platform"] == "telegram" and a.get("is_active")]
    target = telegram_agents[0] if telegram_agents else (agents[0] if agents else None)
    if target is None:
        await message.answer("Нет активных агентов.")
        return

    action = parse_intent(caption)
    # Media file_id as placeholder URL for MVP (real upload pipeline later)
    media_ref = None
    if message.photo:
        media_ref = message.photo[-1].file_id
    elif message.video:
        media_ref = message.video.file_id

    job = await api.post(
        "/commands",
        {
            "agent_id": target["id"],
            "action": action,
            "payload": {"text": caption, "media_url": media_ref, "source": "control_bot"},
        },
    )
    await message.answer(
        f"→ агент *{target['name']}*\nдействие: `{action}`\njob: `{job['id'][:8]}…`",
        parse_mode="Markdown",
    )


async def main() -> None:
    if not BOT_TOKEN:
        raise SystemExit("Set CONTROL_BOT_TOKEN env var")
    await api.login()
    bot = Bot(BOT_TOKEN)
    logger.info("Control bot starting")
    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
