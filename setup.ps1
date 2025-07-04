# Trading Analysis System - Setup Script (PowerShell Version)
# Sets up environment and dependencies only

$ErrorActionPreference = "Stop"

Write-Host "🔧 Setting up Trading Analysis System..." -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Green

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

# Setup .env file
if (-not (Test-Path ".env")) {
    Write-Host "📋 Setting up .env file from template..." -ForegroundColor Yellow
    
    if (Test-Path "env.template") {
        Copy-Item "env.template" ".env"
        Write-Host "✅ .env file created from template" -ForegroundColor Green
        Write-Host "⚠️  IMPORTANT: Edit .env file with your actual API keys!" -ForegroundColor Yellow
        Write-Host "   Required: FMP_API_KEY, SUPABASE_URL, SUPABASE_KEY" -ForegroundColor Yellow
    } else {
        Write-Host "❌ env.template not found" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✅ .env file already exists" -ForegroundColor Green
}

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

# Deactivate virtual environment
deactivate

# Create logs directory
if (-not (Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs"
}

Write-Host ""
Write-Host "✅ SETUP COMPLETED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Edit .env file with your actual API keys" -ForegroundColor Yellow
Write-Host "2. Run: .\run_all.ps1" -ForegroundColor Yellow
Write-Host ""
Write-Host "📁 Files created:" -ForegroundColor Yellow
Write-Host "  • .env - Environment variables (edit with your keys)" -ForegroundColor Gray
Write-Host "  • venv\ - Python virtual environment" -ForegroundColor Gray
Write-Host "  • logs\ - Log files directory" -ForegroundColor Gray
Write-Host ""
Write-Host "🚀 System ready for first run!" -ForegroundColor Green 