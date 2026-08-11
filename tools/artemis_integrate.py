#!/usr/bin/env python3
"""
Artemis Integrate Tool - Multi-system integration and data synchronization

Coordinates connections to external APIs, services, and systems.
Manages authentication, data transformation, and bi-directional sync.
"""

import json
from typing import Dict, Any, List, Optional
from tools import registry


ARTEMIS_INTEGRATE_SCHEMA = {
    "type": "object",
    "properties": {
        "service": {
            "type": "string",
            "description": "Target service/API (e.g., 'github', 'slack', 'notion', 'stripe', 'custom_api')"
        },
        "action": {
            "type": "string",
            "enum": ["query", "write", "sync", "webhook", "auth"],
            "description": "Integration action type"
        },
        "endpoint": {
            "type": "string",
            "description": "API endpoint or data source path"
        },
        "payload": {
            "type": "object",
            "description": "Data payload for query or write operations"
        },
        "auth_provider": {
            "type": "string",
            "description": "Authentication method: 'oauth', 'api_key', 'bearer', 'custom'"
        },
        "transform": {
            "type": "string",
            "description": "Optional: data transformation (e.g., 'json_to_csv', 'markdown_to_json')"
        },
        "bidirectional": {
            "type": "boolean",
            "description": "Enable two-way sync if supported",
            "default": False
        }
    },
    "required": ["service", "action", "endpoint"],
    "additionalProperties": False
}


def artemis_integrate_handler(args: Dict[str, Any], **kw) -> str:
    """
    Manage integrations with external services and APIs.

    Handles authentication, data transformation, and synchronization.
    """
    service = args.get("service", "").strip().lower()
    action = args.get("action", "").strip().lower()
    endpoint = args.get("endpoint", "").strip()
    payload = args.get("payload", {})
    auth_provider = args.get("auth_provider", "oauth")
    transform = args.get("transform", "")
    bidirectional = args.get("bidirectional", False)

    if not service or not endpoint:
        return json.dumps({
            "error": "service and endpoint are required"
        })

    # Validate action
    valid_actions = {"query", "write", "sync", "webhook", "auth"}
    if action not in valid_actions:
        return json.dumps({
            "error": f"action must be one of: {', '.join(valid_actions)}"
        })

    try:
        integration_config = {
            "service": service,
            "action": action,
            "endpoint": endpoint,
            "auth_method": auth_provider,
            "has_payload": bool(payload),
            "transformation": transform if transform else "none",
            "bidirectional_sync": bidirectional
        }

        # Different handling based on action
        if action == "auth":
            return json.dumps({
                "status": "auth_required",
                "service": service,
                "auth_provider": auth_provider,
                "next_step": f"Authenticate with {service} using {auth_provider}"
            })

        elif action == "query":
            return json.dumps({
                "status": "ready",
                "operation": "fetch_data",
                "config": integration_config,
                "instructions": f"Query {service} endpoint: {endpoint}"
            })

        elif action == "write":
            return json.dumps({
                "status": "ready",
                "operation": "send_data",
                "config": integration_config,
                "payload_size": len(json.dumps(payload)) if payload else 0,
                "instructions": f"Write to {service} endpoint: {endpoint}"
            })

        elif action == "sync":
            return json.dumps({
                "status": "ready",
                "operation": "synchronize",
                "config": integration_config,
                "bidirectional": bidirectional,
                "instructions": f"Sync data with {service}"
            })

        elif action == "webhook":
            return json.dumps({
                "status": "ready",
                "operation": "setup_webhook",
                "service": service,
                "endpoint": endpoint,
                "instructions": f"Configure webhook from {service} to {endpoint}"
            })

        return json.dumps({
            "status": "ready",
            "config": integration_config
        })

    except Exception as e:
        return json.dumps({
            "error": f"Integration setup failed: {str(e)}"
        })


def check_artemis_integrate_requirements() -> bool:
    """Integration tool requires API access capabilities."""
    return True  # Always available


registry.register(
    name="artemis_integrate",
    toolset="artemis",
    schema=ARTEMIS_INTEGRATE_SCHEMA,
    handler=artemis_integrate_handler,
    check_fn=check_artemis_integrate_requirements,
    emoji="🔗"
)
