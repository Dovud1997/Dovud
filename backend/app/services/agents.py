from __future__ import annotations

import re
from typing import Any

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.agents.registry import get_registry
from app.core.security import decrypt_secret, encrypt_secret
from app.llm.client import llm_client
from app.models.entities import (
    AIMode,
    Agent,
    AgentEvent,
    AgentLog,
    AgentSecret,
    AgentStatus,
    CommandJob,
    JobStatus,
    LogLevel,
    ReplyTemplate,
    StyleExample,
)
from app.services.events import event_hub
from app.workers.queue import task_queue

ZONE_DEFAULTS: dict[str, tuple[float, float]] = {
    "telegram": (140.0, 200.0),
    "instagram": (420.0, 200.0),
    "youtube": (700.0, 200.0),
}


def agent_to_dict(agent: Agent) -> dict[str, Any]:
    return {
        "id": agent.id,
        "name": agent.name,
        "platform": agent.platform,
        "status": agent.status.value if hasattr(agent.status, "value") else agent.status,
        "status_message": agent.status_message,
        "ai_mode": agent.ai_mode.value if hasattr(agent.ai_mode, "value") else agent.ai_mode,
        "system_prompt": agent.system_prompt,
        "zone": agent.zone,
        "pos_x": agent.pos_x,
        "pos_y": agent.pos_y,
        "is_active": agent.is_active,
        "created_at": agent.created_at,
        "updated_at": agent.updated_at,
        "has_secrets": bool(agent.secrets) if agent.secrets is not None else False,
    }


async def get_credentials(db: AsyncSession, agent_id: str) -> dict[str, str]:
    result = await db.execute(select(AgentSecret).where(AgentSecret.agent_id == agent_id))
    secrets = result.scalars().all()
    return {s.key: decrypt_secret(s.value_encrypted) for s in secrets}


async def upsert_secrets(db: AsyncSession, agent_id: str, credentials: dict[str, str]) -> None:
    manifest = None
    agent = await db.get(Agent, agent_id)
    if agent:
        manifest = get_registry().get_manifest(agent.platform)
    secret_keys = {f.key for f in manifest.fields if f.secret} if manifest else set(credentials)

    for key, value in credentials.items():
        if value is None or value == "":
            continue
        # Encrypt all provided values; mark non-secret keys too for simplicity in MVP.
        result = await db.execute(
            select(AgentSecret).where(AgentSecret.agent_id == agent_id, AgentSecret.key == key)
        )
        existing = result.scalar_one_or_none()
        enc = encrypt_secret(value)
        if existing:
            existing.value_encrypted = enc
        else:
            db.add(AgentSecret(agent_id=agent_id, key=key, value_encrypted=enc))
        _ = secret_keys  # reserved for future non-encrypted config table


