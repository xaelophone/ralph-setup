#!/bin/bash
# Install Ralph setup script globally
# Usage: curl -fsSL https://raw.githubusercontent.com/xaelophone/ralph-setup/main/install.sh | bash

set -e

INSTALL_DIR="$HOME/bin/ralph"

echo "🚀 Installing Ralph..."

# Create directory
mkdir -p "$INSTALL_DIR"

# Download setup-ralph script
if command -v curl &> /dev/null; then
    curl -fsSL "https://raw.githubusercontent.com/xaelophone/ralph-setup/main/setup-ralph" -o "$INSTALL_DIR/setup-ralph"
elif command -v wget &> /dev/null; then
    wget -q "https://raw.githubusercontent.com/xaelophone/ralph-setup/main/setup-ralph" -O "$INSTALL_DIR/setup-ralph"
else
    echo "❌ Error: curl or wget required"
    exit 1
fi

chmod +x "$INSTALL_DIR/setup-ralph"

# Check if already in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "📝 Add this to your shell config (~/.zshrc or ~/.bashrc):"
    echo ""
    echo "    export PATH=\"\$HOME/bin/ralph:\$PATH\""
    echo ""
fi

echo "✅ Ralph installed to $INSTALL_DIR/setup-ralph"
echo ""
echo "Usage:"
echo "  cd your-project"
echo "  setup-ralph"
echo "  claude"
