from __future__ import annotations

import asyncio
from collections.abc import AsyncIterator
from typing import Any


class EventHub:
    """In-process pub/sub for WebSocket clients and internal listeners."""

    def __init__(self) -> None:
        self._subscribers: set[asyncio.Queue[dict[str, Any]]] = set()
        self._lock = asyncio.Lock()

    async def publish(self, event: dict[str, Any]) -> None:
        async with self._lock:
            dead: list[asyncio.Queue[dict[str, Any]]] = []
            for q in self._subscribers:
                try:
                    q.put_nowait(event)
                except asyncio.QueueFull:
                    dead.append(q)
            for q in dead:
                self._subscribers.discard(q)

    async def subscribe(self) -> AsyncIterator[dict[str, Any]]:
        queue: asyncio.Queue[dict[str, Any]] = asyncio.Queue(maxsize=200)
        async with self._lock:
            self._subscribers.add(queue)
        try:
            while True:
                item = await queue.get()
                yield item
        finally:
            async with self._lock:
                self._subscribers.discard(queue)


event_hub = EventHub()
