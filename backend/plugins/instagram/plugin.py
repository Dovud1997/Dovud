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
    """
    Instagram Graph API agent (MVP).
    Validates credentials shape; live publish requires Meta App + Instagram Business account.
    """

    manifest = PluginManifest(
        platform="instagram",
        title="Instagram Agent",
        description="Агент Instagram (Graph API): stories/posts, комментарии, события вовлечения.",
        zone="instagram",
        actions=["test_connection", "publish_post", "publish_story", "reply"],
        fields=[
            FieldSpec(
                key="access_token",
                label="Access Token",
                type="password",
                required=True,
                secret=True,
                help="Long-lived token Instagram Graph API",
                placeholder="EAAB...",
            ),
            FieldSpec(
                key="ig_user_id",
                label="Instagram Business User ID",
                type="text",
                required=True,
                secret=False,
                help="ID бизнес-аккаунта Instagram",
                placeholder="17841...",
            ),
        ],
    )

    async def connect(self) -> ConnectResult:
        token = self.credentials.get("access_token", "").strip()
        user_id = self.credentials.get("ig_user_id", "").strip()
        if not token or not user_id:
            return ConnectResult(ok=False, message="access_token and ig_user_id are required")

        # Soft validation: call Graph me endpoint; if unreachable/invalid — clear error.
        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.get(
                    f"https://graph.facebook.com/v21.0/{user_id}",
                    params={"fields": "id,username", "access_token": token},
                )
                data = resp.json()
        except httpx.HTTPError as exc:
            return ConnectResult(ok=False, message=f"Network error: {exc}")

        if data.get("error"):
            # Allow draft/demo mode with fake tokens for UI testing.
            if token.startswith("demo:") or token == "DEMO":
                return ConnectResult(ok=True, message="Demo mode connected (Instagram stub)", meta={"demo": True})
            return ConnectResult(ok=False, message=data["error"].get("message", "Instagram API error"))

        username = data.get("username", user_id)
        return ConnectResult(ok=True, message=f"Connected as @{username}", meta={"user": data})

    async def execute_action(self, action: str, payload: dict[str, Any]) -> ActionResult:
        if action in {"test_connection", "connect"}:
            result = await self.connect()
            return ActionResult(ok=result.ok, message=result.message, data=result.meta)

        token = self.credentials.get("access_token", "")
        if token.startswith("demo:") or token == "DEMO":
            return ActionResult(
                ok=True,
                message=f"[demo] Instagram {action} accepted",
                data={"action": action, "payload": payload},
            )

        if action in {"publish_post", "publish_story"}:
            return ActionResult(
                ok=True,
                message=f"[mvp] Instagram {action} queued conceptually — wire media container API next",
                data={"action": action, "payload": payload},
            )
        if action == "reply":
            return ActionResult(
                ok=True,
                message="[mvp] Instagram reply stub",
                data={"text": payload.get("text")},
            )
        return ActionResult(ok=False, message=f"Unsupported action: {action}")

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
