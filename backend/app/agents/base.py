from __future__ import annotations

from abc import ABC, abstractmethod
from collections.abc import AsyncIterator
from dataclasses import dataclass, field
from typing import Any


@dataclass
class FieldSpec:
    key: str
    label: str
    type: str = "text"  # text | password | textarea | url
    required: bool = True
    secret: bool = False
    help: str = ""
    placeholder: str = ""


@dataclass
class PluginManifest:
    platform: str
    title: str
    description: str
    zone: str
    fields: list[FieldSpec]
    actions: list[str] = field(default_factory=list)


@dataclass
class ConnectResult:
    ok: bool
    message: str
    meta: dict[str, Any] = field(default_factory=dict)


@dataclass
class ActionResult:
    ok: bool
    message: str
    data: dict[str, Any] = field(default_factory=dict)


@dataclass
class AgentStatusDTO:
    status: str
    message: str = ""
    details: dict[str, Any] = field(default_factory=dict)


@dataclass
class AgentEventDTO:
    type: str
    payload: dict[str, Any]
    summary: str = ""


class BaseAgentPlugin(ABC):
    """Единый интерфейс всех платформенных агентов-плагинов."""

    manifest: PluginManifest

    def __init__(self, agent_id: str, credentials: dict[str, str]):
        self.agent_id = agent_id
        self.credentials = credentials

    @abstractmethod
    async def connect(self) -> ConnectResult:
        ...

    @abstractmethod
    async def execute_action(self, action: str, payload: dict[str, Any]) -> ActionResult:
        ...

    @abstractmethod
    async def listen_events(self) -> AsyncIterator[AgentEventDTO]:
        ...

    @abstractmethod
    async def get_status(self) -> AgentStatusDTO:
        ...

    @abstractmethod
    async def disconnect(self) -> None:
        ...
