"""MkDocs plugin for rendering PIDL diagrams."""

import json
import re
import subprocess
from pathlib import Path
from typing import Any

from mkdocs.config import config_options
from mkdocs.config.defaults import MkDocsConfig
from mkdocs.plugins import BasePlugin
from mkdocs.structure.files import Files
from mkdocs.structure.pages import Page


class PidlPluginConfig:
    """Configuration options for the PIDL plugin."""

    pidl_path = config_options.Type(str, default="pidl")
    default_format = config_options.Choice(
        ["mermaid", "svg", "infographic"], default="mermaid"
    )
    default_theme = config_options.Choice(
        ["bold", "minimal", "dark", "tech", "corporate", "accessible"],
        default="bold",
    )
    default_direction = config_options.Choice(
        ["horizontal", "vertical"], default="horizontal"
    )
    cache_enabled = config_options.Type(bool, default=True)


class PidlPlugin(BasePlugin[PidlPluginConfig]):
    """MkDocs plugin for rendering PIDL protocol diagrams.

    This plugin processes PIDL code blocks in markdown files and renders them
    as Mermaid diagrams, SVG images, or infographics.

    Usage in markdown:
        ```pidl
        {
            "id": "example",
            "name": "Example Protocol",
            "entities": [...],
            "flows": [...]
        }
        ```

        Or with options:
        ```pidl format=infographic theme=dark direction=vertical
        {...}
        ```

        Or reference an external file:
        ```pidl file=protocols/oauth2.pidl.json format=mermaid
        ```
    """

    def __init__(self) -> None:
        """Initialize the plugin."""
        super().__init__()
        self._cache: dict[str, str] = {}

    def on_page_markdown(
        self, markdown: str, page: Page, config: MkDocsConfig, files: Files
    ) -> str:
        """Process PIDL code blocks in the markdown content."""
        # Pattern to match pidl code blocks with optional attributes
        pattern = r"```pidl\s*([^\n]*)\n(.*?)```"

        def replace_pidl_block(match: re.Match[str]) -> str:
            attrs_str = match.group(1).strip()
            content = match.group(2).strip()

            # Parse attributes
            attrs = self._parse_attributes(attrs_str)

            # Check if referencing external file
            if "file" in attrs:
                file_path = attrs["file"]
                # Resolve relative to docs directory
                docs_dir = Path(config["docs_dir"])
                full_path = docs_dir / file_path
                if full_path.exists():
                    content = full_path.read_text()
                else:
                    return f"<!-- PIDL Error: File not found: {file_path} -->"

            # Get rendering options
            fmt = attrs.get("format", self.config.default_format)
            theme = attrs.get("theme", self.config.default_theme)
            direction = attrs.get("direction", self.config.default_direction)

            # Generate cache key
            cache_key = f"{content}:{fmt}:{theme}:{direction}"

            # Check cache
            if self.config.cache_enabled and cache_key in self._cache:
                return self._cache[cache_key]

            # Render the PIDL content
            result = self._render_pidl(content, fmt, theme, direction)

            # Cache the result
            if self.config.cache_enabled:
                self._cache[cache_key] = result

            return result

        return re.sub(pattern, replace_pidl_block, markdown, flags=re.DOTALL)

    def _parse_attributes(self, attrs_str: str) -> dict[str, str]:
        """Parse attribute string into a dictionary."""
        attrs: dict[str, str] = {}
        if not attrs_str:
            return attrs

        # Match key=value pairs
        pattern = r'(\w+)=(["\']?)([^"\'\s]+)\2'
        for match in re.finditer(pattern, attrs_str):
            key = match.group(1)
            value = match.group(3)
            attrs[key] = value

        return attrs

    def _render_pidl(
        self, content: str, fmt: str, theme: str, direction: str
    ) -> str:
        """Render PIDL content using the pidl CLI."""
        try:
            # Validate JSON
            json.loads(content)
        except json.JSONDecodeError as e:
            return f"<!-- PIDL Error: Invalid JSON: {e} -->"

        if fmt == "mermaid":
            return self._render_mermaid(content)
        elif fmt == "svg":
            return self._render_svg(content)
        elif fmt == "infographic":
            return self._render_infographic(content, theme, direction)
        else:
            return f"<!-- PIDL Error: Unknown format: {fmt} -->"

    def _render_mermaid(self, content: str) -> str:
        """Render PIDL as Mermaid diagram."""
        try:
            result = subprocess.run(
                [self.config.pidl_path, "render", "-f", "mermaid", "-"],
                input=content,
                capture_output=True,
                text=True,
                timeout=30,
            )

            if result.returncode != 0:
                return f"<!-- PIDL Error: {result.stderr} -->"

            # Wrap in mermaid code block for MkDocs
            mermaid_code = result.stdout.strip()
            return f"```mermaid\n{mermaid_code}\n```"

        except subprocess.TimeoutExpired:
            return "<!-- PIDL Error: Rendering timed out -->"
        except FileNotFoundError:
            return f"<!-- PIDL Error: pidl CLI not found at {self.config.pidl_path} -->"

    def _render_svg(self, content: str) -> str:
        """Render PIDL as SVG."""
        try:
            result = subprocess.run(
                [self.config.pidl_path, "render", "-f", "svg", "-"],
                input=content,
                capture_output=True,
                text=True,
                timeout=30,
            )

            if result.returncode != 0:
                return f"<!-- PIDL Error: {result.stderr} -->"

            # Return SVG directly (will be embedded in HTML)
            return result.stdout.strip()

        except subprocess.TimeoutExpired:
            return "<!-- PIDL Error: Rendering timed out -->"
        except FileNotFoundError:
            return f"<!-- PIDL Error: pidl CLI not found at {self.config.pidl_path} -->"

    def _render_infographic(
        self, content: str, theme: str, direction: str
    ) -> str:
        """Render PIDL as infographic SVG."""
        try:
            cmd = [
                self.config.pidl_path,
                "render",
                "-f", "infographic",
                "--ig-theme", theme,
                "--ig-direction", direction,
                "-",
            ]

            result = subprocess.run(
                cmd,
                input=content,
                capture_output=True,
                text=True,
                timeout=30,
            )

            if result.returncode != 0:
                return f"<!-- PIDL Error: {result.stderr} -->"

            # Return SVG directly
            return result.stdout.strip()

        except subprocess.TimeoutExpired:
            return "<!-- PIDL Error: Rendering timed out -->"
        except FileNotFoundError:
            return f"<!-- PIDL Error: pidl CLI not found at {self.config.pidl_path} -->"

    def on_env(self, env: Any, config: MkDocsConfig, files: Files) -> Any:
        """Clear cache when environment is rebuilt."""
        self._cache.clear()
        return env
