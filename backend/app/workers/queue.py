from __future__ import annotations

import asyncio
import logging
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Protocol

from sqlalchemy import select

from app.agents.registry import get_registry
from app.core.config import get_settings
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


class TaskQueue(Protocol):
    async def start(self) -> None: ...
    async def stop(self) -> None: ...
    async def enqueue_job(self, job_id: str) -> None: ...


@dataclass
class InProcessQueue:
    """MVP async queue. Production can swap to ArqQueue via QUEUE_BACKEND=arq."""

    _queue: asyncio.Queue[str] | None = None
    _worker_task: asyncio.Task[None] | None = None
    _running: bool = False

    async def start(self) -> None:
        if self._running:
            return
        # Recreate queue in the current event loop (important for pytest).
        self._queue = asyncio.Queue()
        self._running = True
        self._worker_task = asyncio.create_task(self._loop(), name="inprocess-queue-worker")

    async def stop(self) -> None:
        self._running = False
        if self._worker_task:
            self._worker_task.cancel()
            try:
                await self._worker_task
            except (asyncio.CancelledError, RuntimeError):
                pass
            self._worker_task = None
        self._queue = None

    async def enqueue_job(self, job_id: str) -> None:
        if self._queue is None:
            await self.start()
        assert self._queue is not None
        await self._queue.put(job_id)

    async def _loop(self) -> None:
        assert self._queue is not None
        queue = self._queue
        while self._running:
            try:
                job_id = await asyncio.wait_for(queue.get(), timeout=1.0)
            except (TimeoutError, asyncio.CancelledError):
                if not self._running:
                    break
                continue
            try:
                await process_command_job(job_id)
            except Exception:  # noqa: BLE001
                logger.exception("Failed processing job %s", job_id)


class ArqQueue:
    """
    Redis/ARQ-backed queue.
    Requires `arq` and a running Redis. Falls back to in-process enqueue of local processing
    if Redis is unavailable at enqueue time.
    """

    def __init__(self, redis_url: str) -> None:
        self.redis_url = redis_url
        self._fallback = InProcessQueue()
        self._redis = None

    async def start(self) -> None:
        try:
            from arq import create_pool
            from arq.connections import RedisSettings

            self._redis = await create_pool(RedisSettings.from_dsn(self.redis_url))
            logger.info("ARQ queue connected to %s", self.redis_url)
        except Exception:  # noqa: BLE001
            logger.warning("ARQ unavailable, using in-process fallback", exc_info=True)
            self._redis = None
            await self._fallback.start()

    async def stop(self) -> None:
        if self._redis is not None:
            await self._redis.close()
            self._redis = None
        await self._fallback.stop()

    async def enqueue_job(self, job_id: str) -> None:
        if self._redis is None:
            await self._fallback.enqueue_job(job_id)
            return
        try:
            await self._redis.enqueue_job("process_command_job_arq", job_id)
        except Exception:  # noqa: BLE001
            logger.exception("ARQ enqueue failed, fallback in-process")
            await self._fallback.start()
            await self._fallback.enqueue_job(job_id)


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


def build_queue() -> TaskQueue:
    settings = get_settings()
    if settings.queue_backend == "arq":
        return ArqQueue(settings.redis_url)
    return InProcessQueue()


task_queue: TaskQueue = build_queue()


async def process_command_job_arq(ctx, job_id: str) -> None:  # noqa: ANN001
    await process_command_job(job_id)
