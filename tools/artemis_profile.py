#!/usr/bin/env python3
"""
Artemis Profile Tool - View and configure Artemis personality and settings

Allows users to:
- View Artemis' personality traits and characteristics
- Check current configuration
- Get system messages and interaction guidelines
- See communication preferences
"""

import json
from typing import Dict, Any
from tools.registry import registry
from tools.artemis_personality import ArtemisPersonality


ARTEMIS_PROFILE_SCHEMA = {
    "type": "object",
    "properties": {
        "action": {
            "type": "string",
            "enum": ["show_profile", "show_traits", "show_voice", "show_introduction", "show_all"],
            "description": "What aspect of Artemis to show",
            "default": "show_all"
        }
    },
    "required": [],
    "additionalProperties": False
}


def artemis_profile_handler(args: Dict[str, Any], **kw) -> str:
    """
    Display Artemis profile, personality traits, and configuration.
    """
    action = args.get("action", "show_all").strip().lower()

    try:
        if action == "show_introduction":
            return json.dumps({
                "status": "introduction",
                "content": ArtemisPersonality.introduction()
            })

        elif action == "show_traits":
            traits = ArtemisPersonality.get_all_traits()
            return json.dumps({
                "status": "traits",
                "traits": {
                    name: {
                        "description": trait.description,
                        "intensity": trait.intensity
                    }
                    for name, trait in traits.items()
                }
            })

        elif action == "show_voice":
            voice = ArtemisPersonality.VOICE_STYLE
            return json.dumps({
                "status": "voice_style",
                "voice": {
                    "formality": voice.formality,
                    "proactivity": voice.proactivity,
                    "uses_emoji": voice.emoji_usage,
                    "explanation_depth": voice.explanation_depth
                }
            })

        elif action == "show_profile":
            profile = ArtemisPersonality.to_dict()
            return json.dumps({
                "status": "profile",
                "profile": profile
            })

        elif action == "show_all":
            profile = ArtemisPersonality.to_dict()
            return json.dumps({
                "status": "complete_profile",
                "profile": profile,
                "system_prompt": ArtemisPersonality.get_system_prompt("default"),
                "ready_message": "Artemis is ready to assist. All systems nominal."
            })

        else:
            return json.dumps({
                "error": f"Unknown action: {action}"
            })

    except Exception as e:
        return json.dumps({
            "error": f"Failed to retrieve profile: {str(e)}"
        })


def check_artemis_profile_requirements() -> bool:
    """Profile tool always available."""
    return True


registry.register(
    name="artemis_profile",
    toolset="artemis",
    schema=ARTEMIS_PROFILE_SCHEMA,
    handler=artemis_profile_handler,
    check_fn=check_artemis_profile_requirements,
    emoji="👤"
)
