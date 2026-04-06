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

# Setup shell aliases (gcu command)
echo "🔧 Setting up shell aliases..."
./setup-aliases.sh

# Check if installation was successful
if command -v claude-usage >/dev/null 2>&1; then
    echo "✅ Successfully installed!"
    echo ""
    echo "You can now use either:"
    echo "  claude-usage daily --instances        # Full command"
    echo "  gcu daily --instances                 # Short alias"
    echo ""
    echo "Run 'make help' for more commands."
    echo "Run './setup-aliases.sh' to update aliases later."
else
    echo "⚠️  Installation completed but binary not found in PATH."
    echo "You can run it with: ./claude-usage or add it to your PATH."
fi