from __future__ import annotations

import re

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.models.entities import MemberRole, Membership, Organization, User


def slugify(value: str) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
    return slug[:80] or "org"


async def list_user_orgs(db: AsyncSession, user_id: str) -> list[dict]:
    result = await db.execute(
        select(Membership)
        .options(selectinload(Membership.organization))
        .where(Membership.user_id == user_id)
    )
    rows = result.scalars().all()
    return [
        {
            "id": m.organization.id,
            "name": m.organization.name,
            "slug": m.organization.slug,
            "role": m.role.value,
        }
        for m in rows
    ]


async def create_org(
    db: AsyncSession,
    *,
    user: User,
    name: str,
    slug: str | None = None,
) -> Organization:
    base = slugify(slug or name)
    candidate = base
    i = 1
    while True:
        exists = await db.execute(select(Organization).where(Organization.slug == candidate))
        if exists.scalar_one_or_none() is None:
            break
        i += 1
        candidate = f"{base}-{i}"

    org = Organization(name=name, slug=candidate)
    db.add(org)
    await db.flush()
    db.add(Membership(user_id=user.id, org_id=org.id, role=MemberRole.owner))
    await db.commit()
    await db.refresh(org)
    return org


async def ensure_membership(db: AsyncSession, user_id: str, org_id: str) -> Membership:
    result = await db.execute(
        select(Membership).where(Membership.user_id == user_id, Membership.org_id == org_id)
    )
    membership = result.scalar_one_or_none()
    if membership is None:
        raise PermissionError("Not a member of this organization")
    return membership
