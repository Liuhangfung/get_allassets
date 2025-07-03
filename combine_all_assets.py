#!/usr/bin/env python3
"""
Asset Combiner and Ranker with Supabase Integration
Combines traditional securities (from Go/FMP) with cryptocurrency data (from Python/CCXT)
and ranks all assets by market cap, then uploads to Supabase database
"""

import json
import glob
import os
import subprocess
import sys
from datetime import datetime, date
from typing import List, Dict, Any, Optional
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

def install_and_import(package: str):
    """Install and import a package if not available"""
    try:
        __import__(package)
    except ImportError:
        print(f"📦 Installing {package}...")
        subprocess.check_call([sys.executable, "-m", "pip", "install", package])
        __import__(package)

# Install required packages if not available
try:
    from supabase import create_client, Client
except ImportError:
    install_and_import('supabase')
    from supabase import create_client, Client

def get_supabase_client() -> Optional[Client]:
    """Initialize Supabase client"""
    url = os.getenv('SUPABASE_URL')
    key = os.getenv('SUPABASE_KEY')
    
    if not url or not key:
        print("❌ Supabase credentials not found in environment variables!")
        print("💡 Make sure SUPABASE_URL and SUPABASE_KEY are set in your .env file")
        return None
    
    try:
        supabase = create_client(url, key)
        print("✅ Supabase client initialized successfully")
        return supabase
    except Exception as e:
        print(f"❌ Failed to initialize Supabase client: {e}")
        return None

def check_dependencies():
    """Check if all required data files exist"""
    print("🔍 Checking for data files...")
    
    # Check for Go program and run it if needed
    if not os.path.exists("global_assets_fmp.json"):
        print("📈 No FMP assets file found, running Go program...")
        try:
            result = subprocess.run(["go", "run", "stock_fmp_global.go"], 
                                  capture_output=True, text=True, timeout=300)
            if result.returncode != 0:
                print(f"❌ Go program failed: {result.stderr}")
                return False
            print("✅ Go program completed successfully")
        except subprocess.TimeoutExpired:
            print("❌ Go program timed out after 5 minutes")
            return False
        except FileNotFoundError:
            print("❌ Go not found. Please install Go or add it to PATH")
            return False
    
    # Check for Python crypto program and run it if needed
    if not os.path.exists("crypto_data.json"):
        print("₿ No crypto data file found, running crypto fetcher...")
        try:
            result = subprocess.run([sys.executable, "get_crypto_ccxt.py"], 
                                  capture_output=True, text=True, timeout=300)
            if result.returncode != 0:
                print(f"❌ Crypto fetcher failed: {result.stderr}")
                return False
            print("✅ Crypto fetcher completed successfully")
        except subprocess.TimeoutExpired:
            print("❌ Crypto fetcher timed out after 5 minutes")
            return False
        except FileNotFoundError:
            print("❌ get_crypto_ccxt.py not found")
            return False
    
    return True

def load_json_file(filename: str) -> List[Dict[str, Any]]:
    """Load JSON file by filename"""
    # Handle both direct filenames and wildcard patterns for backward compatibility
    if "*" in filename:
        files = glob.glob(filename)
        if not files:
            print(f"❌ No files found matching pattern: {filename}")
            return []
        # Get the most recent file
        latest_file = max(files, key=os.path.getctime)
        print(f"📂 Loading: {latest_file}")
        filename = latest_file
    else:
        if not os.path.exists(filename):
            print(f"❌ File not found: {filename}")
            return []
        print(f"📂 Loading: {filename}")
    
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            data = json.load(f)
        print(f"✅ Loaded {len(data)} assets from {filename}")
        return data
    except Exception as e:
        print(f"❌ Error loading {filename}: {e}")
        return []

def validate_asset_data(asset: Dict[str, Any]) -> bool:
    """Validate that asset has all required fields"""
    required_fields = [
        'ticker', 'name', 'market_cap', 'current_price', 
        'previous_close', 'percentage_change', 'volume', 
        'primary_exchange', 'asset_type'
    ]
    
    for field in required_fields:
        if field not in asset:
            return False
        if field == 'market_cap' and asset['market_cap'] <= 0:  # Must have positive market cap
            return False
    
    return True

