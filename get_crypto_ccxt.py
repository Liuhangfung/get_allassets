#!/usr/bin/env python3
"""
Cryptocurrency data fetcher using CCXT
Fetches top cryptocurrencies by market cap and exports to JSON
"""

import ccxt
import json
import time
from datetime import datetime
from typing import List, Dict, Any
import asyncio
import aiohttp

def get_crypto_data_ccxt():
    """Fetch cryptocurrency data using CCXT from multiple exchanges"""
    
    print("🔄 Initializing CCXT exchanges...")
    
    # Initialize multiple exchanges for comprehensive data
    exchanges = []
    
    # Binance - most comprehensive
    try:
        binance = ccxt.binance({
            'rateLimit': 1200,
            'enableRateLimit': True,
        })
        exchanges.append(('binance', binance))
        print("✅ Binance initialized")
    except Exception as e:
        print(f"❌ Binance failed: {e}")
    
    # Coinbase Pro
    try:
        coinbase = ccxt.coinbasepro({
            'rateLimit': 1000,
            'enableRateLimit': True,
        })
        exchanges.append(('coinbase', coinbase))
        print("✅ Coinbase Pro initialized")
    except Exception as e:
        print(f"❌ Coinbase Pro failed: {e}")
    
    if not exchanges:
        print("❌ No exchanges available, falling back to CoinGecko...")
        return get_crypto_data_coingecko()
    
    print("⚠️  Note: CCXT provides exchange trading data but NOT circulating supply")
    print("⚠️  Cannot calculate accurate Market Cap without circulating supply data")
    print("⚠️  Using CoinGecko as primary source for accurate market caps...")
    
    print(f"📊 Using {len(exchanges)} exchanges for crypto data")
    
    # Get tickers from primary exchange (Binance)
    exchange_name, exchange = exchanges[0]
    
    try:
        print(f"🔄 Fetching tickers from {exchange_name}...")
        tickers = exchange.fetch_tickers()
        print(f"✅ Fetched {len(tickers)} tickers from {exchange_name}")
        
        # Filter to USD pairs and sort by volume
        usd_pairs = []
        for symbol, ticker in tickers.items():
            if '/USDT' in symbol or '/USD' in symbol or '/BUSD' in symbol:
                if ticker['quoteVolume'] and ticker['quoteVolume'] > 1000000:  # Min $1M volume
                    usd_pairs.append((symbol, ticker))
        
        # Sort by volume descending
        usd_pairs.sort(key=lambda x: x[1]['quoteVolume'], reverse=True)
        
        print(f"🎯 Found {len(usd_pairs)} USD pairs with significant volume")
        
        crypto_data = []
        processed = 0
        
        for symbol, ticker in usd_pairs[:500]:  # Top 500 by volume
            try:
                # Extract base currency (e.g., BTC from BTC/USDT)
                base_currency = symbol.split('/')[0]
                
                # Skip stablecoins and wrapped tokens
                if base_currency.upper() in ['USDT', 'USDC', 'BUSD', 'DAI', 'TUSD', 'WBTC', 'WETH']:
                    continue
                
                # Get price data
                current_price = ticker['last'] if ticker['last'] else ticker['close']
                volume_24h = ticker['quoteVolume']
                
                # NOTE: CCXT doesn't provide circulating supply or market cap
                # We can't calculate accurate market cap from exchange data alone
                # Setting to 0 - this is why we use CoinGecko as primary source
                estimated_market_cap = 0  # Cannot calculate without circulating supply
                
                # Calculate percentage change
                percentage_change = ticker['percentage'] if ticker['percentage'] else 0
                
                # Calculate previous close
                previous_close = current_price * (1 - percentage_change / 100) if percentage_change else current_price
                
                stock_data = {
                    "ticker": base_currency,
                    "name": f"{base_currency} Cryptocurrency",
                    "market_cap": estimated_market_cap,
                    "current_price": current_price,
                    "previous_close": previous_close,
                    "percentage_change": percentage_change,
                    "volume": volume_24h,
                    "primary_exchange": exchange_name.title(),
                    "asset_type": "crypto",
                    "image": "",  # No images available from CCXT
                }
                
                crypto_data.append(stock_data)
                processed += 1
                
                if processed % 50 == 0:
                    print(f"✅ Processed {processed} cryptocurrencies...")
                
            except Exception as e:
                print(f"❌ Error processing {symbol}: {e}")
                continue
        
        print(f"🎉 Successfully processed {len(crypto_data)} cryptocurrencies")
        return crypto_data
        
    except Exception as e:
        print(f"❌ CCXT failed: {e}")
        print("🔄 Falling back to CoinGecko...")
        return get_crypto_data_coingecko()

