from __future__ import annotations

from typing import Any

import httpx

from app.core.config import get_settings


class LLMClient:
    """OpenAI-compatible chat completions client."""

    async def generate_reply(
        self,
        *,
        system_prompt: str,
        examples: list[tuple[str, str]],
        user_message: str,
    ) -> str:
        settings = get_settings()
        if not settings.openai_api_key:
            # Deterministic offline fallback for MVP demos without API key.
            if examples:
                return examples[0][1]
            return f"[llm-stub] {user_message[:200]}"

        messages: list[dict[str, Any]] = [{"role": "system", "content": system_prompt or "You are a helpful assistant."}]
        for user, assistant in examples:
            messages.append({"role": "user", "content": user})
            messages.append({"role": "assistant", "content": assistant})
        messages.append({"role": "user", "content": user_message})

        async with httpx.AsyncClient(timeout=45) as client:
            resp = await client.post(
                f"{settings.openai_base_url.rstrip('/')}/chat/completions",
                headers={"Authorization": f"Bearer {settings.openai_api_key}"},
                json={"model": settings.openai_model, "messages": messages, "temperature": 0.7},
            )
            resp.raise_for_status()
            data = resp.json()
        return data["choices"][0]["message"]["content"]


llm_client = LLMClient()
