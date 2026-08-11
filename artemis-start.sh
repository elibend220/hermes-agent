#!/bin/bash
#
# Artemis Personal Agent Startup Script
# For: Eliyahu
# Usage: ./artemis-start.sh
#

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}🤖 ARTEMIS - Personal AI Agent${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "hermes" ]; then
    echo -e "${YELLOW}⚠️  Not in hermes-agent directory${NC}"
    echo "Please cd to the hermes-agent directory first:"
    echo "  cd ~/projects/hermes-agent"
    exit 1
fi

# Activate virtual environment if it exists
if [ -d ".venv" ]; then
    echo -e "${GREEN}✓${NC} Activating virtual environment..."
    source .venv/bin/activate
elif [ -d "venv" ]; then
    echo -e "${GREEN}✓${NC} Activating virtual environment..."
    source venv/bin/activate
else
    echo -e "${YELLOW}⚠️  No virtual environment found${NC}"
    echo "Run './setup-hermes.sh' first to set up the environment"
    exit 1
fi

# Check if Artemis tools are available
echo -e "${GREEN}✓${NC} Checking Artemis tools..."
ARTEMIS_COUNT=$(python3 -c "from tools.registry import discover_builtin_tools; tools=discover_builtin_tools(); print(len([t for t in tools if 'artemis' in t]))" 2>/dev/null || echo "0")

if [ "$ARTEMIS_COUNT" = "0" ]; then
    echo -e "${YELLOW}⚠️  Artemis tools not found${NC}"
    echo "Try: python3 -c \"from tools.registry import discover_builtin_tools; discover_builtin_tools()\""
    exit 1
fi

echo -e "${GREEN}✓${NC} Found $ARTEMIS_COUNT Artemis tools"

# Load user configuration if exists
if [ -f "artemis-config.json" ]; then
    echo -e "${GREEN}✓${NC} Loading your Artemis configuration..."
    export ARTEMIS_CONFIG="$(pwd)/artemis-config.json"
fi

# Show greeting
echo ""
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
python3 << 'EOF'
try:
    from tools.artemis_user_identity import ArtemisUserIdentity
    identity = ArtemisUserIdentity()
    print(identity.get_greeting())
except Exception as e:
    print(f"Artemis ready! Initializing...\n")
EOF
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""

# Start Hermes with Artemis
echo -e "${GREEN}Starting Artemis...${NC}"
echo ""

# Check if running with specific model or config
if [ -n "$1" ]; then
    hermes "$@"
else
    # Default: start hermes
    hermes
fi
