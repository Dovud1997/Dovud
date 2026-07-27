from __future__ import annotations

import asyncio
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
from app.services.media import is_video_url

GRAPH_BASE = "https://graph.facebook.com/v21.0"


class Plugin(BaseAgentPlugin):
    """
    Instagram Graph API agent.

    Live publish uses Content Publishing API:
      1) create media container (image_url / video_url)
      2) wait until FINISHED (video)
      3) media_publish

    Demo tokens (`demo:*`) short-circuit without calling Meta.
    PUBLIC_BASE_URL must be reachable by Meta for real publishes.
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
                help="Long-lived token Instagram Graph API (или demo:local)",
                placeholder="EAAB... или demo:local",
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

    @property
    def _token(self) -> str:
        return self.credentials.get("access_token", "").strip()

    @property
    def _user_id(self) -> str:
        return self.credentials.get("ig_user_id", "").strip()

    @property
    def _demo(self) -> bool:
        return self._token.startswith("demo:") or self._token == "DEMO"

    async def connect(self) -> ConnectResult:
        token = self._token
        user_id = self._user_id
        if not token or not user_id:
            return ConnectResult(ok=False, message="access_token and ig_user_id are required")

        if self._demo:
            return ConnectResult(
                ok=True,
                message="Demo mode connected (Instagram stub)",
                meta={"demo": True, "ig_user_id": user_id},
            )

        try:
            async with httpx.AsyncClient(timeout=15) as client:
                resp = await client.get(
                    f"{GRAPH_BASE}/{user_id}",
                    params={"fields": "id,username", "access_token": token},
                )
                data = resp.json()
        except httpx.HTTPError as exc:
            return ConnectResult(ok=False, message=f"Network error: {exc}")

        if data.get("error"):
            return ConnectResult(ok=False, message=data["error"].get("message", "Instagram API error"))

        username = data.get("username", user_id)
        return ConnectResult(ok=True, message=f"Connected as @{username}", meta={"user": data})

    async def execute_action(self, action: str, payload: dict[str, Any]) -> ActionResult:
        if action in {"test_connection", "connect"}:
            result = await self.connect()
            return ActionResult(ok=result.ok, message=result.message, data=result.meta)

        if self._demo:
            return ActionResult(
                ok=True,
                message=f"[demo] Instagram {action} accepted",
                data={"action": action, "payload": payload, "demo": True},
            )

        if action in {"publish_post", "publish_story"}:
            return await self._publish(action=action, payload=payload)

        if action == "reply":
            return await self._reply_comment(payload)

        return ActionResult(ok=False, message=f"Unsupported action: {action}")

    async def _publish(self, *, action: str, payload: dict[str, Any]) -> ActionResult:
        media_url = str(payload.get("media_url") or payload.get("image_url") or payload.get("video_url") or "").strip()
        caption = str(payload.get("text") or payload.get("caption") or "").strip()
        if not media_url:
            return ActionResult(
                ok=False,
                message="media_url is required (publicly reachable image/video URL)",
            )
        if not self._token or not self._user_id:
            return ActionResult(ok=False, message="Missing Instagram credentials")

        is_story = action == "publish_story"
        is_video = bool(payload.get("is_video")) or is_video_url(media_url)
        container_params: dict[str, Any] = {"access_token": self._token}

        if is_story:
            container_params["media_type"] = "STORIES"
            if is_video:
                container_params["video_url"] = media_url
            else:
                container_params["image_url"] = media_url
        elif is_video:
            # Feed video → Reels container (Graph Content Publishing).
            container_params["media_type"] = "REELS"
            container_params["video_url"] = media_url
            if caption:
                container_params["caption"] = caption
        else:
            container_params["image_url"] = media_url
            if caption:
                container_params["caption"] = caption

        try:
            async with httpx.AsyncClient(timeout=60) as client:
                create = await client.post(f"{GRAPH_BASE}/{self._user_id}/media", data=container_params)
                created = create.json()
                if created.get("error"):
                    return ActionResult(
                        ok=False,
                        message=created["error"].get("message", "Failed to create media container"),
                        data=created,
                    )
                creation_id = created.get("id")
                if not creation_id:
                    return ActionResult(ok=False, message="No creation_id from Instagram", data=created)

                if is_video:
                    status = await self._wait_container_ready(client, creation_id)
                    if status.get("error"):
                        return ActionResult(
                            ok=False,
                            message=status["error"].get("message", "Container processing failed"),
                            data=status,
                        )
                    code = (status.get("status_code") or "").upper()
                    if code not in {"FINISHED", "PUBLISHED"}:
                        return ActionResult(
                            ok=False,
                            message=f"Media not ready: {code or 'unknown'}",
                            data=status,
                        )

                publish = await client.post(
                    f"{GRAPH_BASE}/{self._user_id}/media_publish",
                    data={"creation_id": creation_id, "access_token": self._token},
                )
                published = publish.json()
        except httpx.HTTPError as exc:
            return ActionResult(ok=False, message=f"Network error: {exc}")

        if published.get("error"):
            return ActionResult(
                ok=False,
                message=published["error"].get("message", "Publish failed"),
                data=published,
            )

        media_id = published.get("id")
        kind = "story" if is_story else ("reel" if is_video else "post")
        return ActionResult(
            ok=True,
            message=f"Instagram {kind} published",
            data={
                "media_id": media_id,
                "creation_id": creation_id,
                "kind": kind,
                "media_url": media_url,
            },
        )

    async def _wait_container_ready(
        self,
        client: httpx.AsyncClient,
        creation_id: str,
        *,
        attempts: int = 12,
        delay_sec: float = 2.0,
    ) -> dict[str, Any]:
        last: dict[str, Any] = {}
        for _ in range(attempts):
            resp = await client.get(
                f"{GRAPH_BASE}/{creation_id}",
                params={"fields": "status_code", "access_token": self._token},
            )
            last = resp.json()
            if last.get("error"):
                return last
            code = (last.get("status_code") or "").upper()
            if code in {"FINISHED", "PUBLISHED", "ERROR", "EXPIRED"}:
                return last
            await asyncio.sleep(delay_sec)
        return last or {"status_code": "TIMEOUT"}

    async def _reply_comment(self, payload: dict[str, Any]) -> ActionResult:
        text = str(payload.get("text") or "").strip()
        comment_id = str(payload.get("comment_id") or payload.get("reply_to") or "").strip()
        if not text:
            return ActionResult(ok=False, message="reply text is required")
        if not comment_id:
            return ActionResult(
                ok=True,
                message="[dry-run] Instagram reply without comment_id",
                data={"dry_run": True, "text": text},
            )
        try:
            async with httpx.AsyncClient(timeout=20) as client:
                resp = await client.post(
                    f"{GRAPH_BASE}/{comment_id}/replies",
                    data={"message": text, "access_token": self._token},
                )
                data = resp.json()
        except httpx.HTTPError as exc:
            return ActionResult(ok=False, message=str(exc))
        if data.get("error"):
            return ActionResult(ok=False, message=data["error"].get("message", "Reply failed"), data=data)
        return ActionResult(ok=True, message="Comment reply sent", data=data)

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
