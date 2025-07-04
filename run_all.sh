#!/bin/bash

# Trading Analysis Data Collection System
# Complete automation script for global asset ranking

set -e  # Exit on any error

echo "🚀 Starting Trading Analysis Data Collection System"
echo "=================================================="

# Check if running as root or with proper permissions
if [ "$EUID" -eq 0 ]; then
    echo "⚠️  Running as root. Consider using a non-root user for security."
fi

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

# Check Go version
GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
echo "✅ Go version: $GO_VERSION"

# Check Python version  
PYTHON_VERSION=$(python3 --version | grep -oP 'Python \K[0-9]+\.[0-9]+')
echo "✅ Python version: $PYTHON_VERSION"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found!"
    echo "📋 Setting up .env file from template..."
    
    if [ -f "env.template" ]; then
        cp env.template .env
        echo "✅ .env file created from template"
        echo "⚠️  IMPORTANT: Edit .env file with your actual API keys before continuing!"
        echo "   Required: FMP_API_KEY, SUPABASE_URL, SUPABASE_KEY"
        echo ""
        echo "Press Enter after you've updated the .env file..."
        read -r
    else
        echo "❌ env.template not found. Please create .env manually."
        exit 1
    fi
fi

# Source environment variables
echo "🔧 Loading environment variables..."
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
    echo "✅ Environment variables loaded"
else
    echo "❌ .env file not found"
    exit 1
fi

# Verify required environment variables
echo "🔒 Verifying API credentials..."
if [ -z "$FMP_API_KEY" ] || [ "$FMP_API_KEY" = "your_fmp_api_key_here" ]; then
    echo "❌ FMP_API_KEY not set in .env file"
    exit 1
fi

if [ -z "$SUPABASE_URL" ] || [ "$SUPABASE_URL" = "https://your-project-id.supabase.co" ]; then
    echo "❌ SUPABASE_URL not set in .env file"
    exit 1
fi

if [ -z "$SUPABASE_KEY" ] || [ "$SUPABASE_KEY" = "your_supabase_service_role_key_here" ]; then
    echo "❌ SUPABASE_KEY not set in .env file"
    exit 1
fi

echo "✅ All required credentials configured"

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

# Create logs directory
mkdir -p logs

echo ""
echo "🎯 Starting data collection process..."
echo "=================================================="

# Step 1: Fetch global stock data with Go
echo "📈 Step 1: Fetching global stock data (Go)..."
echo "Running: go run stock_fmp_global.go"

if go run stock_fmp_global.go; then
    echo "✅ Global stock data fetched successfully"
    if [ -f "global_assets_fmp.json" ]; then
        STOCK_COUNT=$(jq length global_assets_fmp.json 2>/dev/null || echo "unknown")
        echo "   📊 Fetched $STOCK_COUNT stocks"
    fi
else
    echo "❌ Failed to fetch global stock data"
    exit 1
fi

echo ""

# Step 2: Fetch cryptocurrency data with Python
echo "🪙 Step 2: Fetching cryptocurrency data (Python)..."
echo "Running: python get_crypto_ccxt.py"

if python get_crypto_ccxt.py; then
    echo "✅ Cryptocurrency data fetched successfully"
    if [ -f "crypto_data.json" ]; then
        CRYPTO_COUNT=$(jq length crypto_data.json 2>/dev/null || echo "unknown")
        echo "   🪙 Fetched $CRYPTO_COUNT cryptocurrencies"
    fi
else
    echo "❌ Failed to fetch cryptocurrency data"
    exit 1
fi

echo ""

# Step 3: Combine all data and upload to Supabase
echo "🔄 Step 3: Combining data and uploading to Supabase..."
echo "Running: python combine_all_assets.py"

if python combine_all_assets.py; then
    echo "✅ Data combined and uploaded successfully"
    if [ -f "all_assets_combined.json" ]; then
        TOTAL_COUNT=$(jq length all_assets_combined.json 2>/dev/null || echo "unknown")
        echo "   🎯 Total assets processed: $TOTAL_COUNT"
    fi
else
    echo "❌ Failed to combine and upload data"
    exit 1
fi

echo ""
echo "🎉 DATA COLLECTION COMPLETED SUCCESSFULLY!"
echo "=================================================="
echo "Summary:"
echo "  • Global stocks: ✅ Fetched ($STOCK_COUNT stocks)"
echo "  • Cryptocurrencies: ✅ Fetched ($CRYPTO_COUNT cryptos)"  
echo "  • Combined ranking: ✅ Uploaded ($TOTAL_COUNT total assets)"
echo "  • Database: ✅ Updated in Supabase"
echo ""
echo "📁 Generated files:"
echo "  • global_assets_fmp.json - Raw stock data"
echo "  • crypto_data.json - Raw crypto data"
echo "  • all_assets_combined.json - Final ranked data"
echo "  • combine_all_assets.log - Execution log"
echo ""
echo "🔗 Check your Supabase dashboard for the latest data!"

# Deactivate virtual environment
deactivate

echo "✨ System ready for next run!" 