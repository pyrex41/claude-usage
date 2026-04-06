#!/bin/bash
# setup-aliases.sh - Set up shell aliases for claude-usage (gcu)

set -e

BINARY_NAME="claude-usage"
ALIAS_NAME="gcu"
INSTALL_DIR="$HOME/.local/bin"

echo "🔧 Setting up shell aliases for $BINARY_NAME..."

# Ensure binary is in PATH
if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ] && [ ! -f "./$BINARY_NAME" ]; then
    echo "⚠️  Binary not found. Building first..."
    go build -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/claude-usage
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Function to add alias if not exists
add_alias() {
    local rc_file="$1"
    local shell_name="$2"
    
    if [ ! -f "$rc_file" ]; then
        echo "Creating $rc_file..."
        touch "$rc_file"
    fi
    
    if grep -q "alias $ALIAS_NAME=" "$rc_file" 2>/dev/null; then
        echo "✅ Alias already exists in $shell_name config"
        return 0
    fi
    
    echo "" >> "$rc_file"
    echo "# Claude Usage alias" >> "$rc_file"
    echo "alias $ALIAS_NAME='$BINARY_NAME'" >> "$rc_file"
    echo "✅ Added $ALIAS_NAME alias to $shell_name"
}

# Setup for different shells
if [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bashrc" ]; then
    add_alias "$HOME/.bashrc" "Bash"
fi

if [ -f "$HOME/.zshrc" ]; then
    add_alias "$HOME/.zshrc" "Zsh"
fi

# Fish shell
FISH_CONFIG="$HOME/.config/fish/config.fish"
if [ -f "$FISH_CONFIG" ] || command -v fish >/dev/null 2>&1; then
    mkdir -p "$(dirname "$FISH_CONFIG")"
    if [ ! -f "$FISH_CONFIG" ]; then
        touch "$FISH_CONFIG"
    fi
    
    if grep -q "alias $ALIAS_NAME " "$FISH_CONFIG" 2>/dev/null; then
        echo "✅ Alias already exists in Fish config"
    else
        echo "" >> "$FISH_CONFIG"
        echo "# Claude Usage alias" >> "$FISH_CONFIG"
        echo "alias $ALIAS_NAME $BINARY_NAME" >> "$FISH_CONFIG"
        echo "✅ Added $ALIAS_NAME alias to Fish"
    fi
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "You can now use:"
echo "  $ALIAS_NAME daily --instances     # instead of claude-usage daily --instances"
echo ""
echo "To apply changes immediately:"
echo "  Bash:   source ~/.bashrc"
echo "  Zsh:    source ~/.zshrc" 
echo "  Fish:   source ~/.config/fish/config.fish"
echo ""
echo "Or just restart your terminal."