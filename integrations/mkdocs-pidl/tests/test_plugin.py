"""Tests for the PIDL MkDocs plugin."""

import pytest

from mkdocs_pidl.plugin import PidlPlugin


class TestPidlPlugin:
    """Test the PidlPlugin class."""

    def test_parse_attributes_empty(self) -> None:
        """Test parsing empty attributes."""
        plugin = PidlPlugin()
        attrs = plugin._parse_attributes("")
        assert attrs == {}

    def test_parse_attributes_single(self) -> None:
        """Test parsing a single attribute."""
        plugin = PidlPlugin()
        attrs = plugin._parse_attributes("format=mermaid")
        assert attrs == {"format": "mermaid"}

    def test_parse_attributes_multiple(self) -> None:
        """Test parsing multiple attributes."""
        plugin = PidlPlugin()
        attrs = plugin._parse_attributes("format=infographic theme=dark direction=vertical")
        assert attrs == {
            "format": "infographic",
            "theme": "dark",
            "direction": "vertical",
        }

    def test_parse_attributes_quoted(self) -> None:
        """Test parsing quoted attributes."""
        plugin = PidlPlugin()
        attrs = plugin._parse_attributes('file="path/to/file.json"')
        assert attrs == {"file": "path/to/file.json"}

    def test_parse_attributes_single_quotes(self) -> None:
        """Test parsing single-quoted attributes."""
        plugin = PidlPlugin()
        attrs = plugin._parse_attributes("file='path/to/file.json'")
        assert attrs == {"file": "path/to/file.json"}


class TestPidlPluginRender:
    """Test rendering functionality."""

    def test_render_invalid_json(self) -> None:
        """Test rendering invalid JSON returns error comment."""
        plugin = PidlPlugin()
        result = plugin._render_pidl("not valid json", "mermaid", "bold", "horizontal")
        assert "PIDL Error" in result
        assert "Invalid JSON" in result

    def test_render_unknown_format(self) -> None:
        """Test rendering unknown format returns error comment."""
        plugin = PidlPlugin()
        result = plugin._render_pidl('{"id": "test"}', "unknown", "bold", "horizontal")
        assert "PIDL Error" in result
        assert "Unknown format" in result
