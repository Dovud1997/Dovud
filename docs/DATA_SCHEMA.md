# Схема данных

## ER (логическая)

```
User 1──* Agent
Agent 1──* AgentCredential (encrypted)
Agent 1──* ReplyTemplate
Agent 1──* StyleExample
Agent 1──* AgentEvent
Agent 1──* AgentLog
Agent 1──* CommandJob
```

## Таблицы

### users
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| email | str unique | Логин админа |
| password_hash | str | bcrypt |
| created_at | datetime | |

### agents
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| name | str | Человекочитаемое имя |
| platform | str | `telegram`, `instagram`, … (plugin id) |
| status | enum | `draft`, `connecting`, `online`, `offline`, `error`, `busy` |
| status_message | str? | Детали ошибки / текущее действие |
| ai_mode | enum | `off`, `template`, `llm` |
| system_prompt | text? | Инструкция «в стиле пользователя» |
| zone | str | Локация на сцене (`telegram`, `instagram`, …) |
| pos_x, pos_y | float | Координаты спрайта на сцене |
| is_active | bool | Активирован ли |
| owner_id | FK users | |
| created_at / updated_at | datetime | |

### agent_secrets
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| agent_id | FK | |
| key | str | Например `bot_token` |
| value_encrypted | bytes | Fernet ciphertext |
| created_at | datetime | |

Чувствительные поля формы плагина пишутся сюда, не в `agents`.

### reply_templates
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| agent_id | FK | |
| name | str | |
| trigger_pattern | str? | regex / keyword; null = default |
| body | text | Текст шаблона, поддержка `{{user_name}}` |
| is_default | bool | |
| created_at | datetime | |

### style_examples
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| agent_id | FK | |
| user_message | text | Пример входящего |
| assistant_reply | text | Как ответил бы владелец |
| created_at | datetime | |

Few-shot для LLM-режима.

### agent_events
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| agent_id | FK | |
| type | str | `like`, `follow`, `unfollow`, `message`, `comment`, `status`, `action_result` |
| payload | JSON | Сырые данные события |
| notified | bool | Ушло ли в бот/push |
| created_at | datetime | |

### agent_logs
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| agent_id | FK? | null для системных |
| level | enum | `info`, `warn`, `error` |
| message | text | |
| meta | JSON? | |
| created_at | datetime | |

### command_jobs
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID PK | |
| agent_id | FK | |
| action | str | `publish_story`, `publish_post`, `reply`, `test_connection` |
| payload | JSON | Медиа URL, текст, … |
| status | enum | `pending`, `running`, `done`, `failed` |
| result | JSON? | |
| error | text? | |
| created_at / finished_at | datetime | |

## Статусы на пиксельной сцене
Маппинг `agents.status` → анимация персонажа:
- `online` → idle / ждёт задачу
- `busy` → публикует / отвечает (walk + work)
- `error` → shake / alert
- `offline` / `draft` → серый idle
