#!/bin/bash
# Install Ralph setup script globally
# Usage: curl -fsSL https://raw.githubusercontent.com/xaelophone/ralph-setup/main/install.sh | bash

set -e

INSTALL_DIR="$HOME/bin/ralph"

echo "🚀 Installing Ralph..."

# Create directory
mkdir -p "$INSTALL_DIR"

# Download scripts
download_file() {
    local file="$1"
    if command -v curl &> /dev/null; then
        curl -fsSL "https://raw.githubusercontent.com/xaelophone/ralph-setup/main/$file" -o "$INSTALL_DIR/$file"
    elif command -v wget &> /dev/null; then
        wget -q "https://raw.githubusercontent.com/xaelophone/ralph-setup/main/$file" -O "$INSTALL_DIR/$file"
    else
        echo "❌ Error: curl or wget required"
        exit 1
    fi
    chmod +x "$INSTALL_DIR/$file"
}

download_file "setup-ralph"
download_file "ralph-loop"

# Check if already in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "📝 Add this to your shell config (~/.zshrc or ~/.bashrc):"
    echo ""
    echo "    export PATH=\"\$HOME/bin/ralph:\$PATH\""
    echo ""
fi

echo "✅ Ralph tools installed to $INSTALL_DIR/"
echo "   • setup-ralph  - Initialize Ralph workflow in a project"
echo "   • ralph-loop   - Run Claude autonomously (for overnight runs)"
echo ""
echo "Usage:"
echo "  cd your-project"
echo "  setup-ralph       # Initialize Ralph files"
echo "  claude            # Manual mode"
echo "  ralph-loop        # Autonomous mode (recommended for overnight)"
