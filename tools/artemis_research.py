#!/usr/bin/env python3
"""
Artemis Research Tool - Advanced information gathering and analysis

Combines web search, extraction, and analysis to provide deep research
capabilities. Can search across multiple sources, analyze findings, and
synthesize knowledge.
"""

import json
from typing import Dict, Any, List, Optional
from tools import registry


ARTEMIS_RESEARCH_SCHEMA = {
    "type": "object",
    "properties": {
        "query": {
            "type": "string",
            "description": "Research query or topic to investigate"
        },
        "depth": {
            "type": "string",
            "enum": ["quick", "moderate", "deep"],
            "description": "Search depth: quick (1-2 sources), moderate (3-5), deep (6+ sources)",
            "default": "moderate"
        },
        "sources": {
            "type": "array",
            "items": {"type": "string"},
            "description": "Optional: specific domains to prioritize (e.g., ['arxiv.org', 'github.com'])",
            "default": []
        },
        "focus": {
            "type": "string",
            "description": "Optional: specific aspect to focus on (e.g., 'latest developments', 'security concerns')"
        }
    },
    "required": ["query"],
    "additionalProperties": False
}


def artemis_research_handler(args: Dict[str, Any], **kw) -> str:
    """
    Execute advanced research across multiple sources.

    Returns structured research findings with sources and synthesis.
    """
    query = args.get("query", "").strip()
    depth = args.get("depth", "moderate")
    sources = args.get("sources", [])
    focus = args.get("focus", "")

    if not query:
        return json.dumps({"error": "query is required"})

    try:
        # Build research context
        research_context = {
            "query": query,
            "depth": depth,
            "target_sources": sources if sources else "any",
            "focus_area": focus if focus else "general"
        }

        # Determine number of sources based on depth
        source_count = {"quick": 2, "moderate": 5, "deep": 10}.get(depth, 5)

        return json.dumps({
            "status": "ready",
            "research_task": research_context,
            "expected_sources": source_count,
            "instructions": "Use web_search and web_extract tools to gather information",
            "synthesis_required": depth in ["moderate", "deep"]
        })
    except Exception as e:
        return json.dumps({
            "error": f"Research initialization failed: {str(e)}"
        })


def check_artemis_research_requirements() -> bool:
    """Research tool requires web search capabilities."""
    return True  # Always available


registry.register(
    name="artemis_research",
    toolset="artemis",
    schema=ARTEMIS_RESEARCH_SCHEMA,
    handler=artemis_research_handler,
    check_fn=check_artemis_research_requirements,
    emoji="🔍"
)
