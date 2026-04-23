#!/bin/bash

# setup-hooks.sh
# This script sets up a Git hook to enforce Conventional Commits.

HOOK_DIR=".git/hooks"
HOOK_FILE="$HOOK_DIR/commit-msg"

# Ensure the hooks directory exists
if [ ! -d "$HOOK_DIR" ]; then
    echo "Error: .git directory not found. Are you in the root of your Git repository?"
    exit 1
fi

# Create the commit-msg hook
cat << 'EOF' > "$HOOK_FILE"
#!/bin/bash

# Git hook to verify commit messages for Conventional Commits compliance.

COMMIT_MSG_FILE=$1
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")
# Extract only the first line (subject) for regex validation
SUBJECT=$(head -n 1 "$COMMIT_MSG_FILE")

# 1. Try using cocogitto (cog) if installed
if command -v cog &> /dev/null; then
    cog verify "$COMMIT_MSG"
    exit $?
fi

# 2. Fallback: Simple regex check if cog is not installed
# Regex for: type(scope)!: description
# Note: Bash uses Extended Regular Expressions (ERE) which don't support non-capturing groups (?:...)
REGEX_PATTERN="^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?(!?): .+"

if [[ ! "$SUBJECT" =~ $REGEX_PATTERN ]]; then
    echo "Error: Invalid commit message format."
    echo "Your commit message must follow Conventional Commits: <type>(<scope>)!: <description>"
    echo "Example: feat(api): add new endpoint"
    echo "See https://www.conventionalcommits.org/ for more details."
    exit 1
fi
EOF

# Make the hook executable
chmod +x "$HOOK_FILE"

echo "Success: Git commit-msg hook has been set up."
