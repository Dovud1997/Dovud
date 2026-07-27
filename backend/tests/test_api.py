import asyncio
import uuid

import pytest
from httpx import ASGITransport, AsyncClient

from app.bootstrap import shutdown, startup
from app.main import app


@pytest.fixture
async def client():
    await startup()
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac
    await shutdown()


@pytest.mark.asyncio
async def test_health_and_login(client: AsyncClient):
    r = await client.get("/health")
    assert r.status_code == 200
    r = await client.post("/api/auth/login", json={"email": "admin@example.com", "password": "admin123"})
    assert r.status_code == 200
    body = r.json()
    assert body["access_token"]
    assert body["orgs"]


@pytest.mark.asyncio
async def test_plugins_and_demo_agent_flow(client: AsyncClient):
    login = await client.post("/api/auth/login", json={"email": "admin@example.com", "password": "admin123"})
    token = login.json()["access_token"]
    org_id = login.json()["orgs"][0]["id"]
    h = {"Authorization": f"Bearer {token}", "X-Org-Id": org_id}

    platforms = await client.get("/api/platforms", headers=h)
    names = {p["platform"] for p in platforms.json()}
    assert {"telegram", "instagram", "youtube"} <= names

    created = await client.post(
        "/api/agents",
        headers=h,
        json={
            "name": "TG Demo",
            "platform": "telegram",
            "credentials": {"bot_token": "demo:local"},
            "ai_mode": "template",
            "activate": True,
        },
    )
    assert created.status_code == 200, created.text
    agent = created.json()
    assert agent["status"] == "online"
    agent_id = agent["id"]

    tpl = await client.post(
        f"/api/agents/{agent_id}/templates",
        headers=h,
        json={"name": "hi", "body": "Привет! {{message}}", "is_default": True},
    )
    assert tpl.status_code == 200

    preview = await client.post(
        f"/api/agents/{agent_id}/auto-reply/preview",
        headers=h,
        json={"message": "как дела"},
    )
    assert preview.status_code == 200
    assert "как дела" in (preview.json()["reply"] or "")

    ev = await client.post(
        f"/api/agents/{agent_id}/simulate-event",
        headers=h,
        json={"type": "message", "payload": {"text": "как дела"}, "auto_reply": True},
    )
    assert ev.status_code == 200
    await asyncio.sleep(0.5)

    jobs = await client.get("/api/jobs", headers=h)
    assert jobs.status_code == 200
    assert any(j["action"] == "reply" for j in jobs.json())


@pytest.mark.asyncio
async def test_register_multi_tenant(client: AsyncClient):
    email = f"studio-{uuid.uuid4().hex[:8]}@example.com"
    r = await client.post(
        "/api/auth/register",
        json={"email": email, "password": "secret12", "org_name": "Studio One"},
    )
    assert r.status_code == 200, r.text
    assert r.json()["orgs"][0]["name"] == "Studio One"


@pytest.mark.asyncio
async def test_telegram_webhook_inbound(client: AsyncClient):
    login = await client.post("/api/auth/login", json={"email": "admin@example.com", "password": "admin123"})
    token = login.json()["access_token"]
    org_id = login.json()["orgs"][0]["id"]
    h = {"Authorization": f"Bearer {token}", "X-Org-Id": org_id}

    created = await client.post(
        "/api/agents",
        headers=h,
        json={
            "name": "TG Hook",
            "platform": "telegram",
            "credentials": {"bot_token": "demo:hook"},
            "ai_mode": "off",
            "activate": True,
        },
    )
    assert created.status_code == 200, created.text
    agent_id = created.json()["id"]

    # Public webhook — no JWT
    hook = await client.post(
        f"/api/webhooks/telegram/{agent_id}",
        json={
            "update_id": 1001,
            "message": {
                "message_id": 7,
                "text": "webhook hello",
                "chat": {"id": 42},
                "from": {"id": 1, "username": "hookuser"},
            },
        },
    )
    assert hook.status_code == 200, hook.text
    body = hook.json()
    assert body["ok"] is True
    assert body["type"] == "message"

    status = await client.get("/api/telegram/listener-status", headers=h)
    assert status.status_code == 200
    assert status.json()["enabled"] is True
