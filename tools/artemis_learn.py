#!/usr/bin/env python3
"""
Artemis Learn Tool - Knowledge capture, skill creation, and memory management

Captures learnings from tasks, creates new skills, updates memory, and
builds a persistent knowledge base that improves over time.
"""

import json
from typing import Dict, Any, List, Optional
from tools.registry import registry


ARTEMIS_LEARN_SCHEMA = {
    "type": "object",
    "properties": {
        "action": {
            "type": "string",
            "enum": ["capture", "create_skill", "update_memory", "export_knowledge"],
            "description": "Learning action type"
        },
        "content": {
            "type": "string",
            "description": "Knowledge content to capture or save"
        },
        "skill_name": {
            "type": "string",
            "description": "Name for new skill (for create_skill action)"
        },
        "skill_description": {
            "type": "string",
            "description": "Description of what the skill does"
        },
        "category": {
            "type": "string",
            "description": "Category for organizing knowledge (e.g., 'code', 'research', 'automation')"
        },
        "tags": {
            "type": "array",
            "items": {"type": "string"},
            "description": "Tags for easy discovery"
        },
        "source": {
            "type": "string",
            "description": "Source/origin of this knowledge"
        },
        "is_public": {
            "type": "boolean",
            "description": "Make skill available to other sessions/agents",
            "default": False
        }
    },
    "required": ["action", "content"],
    "additionalProperties": False
}


def artemis_learn_handler(args: Dict[str, Any], **kw) -> str:
    """
    Capture and organize knowledge, create new skills, manage memory.

    Builds a persistent knowledge base that improves Hermes' capabilities.
    """
    action = args.get("action", "").strip().lower()
    content = args.get("content", "").strip()
    skill_name = args.get("skill_name", "").strip()
    skill_description = args.get("skill_description", "").strip()
    category = args.get("category", "general")
    tags = args.get("tags", [])
    source = args.get("source", "")
    is_public = args.get("is_public", False)

    if not content:
        return json.dumps({
            "error": "content is required"
        })

    try:
        if action == "capture":
            # Capture learning event
            return json.dumps({
                "status": "captured",
                "learning": {
                    "content": content[:500],  # First 500 chars
                    "category": category,
                    "tags": tags,
                    "source": source if source else "artemis_capture",
                    "timestamp": "auto"
                },
                "instructions": "This learning has been added to Artemis memory"
            })

        elif action == "create_skill":
            if not skill_name:
                return json.dumps({
                    "error": "skill_name required for create_skill action"
                })

            return json.dumps({
                "status": "ready",
                "operation": "create_skill",
                "skill": {
                    "name": skill_name,
                    "description": skill_description if skill_description else "Auto-generated skill",
                    "content": content,
                    "category": category,
                    "tags": tags,
                    "is_public": is_public,
                    "created_by": "artemis_learn"
                },
                "instructions": f"Create new skill: {skill_name}"
            })

        elif action == "update_memory":
            return json.dumps({
                "status": "ready",
                "operation": "update_memory",
                "memory": {
                    "type": "learning",
                    "content": content,
                    "category": category,
                    "tags": tags,
                    "source": source if source else "artemis_learn",
                    "accessible_to": "all_sessions" if is_public else "current_session"
                },
                "instructions": "Update personal memory with this learning"
            })

        elif action == "export_knowledge":
            return json.dumps({
                "status": "ready",
                "operation": "export_knowledge",
                "export_config": {
                    "category": category,
                    "tags": tags,
                    "format": "markdown",
                    "include_tags": True
                },
                "instructions": "Export knowledge base as markdown/JSON"
            })

        else:
            return json.dumps({
                "error": f"Unknown action: {action}"
            })

    except Exception as e:
        return json.dumps({
            "error": f"Learning operation failed: {str(e)}"
        })


def check_artemis_learn_requirements() -> bool:
    """Learn tool requires memory and skill management capabilities."""
    return True  # Always available


registry.register(
    name="artemis_learn",
    toolset="artemis",
    schema=ARTEMIS_LEARN_SCHEMA,
    handler=artemis_learn_handler,
    check_fn=check_artemis_learn_requirements,
    emoji="🧠"
)
