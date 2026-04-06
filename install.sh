#!/bin/bash
# install.sh - Simple installation script for claude-usage

set -e

echo "🚀 Installing claude-usage..."

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo "❌ Go is not installed. Please install Go first."
    echo "Visit: https://go.dev/doc/install"
    exit 1
fi

echo "✅ Go detected: $(go version)"

# Build and install
echo "📦 Building and installing..."
go install ./cmd/claude-usage

# Check if installation was successful
if command -v claude-usage >/dev/null 2>&1; then
    echo "✅ Successfully installed!"
    echo ""
    echo "Usage:"
    echo "  claude-usage daily                    # Daily report"
    echo "  claude-usage daily --instances        # By project"
    echo "  claude-usage daily --compact          # Compact view"
    echo "  claude-usage daily --help             # Show help"
    echo ""
    echo "Run 'make help' for more commands."
else
    echo "⚠️  Installation completed but binary not found in PATH."
    echo "You can run it with: ./claude-usage or add it to your PATH."
fi