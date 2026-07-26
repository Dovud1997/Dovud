from __future__ import annotations

import logging
from typing import Any

import httpx
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.config import get_settings
from app.models.entities import NotificationTarget
from app.services.events import event_hub

logger = logging.getLogger(__name__)


async def notify_org(
    db: AsyncSession,
    *,
    org_id: str,
    title: str,
    body: str,
    meta: dict[str, Any] | None = None,
) -> None:
    """Fan-out notifications to org targets (Telegram control chat / webhook)."""
    result = await db.execute(
        select(NotificationTarget).where(
            NotificationTarget.org_id == org_id,
            NotificationTarget.is_active.is_(True),
        )
    )
    targets = list(result.scalars().all())
    payload = {"title": title, "body": body, "meta": meta or {}, "org_id": org_id}
    await event_hub.publish({"type": "notification", **payload})

    settings = get_settings()
    for target in targets:
        try:
            if target.channel == "telegram":
                await _send_telegram(settings.control_bot_token, target.address, f"*{title}*\n{body}")
            elif target.channel == "webhook":
                async with httpx.AsyncClient(timeout=15) as client:
                    await client.post(target.address, json=payload)
        except Exception:  # noqa: BLE001
            logger.exception("Notification failed for %s:%s", target.channel, target.address)


async def _send_telegram(token: str, chat_id: str, text: str) -> None:
    if not token or not chat_id:
        return
    async with httpx.AsyncClient(timeout=15) as client:
        await client.post(
            f"https://api.telegram.org/bot{token}/sendMessage",
            json={"chat_id": chat_id, "text": text, "parse_mode": "Markdown"},
        )
