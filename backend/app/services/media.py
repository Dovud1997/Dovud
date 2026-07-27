from __future__ import annotations

import mimetypes
import uuid
from pathlib import Path

from app.core.config import BACKEND_ROOT, get_settings

MEDIA_DIR = BACKEND_ROOT / "media"
MEDIA_DIR.mkdir(parents=True, exist_ok=True)

ALLOWED_EXTENSIONS = {
    ".jpg",
    ".jpeg",
    ".png",
    ".gif",
    ".webp",
    ".mp4",
    ".mov",
    ".m4v",
}


def _safe_ext(filename: str | None, content_type: str | None) -> str:
    if filename:
        ext = Path(filename).suffix.lower()
        if ext in ALLOWED_EXTENSIONS:
            return ext
    if content_type:
        guessed = mimetypes.guess_extension(content_type.split(";")[0].strip())
        if guessed == ".jpe":
            guessed = ".jpg"
        if guessed and guessed.lower() in ALLOWED_EXTENSIONS:
            return guessed.lower()
    return ".bin"


def is_video_path(path: str | Path) -> bool:
    return Path(path).suffix.lower() in {".mp4", ".mov", ".m4v"}


def is_video_url(url: str) -> bool:
    clean = url.split("?", 1)[0].lower()
    return any(clean.endswith(ext) for ext in (".mp4", ".mov", ".m4v"))


def save_bytes(
    data: bytes,
    *,
    filename: str | None = None,
    content_type: str | None = None,
) -> dict[str, str]:
    if not data:
        raise ValueError("Empty media payload")
    ext = _safe_ext(filename, content_type)
    media_id = uuid.uuid4().hex
    stored_name = f"{media_id}{ext}"
    path = MEDIA_DIR / stored_name
    path.write_bytes(data)
    settings = get_settings()
    public_url = f"{settings.public_base_url.rstrip('/')}/media/{stored_name}"
    return {
        "id": media_id,
        "filename": stored_name,
        "content_type": content_type or mimetypes.guess_type(stored_name)[0] or "application/octet-stream",
        "public_url": public_url,
        "path": str(path),
        "media_kind": "video" if is_video_path(path) else "image",
    }


def resolve_media_file(filename: str) -> Path | None:
    # Prevent path traversal — only basename under MEDIA_DIR.
    name = Path(filename).name
    if name != filename or ".." in filename:
        return None
    path = MEDIA_DIR / name
    if not path.is_file():
        return None
    return path
