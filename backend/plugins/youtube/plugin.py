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
    """YouTube Data API agent (MVP: demo mode + live channel lookup)."""

    manifest = PluginManifest(
        platform="youtube",
        title="YouTube Agent",
        description="Агент YouTube: канал, комментарии, публикации (Data API).",
        zone="youtube",
        actions=["test_connection", "publish_post", "reply"],
        fields=[
            FieldSpec(
                key="api_key",
                label="API Key",
                type="password",
                required=False,
                secret=True,
                help="Google API key. Для локальных тестов: demo:local",
                placeholder="AIza... или demo:local",
            ),
            FieldSpec(
                key="channel_id",
                label="Channel ID",
                type="text",
                required=True,
                secret=False,
                help="ID канала YouTube",
                placeholder="UC...",
            ),
            FieldSpec(
                key="oauth_access_token",
                label="OAuth Access Token",
                type="password",
                required=False,
                secret=True,
                help="Нужен для записи/комментариев",
                placeholder="ya29...",
            ),
        ],
    )

    async def connect(self) -> ConnectResult:
        channel_id = self.credentials.get("channel_id", "").strip()
        api_key = self.credentials.get("api_key", "").strip()
        if not channel_id:
            return ConnectResult(ok=False, message="channel_id is required")

        if not api_key or api_key.startswith("demo:") or api_key == "DEMO":
            return ConnectResult(
                ok=True,
                message=f"Demo mode for channel {channel_id}",
                meta={"demo": True, "channel_id": channel_id},
            )

        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.get(
                    "https://www.googleapis.com/youtube/v3/channels",
                    params={"part": "snippet", "id": channel_id, "key": api_key},
                )
                data = resp.json()
        except httpx.HTTPError as exc:
            return ConnectResult(ok=False, message=f"Network error: {exc}")

        items = data.get("items") or []
        if not items:
            err = (data.get("error") or {}).get("message")
            return ConnectResult(ok=False, message=err or "Channel not found / invalid API key")
        title = items[0].get("snippet", {}).get("title", channel_id)
        return ConnectResult(ok=True, message=f"Connected to {title}", meta={"channel": items[0]})

    async def execute_action(self, action: str, payload: dict[str, Any]) -> ActionResult:
        if action in {"test_connection", "connect"}:
            result = await self.connect()
            return ActionResult(ok=result.ok, message=result.message, data=result.meta)
        return ActionResult(
            ok=True,
            message=f"[mvp] YouTube {action} stub accepted",
            data={"action": action, "payload": payload},
        )

    async def listen_events(self) -> AsyncIterator[AgentEventDTO]:
        if False:  # pragma: no cover
            yield AgentEventDTO(type="comment", payload={}, summary="")
        return
        yield

    async def get_status(self) -> AgentStatusDTO:
        result = await self.connect()
        return AgentStatusDTO(
            status="online" if result.ok else "error",
            message=result.message,
            details=result.meta,
        )

    async def disconnect(self) -> None:
        return None
