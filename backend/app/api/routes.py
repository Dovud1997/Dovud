from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, WebSocket, WebSocketDisconnect
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.agents.registry import get_registry
from app.api.deps import get_current_org, get_current_user
from app.core.db import SessionLocal, get_db
from app.core.security import create_access_token, hash_password, verify_password
from app.models.entities import (
    AIMode,
    Agent,
    AgentEvent,
    AgentLog,
    AgentStatus,
    CommandJob,
    Membership,
    NotificationTarget,
    Organization,
    ReplyTemplate,
    StyleExample,
    User,
)
from app.schemas import (
    AgentCreateIn,
    AgentOut,
    AgentUpdateIn,
    AutoReplyPreviewIn,
    AutoReplyPreviewOut,
    CommandIn,
    EventOut,
    JobOut,
    LoginIn,
    LogOut,
    NotificationTargetIn,
    NotificationTargetOut,
    OrgCreateIn,
    OrgOut,
    PlatformOut,
    RegisterIn,
    SceneAgentOut,
    SimulateEventIn,
    StyleExampleIn,
    StyleExampleOut,
    TemplateIn,
    TemplateOut,
    TestConnectionIn,
    TokenOut,
)
from app.services import agents as agent_service
from app.services import orgs as org_service
from app.services.events import event_hub

router = APIRouter()


async def _token_for_user(db: AsyncSession, user: User) -> TokenOut:
    orgs = await org_service.list_user_orgs(db, user.id)
    return TokenOut(
        access_token=create_access_token(user.id, {"email": user.email}),
        user_id=user.id,
        email=user.email,
        orgs=[OrgOut(**o) for o in orgs],
    )


@router.post("/auth/register", response_model=TokenOut)
async def register(body: RegisterIn, db: AsyncSession = Depends(get_db)) -> TokenOut:
    exists = await db.execute(select(User).where(User.email == body.email))
    if exists.scalar_one_or_none():
        raise HTTPException(status_code=400, detail="Email already registered")
    user = User(
        email=body.email,
        password_hash=hash_password(body.password),
        display_name=body.display_name or body.email.split("@")[0],
    )
    db.add(user)
    await db.flush()
    await org_service.create_org(db, user=user, name=body.org_name)
    return await _token_for_user(db, user)


@router.post("/auth/login", response_model=TokenOut)
async def login(body: LoginIn, db: AsyncSession = Depends(get_db)) -> TokenOut:
    result = await db.execute(select(User).where(User.email == body.email))
    user = result.scalar_one_or_none()
    if user is None or not verify_password(body.password, user.password_hash):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    return await _token_for_user(db, user)


@router.get("/me/orgs", response_model=list[OrgOut])
async def my_orgs(db: AsyncSession = Depends(get_db), user: User = Depends(get_current_user)) -> list[OrgOut]:
    return [OrgOut(**o) for o in await org_service.list_user_orgs(db, user.id)]


@router.post("/orgs", response_model=OrgOut)
async def create_org(
    body: OrgCreateIn,
    db: AsyncSession = Depends(get_db),
    user: User = Depends(get_current_user),
) -> OrgOut:
    org = await org_service.create_org(db, user=user, name=body.name, slug=body.slug)
    return OrgOut(id=org.id, name=org.name, slug=org.slug, role="owner")


@router.get("/platforms", response_model=list[PlatformOut])
async def list_platforms(_: User = Depends(get_current_user)) -> list[PlatformOut]:
    manifests = get_registry().list_manifests()
    return [
        PlatformOut(
            platform=m.platform,
            title=m.title,
            description=m.description,
            zone=m.zone,
            fields=[
                {
                    "key": f.key,
                    "label": f.label,
                    "type": f.type,
                    "required": f.required,
                    "secret": f.secret,
                    "help": f.help,
                    "placeholder": f.placeholder,
                }
                for f in m.fields
            ],
            actions=m.actions,
        )
        for m in manifests
    ]


