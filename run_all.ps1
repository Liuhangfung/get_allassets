# Trading Analysis Data Collection System - PowerShell Version
# Complete automation script for global asset ranking

$ErrorActionPreference = "Stop"

Write-Host "🚀 Starting Trading Analysis Data Collection System" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Green

# Function to check if command exists
function Test-Command {
    param($CommandName)
    return $null -ne (Get-Command $CommandName -ErrorAction SilentlyContinue)
}

# Check required tools
Write-Host "🔍 Checking system requirements..." -ForegroundColor Yellow

if (-not (Test-Command "go")) {
    Write-Host "❌ Go is not installed. Please install Go 1.19+ first." -ForegroundColor Red
    Write-Host "   Download from: https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

if (-not (Test-Command "python")) {
    Write-Host "❌ Python is not installed. Please install Python 3.8+ first." -ForegroundColor Red
    Write-Host "   Download from: https://www.python.org/downloads/" -ForegroundColor Red
    exit 1
}

# Check versions
$goVersion = (go version) -replace "go version go([0-9.]+).*", '$1'
$pythonVersion = (python --version) -replace "Python ([0-9.]+).*", '$1'

Write-Host "✅ Go version: $goVersion" -ForegroundColor Green
Write-Host "✅ Python version: $pythonVersion" -ForegroundColor Green

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "❌ .env file not found!" -ForegroundColor Red
    Write-Host "📋 Setting up .env file from template..." -ForegroundColor Yellow
    
    if (Test-Path "env.template") {
        Copy-Item "env.template" ".env"
        Write-Host "✅ .env file created from template" -ForegroundColor Green
        Write-Host "⚠️  IMPORTANT: Edit .env file with your actual API keys before continuing!" -ForegroundColor Yellow
        Write-Host "   Required: FMP_API_KEY, SUPABASE_URL, SUPABASE_KEY" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Press Enter after you've updated the .env file..."
        Read-Host
    } else {
        Write-Host "❌ env.template not found. Please create .env manually." -ForegroundColor Red
        exit 1
    }
}

# Load environment variables
Write-Host "🔧 Loading environment variables..." -ForegroundColor Yellow
if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match "^([^#].*?)=(.*)$") {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2], "Process")
        }
    }
    Write-Host "✅ Environment variables loaded" -ForegroundColor Green
} else {
    Write-Host "❌ .env file not found" -ForegroundColor Red
    exit 1
}

# Verify required environment variables
Write-Host "🔒 Verifying API credentials..." -ForegroundColor Yellow
$fmpKey = [Environment]::GetEnvironmentVariable("FMP_API_KEY")
$supabaseUrl = [Environment]::GetEnvironmentVariable("SUPABASE_URL")
$supabaseKey = [Environment]::GetEnvironmentVariable("SUPABASE_KEY")

if (-not $fmpKey -or $fmpKey -eq "your_fmp_api_key_here") {
    Write-Host "❌ FMP_API_KEY not set in .env file" -ForegroundColor Red
    exit 1
}

if (-not $supabaseUrl -or $supabaseUrl -eq "https://your-project-id.supabase.co") {
    Write-Host "❌ SUPABASE_URL not set in .env file" -ForegroundColor Red
    exit 1
}

if (-not $supabaseKey -or $supabaseKey -eq "your_supabase_service_role_key_here") {
    Write-Host "❌ SUPABASE_KEY not set in .env file" -ForegroundColor Red
    exit 1
}

Write-Host "✅ All required credentials configured" -ForegroundColor Green

# Setup Go dependencies
Write-Host "🔧 Setting up Go dependencies..." -ForegroundColor Yellow
if (-not (Test-Path "go.mod")) {
    Write-Host "Initializing Go module..." -ForegroundColor Yellow
    go mod init trading-analysis
}

try {
    go mod tidy
    Write-Host "✅ Go dependencies ready" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to install Go dependencies" -ForegroundColor Red
    exit 1
}

