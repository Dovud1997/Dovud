from __future__ import annotations

import importlib.util
import json
import sys
from pathlib import Path
from typing import Any

from app.agents.base import BaseAgentPlugin, PluginManifest
from app.core.config import get_settings


class AgentRegistry:
    """Discovers plugins from backend/plugins/<name>/ without touching core."""

    def __init__(self) -> None:
        self._plugins: dict[str, type[BaseAgentPlugin]] = {}
        self._manifests: dict[str, PluginManifest] = {}

    def discover(self, plugins_dir: str | None = None) -> None:
        root = Path(plugins_dir or get_settings().plugins_dir)
        if not root.exists():
            return
        for entry in sorted(root.iterdir()):
            if not entry.is_dir() or entry.name.startswith("_"):
                continue
            plugin_py = entry / "plugin.py"
            if not plugin_py.exists():
                continue
            module_name = f"agent_plugin_{entry.name}"
            spec = importlib.util.spec_from_file_location(module_name, plugin_py)
            if spec is None or spec.loader is None:
                continue
            module = importlib.util.module_from_spec(spec)
            sys.modules[module_name] = module
            spec.loader.exec_module(module)
            plugin_cls = getattr(module, "Plugin", None)
            if plugin_cls is None:
                continue
            manifest = getattr(plugin_cls, "manifest", None)
            if manifest is None and (entry / "manifest.json").exists():
                manifest = self._manifest_from_json(entry / "manifest.json")
                plugin_cls.manifest = manifest
            if manifest is None:
                continue
            self.register(plugin_cls)

    def register(self, plugin_cls: type[BaseAgentPlugin]) -> None:
        platform = plugin_cls.manifest.platform
        self._plugins[platform] = plugin_cls
        self._manifests[platform] = plugin_cls.manifest

    def list_manifests(self) -> list[PluginManifest]:
        return list(self._manifests.values())

    def get_manifest(self, platform: str) -> PluginManifest | None:
        return self._manifests.get(platform)

    def create(self, platform: str, agent_id: str, credentials: dict[str, str]) -> BaseAgentPlugin:
        if platform not in self._plugins:
            raise KeyError(f"Unknown platform plugin: {platform}")
        return self._plugins[platform](agent_id=agent_id, credentials=credentials)

    @staticmethod
    def _manifest_from_json(path: Path) -> PluginManifest:
        from app.agents.base import FieldSpec

        raw: dict[str, Any] = json.loads(path.read_text(encoding="utf-8"))
        fields = [FieldSpec(**f) for f in raw.get("fields", [])]
        return PluginManifest(
            platform=raw["platform"],
            title=raw["title"],
            description=raw.get("description", ""),
            zone=raw.get("zone", raw["platform"]),
            fields=fields,
            actions=raw.get("actions", []),
        )


_registry: AgentRegistry | None = None


def get_registry() -> AgentRegistry:
    global _registry
    if _registry is None:
        _registry = AgentRegistry()
        _registry.discover()
    return _registry
