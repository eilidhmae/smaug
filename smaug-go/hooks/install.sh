#!/usr/bin/env bash
#
# install.sh -- Point this clone's git hooks at smaug-go/hooks.
# Idempotent: safe to run multiple times.

set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

chmod +x smaug-go/hooks/pre-commit
chmod +x smaug-go/hooks/pre-push
chmod +x smaug-go/hooks/post-commit
git config core.hooksPath smaug-go/hooks

echo "core.hooksPath = $(git config --get core.hooksPath)"
echo "Pre-commit hook installed.  Bypass with: git commit --no-verify"
echo "Pre-push hook installed.    Mandatory adversary FAIL gate; no bypass."
echo "Post-commit hook installed. Non-blocking corpus-capture scan."
echo "                            Bypass: SMAUG_SKIP_POST_COMMIT_SCAN=1"
