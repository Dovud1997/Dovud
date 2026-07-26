from __future__ import annotations

import asyncio
import logging
from collections.abc import Awaitable, Callable
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any

from sqlalchemy import select

from app.agents.registry import get_registry
from app.core.db import SessionLocal
from app.core.security import decrypt_secret
from app.models.entities import (
    Agent,
    AgentEvent,
    AgentLog,
    AgentSecret,
    AgentStatus,
    CommandJob,
    JobStatus,
    LogLevel,
)
from app.services.events import event_hub

logger = logging.getLogger(__name__)

JobHandler = Callable[[str], Awaitable[None]]


@dataclass
class InProcessQueue:
    """MVP async queue. Swap for ARQ/Redis in production via same enqueue API."""

    _queue: asyncio.Queue[str] = field(default_factory=asyncio.Queue)
    _worker_task: asyncio.Task[None] | None = None
    _running: bool = False

    async def start(self) -> None:
        if self._running:
            return
        self._running = True
        self._worker_task = asyncio.create_task(self._loop(), name="inprocess-queue-worker")

    async def stop(self) -> None:
        self._running = False
        if self._worker_task:
            self._worker_task.cancel()
            try:
                await self._worker_task
            except asyncio.CancelledError:
                pass
            self._worker_task = None

    async def enqueue_job(self, job_id: str) -> None:
        await self._queue.put(job_id)

    async def _loop(self) -> None:
        while self._running:
            try:
                job_id = await asyncio.wait_for(self._queue.get(), timeout=1.0)
            except TimeoutError:
                continue
            try:
                await process_command_job(job_id)
            except Exception:  # noqa: BLE001
                logger.exception("Failed processing job %s", job_id)


task_queue = InProcessQueue()


async def process_command_job(job_id: str) -> None:
    async with SessionLocal() as db:
        job = await db.get(CommandJob, job_id)
        if job is None:
            return
        agent = await db.get(Agent, job.agent_id)
        if agent is None:
            job.status = JobStatus.failed
            job.error = "Agent not found"
            job.finished_at = datetime.now(timezone.utc)
            await db.commit()
            return

        job.status = JobStatus.running
        agent.status = AgentStatus.busy
        agent.status_message = f"Running {job.action}"
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

        secrets = (
            await db.execute(select(AgentSecret).where(AgentSecret.agent_id == agent.id))
        ).scalars().all()
        credentials = {s.key: decrypt_secret(s.value_encrypted) for s in secrets}

        try:
            plugin = get_registry().create(agent.platform, agent.id, credentials)
            result = await plugin.execute_action(job.action, job.payload or {})
            job.status = JobStatus.done if result.ok else JobStatus.failed
            job.result = {"message": result.message, **(result.data or {})}
            job.error = None if result.ok else result.message
            agent.status = AgentStatus.online if result.ok else AgentStatus.error
            agent.status_message = result.message
            db.add(
                AgentLog(
                    agent_id=agent.id,
                    level=LogLevel.info if result.ok else LogLevel.error,
                    message=f"{job.action}: {result.message}",
                    meta={"job_id": job.id, "result": result.data},
                )
            )
            db.add(
                AgentEvent(
                    agent_id=agent.id,
                    type="action_result",
                    payload={"action": job.action, "ok": result.ok, "message": result.message},
                )
            )
        except Exception as exc:  # noqa: BLE001
            job.status = JobStatus.failed
            job.error = str(exc)
            agent.status = AgentStatus.error
            agent.status_message = str(exc)
            db.add(
                AgentLog(
                    agent_id=agent.id,
                    level=LogLevel.error,
                    message=f"{job.action} failed: {exc}",
                    meta={"job_id": job.id},
                )
            )
        finally:
            job.finished_at = datetime.now(timezone.utc)
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
        await event_hub.publish({"type": "job_update", "job_id": job.id, "status": job.status.value})


async def load_agent_credentials(agent_id: str) -> dict[str, str]:
    async with SessionLocal() as db:
        result = await db.execute(select(Agent).where(Agent.id == agent_id))
        agent = result.scalar_one_or_none()
        if agent is None:
            return {}
        await db.refresh(agent, attribute_names=["secrets"])
        return {s.key: decrypt_secret(s.value_encrypted) for s in agent.secrets}
