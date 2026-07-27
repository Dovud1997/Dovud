from __future__ import annotations

from pathlib import Path

from sqlalchemy import select, text

from app.agents.registry import get_registry
from app.core.config import DATA_DIR, get_settings
from app.core.db import SessionLocal, engine, init_db
from app.core.security import hash_password
from app.models.entities import MemberRole, Membership, Organization, User
from app.workers.queue import task_queue
from app.workers.telegram_listener import telegram_listener


async def _sqlite_needs_reset() -> bool:
    """Dev helper: recreate SQLite DB if schema is from pre-tenant MVP."""
    settings = get_settings()
    if not settings.database_url.startswith("sqlite"):
        return False
    db_path = Path(settings.database_url.split("///")[-1])
    if not db_path.exists():
        return False
    async with engine.connect() as conn:
        result = await conn.execute(text("SELECT name FROM sqlite_master WHERE type='table' AND name='organizations'"))
        return result.first() is None


async def startup() -> None:
    if await _sqlite_needs_reset():
        for path in DATA_DIR.glob("platform.db*"):
            path.unlink(missing_ok=True)
        await engine.dispose()

    await init_db()
    # Reset registry each boot so plugin hot-add works in reload.
    from app.agents import registry as registry_mod

    registry_mod._registry = None
    get_registry().discover()

    settings = get_settings()
    async with SessionLocal() as db:
        result = await db.execute(select(User).where(User.email == settings.admin_email))
        user = result.scalar_one_or_none()
        if user is None:
            user = User(
                email=settings.admin_email,
                password_hash=hash_password(settings.admin_password),
                display_name="Admin",
            )
            db.add(user)
            await db.flush()

        org_result = await db.execute(select(Organization).where(Organization.slug == settings.admin_org_slug))
        org = org_result.scalar_one_or_none()
        if org is None:
            org = Organization(name=settings.admin_org_name, slug=settings.admin_org_slug)
            db.add(org)
            await db.flush()

        mem = await db.execute(
            select(Membership).where(Membership.user_id == user.id, Membership.org_id == org.id)
        )
        if mem.scalar_one_or_none() is None:
            db.add(Membership(user_id=user.id, org_id=org.id, role=MemberRole.owner))
        await db.commit()

    await task_queue.start()
    await telegram_listener.start()


async def shutdown() -> None:
    await telegram_listener.stop()
    await task_queue.stop()
