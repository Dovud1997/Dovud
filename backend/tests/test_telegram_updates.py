from app.services.telegram_updates import parse_telegram_update


def test_parse_message():
    parsed = parse_telegram_update(
        {
            "update_id": 1,
            "message": {
                "message_id": 10,
                "text": "Привет",
                "chat": {"id": 123},
                "from": {"id": 5, "username": "alice"},
            },
        }
    )
    assert parsed is not None
    event_type, payload = parsed
    assert event_type == "message"
    assert payload["text"] == "Привет"
    assert payload["chat_id"] == 123


def test_parse_follow_and_skip():
    parsed = parse_telegram_update(
        {
            "update_id": 2,
            "message": {
                "message_id": 11,
                "chat": {"id": 1},
                "from": {"username": "bob"},
                "new_chat_members": [{"username": "bob"}],
            },
        }
    )
    assert parsed is not None
    assert parsed[0] == "follow"
    assert parse_telegram_update({"update_id": 3}) is None