def standardize_asset_data(assets: List[Dict[str, Any]], source: str) -> List[Dict[str, Any]]:
    """Standardize asset data format and add metadata"""
    standardized = []
    
    for asset in assets:
        if not validate_asset_data(asset):
            continue
            
        # Standardize the asset data
        standardized_asset = {
            'ticker': str(asset['ticker']).upper(),
            'name': str(asset['name']),
            'market_cap': float(asset['market_cap']),
            'current_price': float(asset['current_price']),
            'previous_close': float(asset['previous_close']),
            'percentage_change': float(asset['percentage_change']),
            'volume': float(asset['volume']),
            'primary_exchange': str(asset['primary_exchange']),
            'asset_type': str(asset['asset_type']),
            'data_source': source,  # Track where data came from
            'country': asset.get('country', ''),  # Optional FMP fields
            'sector': asset.get('sector', ''),
            'industry': asset.get('industry', ''),
            'circulating_supply': asset.get('circulating_supply', 0),  # For crypto
            'image': asset.get('image', '')  # Company/asset image URL
        }
        
        standardized.append(standardized_asset)
    
    print(f"📊 Standardized {len(standardized)} assets from {source}")
    return standardized

def detect_duplicates(assets: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Remove duplicate assets, preferring traditional securities over crypto for conflicts"""
    seen_tickers = {}
    deduplicated = []
    duplicates_found = 0
    
    # Sort to prioritize traditional securities over crypto
    assets_sorted = sorted(assets, key=lambda x: (x['asset_type'] == 'crypto', x['ticker']))
    
    for asset in assets_sorted:
        ticker = asset['ticker']
        
        if ticker in seen_tickers:
            # Duplicate found
            existing = seen_tickers[ticker]
            duplicates_found += 1
            
            # Keep the one with higher market cap (usually traditional security)
            if asset['market_cap'] > existing['market_cap']:
                # Replace with higher market cap version
                for i, a in enumerate(deduplicated):
                    if a['ticker'] == ticker:
                        deduplicated[i] = asset
                        seen_tickers[ticker] = asset
                        print(f"🔄 Replaced {ticker} with higher market cap version")
                        break
            else:
                print(f"⚠️  Skipping duplicate {ticker} (lower market cap)")
        else:
            seen_tickers[ticker] = asset
            deduplicated.append(asset)
    
    print(f"🧹 Removed {duplicates_found} duplicates, kept {len(deduplicated)} unique assets")
    return deduplicated

def format_large_number(num: float) -> str:
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

def upload_to_supabase(assets: List[Dict[str, Any]], supabase: Client) -> bool:
    """Upload assets to Supabase database (matching existing schema)"""
    print(f"\n🗄️  Uploading {len(assets)} assets to Supabase...")
    
    try:
        # First, clear today's data to avoid duplicates (using snapshot_date)
        today = date.today().isoformat()
        
        print("🧹 Clearing existing data for today...")
        delete_result = supabase.table('assets').delete().eq('snapshot_date', today).execute()
        if hasattr(delete_result, 'data') and delete_result.data:
            print(f"   Deleted {len(delete_result.data)} existing records for {today}")
        
        # Prepare data for upload (matching existing schema)
        upload_data = []
        for i, asset in enumerate(assets):
            # Format percentage change
            change_24h = None
            if asset.get('percentage_change') is not None:
                try:
                    change_val = float(asset['percentage_change'])
                    change_24h = f"{change_val:.2f}%"
                except (ValueError, TypeError):
                    change_24h = str(asset['percentage_change']) if asset['percentage_change'] else None
            
            # Convert data for Supabase (including all new columns from data structure)
            db_asset = {
                # Your existing schema columns
                'rank': i + 1,  # Ranking position (1-based)
                'symbol': str(asset['ticker']).upper(),  # Use 'symbol' instead of 'ticker'
                'name': str(asset['name']),
                'market_cap': format_large_number(float(asset['market_cap'])),  # Formatted string
                'market_cap_raw': int(asset['market_cap']),  # Raw numeric value
                'price': f"${float(asset['current_price']):.8f}".rstrip('0').rstrip('.'),  # Formatted price
                'price_raw': float(asset['current_price']),  # Raw numeric price
                'change_24h': change_24h,  # Formatted percentage change
                'category': str(asset['asset_type']),  # Use 'category' instead of 'asset_type'
                'snapshot_date': today,  # Use 'snapshot_date' instead of 'date_updated'
                'today': change_24h,  # Today's change (same as change_24h for now)
                'image': asset.get('image', ''),  # Company/asset image URL
                
                # All additional columns from your actual data structure
                'ticker': str(asset['ticker']).upper(),  # Original ticker field
                'current_price': float(asset['current_price']),  # Raw current price
                'previous_close': float(asset['previous_close']) if asset.get('previous_close') else None,
                'percentage_change': float(asset['percentage_change']) if asset.get('percentage_change') is not None else None,
                'volume': int(asset['volume']) if asset.get('volume') else 0,
                'primary_exchange': asset.get('primary_exchange', '')[:100] if asset.get('primary_exchange') else None,
                'asset_type': str(asset['asset_type']),  # Original asset_type field
                'data_source': asset.get('data_source', ''),
                'country': asset.get('country', '')[:100] if asset.get('country') else None,
                'sector': asset.get('sector', '')[:100] if asset.get('sector') else None,
                'industry': asset.get('industry', '')[:100] if asset.get('industry') else None,
                'circulating_supply': int(asset['circulating_supply']) if asset.get('circulating_supply') else 0,
            }
            upload_data.append(db_asset)
        
        # Upload in batches to avoid hitting limits
        batch_size = 1000
        total_uploaded = 0
        
        for i in range(0, len(upload_data), batch_size):
            batch = upload_data[i:i + batch_size]
            
            print(f"   Uploading batch {i//batch_size + 1} ({len(batch)} records)...")
            
            result = supabase.table('assets').insert(batch).execute()
            
            if hasattr(result, 'data') and result.data:
                total_uploaded += len(result.data)
                print(f"   ✅ Uploaded {len(result.data)} records")
            else:
                print(f"   ❌ Batch upload failed")
                return False
        
        print(f"🎉 Successfully uploaded {total_uploaded} assets to Supabase!")
        print(f"📊 Data available in public.assets table for snapshot_date: {today}")
        return True
        
    except Exception as e:
        print(f"❌ Failed to upload to Supabase: {e}")
        return False

def combine_and_rank_assets() -> List[Dict[str, Any]]:
    """Main function to combine and rank all assets"""
    print("🚀 COMPREHENSIVE ASSET COMBINER AND RANKER")
    print("=" * 60)
    
    # Load traditional securities data
    print("\n📈 Loading traditional securities data...")
    traditional_assets = load_json_file("global_assets_fmp.json")
    if not traditional_assets:
        # Fallback to old formats
        traditional_assets = load_json_file("traditional_securities_*.json")
        if not traditional_assets:
            traditional_assets = load_json_file("stock_data_*.json")
    
    # Load cryptocurrency data
    print("\n₿ Loading cryptocurrency data...")
    crypto_assets = load_json_file("crypto_data.json")
    
    if not traditional_assets and not crypto_assets:
        print("❌ No data files found!")
        return []
    
    # Standardize data formats
    all_assets = []
    
    if traditional_assets:
        traditional_standardized = standardize_asset_data(traditional_assets, "fmp")
        all_assets.extend(traditional_standardized)
    
    if crypto_assets:
        crypto_standardized = standardize_asset_data(crypto_assets, "coingecko")
        all_assets.extend(crypto_standardized)
    
    print(f"\n🔗 Combined total: {len(all_assets)} assets from all sources")
    
    # Remove duplicates
    all_assets = detect_duplicates(all_assets)
    
    # Sort by market cap descending
    all_assets.sort(key=lambda x: x['market_cap'], reverse=True)
    
    print(f"\n📊 Final dataset: {len(all_assets)} unique assets ranked by market cap")
    
    return all_assets

def print_asset_summary(assets: List[Dict[str, Any]], top_n: int = 50):
    """Print a summary of the top assets"""
    if not assets:
        print("No assets to display")
        return
    
    print(f"\n=== TOP {min(top_n, len(assets))} ASSETS BY MARKET CAP (ALL MARKETS) ===")
    print("-" * 130)
    print(f"{'Rank':<5} {'Ticker':<8} {'Name':<22} {'Country':<8} {'Price':<12} {'Change%':<10} {'Market Cap':<15} {'Type':<12} {'Source':<15}")
    print("-" * 130)
    
    for i, asset in enumerate(assets[:top_n]):
        rank = i + 1
        name = asset['name'][:20] + ".." if len(asset['name']) > 22 else asset['name']
        country = asset.get('country', '')[:6] if asset.get('country') else ''
        market_cap_str = format_large_number(asset['market_cap'])
        
        # Special display for crypto and commodities
        type_display = asset['asset_type']
        if asset['asset_type'] == 'crypto':
            type_display = "₿ crypto"
        elif asset['asset_type'] == 'commodity':
            type_display = "🥇 commodity"
        
        print(f"{rank:<5} {asset['ticker']:<8} {name:<22} {country:<8} ${asset['current_price']:<11.2f} "
              f"{asset['percentage_change']:<9.2f}% {market_cap_str:<15} {type_display:<12} "
              f"{asset['data_source']:<15}")
    
    # Asset breakdown
    print(f"\n📈 ASSET BREAKDOWN:")
    asset_counts = {}
    source_counts = {}
    
    for asset in assets:
        asset_type = asset['asset_type']
        source = asset['data_source']
        
        asset_counts[asset_type] = asset_counts.get(asset_type, 0) + 1
        source_counts[source] = source_counts.get(source, 0) + 1
    
    print("By Type:")
    for asset_type, count in sorted(asset_counts.items(), key=lambda x: x[1], reverse=True):
        if asset_type == 'crypto':
            print(f"  ₿ {count:,} cryptocurrencies")
        elif asset_type == 'commodity':
            print(f"  🥇 {count:,} commodities")
        elif asset_type == 'stock':
            print(f"  📈 {count:,} individual stocks")
        else:
            print(f"  🏢 {count:,} {asset_type}s")
    
    print("\nBy Data Source:")
    for source, count in sorted(source_counts.items(), key=lambda x: x[1], reverse=True):
        print(f"  📡 {count:,} from {source}")

def save_combined_data(assets: List[Dict[str, Any]], top_n: int = 500):
    """Save the combined and ranked data to JSON file"""
    # Take top N assets
    top_assets = assets[:top_n] if len(assets) > top_n else assets
    
    # Create output filename
    filename = "all_assets_combined.json"
    
    # Save to file
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(top_assets, f, indent=2, ensure_ascii=False)
    
    print(f"\n💾 Saved top {len(top_assets)} assets to: {filename}")
    return filename

def main():
    """Main execution function"""
    start_time = datetime.now()
    
    print("🌟 GLOBAL ASSET RANKING SYSTEM WITH SUPABASE INTEGRATION")
    print("=" * 65)
    
    # Check dependencies and run data collection if needed
    if not check_dependencies():
        print("❌ Dependency check failed!")
        return
    
    # Initialize Supabase client
    supabase = get_supabase_client()
    if not supabase:
        print("⚠️  Continuing without Supabase integration...")
    
    # Combine and rank all assets
    all_assets = combine_and_rank_assets()
    
    if not all_assets:
        print("❌ No assets to process!")
        return
    
    # Print summary
    print_asset_summary(all_assets, top_n=50)
    
    # Save combined data
    filename = save_combined_data(all_assets, top_n=500)
    
    # Upload to Supabase if available
    supabase_success = False
    if supabase:
        # Only upload top 500 to database
        top_500 = all_assets[:500]
        supabase_success = upload_to_supabase(top_500, supabase)
    
    # Final stats
    duration = datetime.now() - start_time
    print(f"\n🎉 Processing completed in {duration.total_seconds():.1f} seconds!")
    print(f"📊 Total unique assets processed: {len(all_assets):,}")
    print(f"📁 JSON output file: {filename}")
    
    if supabase_success:
        print(f"🗄️  Database: Successfully uploaded to Supabase public.assets table")
        print(f"📅 Snapshot Date: {date.today().isoformat()}")
    elif supabase:
        print(f"❌ Database: Failed to upload to Supabase")
    else:
        print(f"⚠️  Database: Skipped (no Supabase credentials)")
    
    # Show top 10 quick preview
    print(f"\n🏆 TOP 10 ASSETS GLOBALLY:")
    for i, asset in enumerate(all_assets[:10]):
        emoji = "₿" if asset['asset_type'] == 'crypto' else "🥇" if asset['asset_type'] == 'commodity' else "📈"
        print(f"  {i+1:2d}. {emoji} {asset['ticker']:<8} {asset['name'][:30]:<30} "
              f"{format_large_number(asset['market_cap']):<10} market cap")
    
    print(f"\n🎯 System ready! Top 500 individual assets ranked by market cap.")
    if supabase_success:
        print(f"💡 Query your data: SELECT * FROM public.assets WHERE snapshot_date = '{date.today().isoformat()}' ORDER BY rank;")

if __name__ == "__main__":
    main() 
    