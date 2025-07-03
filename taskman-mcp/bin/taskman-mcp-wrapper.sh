#!/bin/bash
# Wrapper script for taskman-mcp to redirect stderr to a log file
# This ensures clean stdio communication with Claude Desktop

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Set debug logging
export TASKMAN_LOG_LEVEL=DEBUG

# Ensure clean termination
trap 'exit 0' SIGTERM SIGINT

# Create log file with timestamp
LOG_FILE="/tmp/taskman-mcp-$(date +%Y%m%d-%H%M%S).log"

# Run the MCP server with stderr redirected to a log file
# Use unbuffered output to ensure logs are written immediately
exec "$SCRIPT_DIR/taskman-mcp" 2>"$LOG_FILE"