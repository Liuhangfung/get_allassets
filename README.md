# Trading Analysis System 📈

A comprehensive global asset ranking system that fetches real-time data for 500+ assets from worldwide markets and cryptocurrencies, ranks them by market capitalization, and stores the results in Supabase.

## 🌟 Features

- **Global Stock Coverage**: 800+ stocks from major exchanges worldwide
- **Cryptocurrency Data**: Real-time crypto prices and market caps
- **Intelligent Ranking**: Combined ranking by USD market cap
- **Currency Conversion**: Automatic conversion to USD for fair comparison
- **Database Integration**: Automated upload to Supabase
- **Real-time Processing**: Parallel API calls for optimal performance

## 🏗️ System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   FMP API       │    │   CoinGecko     │    │   Supabase      │
│   (Stocks)      │    │   (Crypto)      │    │   (Database)    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ stock_fmp_      │    │ get_crypto_     │    │ combine_all_    │
│ global.go       │    │ ccxt.py         │    │ assets.py       │
│ (Go Program)    │    │ (Python Script) │    │ (Python Script) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ global_assets_  │    │ crypto_data.    │    │ all_assets_     │
│ fmp.json        │    │ json            │    │ combined.json   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### 1. Clone Repository
```bash
git clone <your-repo-url>
cd get_allassets
```

### 2. Setup Environment
```bash
# Copy environment template
cp env.template .env

# Edit with your API keys
nano .env
```

### 3. Run the System
```bash
# Make script executable
chmod +x run_all.sh

# Run complete system
./run_all.sh
```

## 📋 Requirements

### System Requirements
- **Go**: 1.19+ for stock data fetching
- **Python**: 3.8+ for crypto data and processing
- **Linux/Unix**: Ubuntu, Debian, CentOS, etc.
- **Internet**: For API calls to FMP and CoinGecko

