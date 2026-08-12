#!/bin/sh
# NeetoPlanner CLI — session-start hook for Claude Code
# Lightweight auth liveness check. Always exits 0 (informational).

if ! command -v neetoplanner >/dev/null 2>&1; then
  echo "NeetoPlanner CLI is not installed or not on PATH."
  exit 0
fi

if neetoplanner whoami >/dev/null 2>&1; then
  echo "NeetoPlanner plugin active."
else
  echo "NeetoPlanner CLI installed but not authenticated. Run 'neetoplanner login' to authenticate."
fi

exit 0
