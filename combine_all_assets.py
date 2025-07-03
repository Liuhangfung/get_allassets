#!/usr/bin/env python3
"""
Global Asset Combiner and Ranker
Combines stocks, crypto, and commodities into a single ranked list
"""

import json
import os
import sys
import logging
from datetime import datetime
from typing import List, Dict, Any, Optional
import subprocess

# Auto-install required packages
def install_and_import(package_name: str, import_name: str = None):
    """Install and import a package if not available"""
    if import_name is None:
        import_name = package_name
    
    try:
        __import__(import_name)
    except ImportError:
        print(f"📦 Installing {package_name}...")
        subprocess.check_call([sys.executable, "-m", "pip", "install", package_name])
        __import__(import_name)

# Install required packages
install_and_import("python-dotenv", "dotenv")
install_and_import("supabase")
install_and_import("requests")

from dotenv import load_dotenv
from supabase import create_client, Client

# Load environment variables
load_dotenv()

# Set up logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class GlobalAssetCombiner:
    def __init__(self):
        self.supabase_url = os.getenv('SUPABASE_URL')
        self.supabase_key = os.getenv('SUPABASE_ANON_KEY')
        self.supabase_client = None
        
        if self.supabase_url and self.supabase_key:
            try:
                self.supabase_client = create_client(self.supabase_url, self.supabase_key)
                logger.info("✅ Supabase client initialized")
            except Exception as e:
                logger.warning(f"⚠️ Supabase initialization failed: {e}")
    
    def load_json_file(self, filename: str) -> List[Dict]:
        """Load JSON file with error handling"""
        try:
            if os.path.exists(filename):
                with open(filename, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    logger.info(f"📄 Loaded {len(data)} items from {filename}")
                    return data
            else:
                logger.warning(f"⚠️ File not found: {filename}")
                return []
        except Exception as e:
            logger.error(f"❌ Error loading {filename}: {e}")
            return []
    
    def standardize_asset_data(self, asset: Dict, data_source: str) -> Dict:
        """Standardize asset data format for database insertion"""
        
        # Handle different data formats
        if data_source == "stock":
            return {
                'rank': 0,  # Will be assigned after sorting
                'symbol': asset.get('ticker', ''),
                'ticker': asset.get('ticker', ''),
                'name': asset.get('name', ''),
                'market_cap': float(asset.get('market_cap', 0)),
                'market_cap_raw': float(asset.get('market_cap', 0)),
                'current_price': float(asset.get('current_price', 0)),
                'price_raw': float(asset.get('current_price', 0)),
                'previous_close': float(asset.get('previous_close', 0)),
                'percentage_change': float(asset.get('percentage_change', 0)),
                'volume': float(asset.get('volume', 0)),
                'primary_exchange': asset.get('primary_exchange', ''),
                'country': asset.get('country', ''),
                'sector': asset.get('sector', ''),
                'industry': asset.get('industry', ''),
                'circulating_supply': None,  # Not applicable for stocks
                'category': asset.get('asset_type', 'stock'),
                'asset_type': asset.get('asset_type', 'stock'),
                'data_source': data_source,
                'image': asset.get('image', ''),
                'snapshot_date': datetime.now().isoformat()
            }
        
        elif data_source == "crypto":
            return {
                'rank': 0,  # Will be assigned after sorting
                'symbol': asset.get('symbol', ''),
                'ticker': asset.get('symbol', ''),
                'name': asset.get('name', ''),
                'market_cap': float(asset.get('market_cap', 0)),
                'market_cap_raw': float(asset.get('market_cap', 0)),
                'current_price': float(asset.get('current_price', 0)),
                'price_raw': float(asset.get('current_price', 0)),
                'previous_close': float(asset.get('current_price', 0)),  # Crypto doesn't have previous close
                'percentage_change': float(asset.get('price_change_percentage_24h', 0)),
                'volume': float(asset.get('total_volume', 0)),
                'primary_exchange': 'Multiple',
                'country': 'Global',
                'sector': 'Cryptocurrency',
                'industry': 'Blockchain',
                'circulating_supply': float(asset.get('circulating_supply', 0)),
                'category': 'crypto',
                'asset_type': 'crypto',
                'data_source': data_source,
                'image': asset.get('image', ''),
                'snapshot_date': datetime.now().isoformat()
            }
        
        else:
            # Generic fallback
            return {
                'rank': 0,
                'symbol': asset.get('symbol', asset.get('ticker', '')),
                'ticker': asset.get('ticker', asset.get('symbol', '')),
                'name': asset.get('name', ''),
                'market_cap': float(asset.get('market_cap', 0)),
                'market_cap_raw': float(asset.get('market_cap', 0)),
                'current_price': float(asset.get('current_price', asset.get('price', 0))),
                'price_raw': float(asset.get('current_price', asset.get('price', 0))),
                'previous_close': float(asset.get('previous_close', asset.get('current_price', 0))),
                'percentage_change': float(asset.get('percentage_change', 0)),
                'volume': float(asset.get('volume', 0)),
                'primary_exchange': asset.get('primary_exchange', ''),
                'country': asset.get('country', ''),
                'sector': asset.get('sector', ''),
                'industry': asset.get('industry', ''),
                'circulating_supply': asset.get('circulating_supply'),
                'category': asset.get('asset_type', 'unknown'),
                'asset_type': asset.get('asset_type', 'unknown'),
                'data_source': data_source,
                'image': asset.get('image', ''),
                'snapshot_date': datetime.now().isoformat()
            }
    
    def remove_duplicates(self, assets: List[Dict]) -> List[Dict]:
        """Remove duplicate assets, keeping the one with highest market cap"""
        seen = {}
        
        for asset in assets:
            symbol = asset.get('symbol', '').upper()
            name = asset.get('name', '').upper()
            
            # Create a key for deduplication
            key = f"{symbol}|{name}"
            
            if key not in seen or asset['market_cap'] > seen[key]['market_cap']:
                seen[key] = asset
        
        result = list(seen.values())
        logger.info(f"🔄 Removed duplicates: {len(assets)} → {len(result)} unique assets")
        return result
    
    def format_large_number(self, num: float) -> str:
        """Format large numbers for display"""
        if num >= 1e12:
            return f"${num/1e12:.1f}T"
        elif num >= 1e9:
            return f"${num/1e9:.1f}B"
        elif num >= 1e6:
            return f"${num/1e6:.1f}M"
        elif num >= 1e3:
            return f"${num/1e3:.1f}K"
        else:
            return f"${num:.0f}"
    
    def combine_and_rank_assets(self) -> List[Dict]:
        """Combine all asset types and rank by market cap"""
        logger.info("🌍 Starting global asset combination and ranking...")
        
        all_assets = []
        
        # Load traditional securities (stocks, commodities)
        stock_files = [
            "global_assets_fmp.json",
            "global_assets_fmp_2025-07-03.json",
            "global_assets_fmp_2025-07-02.json"
        ]
        
        for filename in stock_files:
            stock_data = self.load_json_file(filename)
            if stock_data:
                for asset in stock_data:
                    standardized = self.standardize_asset_data(asset, "stock")
                    all_assets.append(standardized)
                break  # Use first available file
        
        # Load cryptocurrency data
        crypto_files = [
            "crypto_data.json",
            "crypto_data_2025-07-03.json",
            "crypto_data_2025-07-02.json"
        ]
        
        for filename in crypto_files:
            crypto_data = self.load_json_file(filename)
            if crypto_data:
                for asset in crypto_data:
                    standardized = self.standardize_asset_data(asset, "crypto")
                    all_assets.append(standardized)
                break  # Use first available file
        
        # Remove duplicates
        all_assets = self.remove_duplicates(all_assets)
        
        # Sort by market cap (descending)
        all_assets.sort(key=lambda x: x['market_cap'], reverse=True)
        
        # Assign ranks
        for i, asset in enumerate(all_assets, 1):
            asset['rank'] = i
        
        # Limit to top 500
        top_assets = all_assets[:500]
        
        logger.info(f"📊 Final ranking: {len(top_assets)} assets")
        
        # Show top 10
        logger.info("🏆 Top 10 Global Assets:")
        for i, asset in enumerate(top_assets[:10], 1):
            logger.info(f"   {i:2d}. {asset['symbol']:8s} | {asset['name']:30s} | {self.format_large_number(asset['market_cap'])}")
        
        return top_assets
    
    def save_combined_data(self, assets: List[Dict], filename: str = "all_assets_combined.json"):
        """Save combined asset data to JSON file"""
        try:
            with open(filename, 'w', encoding='utf-8') as f:
                json.dump(assets, f, indent=2, ensure_ascii=False)
            logger.info(f"💾 Saved {len(assets)} assets to {filename}")
        except Exception as e:
            logger.error(f"❌ Error saving to {filename}: {e}")
    
    def upload_to_supabase(self, assets: List[Dict]) -> bool:
        """Upload assets to Supabase database"""
        if not self.supabase_client:
            logger.warning("⚠️ Supabase client not available, skipping upload")
            return False
        
        try:
            # Clear existing data
            logger.info("🗑️ Clearing existing data...")
            self.supabase_client.table('assets').delete().neq('id', 0).execute()
            
            # Insert new data in batches
            batch_size = 100
            total_uploaded = 0
            
            for i in range(0, len(assets), batch_size):
                batch = assets[i:i+batch_size]
                
                # Remove None values and convert to proper format
                clean_batch = []
                for asset in batch:
                    clean_asset = {k: v for k, v in asset.items() if v is not None}
                    clean_batch.append(clean_asset)
                
                self.supabase_client.table('assets').insert(clean_batch).execute()
                total_uploaded += len(batch)
                logger.info(f"📤 Uploaded batch {i//batch_size + 1}: {total_uploaded}/{len(assets)} assets")
            
            logger.info(f"✅ Successfully uploaded {total_uploaded} assets to Supabase")
            return True
            
        except Exception as e:
            logger.error(f"❌ Error uploading to Supabase: {e}")
            return False
    
    def run(self):
        """Main execution function"""
        logger.info("🚀 Starting Global Asset Ranking System")
        
        # Combine and rank assets
        combined_assets = self.combine_and_rank_assets()
        
        if not combined_assets:
            logger.error("❌ No assets to process")
            return
        
        # Save to JSON file
        self.save_combined_data(combined_assets)
        
        # Upload to Supabase
        if self.supabase_client:
            upload_success = self.upload_to_supabase(combined_assets)
            if upload_success:
                logger.info("✅ Data successfully uploaded to Supabase")
            else:
                logger.warning("⚠️ Supabase upload failed")
        
        # Summary
        logger.info("\n🎯 SUMMARY:")
        logger.info(f"   📊 Total assets processed: {len(combined_assets)}")
        
        # Count by category
        categories = {}
        for asset in combined_assets:
            category = asset.get('category', 'unknown')
            categories[category] = categories.get(category, 0) + 1
        
        logger.info(f"   📈 Asset breakdown:")
        for category, count in sorted(categories.items(), key=lambda x: x[1], reverse=True):
            logger.info(f"      {category}: {count}")
        
        # Check for major stocks
        major_stocks = ['NFLX', 'MC.PA', 'RMS.PA', 'ASML.AS', 'NOVO-B.CO']
        found_stocks = []
        
        for asset in combined_assets[:50]:  # Check top 50
            if asset.get('symbol') in major_stocks:
                found_stocks.append(f"{asset['symbol']} (#{asset['rank']})")
        
        if found_stocks:
            logger.info(f"   🎯 Found major stocks: {', '.join(found_stocks)}")
        else:
            logger.info(f"   ⚠️ Major stocks not found in top 50")
        
        logger.info("✅ Global Asset Ranking System completed successfully!")

def main():
    combiner = GlobalAssetCombiner()
    combiner.run()

if __name__ == "__main__":
    main() 