async def create_agent(
    db: AsyncSession,
    *,
    owner_id: str,
    name: str,
    platform: str,
    credentials: dict[str, str],
    ai_mode: str = "off",
    system_prompt: str | None = None,
    activate: bool = False,
) -> Agent:
    registry = get_registry()
    manifest = registry.get_manifest(platform)
    if manifest is None:
        raise ValueError(f"Unknown platform: {platform}")

    pos = ZONE_DEFAULTS.get(manifest.zone, (160.0, 200.0))
    # Offset overlapping agents slightly
    existing = await db.execute(select(Agent).where(Agent.zone == manifest.zone))
    count = len(existing.scalars().all())
    agent = Agent(
        name=name,
        platform=platform,
        zone=manifest.zone,
        pos_x=pos[0] + (count % 4) * 28,
        pos_y=pos[1] + (count // 4) * 24,
        ai_mode=AIMode(ai_mode),
        system_prompt=system_prompt,
        owner_id=owner_id,
        status=AgentStatus.draft,
        is_active=False,
    )
    db.add(agent)
    await db.flush()
    await upsert_secrets(db, agent.id, credentials)
    db.add(AgentLog(agent_id=agent.id, level=LogLevel.info, message="Agent created"))
    await db.commit()

    agent = (
        await db.execute(select(Agent).options(selectinload(Agent.secrets)).where(Agent.id == agent.id))
    ).scalar_one()

    if activate:
        await activate_agent(db, agent)
    return agent


async def activate_agent(db: AsyncSession, agent: Agent) -> Agent:
    credentials = await get_credentials(db, agent.id)
    plugin = get_registry().create(agent.platform, agent.id, credentials)
    agent.status = AgentStatus.connecting
    agent.status_message = "Testing connection..."
    await db.commit()
    await event_hub.publish(
        {
            "type": "agent_status",
            "agent_id": agent.id,
            "status": agent.status.value,
            "status_message": agent.status_message,
            "pos_x": agent.pos_x,
            "pos_y": agent.pos_y,
            "zone": agent.zone,
        }
    )

    result = await plugin.connect()
    if result.ok:
        agent.status = AgentStatus.online
        agent.is_active = True
        agent.status_message = result.message
        db.add(AgentLog(agent_id=agent.id, level=LogLevel.info, message=f"Activated: {result.message}"))
    else:
        agent.status = AgentStatus.error
        agent.is_active = False
        agent.status_message = result.message
        db.add(AgentLog(agent_id=agent.id, level=LogLevel.error, message=f"Activation failed: {result.message}"))
    await db.commit()
    await db.refresh(agent)
    await event_hub.publish(
        {
            "type": "agent_status",
            "agent_id": agent.id,
            "status": agent.status.value,
            "status_message": agent.status_message,
            "pos_x": agent.pos_x,
            "pos_y": agent.pos_y,
            "zone": agent.zone,
        }
    )
    return agent


async def enqueue_command(
    db: AsyncSession,
    *,
    agent_id: str,
    action: str,
    payload: dict[str, Any],
) -> CommandJob:
    agent = await db.get(Agent, agent_id)
    if agent is None:
        raise ValueError("Agent not found")
    job = CommandJob(agent_id=agent_id, action=action, payload=payload, status=JobStatus.pending)
    db.add(job)
    db.add(
        AgentLog(
            agent_id=agent_id,
            level=LogLevel.info,
            message=f"Queued command: {action}",
            meta={"payload_keys": list(payload.keys())},
        )
    )
    await db.commit()
    await db.refresh(job)
    await task_queue.enqueue_job(job.id)
    return job


async def test_connection(platform: str, credentials: dict[str, str]) -> dict[str, Any]:
    plugin = get_registry().create(platform, agent_id="temp", credentials=credentials)
    result = await plugin.connect()
    return {"ok": result.ok, "message": result.message, "meta": result.meta}


async def build_auto_reply(db: AsyncSession, agent: Agent, incoming_text: str) -> str | None:
    if agent.ai_mode == AIMode.off:
        return None
    if agent.ai_mode == AIMode.template:
        result = await db.execute(select(ReplyTemplate).where(ReplyTemplate.agent_id == agent.id))
        templates = list(result.scalars().all())
        chosen = None
        for tpl in templates:
            if tpl.trigger_pattern and re.search(tpl.trigger_pattern, incoming_text, re.I):
                chosen = tpl
                break
        if chosen is None:
            chosen = next((t for t in templates if t.is_default), templates[0] if templates else None)
        if chosen is None:
            return None
        return chosen.body.replace("{{message}}", incoming_text)

    examples_result = await db.execute(select(StyleExample).where(StyleExample.agent_id == agent.id))
    examples = [(e.user_message, e.assistant_reply) for e in examples_result.scalars().all()]
    return await llm_client.generate_reply(
        system_prompt=agent.system_prompt or "Отвечай коротко в стиле владельца аккаунта.",
        examples=examples,
        user_message=incoming_text,
    )


async def record_event(
    db: AsyncSession,
    *,
    agent_id: str,
    event_type: str,
    payload: dict[str, Any],
) -> AgentEvent:
    event = AgentEvent(agent_id=agent_id, type=event_type, payload=payload)
    db.add(event)
    await db.commit()
    await db.refresh(event)
    await event_hub.publish(
        {
            "type": "agent_event",
            "event_id": event.id,
            "agent_id": agent_id,
            "event_type": event_type,
            "payload": payload,
            "created_at": event.created_at.isoformat(),
        }
    )
    return event
