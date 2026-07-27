from __future__ import annotations

from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse

from app.api.routes import router
from app.bootstrap import shutdown, startup
from app.core.config import get_settings
from app.services.media import MEDIA_DIR, resolve_media_file


@asynccontextmanager
async def lifespan(_: FastAPI):
    MEDIA_DIR.mkdir(parents=True, exist_ok=True)
    await startup()
    yield
    await shutdown()


def create_app() -> FastAPI:
    settings = get_settings()
    app = FastAPI(title=settings.app_name, lifespan=lifespan)
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origin_list or ["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    app.include_router(router, prefix=settings.api_prefix)

    @app.get("/health")
    async def health() -> dict[str, str]:
        return {"status": "ok"}

    @app.get("/media/{filename}")
    async def serve_media(filename: str) -> FileResponse:
        path = resolve_media_file(filename)
        if path is None:
            raise HTTPException(status_code=404, detail="Media not found")
        return FileResponse(path)

    return app


app = create_app()
