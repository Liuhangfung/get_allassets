#!/bin/bash

# Trading Analysis System - Setup Script
# Sets up environment and dependencies only

echo "🔧 Setting up Trading Analysis System..."
echo "======================================"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check required tools
echo "🔍 Checking system requirements..."

if ! command_exists go; then
    echo "❌ Go is not installed. Please install Go 1.19+ first."
    echo "   Ubuntu/Debian: sudo apt install golang-go"
    echo "   CentOS/RHEL: sudo yum install golang"
    exit 1
fi

if ! command_exists python3; then
    echo "❌ Python3 is not installed. Please install Python 3.8+ first."
    echo "   Ubuntu/Debian: sudo apt install python3 python3-pip python3-venv"
    exit 1
fi

echo "✅ Go version: $(go version | grep -oP 'go\K[0-9]+\.[0-9]+')"
echo "✅ Python version: $(python3 --version | grep -oP 'Python \K[0-9]+\.[0-9]+')"

# Setup .env file
if [ ! -f ".env" ]; then
    echo "📋 Setting up .env file from template..."
    
    if [ -f "env.template" ]; then
        cp env.template .env
        echo "✅ .env file created from template"
        echo "⚠️  IMPORTANT: Edit .env file with your actual API keys!"
        echo "   Required: FMP_API_KEY, SUPABASE_URL, SUPABASE_KEY"
    else
        echo "❌ env.template not found"
        exit 1
    fi
else
    echo "✅ .env file already exists"
fi

# Setup Go dependencies
echo "🔧 Setting up Go dependencies..."
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init trading-analysis
fi

if ! go mod tidy; then
    echo "❌ Failed to install Go dependencies"
    exit 1
fi
echo "✅ Go dependencies ready"

# Setup Python virtual environment
echo "🐍 Setting up Python environment..."
if [ ! -d "venv" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install Python dependencies
echo "Installing Python dependencies..."
if ! pip install -r requirements.txt; then
    echo "❌ Failed to install Python dependencies"
    exit 1
fi
echo "✅ Python dependencies installed"

# Deactivate virtual environment
deactivate

# Create logs directory
mkdir -p logs

# Make run scripts executable
chmod +x run_all.sh

echo ""
echo "✅ SETUP COMPLETED SUCCESSFULLY!"
echo "================================"
echo "Next steps:"
echo "1. Edit .env file with your actual API keys"
echo "2. Run: ./run_all.sh"
echo ""
echo "📁 Files created:"
echo "  • .env - Environment variables (edit with your keys)"
echo "  • venv/ - Python virtual environment"
echo "  • logs/ - Log files directory"
echo ""
echo "🚀 System ready for first run!" 