@router.post("/agents/test-connection")
async def test_connection(body: TestConnectionIn, _: User = Depends(get_current_user)) -> dict:
    try:
        return await agent_service.test_connection(body.platform, body.credentials)
    except KeyError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc


@router.get("/agents", response_model=list[AgentOut])
async def list_agents(
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[AgentOut]:
    result = await db.execute(
        select(Agent)
        .options(selectinload(Agent.secrets))
        .where(Agent.org_id == org.id)
        .order_by(Agent.created_at)
    )
    return [AgentOut(**agent_service.agent_to_dict(a)) for a in result.scalars().all()]


@router.post("/agents", response_model=AgentOut)
async def create_agent(
    body: AgentCreateIn,
    db: AsyncSession = Depends(get_db),
    user: User = Depends(get_current_user),
    org: Organization = Depends(get_current_org),
) -> AgentOut:
    try:
        agent = await agent_service.create_agent(
            db,
            org_id=org.id,
            created_by=user.id,
            name=body.name,
            platform=body.platform,
            credentials=body.credentials,
            ai_mode=body.ai_mode,
            llm_provider=body.llm_provider,
            system_prompt=body.system_prompt,
            activate=body.activate,
        )
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    agent = (
        await db.execute(select(Agent).options(selectinload(Agent.secrets)).where(Agent.id == agent.id))
    ).scalar_one()
    return AgentOut(**agent_service.agent_to_dict(agent))


@router.get("/agents/{agent_id}", response_model=AgentOut)
async def get_agent(
    agent_id: str,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> AgentOut:
    agent = await _org_agent(db, org.id, agent_id)
    return AgentOut(**agent_service.agent_to_dict(agent))


@router.patch("/agents/{agent_id}", response_model=AgentOut)
async def update_agent(
    agent_id: str,
    body: AgentUpdateIn,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> AgentOut:
    agent = await _org_agent(db, org.id, agent_id)
    if body.name is not None:
        agent.name = body.name
    if body.ai_mode is not None:
        agent.ai_mode = AIMode(body.ai_mode)
    if body.llm_provider is not None:
        agent.llm_provider = body.llm_provider
    if body.system_prompt is not None:
        agent.system_prompt = body.system_prompt
    if body.pos_x is not None:
        agent.pos_x = body.pos_x
    if body.pos_y is not None:
        agent.pos_y = body.pos_y
    if body.credentials:
        await agent_service.upsert_secrets(db, agent.id, body.credentials)
    await db.commit()

    if body.is_active is True:
        agent = await agent_service.activate_agent(db, agent)
    elif body.is_active is False:
        agent.is_active = False
        agent.status = AgentStatus.offline
        agent.status_message = "Deactivated"
        await db.commit()

    agent = (
        await db.execute(select(Agent).options(selectinload(Agent.secrets)).where(Agent.id == agent.id))
    ).scalar_one()
    return AgentOut(**agent_service.agent_to_dict(agent))


@router.post("/agents/{agent_id}/activate", response_model=AgentOut)
async def activate_agent(
    agent_id: str,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> AgentOut:
    agent = await _org_agent(db, org.id, agent_id)
    agent = await agent_service.activate_agent(db, agent)
    agent = (
        await db.execute(select(Agent).options(selectinload(Agent.secrets)).where(Agent.id == agent.id))
    ).scalar_one()
    return AgentOut(**agent_service.agent_to_dict(agent))


@router.post("/agents/{agent_id}/simulate-event", response_model=EventOut)
async def simulate_event(
    agent_id: str,
    body: SimulateEventIn,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> EventOut:
    agent = await _org_agent(db, org.id, agent_id)
    event = await agent_service.record_event(
        db,
        agent=agent,
        event_type=body.type,
        payload=body.payload or {"text": "Привет! Это тестовое сообщение"},
        auto_reply=body.auto_reply,
    )
    return EventOut.model_validate(event)


@router.post("/agents/{agent_id}/auto-reply/preview", response_model=AutoReplyPreviewOut)
async def preview_auto_reply(
    agent_id: str,
    body: AutoReplyPreviewIn,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> AutoReplyPreviewOut:
    agent = await _org_agent(db, org.id, agent_id)
    reply = await agent_service.build_auto_reply(db, agent, body.message)
    return AutoReplyPreviewOut(reply=reply, mode=agent.ai_mode.value)


@router.post("/commands", response_model=JobOut)
async def create_command(
    body: CommandIn,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> JobOut:
    await _org_agent(db, org.id, body.agent_id)
    job = await agent_service.enqueue_command(
        db, agent_id=body.agent_id, action=body.action, payload=body.payload
    )
    return JobOut.model_validate(job)


@router.get("/jobs", response_model=list[JobOut])
async def list_jobs(
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[JobOut]:
    result = await db.execute(
        select(CommandJob)
        .join(Agent, Agent.id == CommandJob.agent_id)
        .where(Agent.org_id == org.id)
        .order_by(CommandJob.created_at.desc())
        .limit(100)
    )
    return [JobOut.model_validate(j) for j in result.scalars().all()]


@router.get("/events", response_model=list[EventOut])
async def list_events(
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[EventOut]:
    result = await db.execute(
        select(AgentEvent)
        .join(Agent, Agent.id == AgentEvent.agent_id)
        .where(Agent.org_id == org.id)
        .order_by(AgentEvent.created_at.desc())
        .limit(100)
    )
    return [EventOut.model_validate(e) for e in result.scalars().all()]


@router.get("/logs", response_model=list[LogOut])
async def list_logs(
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[LogOut]:
    result = await db.execute(
        select(AgentLog)
        .join(Agent, Agent.id == AgentLog.agent_id)
        .where(Agent.org_id == org.id)
        .order_by(AgentLog.created_at.desc())
        .limit(150)
    )
    return [LogOut.model_validate(log) for log in result.scalars().all()]


@router.get("/scene", response_model=list[SceneAgentOut])
async def scene_state(
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[SceneAgentOut]:
    result = await db.execute(select(Agent).where(Agent.org_id == org.id))
    return [
        SceneAgentOut(
            id=a.id,
            name=a.name,
            platform=a.platform,
            status=a.status.value,
            status_message=a.status_message,
            zone=a.zone,
            pos_x=a.pos_x,
            pos_y=a.pos_y,
            is_active=a.is_active,
        )
        for a in result.scalars().all()
    ]


@router.get("/agents/{agent_id}/templates", response_model=list[TemplateOut])
async def list_templates(
    agent_id: str,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[TemplateOut]:
    await _org_agent(db, org.id, agent_id)
    result = await db.execute(select(ReplyTemplate).where(ReplyTemplate.agent_id == agent_id))
    return [TemplateOut.model_validate(t) for t in result.scalars().all()]


@router.post("/agents/{agent_id}/templates", response_model=TemplateOut)
async def create_template(
    agent_id: str,
    body: TemplateIn,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> TemplateOut:
    await _org_agent(db, org.id, agent_id)
    tpl = ReplyTemplate(
        agent_id=agent_id,
        name=body.name,
        body=body.body,
        trigger_pattern=body.trigger_pattern,
        is_default=body.is_default,
    )
    db.add(tpl)
    await db.commit()
    await db.refresh(tpl)
    return TemplateOut.model_validate(tpl)


@router.delete("/agents/{agent_id}/templates/{template_id}")
async def delete_template(
    agent_id: str,
    template_id: str,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> dict:
    await _org_agent(db, org.id, agent_id)
    tpl = await db.get(ReplyTemplate, template_id)
    if tpl is None or tpl.agent_id != agent_id:
        raise HTTPException(status_code=404, detail="Template not found")
    await db.delete(tpl)
    await db.commit()
    return {"ok": True}


@router.get("/agents/{agent_id}/style-examples", response_model=list[StyleExampleOut])
async def list_style_examples(
    agent_id: str,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[StyleExampleOut]:
    await _org_agent(db, org.id, agent_id)
    result = await db.execute(select(StyleExample).where(StyleExample.agent_id == agent_id))
    return [StyleExampleOut.model_validate(e) for e in result.scalars().all()]


@router.post("/agents/{agent_id}/style-examples", response_model=StyleExampleOut)
async def create_style_example(
    agent_id: str,
    body: StyleExampleIn,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> StyleExampleOut:
    await _org_agent(db, org.id, agent_id)
    ex = StyleExample(agent_id=agent_id, user_message=body.user_message, assistant_reply=body.assistant_reply)
    db.add(ex)
    await db.commit()
    await db.refresh(ex)
    return StyleExampleOut.model_validate(ex)


@router.delete("/agents/{agent_id}/style-examples/{example_id}")
async def delete_style_example(
    agent_id: str,
    example_id: str,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> dict:
    await _org_agent(db, org.id, agent_id)
    ex = await db.get(StyleExample, example_id)
    if ex is None or ex.agent_id != agent_id:
        raise HTTPException(status_code=404, detail="Example not found")
    await db.delete(ex)
    await db.commit()
    return {"ok": True}


@router.get("/notifications/targets", response_model=list[NotificationTargetOut])
async def list_notification_targets(
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> list[NotificationTargetOut]:
    result = await db.execute(select(NotificationTarget).where(NotificationTarget.org_id == org.id))
    return [NotificationTargetOut.model_validate(t) for t in result.scalars().all()]


@router.post("/notifications/targets", response_model=NotificationTargetOut)
async def create_notification_target(
    body: NotificationTargetIn,
    db: AsyncSession = Depends(get_db),
    org: Organization = Depends(get_current_org),
) -> NotificationTargetOut:
    target = NotificationTarget(
        org_id=org.id,
        channel=body.channel,
        address=body.address,
        is_active=body.is_active,
    )
    db.add(target)
    await db.commit()
    await db.refresh(target)
    return NotificationTargetOut.model_validate(target)


@router.websocket("/ws/scene")
async def ws_scene(websocket: WebSocket) -> None:
    token = websocket.query_params.get("token")
    org_id = websocket.query_params.get("org_id")
    if not token:
        await websocket.close(code=4401)
        return
    from app.core.security import decode_access_token

    try:
        payload = decode_access_token(token)
    except ValueError:
        await websocket.close(code=4401)
        return

    await websocket.accept()
    async with SessionLocal() as db:
        user_id = payload["sub"]
        if org_id:
            mem = await db.execute(
                select(Membership).where(Membership.user_id == user_id, Membership.org_id == org_id)
            )
            if mem.scalar_one_or_none() is None:
                await websocket.close(code=4403)
                return
            agents_q = select(Agent).where(Agent.org_id == org_id)
        else:
            mem = await db.execute(select(Membership).where(Membership.user_id == user_id).limit(1))
            membership = mem.scalar_one_or_none()
            if membership is None:
                await websocket.send_json({"type": "snapshot", "agents": []})
                return
            agents_q = select(Agent).where(Agent.org_id == membership.org_id)

        agents = (await db.execute(agents_q)).scalars().all()
        await websocket.send_json(
            {
                "type": "snapshot",
                "agents": [
                    {
                        "id": a.id,
                        "name": a.name,
                        "platform": a.platform,
                        "status": a.status.value,
                        "status_message": a.status_message,
                        "zone": a.zone,
                        "pos_x": a.pos_x,
                        "pos_y": a.pos_y,
                        "is_active": a.is_active,
                    }
                    for a in agents
                ],
            }
        )

    try:
        async for event in event_hub.subscribe():
            await websocket.send_json(event)
    except WebSocketDisconnect:
        return


async def _org_agent(db: AsyncSession, org_id: str, agent_id: str) -> Agent:
    result = await db.execute(
        select(Agent).options(selectinload(Agent.secrets)).where(Agent.id == agent_id, Agent.org_id == org_id)
    )
    agent = result.scalar_one_or_none()
    if agent is None:
        raise HTTPException(status_code=404, detail="Agent not found")
    return agent
