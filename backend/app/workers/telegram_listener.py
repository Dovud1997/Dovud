from __future__ import annotations

import asyncio
import logging
from typing import Any

import httpx
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from app.core.config import get_settings
from app.core.db import SessionLocal
from app.core.security import decrypt_secret
from app.models.entities import Agent, AgentSecret
from app.services.telegram_updates import process_telegram_update

logger = logging.getLogger(__name__)


class TelegramListenerHub:
    """Long-polls Telegram getUpdates for every active non-demo Telegram agent."""

    def __init__(self) -> None:
        self._running = False
        self._supervisor: asyncio.Task[None] | None = None
        self._agent_tasks: dict[str, asyncio.Task[None]] = {}
        self._offsets: dict[str, int] = {}

    async def start(self) -> None:
        settings = get_settings()
        if not getattr(settings, "telegram_listen_enabled", True):
            logger.info("Telegram listener disabled")
            return
        if self._running:
            return
        self._running = True
        self._supervisor = asyncio.create_task(self._supervise(), name="telegram-listener-supervisor")
        logger.info("Telegram listener hub started")

    async def stop(self) -> None:
        self._running = False
        for task in list(self._agent_tasks.values()):
            task.cancel()
        self._agent_tasks.clear()
        if self._supervisor:
            self._supervisor.cancel()
            try:
                await self._supervisor
            except (asyncio.CancelledError, RuntimeError):
                pass
            self._supervisor = None

    async def _supervise(self) -> None:
        while self._running:
            try:
                agents = await self._active_telegram_agents()
                live_ids = {a["id"] for a in agents}
                for agent_id in list(self._agent_tasks):
                    if agent_id not in live_ids:
                        self._agent_tasks[agent_id].cancel()
                        self._agent_tasks.pop(agent_id, None)
                        logger.info("Stopped listener for agent %s", agent_id)
                for agent in agents:
                    aid = agent["id"]
                    if aid not in self._agent_tasks or self._agent_tasks[aid].done():
                        self._agent_tasks[aid] = asyncio.create_task(
                            self._poll_agent(aid, agent["token"]),
                            name=f"tg-poll-{aid[:8]}",
                        )
                        logger.info("Started listener for agent %s", aid)
            except Exception:  # noqa: BLE001
                logger.exception("Telegram supervisor iteration failed")
            await asyncio.sleep(5)

    async def _active_telegram_agents(self) -> list[dict[str, str]]:
        async with SessionLocal() as db:
            result = await db.execute(
                select(Agent)
                .options(selectinload(Agent.secrets))
                .where(Agent.platform == "telegram", Agent.is_active.is_(True))
            )
            out: list[dict[str, str]] = []
            for agent in result.scalars().all():
                token = ""
                for secret in agent.secrets:
                    if secret.key == "bot_token":
                        token = decrypt_secret(secret.value_encrypted)
                        break
                if not token or token.startswith("demo:") or token == "DEMO":
                    continue
                out.append({"id": agent.id, "token": token})
            return out

    async def _poll_agent(self, agent_id: str, token: str) -> None:
        api = f"https://api.telegram.org/bot{token}"
        # Drop webhook so getUpdates works.
        try:
            async with httpx.AsyncClient(timeout=20) as client:
                await client.get(f"{api}/deleteWebhook", params={"drop_pending_updates": False})
        except httpx.HTTPError:
            logger.warning("deleteWebhook failed for %s", agent_id)

        while self._running:
            offset = self._offsets.get(agent_id, 0)
            try:
                async with httpx.AsyncClient(timeout=35) as client:
                    resp = await client.post(
                        f"{api}/getUpdates",
                        json={
                            "timeout": 25,
                            "offset": offset,
                            "allowed_updates": ["message", "callback_query", "channel_post"],
                        },
                    )
                    data = resp.json()
            except httpx.HTTPError as exc:
                logger.warning("getUpdates error agent=%s: %s", agent_id, exc)
                await asyncio.sleep(3)
                continue

            if not data.get("ok"):
                logger.warning("getUpdates not ok agent=%s: %s", agent_id, data)
                await asyncio.sleep(5)
                # Refresh token in case credentials rotated.
                fresh = await self._refresh_token(agent_id)
                if not fresh:
                    return
                token = fresh
                api = f"https://api.telegram.org/bot{token}"
                continue

            updates: list[dict[str, Any]] = data.get("result") or []
            for update in updates:
                update_id = int(update.get("update_id", 0))
                self._offsets[agent_id] = update_id + 1
                try:
                    await self._handle_update(agent_id, update)
                except Exception:  # noqa: BLE001
                    logger.exception("Failed handling update %s for %s", update_id, agent_id)

    async def _refresh_token(self, agent_id: str) -> str | None:
        async with SessionLocal() as db:
            result = await db.execute(select(AgentSecret).where(AgentSecret.agent_id == agent_id, AgentSecret.key == "bot_token"))
            secret = result.scalar_one_or_none()
            if secret is None:
                return None
            token = decrypt_secret(secret.value_encrypted)
            if token.startswith("demo:") or token == "DEMO":
                return None
            return token

    async def _handle_update(self, agent_id: str, update: dict[str, Any]) -> None:
        async with SessionLocal() as db:
            agent = await db.get(Agent, agent_id)
            if agent is None or not agent.is_active:
                return
            await process_telegram_update(db, agent=agent, update=update, auto_reply=True)


telegram_listener = TelegramListenerHub()
