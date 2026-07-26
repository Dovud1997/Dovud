from __future__ import annotations

from typing import Any

from sqlalchemy.ext.asyncio import AsyncSession

from app.models.entities import Agent
from app.services import agents as agent_service


def parse_telegram_update(update: dict[str, Any]) -> tuple[str, dict[str, Any]] | None:
    """Map a Telegram Update object to (event_type, payload)."""
    if "message" in update:
        msg = update["message"]
        chat = msg.get("chat") or {}
        sender = msg.get("from") or {}
        text = msg.get("text") or msg.get("caption") or ""

        if msg.get("new_chat_members"):
            return "follow", {
                "chat_id": chat.get("id"),
                "username": sender.get("username"),
                "members": msg.get("new_chat_members"),
                "text": text,
            }
        if msg.get("left_chat_member"):
            return "unfollow", {
                "chat_id": chat.get("id"),
                "username": (msg.get("left_chat_member") or {}).get("username"),
                "text": text,
            }

        # Ignore bot commands meant for control UX when text starts with / — still record as message.
        return "message", {
            "chat_id": chat.get("id"),
            "message_id": msg.get("message_id"),
            "username": sender.get("username"),
            "user_id": sender.get("id"),
            "text": text,
            "has_photo": bool(msg.get("photo")),
            "has_video": bool(msg.get("video")),
        }

    if "callback_query" in update:
        cq = update["callback_query"]
        msg = cq.get("message") or {}
        return "message", {
            "chat_id": (msg.get("chat") or {}).get("id"),
            "message_id": msg.get("message_id"),
            "username": (cq.get("from") or {}).get("username"),
            "text": cq.get("data") or "",
            "callback_query_id": cq.get("id"),
        }

    if "channel_post" in update:
        post = update["channel_post"]
        return "comment", {
            "chat_id": (post.get("chat") or {}).get("id"),
            "message_id": post.get("message_id"),
            "text": post.get("text") or post.get("caption") or "",
        }

    return None


async def process_telegram_update(
    db: AsyncSession,
    *,
    agent: Agent,
    update: dict[str, Any],
    auto_reply: bool = True,
) -> dict[str, Any]:
    parsed = parse_telegram_update(update)
    if parsed is None:
        return {"ok": True, "skipped": True, "reason": "unsupported_update"}

    event_type, payload = parsed
    # Don't auto-reply to slash-commands (control / other bots).
    text = str(payload.get("text") or "")
    should_reply = auto_reply and event_type in {"message", "comment"} and not text.startswith("/")

    event = await agent_service.record_event(
        db,
        agent=agent,
        event_type=event_type,
        payload={**payload, "update_id": update.get("update_id")},
        auto_reply=should_reply,
    )
    return {"ok": True, "event_id": event.id, "type": event_type, "auto_reply": should_reply}
