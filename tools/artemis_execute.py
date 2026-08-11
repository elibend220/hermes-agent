#!/usr/bin/env python3
"""
Artemis Execute Tool - Code and command execution with safety

Coordinates execution of code (Python), shell commands, and complex workflows.
Integrates with execute_code and terminal tools with intelligent batching
and error handling.
"""

import json
from typing import Dict, Any, List, Optional
from tools import registry


ARTEMIS_EXECUTE_SCHEMA = {
    "type": "object",
    "properties": {
        "action": {
            "type": "string",
            "enum": ["python", "shell", "workflow"],
            "description": "Type of execution: python code, shell command, or multi-step workflow"
        },
        "code": {
            "type": "string",
            "description": "Code or command to execute"
        },
        "context": {
            "type": "string",
            "description": "Context for execution (e.g., 'data analysis', 'automation', 'integration')"
        },
        "timeout": {
            "type": "integer",
            "description": "Maximum execution time in seconds",
            "default": 60
        },
        "isolated": {
            "type": "boolean",
            "description": "Run in isolated environment (safer but limited)",
            "default": False
        },
        "workflow_steps": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "step": {"type": "string"},
                    "action": {"type": "string"},
                    "code": {"type": "string"}
                }
            },
            "description": "For workflow action: array of {step, action, code}"
        }
    },
    "required": ["action", "code"],
    "additionalProperties": False
}


def artemis_execute_handler(args: Dict[str, Any], **kw) -> str:
    """
    Execute code or commands in controlled manner.

    Delegates to appropriate tool (execute_code or terminal) based on action type.
    """
    action = args.get("action", "").strip().lower()
    code = args.get("code", "").strip()
    context = args.get("context", "general")
    timeout = args.get("timeout", 60)
    isolated = args.get("isolated", False)
    workflow_steps = args.get("workflow_steps", [])

    if not code and not workflow_steps:
        return json.dumps({"error": "code or workflow_steps required"})

    try:
        if action == "python":
            return json.dumps({
                "status": "ready",
                "executor": "execute_code",
                "payload": {
                    "code": code,
                    "context": context,
                    "timeout": timeout,
                    "isolated": isolated
                }
            })

        elif action == "shell":
            return json.dumps({
                "status": "ready",
                "executor": "terminal",
                "payload": {
                    "command": code,
                    "context": context,
                    "timeout": timeout
                }
            })

        elif action == "workflow":
            if not workflow_steps:
                return json.dumps({"error": "workflow_steps required for workflow action"})

            return json.dumps({
                "status": "ready",
                "executor": "workflow_coordinator",
                "payload": {
                    "steps": workflow_steps,
                    "context": context,
                    "total_timeout": timeout
                }
            })

        else:
            return json.dumps({"error": f"Unknown action: {action}"})

    except Exception as e:
        return json.dumps({
            "error": f"Execution setup failed: {str(e)}"
        })


def check_artemis_execute_requirements() -> bool:
    """Execute tool requires code execution capabilities."""
    return True  # Always available


registry.register(
    name="artemis_execute",
    toolset="artemis",
    schema=ARTEMIS_EXECUTE_SCHEMA,
    handler=artemis_execute_handler,
    check_fn=check_artemis_execute_requirements,
    emoji="⚙️"
)
