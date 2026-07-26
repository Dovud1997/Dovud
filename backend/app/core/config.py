from functools import lru_cache
from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict

BACKEND_ROOT = Path(__file__).resolve().parents[2]
DATA_DIR = BACKEND_ROOT / "data"
DATA_DIR.mkdir(parents=True, exist_ok=True)


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    app_name: str = "AI Agents Platform"
    api_prefix: str = "/api"
    database_url: str = f"sqlite+aiosqlite:///{DATA_DIR / 'platform.db'}"
    secret_key: str = "dev-change-me-jwt-secret-key-32chars!!"
    secret_encryption_key: str = ""
    access_token_expire_minutes: int = 60 * 24
    admin_email: str = "admin@example.com"
    admin_password: str = "admin123"
    admin_org_name: str = "DOVUD HQ"
    admin_org_slug: str = "dovud"
    cors_origins: str = "http://localhost:5173,http://127.0.0.1:5173"

    # LLM
    llm_provider: str = "openai"  # openai | anthropic
    openai_api_key: str = ""
    openai_base_url: str = "https://api.openai.com/v1"
    openai_model: str = "gpt-4o-mini"
    anthropic_api_key: str = ""
    anthropic_base_url: str = "https://api.anthropic.com"
    anthropic_model: str = "claude-sonnet-4-20250514"

    control_bot_token: str = ""
    plugins_dir: str = str(BACKEND_ROOT / "plugins")
    queue_backend: str = "inprocess"  # inprocess | arq
    redis_url: str = "redis://127.0.0.1:6379/0"
    public_base_url: str = "http://127.0.0.1:8000"
    telegram_listen_enabled: bool = True
    telegram_webhook_secret: str = ""  # optional X-Telegram-Bot-Api-Secret-Token

    @property
    def cors_origin_list(self) -> list[str]:
        return [o.strip() for o in self.cors_origins.split(",") if o.strip()]


@lru_cache
def get_settings() -> Settings:
    return Settings()
