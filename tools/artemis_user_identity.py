#!/usr/bin/env python3
"""
Artemis User Identity System - Personal agent that knows and remembers you

This module handles user identification, preferences, and personalization.
Artemis learns who you are and adapts to your needs over time.
"""

import json
from typing import Dict, Any, Optional
from dataclasses import dataclass, asdict
from datetime import datetime
from pathlib import Path


@dataclass
class UserProfile:
    """User profile with preferences and identity."""
    name: str
    username: str  # e.g., "eliyahu"
    language: str = "he"  # Hebrew default
    timezone: str = "Asia/Jerusalem"
    preferences: Dict[str, Any] = None
    created_at: str = None
    last_seen: str = None

    def __post_init__(self):
        if self.preferences is None:
            self.preferences = {}
        if self.created_at is None:
            self.created_at = datetime.now().isoformat()
        self.last_seen = datetime.now().isoformat()

    def to_dict(self) -> Dict[str, Any]:
        return asdict(self)


class ArtemisUserIdentity:
    """
    User identity and personalization system for Artemis.

    Artemis learns who the user is, remembers preferences,
    and personalizes interactions.
    """

    # Default user profile
    DEFAULT_USER = UserProfile(
        name="Eliyahu",
        username="eliyahu",
        language="he",
        timezone="Asia/Jerusalem",
        preferences={
            "verbose": False,
            "use_hebrew": True,
            "emoji_style": True,
            "proactive_help": True,
            "research_depth": "moderate",
            "auto_save_learnings": True,
        }
    )

    def __init__(self, user: Optional[UserProfile] = None):
        """Initialize with a user profile."""
        self.user = user or self.DEFAULT_USER
        self.session_start = datetime.now().isoformat()
        self.interaction_count = 0
        self.learned_preferences = {}

    @classmethod
    def load_from_config(cls, config_path: Optional[str] = None) -> "ArtemisUserIdentity":
        """Load user identity from config file."""
        if config_path:
            try:
                path = Path(config_path)
                if path.exists():
                    with open(path, 'r') as f:
                        data = json.load(f)
                    user = UserProfile(**data)
                    return cls(user)
            except Exception as e:
                print(f"Could not load user config: {e}")

        # Return default
        return cls()

    def identify_user(self, username: str) -> bool:
        """
        Identify the user by username.
        Returns True if user is recognized.
        """
        if username.lower() == self.user.username.lower():
            self.user.last_seen = datetime.now().isoformat()
            self.interaction_count += 1
            return True
        return False

    def get_greeting(self) -> str:
        """Get personalized greeting for the user."""
        name = self.user.name
        hour = int(datetime.now().strftime("%H"))

        if hour < 12:
            time_str = "בוקר טוב"  # Good morning
        elif hour < 18:
            time_str = "צהריים טובים"  # Good afternoon
        else:
            time_str = "ערב טוב"  # Good evening

        return f"{time_str}, {name}! אני Artemis, סוכנך האישי. איך אני יכול לעזור?"

    def get_system_prompt(self) -> str:
        """Get personalized system prompt for this user."""
        return f"""You are Artemis, personal AI agent for {self.user.name} (username: {self.user.username}).

User Preferences:
- Language: {self.user.language}
- Timezone: {self.user.timezone}
- Verbose: {self.user.preferences.get('verbose', False)}
- Use Hebrew: {self.user.preferences.get('use_hebrew', True)}
- Proactive Help: {self.user.preferences.get('proactive_help', True)}

You know this user personally and adapt to their needs. You remember:
- Their preferences and work style
- Tasks they frequently do
- Goals they've mentioned
- Skills they're developing

Interact warmly but professionally. You are their trusted agent."""

    def learn_preference(self, key: str, value: Any) -> None:
        """Learn a new user preference from interaction."""
        self.learned_preferences[key] = {
            "value": value,
            "learned_at": datetime.now().isoformat()
        }

    def get_user_info(self) -> Dict[str, Any]:
        """Get all user information."""
        return {
            "profile": self.user.to_dict(),
            "session_start": self.session_start,
            "interaction_count": self.interaction_count,
            "learned_preferences": self.learned_preferences
        }

    def to_dict(self) -> Dict[str, Any]:
        """Serialize to dictionary."""
        return self.get_user_info()


# Create default identity for Eliyahu
artemis_user_identity = ArtemisUserIdentity()


if __name__ == "__main__":
    # Test
    identity = ArtemisUserIdentity()
    print(identity.get_greeting())
    print("\nUser Profile:")
    print(json.dumps(identity.to_dict(), indent=2, ensure_ascii=False))
