"""
Control Telegram-бот — точка управления агентами.

  python -m bot.control_bot
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
ORG_ID = os.getenv("PLATFORM_ORG_ID", "")

INTENT_MAP = [
    (re.compile(r"истори", re.I), "publish_story"),
    (re.compile(r"пост|опублик|reel", re.I), "publish_post"),
    (re.compile(r"ответ|reply", re.I), "reply"),
    (re.compile(r"сообщ|send", re.I), "send_message"),
]


class PlatformClient:
    def __init__(self) -> None:
        self._token: str | None = None
        self._org_id: str = ORG_ID

    async def login(self) -> None:
        async with httpx.AsyncClient(timeout=20) as client:
            resp = await client.post(
                f"{API_BASE}/auth/login",
                json={"email": API_EMAIL, "password": API_PASSWORD},
            )
            resp.raise_for_status()
            data = resp.json()
            self._token = data["access_token"]
            if not self._org_id and data.get("orgs"):
                self._org_id = data["orgs"][0]["id"]

    async def _headers(self, *, json_body: bool = True) -> dict[str, str]:
        if not self._token:
            await self.login()
        headers = {"Authorization": f"Bearer {self._token}"}
        if self._org_id:
            headers["X-Org-Id"] = self._org_id
        if json_body:
            headers["Content-Type"] = "application/json"
        return headers

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

    async def upload_media(self, data: bytes, filename: str, content_type: str | None = None) -> dict[str, Any]:
        headers = await self._headers(json_body=False)
        files = {"file": (filename, data, content_type or "application/octet-stream")}
        async with httpx.AsyncClient(timeout=60) as client:
            resp = await client.post(f"{API_BASE}/media/upload", headers=headers, files=files)
            if resp.status_code == 401:
                await self.login()
                headers = await self._headers(json_body=False)
                resp = await client.post(f"{API_BASE}/media/upload", headers=headers, files=files)
            resp.raise_for_status()
            return resp.json()


api = PlatformClient()
dp = Dispatcher()
bot_ref: Bot | None = None


def parse_intent(text: str) -> str:
    for pattern, action in INTENT_MAP:
        if pattern.search(text):
            return action
    return "publish_post"


async def resolve_telegram_media(message: Message) -> dict[str, Any] | None:
    """Download Telegram media and re-host on the platform for Instagram Graph API."""
    assert bot_ref is not None
    file_id = None
    filename = "media.bin"
    content_type = None
    if message.photo:
        file_id = message.photo[-1].file_id
        filename = "photo.jpg"
        content_type = "image/jpeg"
    elif message.video:
        file_id = message.video.file_id
        filename = message.video.file_name or "video.mp4"
        content_type = message.video.mime_type or "video/mp4"
    elif message.document:
        file_id = message.document.file_id
        filename = message.document.file_name or "document.bin"
        content_type = message.document.mime_type
    if not file_id:
        return None

    tg_file = await bot_ref.get_file(file_id)
    if not tg_file.file_path:
        return None
    buf = await bot_ref.download_file(tg_file.file_path)
    data = buf.read() if hasattr(buf, "read") else bytes(buf)
    uploaded = await api.upload_media(data, filename=filename, content_type=content_type)
    return {
        "media_url": uploaded["public_url"],
        "media_kind": uploaded.get("media_kind"),
        "is_video": uploaded.get("media_kind") == "video",
        "telegram_file_id": file_id,
    }


@dp.message(Command("start"))
async def cmd_start(message: Message) -> None:
    await message.answer(
        "Control-бот платформы агентов.\n"
        "/agents — список\n"
        "/status — статусы\n"
        "/cmd <agent_id> <action> [text]\n"
        "/notify — сохранить этот чат для уведомлений\n"
        "Фото/видео + подпись «поставь в историю Instagram»."
    )


@dp.message(Command("notify"))
async def cmd_notify(message: Message) -> None:
    chat_id = str(message.chat.id)
    await api.post("/notifications/targets", {"channel": "telegram", "address": chat_id, "is_active": True})
    await message.answer(f"Этот чат ({chat_id}) будет получать уведомления о событиях.")


@dp.message(Command("agents"))
async def cmd_agents(message: Message) -> None:
    agents = await api.get("/agents")
    if not agents:
        await message.answer("Агентов пока нет.")
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
    agents = await api.get("/agents")
    match = next((a for a in agents if a["id"].startswith(agent_id) or a["id"] == agent_id), None)
    if match is None:
        await message.answer("Агент не найден")
        return
    job = await api.post(
        "/commands",
        {"agent_id": match["id"], "action": action, "payload": {"text": text}},
    )
    await message.answer(f"Очередь: job={job['id'][:8]}… status={job['status']}")


@dp.message(F.photo | F.video | F.document | F.text)
async def media_or_text(message: Message) -> None:
    caption = message.caption or message.text or ""
    if caption.startswith("/"):
        return
    if not caption and not (message.photo or message.video or message.document):
        return

    agents = await api.get("/agents")
    active = [a for a in agents if a.get("is_active")]
    # Prefer platform mentioned in caption
    target = None
    lower = caption.lower()
    for platform in ("instagram", "youtube", "telegram"):
        if platform in lower:
            target = next((a for a in active if a["platform"] == platform), None)
            break
    # Default media publish → Instagram when available
    if target is None and (message.photo or message.video or message.document):
        target = next((a for a in active if a["platform"] == "instagram"), None)
    if target is None:
        target = active[0] if active else (agents[0] if agents else None)
    if target is None:
        await message.answer("Нет активных агентов.")
        return

    action = parse_intent(caption)
    payload: dict[str, Any] = {"text": caption, "source": "control_bot"}

    if message.photo or message.video or message.document:
        try:
            media = await resolve_telegram_media(message)
        except Exception as exc:  # noqa: BLE001
            logger.exception("Media resolve failed")
            await message.answer(f"Не удалось загрузить медиа: {exc}")
            return
        if media:
            payload.update(media)
        else:
            await message.answer("Медиафайл пуст или недоступен.")
            return

    job = await api.post(
        "/commands",
        {
            "agent_id": target["id"],
            "action": action,
            "payload": payload,
        },
    )
    media_note = f"\nmedia: `{payload.get('media_url')}`" if payload.get("media_url") else ""
    await message.answer(
        f"→ *{target['name']}* (`{target['platform']}`)\n"
        f"действие: `{action}`\njob: `{job['id'][:8]}…`{media_note}",
        parse_mode="Markdown",
    )


async def main() -> None:
    global bot_ref
    if not BOT_TOKEN:
        raise SystemExit("Set CONTROL_BOT_TOKEN env var")
    await api.login()
    bot_ref = Bot(BOT_TOKEN)
    logger.info("Control bot starting org=%s", api._org_id)
    await dp.start_polling(bot_ref)


if __name__ == "__main__":
    asyncio.run(main())
