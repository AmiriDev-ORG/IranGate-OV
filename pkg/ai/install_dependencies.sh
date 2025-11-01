#!/bin/bash
echo "🤖 Installing IRANGATE AI Agent Dependencies..."
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed. Please install Python 3.7 or higher."
    exit 1
fi
PYTHON_VERSION=$(python3 -c 'import sys; print(".".join(map(str, sys.version_info[:2])))')
echo "✅ Python version: $PYTHON_VERSION"
if ! command -v pip3 &> /dev/null; then
    echo "📦 Installing pip3..."
    sudo apt update
    sudo apt install -y python3-pip
fi
echo "📦 Installing Python packages..."
pip3 install --user -r requirements.txt
echo "🔍 Verifying installation..."
python3 -c "
try:
    import psutil
    print('✅ psutil installed successfully')
except ImportError:
    print('❌ psutil installation failed')
try:
    import requests
    print('✅ requests installed successfully')
except ImportError:
    print('❌ requests installation failed')
try:
    import openai
    print('✅ openai installed successfully')
except ImportError:
    print('❌ openai installation failed')
"
echo "🧪 Testing AI agent syntax..."
python3 -m py_compile ai_agent.py
if [ $? -eq 0 ]; then
    echo "✅ AI agent syntax is valid"
else
    echo "❌ AI agent syntax errors found"
    exit 1
fi
echo ""
echo "🎉 AI Agent dependencies installation complete!"
echo ""
echo "📋 Next steps:"
echo "1. Configure AI settings:"
echo "   cp config.json.example /etc/irangate/ai_config.json"
echo "   nano /etc/irangate/ai_config.json"
echo ""
echo "2. Test the AI agent:"
echo "   python3 ai_agent.py"
echo ""
echo "3. Integrate with IRANGATE:"
echo "   irangate ai start"
echo "   irangate ai status"
echo ""