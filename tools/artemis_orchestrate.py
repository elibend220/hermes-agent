#!/usr/bin/env python3
"""
Artemis Orchestrate Tool - Multi-step task coordination and workflow management

The conductor tool that coordinates research, execution, integration, and learning
into complex workflows. Works with Claude to break down complex tasks and manage
multi-step processes.
"""

import json
from typing import Dict, Any, List, Optional
from tools.registry import registry


ARTEMIS_ORCHESTRATE_SCHEMA = {
    "type": "object",
    "properties": {
        "objective": {
            "type": "string",
            "description": "High-level goal or objective to accomplish"
        },
        "workflow": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "step_id": {"type": "string"},
                    "tool": {
                        "type": "string",
                        "enum": ["research", "execute", "integrate", "learn", "analyze", "delegate"]
                    },
                    "action": {"type": "string"},
                    "description": {"type": "string"},
                    "depends_on": {"type": "array", "items": {"type": "string"}},
                    "params": {"type": "object"}
                }
            },
            "description": "Array of workflow steps with dependencies"
        },
        "parallel": {
            "type": "boolean",
            "description": "Allow steps to run in parallel where dependencies allow",
            "default": True
        },
        "checkpoint": {
            "type": "boolean",
            "description": "Create checkpoints between steps for recovery",
            "default": False
        },
        "notify_on_progress": {
            "type": "boolean",
            "description": "Send progress updates during execution",
            "default": True
        },
        "max_retries": {
            "type": "integer",
            "description": "Maximum retries for failed steps",
            "default": 3
        },
        "timeout": {
            "type": "integer",
            "description": "Total workflow timeout in seconds",
            "default": 3600
        }
    },
    "required": ["objective"],
    "additionalProperties": False
}


def artemis_orchestrate_handler(args: Dict[str, Any], **kw) -> str:
    """
    Orchestrate complex multi-step workflows combining all Artemis capabilities.

    Acts as a task conductor, managing dependencies, parallelization, and recovery.
    """
    objective = args.get("objective", "").strip()
    workflow = args.get("workflow", [])
    parallel = args.get("parallel", True)
    checkpoint = args.get("checkpoint", False)
    notify_on_progress = args.get("notify_on_progress", True)
    max_retries = args.get("max_retries", 3)
    timeout = args.get("timeout", 3600)

    if not objective:
        return json.dumps({
            "error": "objective is required"
        })

    try:
        # If no workflow provided, prepare for Claude to design one
        if not workflow:
            return json.dumps({
                "status": "workflow_design_required",
                "objective": objective,
                "instructions": [
                    "1. Break down objective into concrete steps",
                    "2. Identify which Artemis tools to use for each step",
                    "3. Define dependencies between steps",
                    "4. Structure as workflow array and call artemis_orchestrate again"
                ],
                "available_tools": [
                    "artemis_research - gather information",
                    "artemis_execute - run code/commands",
                    "artemis_integrate - connect to services",
                    "artemis_learn - capture knowledge",
                    "delegate_task - spawn subagent"
                ]
            })

        # Validate workflow structure
        step_ids = set()
        dependencies = {}
        valid_tools = {"research", "execute", "integrate", "learn", "analyze", "delegate"}

        for step in workflow:
            step_id = step.get("step_id", "").strip()
            tool = step.get("tool", "").strip().lower()
            depends_on = step.get("depends_on", [])

            if not step_id:
                return json.dumps({"error": "All workflow steps must have step_id"})

            if step_id in step_ids:
                return json.dumps({"error": f"Duplicate step_id: {step_id}"})

            if tool not in valid_tools:
                return json.dumps({
                    "error": f"Invalid tool: {tool}. Must be one of: {', '.join(valid_tools)}"
                })

            step_ids.add(step_id)
            dependencies[step_id] = depends_on

        # Check for circular dependencies (basic check)
        for step_id, deps in dependencies.items():
            if step_id in deps:
                return json.dumps({
                    "error": f"Circular dependency: {step_id} depends on itself"
                })

        # Build execution plan
        execution_plan = {
            "objective": objective,
            "workflow_steps": len(workflow),
            "configuration": {
                "parallel_execution": parallel,
                "checkpointing": checkpoint,
                "progress_notifications": notify_on_progress,
                "max_retries_per_step": max_retries,
                "total_timeout": timeout
            },
            "workflow": workflow,
            "status": "ready_for_execution",
            "execution_order": "dependency_aware"
        }

        return json.dumps({
            "status": "orchestration_ready",
            "plan": execution_plan,
            "instructions": "Workflow is ready to execute. Starting dependency chain analysis.",
            "next_steps": [
                f"Execute {len(workflow)} steps in order",
                "Monitor for failures and retry",
                "Checkpoint state between steps" if checkpoint else "No checkpointing",
                "Notify on progress" if notify_on_progress else "Silent execution"
            ]
        })

    except Exception as e:
        return json.dumps({
            "error": f"Orchestration failed: {str(e)}"
        })


def check_artemis_orchestrate_requirements() -> bool:
    """Orchestration tool requires delegation capabilities."""
    return True  # Always available


registry.register(
    name="artemis_orchestrate",
    toolset="artemis",
    schema=ARTEMIS_ORCHESTRATE_SCHEMA,
    handler=artemis_orchestrate_handler,
    check_fn=check_artemis_orchestrate_requirements,
    emoji="📅"
)
