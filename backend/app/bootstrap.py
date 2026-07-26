from __future__ import annotations

from sqlalchemy import select

from app.agents.registry import get_registry
from app.core.config import get_settings
from app.core.db import SessionLocal, init_db
from app.core.security import hash_password
from app.models.entities import User
from app.workers.queue import task_queue


async def startup() -> None:
    await init_db()
    get_registry().discover()
    settings = get_settings()
    async with SessionLocal() as db:
        result = await db.execute(select(User).where(User.email == settings.admin_email))
        user = result.scalar_one_or_none()
        if user is None:
            db.add(
                User(
                    email=settings.admin_email,
                    password_hash=hash_password(settings.admin_password),
                )
            )
            await db.commit()
    await task_queue.start()


async def shutdown() -> None:
    await task_queue.stop()