def get_crypto_data_coingecko():
    """Fallback: Fetch cryptocurrency data from CoinGecko API"""
    
    import requests
    
    print("📱 Fetching cryptocurrency data from CoinGecko...")
    
    crypto_data = []
    per_page = 250
    pages = 3  # Top 750 cryptos to ensure we get top 500 after filtering
    
    for page in range(1, pages + 1):
        url = f"https://api.coingecko.com/api/v3/coins/markets"
        params = {
            'vs_currency': 'usd',
            'order': 'market_cap_desc',
            'per_page': per_page,
            'page': page,
            'sparkline': 'false',
            'price_change_percentage': '24h'
        }
        
        print(f"📊 Fetching CoinGecko page {page}/{pages}...")
        
        try:
            response = requests.get(url, params=params, timeout=30)
            response.raise_for_status()
            cryptos = response.json()
            
            print(f"✅ Parsed {len(cryptos)} cryptocurrencies from page {page}")
            
            for crypto in cryptos:
                # Skip if missing essential data
                if not crypto.get('market_cap') or crypto.get('market_cap') <= 0:
                    continue
                if not crypto.get('current_price') or crypto.get('current_price') <= 0:
                    continue
                if not crypto.get('circulating_supply') or crypto.get('circulating_supply') <= 0:
                    continue
                    
                percentage_change = crypto.get('price_change_percentage_24h', 0) or 0
                current_price = crypto.get('current_price')
                circulating_supply = crypto.get('circulating_supply')
                previous_close = current_price * (1 - percentage_change / 100) if percentage_change else current_price
                
                # Market Cap = Current Price × Circulating Supply
                # CoinGecko provides this pre-calculated, but let's validate it
                calculated_market_cap = current_price * circulating_supply
                api_market_cap = crypto.get('market_cap')
                
                # Use the API market cap (it's more accurate as it handles edge cases)
                # but validate it's reasonable
                if abs(calculated_market_cap - api_market_cap) / api_market_cap > 0.1:  # More than 10% difference
                    print(f"⚠️  Market cap mismatch for {crypto['symbol']}: API={api_market_cap:.0f}, Calculated={calculated_market_cap:.0f}")
                
                stock_data = {
                    "ticker": crypto['symbol'].upper(),
                    "name": crypto['name'],
                    "market_cap": api_market_cap,  # Real market cap: Price × Circulating Supply
                    "current_price": current_price,
                    "previous_close": previous_close,
                    "percentage_change": percentage_change,
                    "volume": crypto.get('total_volume', 0) or 0,
                    "circulating_supply": circulating_supply,  # Add this for transparency
                    "primary_exchange": "Cryptocurrency",
                    "asset_type": "crypto",
                    "image": crypto.get('image', ''),  # Crypto logo image URL
                }
                
                crypto_data.append(stock_data)
            
            # Rate limiting
            print("⏳ Waiting 3 seconds to respect CoinGecko rate limits...")
            time.sleep(3)
            
        except Exception as e:
            print(f"❌ Error fetching page {page}: {e}")
            continue
    
    print(f"🎉 Successfully fetched {len(crypto_data)} cryptocurrencies from CoinGecko")
    return crypto_data

def save_to_json(data: List[Dict[str, Any]], filename: str):
    """Save data to JSON file"""
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"💾 Saved {len(data)} cryptocurrencies to {filename}")

def main():
    print("🚀 CRYPTOCURRENCY DATA FETCHER (CCXT + CoinGecko)")
    print("=" * 60)
    
    start_time = time.time()
    
    # Use CoinGecko as primary source since it has accurate market caps
    print("📱 Using CoinGecko as primary source for accurate market cap data...")
    crypto_data = get_crypto_data_coingecko()
    
    if not crypto_data:
        print("❌ No cryptocurrency data collected!")
        return
    
    # Sort by market cap descending
    crypto_data.sort(key=lambda x: x['market_cap'], reverse=True)
    
    # Save to JSON
    filename = "crypto_data.json"
    save_to_json(crypto_data, filename)
    
    # Print summary
    print(f"\n📊 CRYPTOCURRENCY SUMMARY (Top 20)")
    print("Market Cap = Current Price × Circulating Supply")
    print("-" * 100)
    print(f"{'Ticker':<8} {'Name':<20} {'Price':<12} {'Change%':<10} {'Market Cap':<15} {'Circulating Supply':<18}")
    print("-" * 100)
    
    for i, crypto in enumerate(crypto_data[:20]):
        market_cap_str = format_large_number(crypto['market_cap'])
        supply_str = format_large_number(crypto.get('circulating_supply', 0))
        name = crypto['name'][:18] if len(crypto['name']) > 18 else crypto['name']
        print(f"{crypto['ticker']:<8} {name:<20} ${crypto['current_price']:<11.2f} {crypto['percentage_change']:<9.2f}% {market_cap_str:<15} {supply_str:<18}")
    
    duration = time.time() - start_time
    print(f"\n🎉 Completed in {duration:.1f} seconds!")
    print(f"📈 Total cryptocurrencies: {len(crypto_data)}")

def format_large_number(num):
    """Format large numbers with K, M, B, T suffixes"""
    if num >= 1e12:
        return f"{num/1e12:.1f}T"
    elif num >= 1e9:
        return f"{num/1e9:.1f}B"
    elif num >= 1e6:
        return f"{num/1e6:.1f}M"
    elif num >= 1e3:
        return f"{num/1e3:.1f}K"
    else:
        return f"{num:.0f}"

if __name__ == "__main__":
    main() 