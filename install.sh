#!/bin/bash
# SSHer Installation Script
# Developed by Inioluwa Adeyinka

echo "=================================="
echo "  SSHer v3.0.0 Installation"
echo "  Developed by Inioluwa Adeyinka"
echo "=================================="
echo

# Check Python version
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is required but not installed."
    exit 1
fi

PYTHON_VERSION=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
echo "Found Python $PYTHON_VERSION"

# Install dependencies
echo
echo "Installing dependencies..."
pip3 install -r requirements.txt

# Install the package
echo
echo "Installing SSHer..."
pip3 install -e .

# Enable shell completions
echo
echo "To enable tab completion, add one of the following to your shell config:"
echo
echo "  Bash (~/.bashrc):"
echo '    eval "$(ssher completion bash)"'
echo
echo "  Zsh (~/.zshrc):"
echo '    eval "$(ssher completion zsh)"'

echo
echo "=================================="
echo "  Installation Complete!"
echo "=================================="
echo
echo "Run 'ssher' in your terminal to start."
echo "Run 'ssher --version' to verify installation."
echo