### API Keys Required
1. **FMP API Key** - Get from [Financial Modeling Prep](https://financialmodelingprep.com/)
2. **Supabase Credentials** - From your Supabase project dashboard

## ⚙️ Configuration

### Environment Variables (.env)
```bash
# FMP API Configuration
FMP_API_KEY=your_fmp_api_key_here

# Supabase Configuration  
SUPABASE_URL=https://your-project-id.supabase.co
SUPABASE_KEY=your_supabase_service_role_key_here

# Optional settings
CLEAR_EXISTING_DATA=false
```

### Supabase Database Schema
The system expects an `assets` table with these columns:

```sql
CREATE TABLE assets (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50),
    ticker VARCHAR(50),
    name VARCHAR(200),
    current_price NUMERIC,
    previous_close NUMERIC,
    percentage_change NUMERIC,
    market_cap BIGINT,
    volume BIGINT,
    circulating_supply BIGINT,
    primary_exchange VARCHAR(50),
    country VARCHAR(50),
    sector VARCHAR(100),
    industry VARCHAR(100),
    asset_type VARCHAR(50),
    image VARCHAR(500),
    rank INTEGER,
    snapshot_date DATE,
    price_raw NUMERIC,
    market_cap_raw BIGINT,
    category VARCHAR(50),
    data_source VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Add unique constraint (optional but recommended)
ALTER TABLE assets 
ADD CONSTRAINT unique_symbol_snapshot_date 
UNIQUE (symbol, snapshot_date);
```

## 📊 Data Sources

### Global Stocks (Go Program)
- **Coverage**: 800+ stocks from 15+ countries
- **Markets**: NYSE, NASDAQ, LSE, TSE, HKSE, Euronext, etc.
- **Data**: Real-time prices, market caps, volume, sectors
- **Currency**: Auto-converted to USD for ranking

**Covered Markets:**
- 🇺🇸 USA: AAPL, MSFT, GOOGL, NVDA, META, TSLA
- 🇨🇭 Switzerland: NESN.SW, NOVN.SW, ROG.SW
- 🇸🇦 Saudi Arabia: 2222.SR (Aramco), major TADAWUL stocks
- 🇬🇧 UK: SHEL.L, AZN, BP, ULVR.L
- 🇯🇵 Japan: 7203.T (Toyota), major Nikkei stocks
- 🇭🇰 Hong Kong: 700.HK (Tencent), major HKSE stocks
- And many more...

### Cryptocurrencies (Python Script)
- **Source**: CoinGecko API via CCXT
- **Coverage**: Top cryptocurrencies by market cap
- **Data**: Real-time prices, 24h changes, volumes
- **Metrics**: Market cap, circulating supply, ATH/ATL

## 🔄 System Workflow

1. **Stock Data Collection** (Go)
   - Fetches 800+ global stocks from FMP API
   - Processes in parallel batches for speed
   - Converts currencies to USD
   - Saves to `global_assets_fmp.json`

2. **Crypto Data Collection** (Python)
   - Fetches top cryptocurrencies from CoinGecko
   - Gets real-time prices and market data
   - Saves to `crypto_data.json`

3. **Data Combination & Ranking** (Python)
   - Merges stocks and crypto data
   - Ranks all assets by USD market cap
   - Selects top 500 assets
   - Uploads to Supabase database
   - Saves final ranking to `all_assets_combined.json`

## 📁 File Structure

```
├── README.md                 # This file
├── env.template             # Environment variables template
├── requirements.txt         # Python dependencies
├── run_all.sh              # Main execution script
├── go.mod                  # Go module definition
├── go.sum                  # Go dependency checksums
├── stock_fmp_global.go     # Stock data fetcher (Go)
├── get_crypto_ccxt.py      # Crypto data fetcher (Python)
├── combine_all_assets.py   # Data combiner & uploader (Python)
├── pinescript.pine         # TradingView alert script
├── global_assets_fmp.json  # Generated: Stock data
├── crypto_data.json        # Generated: Crypto data
├── all_assets_combined.json # Generated: Final ranking
├── combine_all_assets.log  # Generated: Execution log
└── venv/                   # Generated: Python virtual environment
```

## 🚀 Usage Examples

### Manual Execution
```bash
# Run individual components
go run stock_fmp_global.go           # Fetch stocks only
python get_crypto_ccxt.py            # Fetch crypto only  
python combine_all_assets.py         # Combine & upload only

# Run complete system
./run_all.sh                         # Everything automated
```

### Cron Job Setup
```bash
# Add to crontab for daily execution at 6 AM
0 6 * * * cd /path/to/get_allassets && ./run_all.sh >> logs/daily.log 2>&1
```

### Docker Deployment
```dockerfile
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y golang python3 python3-pip
COPY . /app
WORKDIR /app
RUN chmod +x run_all.sh
CMD ["./run_all.sh"]
```

## 🛠️ Troubleshooting

### Common Issues

**API Key Errors**
```bash
# Check if API keys are set
grep FMP_API_KEY .env
grep SUPABASE_URL .env
```

**Permission Errors**  
```bash
# Make script executable
chmod +x run_all.sh

# Check Go installation
go version

# Check Python installation
python3 --version
```

**Database Connection Issues**
```bash
# Test Supabase connection
python3 -c "
import os
from supabase import create_client
client = create_client(os.getenv('SUPABASE_URL'), os.getenv('SUPABASE_KEY'))
result = client.table('assets').select('id').limit(1).execute()
print('Connection successful!')
"
```

### Performance Optimization

- **Parallel Processing**: Adjust batch sizes in Go program
- **Rate Limiting**: Modify sleep times between API calls
- **Memory Usage**: Monitor for large JSON files
- **Database Performance**: Add indexes on frequently queried columns

## 🔒 Security Notes

- **API Keys**: Never commit `.env` file to version control
- **Database**: Use service role key for server environments
- **Network**: Consider API rate limits and quotas
- **Logs**: Rotate log files to prevent disk space issues

## 📈 Output Format

### Final JSON Structure
```json
{
  "ticker": "AAPL",
  "name": "Apple Inc.",
  "market_cap": 3500000000000,
  "current_price": 185.50,
  "previous_close": 184.20,
  "percentage_change": 0.71,
  "volume": 50000000,
  "rank": 1,
  "asset_type": "stock",
  "country": "US",
  "sector": "Technology",
  "snapshot_date": "2025-01-15"
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch  
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For issues and questions:
1. Check this README
2. Review the generated log files
3. Open an issue on GitHub
4. Contact the development team

---

**Last Updated**: January 2025  
**Version**: 2.0  
**Maintained by**: Trading Analysis Team 