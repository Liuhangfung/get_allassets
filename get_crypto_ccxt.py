#!/usr/bin/env python3
"""
Top 10 Cryptocurrency data fetcher using CoinGecko API
Fetches the top 10 major cryptocurrencies by market cap and exports to JSON
"""
import json
import time
from datetime import datetime
from typing import List, Dict, Any



def get_crypto_data_coingecko():
    """Fetch specific top 10 cryptocurrency data from CoinGecko API"""
    
    import requests
    
    print("📱 Fetching TOP 10 major cryptocurrencies from CoinGecko...")
    
    # Top 10 major cryptocurrencies by market cap and popularity
    target_cryptos = {
        'bitcoin': 'BTC',
        'ethereum': 'ETH', 
        'tether': 'USDT',
        'ripple': 'XRP',
        'binancecoin': 'BNB',
        'solana': 'SOL',
        'usd-coin': 'USDC',
        'tron': 'TRX',
        'dogecoin': 'DOGE',
        'cardano': 'ADA'
    }
    
    print(f"🎯 Targeting {len(target_cryptos)} specific cryptocurrencies:")
    for coin_id, symbol in target_cryptos.items():
        print(f"   • {symbol} ({coin_id})")
    
    crypto_data = []
    
    # Fetch specific cryptocurrencies using CoinGecko IDs
    coin_ids = ','.join(target_cryptos.keys())
    url = f"https://api.coingecko.com/api/v3/coins/markets"
    params = {
        'ids': coin_ids,
        'vs_currency': 'usd',
        'order': 'market_cap_desc',
        'per_page': len(target_cryptos),
        'page': 1,
        'sparkline': 'false',
        'price_change_percentage': '24h'
    }
    
    print(f"📊 Fetching specific crypto data from CoinGecko...")
    
    try:
        response = requests.get(url, params=params, timeout=30)
        response.raise_for_status()
        cryptos = response.json()
        
        print(f"✅ Retrieved {len(cryptos)} targeted cryptocurrencies")
        
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
        
        print(f"🎉 Successfully fetched {len(crypto_data)} cryptocurrencies from CoinGecko")
        
    except Exception as e:
        print(f"❌ Error fetching crypto data: {e}")
        return []
    
    return crypto_data

def save_to_json(data: List[Dict[str, Any]], filename: str):
    """Save data to JSON file"""
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"💾 Saved {len(data)} cryptocurrencies to {filename}")

def main():
    print("🚀 TOP 10 CRYPTOCURRENCY DATA FETCHER (CoinGecko)")
    print("=" * 60)
    
    start_time = time.time()
    
    # Use CoinGecko to fetch only the top 10 major cryptocurrencies
    print("📱 Using CoinGecko to fetch top 10 major cryptocurrencies...")
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
    print(f"\n📊 CRYPTOCURRENCY SUMMARY (Top 10)")
    print("Market Cap = Current Price × Circulating Supply")
    print("-" * 100)
    print(f"{'Ticker':<8} {'Name':<20} {'Price':<12} {'Change%':<10} {'Market Cap':<15} {'Circulating Supply':<18}")
    print("-" * 100)
    
    for i, crypto in enumerate(crypto_data[:10]):
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