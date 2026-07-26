from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

import httpx

from app.agents.base import (
    ActionResult,
    AgentEventDTO,
    AgentStatusDTO,
    BaseAgentPlugin,
    ConnectResult,
    FieldSpec,
    PluginManifest,
)


class Plugin(BaseAgentPlugin):
    """Telegram agent: validates token via getMe, executes basic actions, supports demo tokens."""

    manifest = PluginManifest(
        platform="telegram",
        title="Telegram Agent",
        description="Агент для Telegram-бота/канала: тест соединения, статус, события и команды публикации.",
        zone="telegram",
        actions=["test_connection", "send_message", "publish_post", "publish_story", "reply"],
        fields=[
            FieldSpec(
                key="bot_token",
                label="Bot Token",
                type="password",
                required=True,
                secret=True,
                help="Токен от @BotFather (или demo:local для локальных тестов)",
                placeholder="123456:ABC-DEF... или demo:local",
            ),
            FieldSpec(
                key="chat_id",
                label="Default Chat / Channel ID",
                type="text",
                required=False,
                secret=False,
                help="Куда публиковать по умолчанию (опционально)",
                placeholder="-1001234567890",
            ),
        ],
    )

    def __init__(self, agent_id: str, credentials: dict[str, str]):
        super().__init__(agent_id, credentials)
        self._me: dict[str, Any] | None = None
        self._connected = False

    @property
    def _token(self) -> str:
        return self.credentials.get("bot_token", "").strip()

    @property
    def _demo(self) -> bool:
        return self._token.startswith("demo:") or self._token == "DEMO"

    def _api(self, method: str) -> str:
        return f"https://api.telegram.org/bot{self._token}/{method}"

    async def connect(self) -> ConnectResult:
        if not self._token:
            return ConnectResult(ok=False, message="bot_token is required")
        if self._demo:
            self._connected = True
            self._me = {"username": "demo_bot", "id": 0}
            return ConnectResult(ok=True, message="Connected as @demo_bot (demo mode)", meta={"bot": self._me})

        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.get(self._api("getMe"))
                data = resp.json()
        except httpx.HTTPError as exc:
            return ConnectResult(ok=False, message=f"Network error: {exc}")

        if not data.get("ok"):
            return ConnectResult(ok=False, message=data.get("description", "Telegram API error"))

        self._me = data.get("result", {})
        self._connected = True
        username = self._me.get("username", "?")
        return ConnectResult(ok=True, message=f"Connected as @{username}", meta={"bot": self._me})

    async def execute_action(self, action: str, payload: dict[str, Any]) -> ActionResult:
        if action in {"test_connection", "connect"}:
            result = await self.connect()
            return ActionResult(ok=result.ok, message=result.message, data=result.meta)

        chat_id = str(payload.get("chat_id") or self.credentials.get("chat_id") or "").strip()
        text = str(payload.get("text") or payload.get("caption") or "").strip()
        media_url = payload.get("media_url")

        if self._demo:
            return ActionResult(
                ok=True,
                message=f"[demo] {action} → chat={chat_id or 'n/a'}: {text[:120]}",
                data={"dry_run": True, "demo": True, "action": action, "text": text, "media_url": media_url},
            )

        if action in {"send_message", "reply", "publish_post"}:
            if not chat_id:
                return ActionResult(
                    ok=True,
                    message=f"[dry-run] {action}: no chat_id; would send: {text[:120]}",
                    data={"dry_run": True, "action": action, "text": text, "media_url": media_url},
                )
            if not self._connected:
                conn = await self.connect()
                if not conn.ok:
                    return ActionResult(ok=False, message=conn.message)

            try:
                async with httpx.AsyncClient(timeout=20) as client:
                    if media_url and action == "publish_post":
                        resp = await client.post(
                            self._api("sendPhoto"),
                            json={"chat_id": chat_id, "photo": media_url, "caption": text},
                        )
                    else:
                        resp = await client.post(
                            self._api("sendMessage"),
                            json={"chat_id": chat_id, "text": text or "(empty)"},
                        )
                    data = resp.json()
            except httpx.HTTPError as exc:
                return ActionResult(ok=False, message=str(exc))

            if not data.get("ok"):
                return ActionResult(ok=False, message=data.get("description", "send failed"))
            return ActionResult(ok=True, message="Message sent", data=data.get("result", {}))

        if action == "publish_story":
            return ActionResult(
                ok=True,
                message="[mvp] Stories API stub — command accepted",
                data={"action": action, "media_url": media_url, "text": text, "chat_id": chat_id},
            )

        return ActionResult(ok=False, message=f"Unsupported action: {action}")

    async def listen_events(self) -> AsyncIterator[AgentEventDTO]:
        if False:  # pragma: no cover
            yield AgentEventDTO(type="message", payload={}, summary="")
        return
        yield

    async def get_status(self) -> AgentStatusDTO:
        if not self._token:
            return AgentStatusDTO(status="error", message="Missing bot_token")
        result = await self.connect()
        if result.ok:
            return AgentStatusDTO(status="online", message=result.message, details=result.meta)
        return AgentStatusDTO(status="error", message=result.message)

    async def disconnect(self) -> None:
        self._connected = False
        self._me = None
