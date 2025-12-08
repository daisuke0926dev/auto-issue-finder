#!/bin/bash
# Install Git hooks for the sleepship project

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

echo "Installing Git hooks..."

# Create pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/sh
echo "Running golangci-lint..."
golangci-lint run
if [ $? -ne 0 ]; then
    echo "Lint failed. Please fix the errors before committing."
    exit 1
fi
EOF

# Make pre-commit hook executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "Git hooks installed successfully!"
echo "Pre-commit hook will now run golangci-lint before each commit."
