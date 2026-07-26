from __future__ import annotations

from datetime import datetime
from typing import Any

from pydantic import BaseModel, EmailStr, Field


class TokenOut(BaseModel):
    access_token: str
    token_type: str = "bearer"
    user_id: str
    email: str
    orgs: list["OrgOut"] = Field(default_factory=list)


class LoginIn(BaseModel):
    email: EmailStr
    password: str


class RegisterIn(BaseModel):
    email: EmailStr
    password: str = Field(min_length=6)
    display_name: str = ""
    org_name: str = Field(min_length=2, max_length=160)


class OrgOut(BaseModel):
    id: str
    name: str
    slug: str
    role: str

    model_config = {"from_attributes": True}


class OrgCreateIn(BaseModel):
    name: str = Field(min_length=2, max_length=160)
    slug: str | None = None


class FieldSpecOut(BaseModel):
    key: str
    label: str
    type: str
    required: bool
    secret: bool
    help: str = ""
    placeholder: str = ""


class PlatformOut(BaseModel):
    platform: str
    title: str
    description: str
    zone: str
    fields: list[FieldSpecOut]
    actions: list[str]


class AgentCreateIn(BaseModel):
    name: str = Field(min_length=1, max_length=120)
    platform: str
    credentials: dict[str, str] = Field(default_factory=dict)
    ai_mode: str = "off"
    llm_provider: str = "openai"
    system_prompt: str | None = None
    activate: bool = False


class AgentUpdateIn(BaseModel):
    name: str | None = None
    credentials: dict[str, str] | None = None
    ai_mode: str | None = None
    llm_provider: str | None = None
    system_prompt: str | None = None
    is_active: bool | None = None
    pos_x: float | None = None
    pos_y: float | None = None


class AgentOut(BaseModel):
    id: str
    org_id: str
    name: str
    platform: str
    status: str
    status_message: str | None
    ai_mode: str
    llm_provider: str
    system_prompt: str | None
    zone: str
    pos_x: float
    pos_y: float
    is_active: bool
    created_at: datetime
    updated_at: datetime
    has_secrets: bool = False

    model_config = {"from_attributes": True}


class TemplateIn(BaseModel):
    name: str
    body: str
    trigger_pattern: str | None = None
    is_default: bool = False


class TemplateOut(BaseModel):
    id: str
    agent_id: str
    name: str
    body: str
    trigger_pattern: str | None
    is_default: bool
    created_at: datetime

    model_config = {"from_attributes": True}


class StyleExampleIn(BaseModel):
    user_message: str
    assistant_reply: str


class StyleExampleOut(BaseModel):
    id: str
    agent_id: str
    user_message: str
    assistant_reply: str
    created_at: datetime

    model_config = {"from_attributes": True}


class EventOut(BaseModel):
    id: str
    agent_id: str
    type: str
    payload: dict[str, Any]
    notified: bool
    created_at: datetime

    model_config = {"from_attributes": True}


class LogOut(BaseModel):
    id: str
    agent_id: str | None
    level: str
    message: str
    meta: dict[str, Any] | None
    created_at: datetime

    model_config = {"from_attributes": True}


class CommandIn(BaseModel):
    agent_id: str
    action: str
    payload: dict[str, Any] = Field(default_factory=dict)


class JobOut(BaseModel):
    id: str
    agent_id: str
    action: str
    payload: dict[str, Any]
    status: str
    result: dict[str, Any] | None
    error: str | None
    created_at: datetime
    finished_at: datetime | None

    model_config = {"from_attributes": True}


class TestConnectionIn(BaseModel):
    platform: str
    credentials: dict[str, str]


class SceneAgentOut(BaseModel):
    id: str
    name: str
    platform: str
    status: str
    status_message: str | None
    zone: str
    pos_x: float
    pos_y: float
    is_active: bool


class SimulateEventIn(BaseModel):
    type: str = Field(description="message | like | follow | unfollow | comment")
    payload: dict[str, Any] = Field(default_factory=dict)
    auto_reply: bool = True


class NotificationTargetIn(BaseModel):
    channel: str = "telegram"
    address: str
    is_active: bool = True


class NotificationTargetOut(BaseModel):
    id: str
    org_id: str
    channel: str
    address: str
    is_active: bool
    created_at: datetime

    model_config = {"from_attributes": True}


class AutoReplyPreviewIn(BaseModel):
    message: str


class AutoReplyPreviewOut(BaseModel):
    reply: str | None
    mode: str


TokenOut.model_rebuild()