# Setup Python virtual environment
Write-Host "🐍 Setting up Python environment..." -ForegroundColor Yellow
if (-not (Test-Path "venv")) {
    Write-Host "Creating Python virtual environment..." -ForegroundColor Yellow
    python -m venv venv
}

# Activate virtual environment
Write-Host "Activating virtual environment..." -ForegroundColor Yellow
& "venv\Scripts\Activate.ps1"

# Install Python dependencies
Write-Host "Installing Python dependencies..." -ForegroundColor Yellow
try {
    pip install -r requirements.txt
    Write-Host "✅ Python dependencies installed" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to install Python dependencies" -ForegroundColor Red
    exit 1
}

# Create logs directory
if (-not (Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs"
}

Write-Host ""
Write-Host "🎯 Starting data collection process..." -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Green

# Step 1: Fetch global stock data with Go
Write-Host "📈 Step 1: Fetching global stock data (Go)..." -ForegroundColor Cyan
Write-Host "Running: go run stock_fmp_global.go" -ForegroundColor Gray

try {
    go run stock_fmp_global.go
    Write-Host "✅ Global stock data fetched successfully" -ForegroundColor Green
    if (Test-Path "global_assets_fmp.json") {
        $stockData = Get-Content "global_assets_fmp.json" | ConvertFrom-Json
        $stockCount = $stockData.Count
        Write-Host "   📊 Fetched $stockCount stocks" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Failed to fetch global stock data" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Step 2: Fetch cryptocurrency data with Python
Write-Host "🪙 Step 2: Fetching cryptocurrency data (Python)..." -ForegroundColor Cyan
Write-Host "Running: python get_crypto_ccxt.py" -ForegroundColor Gray

try {
    python get_crypto_ccxt.py
    Write-Host "✅ Cryptocurrency data fetched successfully" -ForegroundColor Green
    if (Test-Path "crypto_data.json") {
        $cryptoData = Get-Content "crypto_data.json" | ConvertFrom-Json
        $cryptoCount = $cryptoData.Count
        Write-Host "   🪙 Fetched $cryptoCount cryptocurrencies" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Failed to fetch cryptocurrency data" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Step 3: Combine all data and upload to Supabase
Write-Host "🔄 Step 3: Combining data and uploading to Supabase..." -ForegroundColor Cyan
Write-Host "Running: python combine_all_assets.py" -ForegroundColor Gray

try {
    python combine_all_assets.py
    Write-Host "✅ Data combined and uploaded successfully" -ForegroundColor Green
    if (Test-Path "all_assets_combined.json") {
        $allData = Get-Content "all_assets_combined.json" | ConvertFrom-Json
        $totalCount = $allData.Count
        Write-Host "   🎯 Total assets processed: $totalCount" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Failed to combine and upload data" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🎉 DATA COLLECTION COMPLETED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Green
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  • Global stocks: ✅ Fetched ($stockCount stocks)" -ForegroundColor Green
Write-Host "  • Cryptocurrencies: ✅ Fetched ($cryptoCount cryptos)" -ForegroundColor Green
Write-Host "  • Combined ranking: ✅ Uploaded ($totalCount total assets)" -ForegroundColor Green
Write-Host "  • Database: ✅ Updated in Supabase" -ForegroundColor Green
Write-Host ""
Write-Host "📁 Generated files:" -ForegroundColor Yellow
Write-Host "  • global_assets_fmp.json - Raw stock data" -ForegroundColor Gray
Write-Host "  • crypto_data.json - Raw crypto data" -ForegroundColor Gray
Write-Host "  • all_assets_combined.json - Final ranked data" -ForegroundColor Gray
Write-Host "  • combine_all_assets.log - Execution log" -ForegroundColor Gray
Write-Host ""
Write-Host "🔗 Check your Supabase dashboard for the latest data!" -ForegroundColor Cyan

# Deactivate virtual environment
deactivate

Write-Host "✨ System ready for next run!" -ForegroundColor Green 