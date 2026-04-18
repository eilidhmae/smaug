#!/usr/bin/env bash
#
# install.sh -- Point this clone's git hooks at smaug-go/hooks.
# Idempotent: safe to run multiple times.

set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

chmod +x smaug-go/hooks/pre-commit
git config core.hooksPath smaug-go/hooks

echo "core.hooksPath = $(git config --get core.hooksPath)"
echo "Pre-commit hook installed. Bypass with: git commit --no-verify"
