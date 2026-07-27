import asyncio
from pathlib import Path
from unittest.mock import AsyncMock, patch

import pytest
from httpx import ASGITransport, AsyncClient, Response

from app.bootstrap import shutdown, startup
from app.main import app
from app.services.media import MEDIA_DIR, save_bytes
from plugins.instagram.plugin import Plugin as InstagramPlugin


@pytest.fixture
async def client():
    await startup()
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac
    await shutdown()


async def _auth_headers(client: AsyncClient) -> dict[str, str]:
    login = await client.post("/api/auth/login", json={"email": "admin@example.com", "password": "admin123"})
    token = login.json()["access_token"]
    org_id = login.json()["orgs"][0]["id"]
    return {"Authorization": f"Bearer {token}", "X-Org-Id": org_id}


@pytest.mark.asyncio
async def test_media_upload_and_serve(client: AsyncClient):
    h = await _auth_headers(client)
    files = {"file": ("shot.jpg", b"\xff\xd8\xfffakejpeg", "image/jpeg")}
    up = await client.post("/api/media/upload", headers=h, files=files)
    assert up.status_code == 200, up.text
    body = up.json()
    assert body["media_kind"] == "image"
    assert body["public_url"].endswith(body["filename"])
    assert (MEDIA_DIR / body["filename"]).is_file()

    served = await client.get(f"/media/{body['filename']}")
    assert served.status_code == 200
    assert served.content.startswith(b"\xff\xd8\xff")


@pytest.mark.asyncio
async def test_instagram_demo_publish_command(client: AsyncClient):
    h = await _auth_headers(client)
    created = await client.post(
        "/api/agents",
        headers=h,
        json={
            "name": "IG Demo",
            "platform": "instagram",
            "credentials": {"access_token": "demo:local", "ig_user_id": "178410000"},
            "ai_mode": "off",
            "activate": True,
        },
    )
    assert created.status_code == 200, created.text
    agent_id = created.json()["id"]
    assert created.json()["status"] == "online"

    job = await client.post(
        "/api/commands",
        headers=h,
        json={
            "agent_id": agent_id,
            "action": "publish_story",
            "payload": {
                "text": "в историю",
                "media_url": "https://example.com/photo.jpg",
            },
        },
    )
    assert job.status_code == 200, job.text
    await asyncio.sleep(0.4)

    jobs = await client.get("/api/jobs", headers=h)
    assert jobs.status_code == 200
    done = next(j for j in jobs.json() if j["id"] == job.json()["id"])
    assert done["status"] == "done"
    assert "demo" in (done.get("result") or {}).get("message", "").lower() or done["result"].get("demo")


@pytest.mark.asyncio
async def test_instagram_graph_publish_mocked():
    plugin = InstagramPlugin(
        agent_id="ig-1",
        credentials={"access_token": "EAAB-test", "ig_user_id": "17841"},
    )

    responses = [
        # create container
        Response(200, json={"id": "creation-99"}),
        # publish
        Response(200, json={"id": "media-55"}),
    ]

    mock_client = AsyncMock()
    mock_client.__aenter__.return_value = mock_client
    mock_client.__aexit__.return_value = None
    mock_client.post = AsyncMock(side_effect=responses)

    with patch("plugins.instagram.plugin.httpx.AsyncClient", return_value=mock_client):
        result = await plugin.execute_action(
            "publish_post",
            {"media_url": "https://cdn.example.com/a.jpg", "text": "hello"},
        )

    assert result.ok is True
    assert result.data["media_id"] == "media-55"
    assert result.data["creation_id"] == "creation-99"
    assert mock_client.post.await_count == 2


@pytest.mark.asyncio
async def test_instagram_story_video_waits_for_finished():
    plugin = InstagramPlugin(
        agent_id="ig-2",
        credentials={"access_token": "EAAB-test", "ig_user_id": "17841"},
    )

    mock_client = AsyncMock()
    mock_client.__aenter__.return_value = mock_client
    mock_client.__aexit__.return_value = None
    mock_client.post = AsyncMock(
        side_effect=[
            Response(200, json={"id": "creation-video"}),
            Response(200, json={"id": "media-video"}),
        ]
    )
    mock_client.get = AsyncMock(
        side_effect=[
            Response(200, json={"status_code": "IN_PROGRESS"}),
            Response(200, json={"status_code": "FINISHED"}),
        ]
    )

    with (
        patch("plugins.instagram.plugin.httpx.AsyncClient", return_value=mock_client),
        patch("plugins.instagram.plugin.asyncio.sleep", new=AsyncMock()),
    ):
        result = await plugin.execute_action(
            "publish_story",
            {"media_url": "https://cdn.example.com/clip.mp4", "is_video": True},
        )

    assert result.ok is True
    assert result.data["kind"] == "story"
    assert mock_client.get.await_count == 2


def test_save_bytes_kinds(tmp_path: Path, monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setattr("app.services.media.MEDIA_DIR", tmp_path)
    img = save_bytes(b"abc", filename="x.png", content_type="image/png")
    assert img["media_kind"] == "image"
    vid = save_bytes(b"vid", filename="clip.mp4", content_type="video/mp4")
    assert vid["media_kind"] == "video"
