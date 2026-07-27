# Agent plugins

Каждая папка — независимый платформенный агент.

Минимальный набор файлов:
- `plugin.py` — класс `Plugin(BaseAgentPlugin)` с `manifest`
- опционально `manifest.json` — дублирование метаданных для UI

Registry (`app.agents.registry`) автоматически обнаруживает плагины при старте API.
Ядро не нужно менять при добавлении Instagram/YouTube/… — только новый каталог здесь.
