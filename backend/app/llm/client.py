from __future__ import annotations

from typing import Any, Literal

import httpx

from app.core.config import get_settings

Provider = Literal["openai", "anthropic"]


class LLMClient:
    """Multi-provider chat client (OpenAI-compatible + Anthropic Messages API)."""

    async def generate_reply(
        self,
        *,
        system_prompt: str,
        examples: list[tuple[str, str]],
        user_message: str,
        provider: str | None = None,
    ) -> str:
        settings = get_settings()
        chosen: str = (provider or settings.llm_provider or "openai").lower()

        if chosen == "anthropic":
            if settings.anthropic_api_key:
                return await self._anthropic(system_prompt, examples, user_message)
            # Fall through to openai if anthropic key missing
            chosen = "openai"

        if chosen == "openai" and settings.openai_api_key:
            return await self._openai(system_prompt, examples, user_message)

        # Offline deterministic stub for demos without API keys.
        if examples:
            return examples[0][1]
        return f"[llm-stub:{chosen}] {user_message[:200]}"

    async def _openai(
        self,
        system_prompt: str,
        examples: list[tuple[str, str]],
        user_message: str,
    ) -> str:
        settings = get_settings()
        messages: list[dict[str, Any]] = [
            {"role": "system", "content": system_prompt or "You are a helpful assistant."}
        ]
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

    async def _anthropic(
        self,
        system_prompt: str,
        examples: list[tuple[str, str]],
        user_message: str,
    ) -> str:
        settings = get_settings()
        messages: list[dict[str, Any]] = []
        for user, assistant in examples:
            messages.append({"role": "user", "content": user})
            messages.append({"role": "assistant", "content": assistant})
        messages.append({"role": "user", "content": user_message})

        async with httpx.AsyncClient(timeout=45) as client:
            resp = await client.post(
                f"{settings.anthropic_base_url.rstrip('/')}/v1/messages",
                headers={
                    "x-api-key": settings.anthropic_api_key,
                    "anthropic-version": "2023-06-01",
                    "content-type": "application/json",
                },
                json={
                    "model": settings.anthropic_model,
                    "max_tokens": 512,
                    "system": system_prompt or "You are a helpful assistant.",
                    "messages": messages,
                },
            )
            resp.raise_for_status()
            data = resp.json()
        blocks = data.get("content") or []
        texts = [b.get("text", "") for b in blocks if b.get("type") == "text"]
        return "\n".join(texts).strip() or str(data)


llm_client = LLMClient()
