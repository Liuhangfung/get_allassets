#!/usr/bin/env python3

import json
import logging
import os
import requests
import time
from datetime import datetime
from typing import Dict, List, Optional
from supabase import create_client, Client

# Load environment variables from .env file
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    # dotenv not installed, will use system environment variables
    pass

# Configure logging with UTF-8 encoding
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('combine_all_assets.log', encoding='utf-8'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class AssetCombiner:
    def __init__(self):
        self.supabase_url = os.environ.get('SUPABASE_URL', '')
        self.supabase_key = os.environ.get('SUPABASE_KEY', '')
        self.supabase: Optional[Client] = None
        
        if self.supabase_url and self.supabase_key:
            self.supabase = create_client(self.supabase_url, self.supabase_key)
            logger.info("Supabase connection configured")
            logger.info(f"Supabase URL: {self.supabase_url[:50]}...")
            logger.info(f"Supabase Key: ...{self.supabase_key[-10:]}")
        else:
            logger.warning("Supabase environment variables not set")
            
        # Emergency currency conversion rates (backup if Go conversion fails)
        self.emergency_rates = {
            'IDR': 0.000065,  # Indonesian Rupiah
            'CLP': 0.0010,    # Chilean Peso
            'SAR': 0.267,     # Saudi Riyal
            'ILS': 0.27,      # Israeli Shekel
            'COP': 0.00025,   # Colombian Peso
            'PEN': 0.27,      # Peruvian Sol
            'EGP': 0.020,     # Egyptian Pound
            'TRY': 0.029,     # Turkish Lira
            'RUB': 0.010,     # Russian Ruble
            'KRW': 0.00075,   # South Korean Won
            'INR': 0.012,     # Indian Rupee
            'BRL': 0.18,      # Brazilian Real
            'MXN': 0.058,     # Mexican Peso
            'ZAR': 0.055,     # South African Rand
            'THB': 0.029,     # Thai Baht
            'MYR': 0.22,      # Malaysian Ringgit
            'PHP': 0.018,     # Philippine Peso
            'VND': 0.000040,  # Vietnamese Dong
            'TWD': 0.031,     # Taiwan Dollar
            'HKD': 0.128,     # Hong Kong Dollar
            'SGD': 0.74,      # Singapore Dollar
            'JPY': 0.0067,    # Japanese Yen
            'CNY': 0.14,      # Chinese Yuan
            'AUD': 0.64,      # Australian Dollar
            'CAD': 0.74,      # Canadian Dollar
            'EUR': 1.08,      # Euro
            'GBP': 1.26,      # British Pound
            'ARS': 0.0010,    # Argentine Peso (~1000 ARS = 1 USD)
        }
        
    def detect_currency_from_symbol(self, symbol: str, country: str = '') -> str:
        """Detect currency from stock symbol or country"""
        symbol_upper = symbol.upper()
        country_upper = country.upper()
        
        # Symbol-based detection (most reliable)
        if symbol_upper.endswith('.JK') or country_upper == 'ID':
            return 'IDR'
        elif symbol_upper.endswith('.SN') or country_upper == 'CL':
            return 'CLP'
        elif symbol_upper.endswith('.SR') or country_upper == 'SA':
            return 'SAR'
        elif symbol_upper.endswith('.TA') or country_upper == 'IL':
            return 'ILS'
        elif symbol_upper.endswith('.BA') or country_upper == 'AR':
            return 'ARS'
        elif symbol_upper.endswith('.L') or country_upper == 'GB':
            return 'GBP'
        elif symbol_upper.endswith('.JO') or country_upper == 'ZA':
            return 'ZAR'
        elif symbol_upper.endswith('.CO') or country_upper == 'CO':
            return 'COP'
        elif symbol_upper.endswith('.LM') or country_upper == 'PE':
            return 'PEN'
        elif symbol_upper.endswith('.EG') or country_upper == 'EG':
            return 'EGP'
        elif symbol_upper.endswith('.IS') or country_upper == 'TR':
            return 'TRY'
        elif symbol_upper.endswith('.ME') or country_upper == 'RU':
            return 'RUB'
        elif symbol_upper.endswith('.KS') or symbol_upper.endswith('.KQ') or country_upper == 'KR':
            return 'KRW'
        elif symbol_upper.endswith('.BO') or symbol_upper.endswith('.NS') or country_upper == 'IN':
            return 'INR'
        elif symbol_upper.endswith('.SA') or country_upper == 'BR':
            return 'BRL'
        elif symbol_upper.endswith('.MX') or country_upper == 'MX':
            return 'MXN'
        elif symbol_upper.endswith('.BK') or country_upper == 'TH':
            return 'THB'
        elif symbol_upper.endswith('.KL') or country_upper == 'MY':
            return 'MYR'
        elif symbol_upper.endswith('.PS') or country_upper == 'PH':
            return 'PHP'
        elif symbol_upper.endswith('.VN') or country_upper == 'VN':
            return 'VND'
        elif symbol_upper.endswith('.TW') or country_upper == 'TW':
            return 'TWD'
        elif symbol_upper.endswith('.HK') or country_upper == 'HK':
            return 'HKD'
        elif symbol_upper.endswith('.SI') or country_upper == 'SG':
            return 'SGD'
        elif symbol_upper.endswith('.T') or country_upper == 'JP':
            return 'JPY'
        elif symbol_upper.endswith('.SS') or symbol_upper.endswith('.SZ') or country_upper == 'CN':
            return 'CNY'
        elif symbol_upper.endswith('.AX') or country_upper == 'AU':
            return 'AUD'
        elif symbol_upper.endswith('.TO') or country_upper == 'CA':
            return 'CAD'
        elif symbol_upper.endswith('.PA') or symbol_upper.endswith('.DE') or country_upper in ['FR', 'DE', 'IT', 'ES', 'NL', 'BE', 'AT', 'PT', 'GR', 'FI', 'IE']:
            return 'EUR'
        else:
            return 'USD'
    
    def validate_and_fix_market_cap(self, asset: Dict) -> Dict:
        """Validate and fix market cap values with emergency currency conversion"""
        
        symbol = asset.get('ticker', '')
        country = asset.get('country', '')
        market_cap = asset.get('market_cap', 0)
        
        # Skip crypto and commodities (they should be in USD already)
        if asset.get('asset_type') in ['crypto', 'commodity']:
            return asset
        
        # IMPORTANT: Skip currency conversion for stocks from Go program
        # The Go program already converts all market caps to USD correctly
        # Only apply emergency conversion to assets that don't have proper USD conversion
        data_source = asset.get('data_source', '')
        
        # Emergency currency conversion ONLY for non-FMP sources or legacy data
        if market_cap > 1e12 and data_source != 'FMP':  # > $1 trillion AND not from FMP/Go
            detected_currency = self.detect_currency_from_symbol(symbol, country)
            
            if detected_currency != 'USD' and detected_currency in self.emergency_rates:
                original_market_cap = market_cap
                market_cap = market_cap * self.emergency_rates[detected_currency]
                
                logger.warning(f"Emergency currency conversion: {symbol} | {original_market_cap/1e12:.1f}T {detected_currency} -> ${market_cap/1e9:.1f}B USD")
                
                # Update all USD values
                asset['market_cap'] = market_cap
                asset['current_price'] = asset.get('current_price', 0) * self.emergency_rates[detected_currency]
                asset['previous_close'] = asset.get('previous_close', 0) * self.emergency_rates[detected_currency]
        
        # Cap market cap at $4 trillion (even Apple is ~$3.5T)
        if market_cap > 4e12:
            logger.warning(f"Capping {symbol} market cap from ${market_cap/1e12:.1f}T to $4.0T")
            asset['market_cap'] = 4e12
        
        # Skip stocks with unrealistic market caps
        if market_cap > 10e12:
            logger.error(f"Removing {symbol}: Market cap too large (${market_cap/1e12:.1f}T)")
            return None
        
        return asset
    
    def load_json_file(self, filename: str) -> List[Dict]:
        """Load and validate JSON file with proper UTF-8 encoding"""
        try:
            with open(filename, 'r', encoding='utf-8') as f:
                data = json.load(f)
                return data if isinstance(data, list) else []
        except FileNotFoundError:
            logger.warning(f"File not found: {filename}")
            return []
        except json.JSONDecodeError as e:
            logger.error(f"JSON decode error in {filename}: {e}")
            return []
    
    def combine_all_assets(self) -> List[Dict]:
        """Combine all asset data from different sources"""
        
        # Load all data sources (matching Go program output filenames)
        stock_data = self.load_json_file('global_assets_fmp.json')
        crypto_data = self.load_json_file('crypto_data.json')
        
        logger.info(f"Loaded: {len(stock_data)} stocks, {len(crypto_data)} crypto")
        
        # If no stock data, proceed with crypto only for testing
        if not stock_data and crypto_data:
            logger.warning("No stock data found, proceeding with crypto only")
        
        # Mark stock data as already converted to USD by Go program
        for asset in stock_data:
            asset['data_source'] = 'FMP'  # Mark as FMP source (already converted)
        
        # Mark crypto data as already in USD
        for asset in crypto_data:
            asset['data_source'] = 'CoinGecko'  # Already in USD
        
        # Combine all assets
        all_assets = []
        all_assets.extend(stock_data)
        all_assets.extend(crypto_data)
        
        logger.info(f"Total assets before validation: {len(all_assets)}")
        
        # Remove duplicates by ticker symbol before validation
        seen_tickers = set()
        unique_assets = []
        for asset in all_assets:
            ticker = asset.get('ticker', '')
            if ticker and ticker not in seen_tickers:
                seen_tickers.add(ticker)
                unique_assets.append(asset)
            elif ticker:
                logger.info(f"Removing duplicate ticker: {ticker}")
        
        logger.info(f"Assets after deduplication: {len(unique_assets)}")
        
        # Validate and fix market caps
        validated_assets = []
        for asset in unique_assets:
            validated_asset = self.validate_and_fix_market_cap(asset)
            if validated_asset:
                validated_assets.append(validated_asset)
        
        logger.info(f"Assets after validation: {len(validated_assets)}")
        
        # Sort by market cap (descending)
        validated_assets.sort(key=lambda x: x.get('market_cap', 0), reverse=True)
        
        # Take top 500
        top_assets = validated_assets[:500]
        
        # Add ranking
        for i, asset in enumerate(top_assets):
            asset['rank'] = i + 1
            asset['snapshot_date'] = datetime.now().strftime('%Y-%m-%d')
            
            # Add missing fields with defaults
            asset.setdefault('circulating_supply', None)
            asset.setdefault('price_raw', asset.get('current_price', 0))
            asset.setdefault('market_cap_raw', asset.get('market_cap', 0))
            asset.setdefault('category', 'Global')
            # data_source already set above
        
        return top_assets
    
    def save_to_json(self, data: List[Dict], filename: str = 'all_assets_combined.json'):
        """Save combined data to JSON file with proper UTF-8 encoding"""
        try:
            with open(filename, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2, ensure_ascii=False)
            logger.info(f"Saved {len(data)} assets to {filename}")
        except Exception as e:
            logger.error(f"Error saving to {filename}: {e}")
    
    def prepare_for_database(self, asset: Dict) -> Dict:
        """Prepare asset data for database insertion with overflow protection"""
        
        # PostgreSQL bigint max value: 9,223,372,036,854,775,807
        MAX_BIGINT = 9_223_372_036_854_775_807
        
        def safe_number(value, max_val=MAX_BIGINT, as_int=False):
            if value is None:
                return None
            try:
                num = float(value)
                if num > max_val:
                    num = max_val
                return int(num) if as_int else num
            except (ValueError, TypeError):
                return None
        
        # Map fields and ensure safe values
        db_asset = {
            'symbol': str(asset.get('ticker', ''))[:50],
            'ticker': str(asset.get('ticker', ''))[:50],
            'name': str(asset.get('name', ''))[:200],
            'current_price': safe_number(asset.get('current_price', 0)),
            'previous_close': safe_number(asset.get('previous_close', 0)),
            'percentage_change': safe_number(asset.get('percentage_change', 0)),
            'market_cap': safe_number(asset.get('market_cap', 0), as_int=True),
            'volume': safe_number(asset.get('volume', 0), as_int=True),
            'circulating_supply': safe_number(asset.get('circulating_supply'), as_int=True),
            'primary_exchange': str(asset.get('primary_exchange', ''))[:50],
            'country': str(asset.get('country', ''))[:50],
            'sector': str(asset.get('sector', ''))[:100],
            'industry': str(asset.get('industry', ''))[:100],
            'asset_type': str(asset.get('asset_type', 'unknown'))[:50],
            'image': str(asset.get('image', ''))[:500],
            'rank': int(asset.get('rank', 0)),
            'snapshot_date': asset.get('snapshot_date', datetime.now().strftime('%Y-%m-%d')),
            'price_raw': safe_number(asset.get('price_raw', 0)),
            'market_cap_raw': safe_number(asset.get('market_cap_raw', 0), as_int=True),
            'category': str(asset.get('category', ''))[:50],
            'data_source': str(asset.get('data_source', ''))[:50],
        }
        
        return db_asset
    
    def upload_to_supabase(self, assets: List[Dict], clear_existing=False):
        """Upload assets to Supabase with upsert handling for duplicates"""
        if not self.supabase:
            logger.warning("No Supabase connection configured")
            return
        
        # Debug: Test connection and check table
        try:
            logger.info("Testing Supabase connection...")
            test_query = self.supabase.table('assets').select('id').limit(1).execute()
            logger.info(f"Connection test successful. Found {len(test_query.data)} existing records")
        except Exception as e:
            logger.error(f"Connection test failed: {e}")
            return
        
        try:
            today = datetime.now().strftime('%Y-%m-%d')
            
            if clear_existing:
                # Clear only today's data, not all historical data
                logger.info(f"Clearing existing data for today ({today})...")
                
                # First check if data exists for today
                existing = self.supabase.table('assets').select('id').eq('snapshot_date', today).limit(1).execute()
                
                if existing.data:
                    # Delete existing data for today
                    result = self.supabase.table('assets').delete().eq('snapshot_date', today).execute()
                    logger.info(f"Deleted existing records for today")
                else:
                    logger.info("No existing records found for today")
            else:
                logger.info("Using upsert mode (update existing, insert new)")
            
            # Prepare data for database
            db_assets = []
            for asset in assets:
                db_asset = self.prepare_for_database(asset)
                db_assets.append(db_asset)
            
            # Upload in batches with upsert
            batch_size = 100
            total_processed = 0
            
            for i in range(0, len(db_assets), batch_size):
                batch = db_assets[i:i+batch_size]
                
                try:
                    if clear_existing:
                        # Use regular insert when database was cleared
                        result = self.supabase.table('assets').insert(batch).execute()
                    else:
                        # Try upsert first, fall back to insert if constraint doesn't exist
                        try:
                            logger.info(f"Attempting upsert for batch {i//batch_size + 1} with {len(batch)} assets")
                            result = self.supabase.table('assets').upsert(batch, on_conflict='symbol,snapshot_date').execute()
                            logger.info(f"Upsert successful for batch {i//batch_size + 1}")
                        except Exception as upsert_error:
                            error_code = str(upsert_error)
                            logger.warning(f"Upsert failed: {error_code}")
                            
                            if '42P10' in error_code:
                                # No unique constraint exists, use regular insert
                                logger.info("Error 42P10: No unique constraint found, switching to insert mode")
                                try:
                                    result = self.supabase.table('assets').insert(batch).execute()
                                    logger.info(f"Insert successful for batch {i//batch_size + 1}")
                                except Exception as insert_error:
                                    logger.error(f"Insert also failed: {insert_error}")
                                    raise insert_error
                            elif '23505' in error_code:
                                logger.info("Error 23505: Duplicate key constraint violation (this should not happen with upsert)")
                                raise upsert_error
                            else:
                                logger.error(f"Unknown error during upsert: {error_code}")
                                raise upsert_error
                    
                    if result.data:
                        total_processed += len(batch)
                        logger.info(f"Processed batch {i//batch_size + 1} ({len(batch)} assets)")
                        
                except Exception as batch_error:
                    # If batch fails, try individual inserts
                    logger.warning(f"Batch operation failed, trying individual inserts: {batch_error}")
                    successful = 0
                    for individual_asset in batch:
                        try:
                            if clear_existing:
                                individual_result = self.supabase.table('assets').insert([individual_asset]).execute()
                            else:
                                try:
                                    individual_result = self.supabase.table('assets').upsert([individual_asset], on_conflict='symbol,snapshot_date').execute()
                                except Exception as upsert_error:
                                    if '42P10' in str(upsert_error):
                                        # No unique constraint exists, use regular insert
                                        individual_result = self.supabase.table('assets').insert([individual_asset]).execute()
                                    else:
                                        raise upsert_error
                            if individual_result.data:
                                successful += 1
                        except Exception as individual_error:
                            logger.warning(f"Failed to insert {individual_asset.get('ticker', 'unknown')}: {individual_error}")
                    logger.info(f"Successfully inserted {successful}/{len(batch)} assets individually")
                    total_processed += successful
                
                time.sleep(0.1)  # Rate limiting
            
            logger.info(f"Successfully processed {total_processed} assets to Supabase")
            
        except Exception as e:
            error_msg = str(e)
            if "duplicate key value violates unique constraint" in error_msg and "snapshot_date" in error_msg:
                today = datetime.now().strftime('%Y-%m-%d')
                logger.info(f"Data for today ({today}) already exists in Supabase")
                logger.info("To replace today's data only, set CLEAR_EXISTING_DATA=true in your .env file")
                logger.info("Note: This will only clear today's data, not historical data")
            else:
                logger.error(f"Error uploading to Supabase: {e}")
                logger.warning("Supabase upload failed")
    
    def print_summary(self, assets: List[Dict]):
        """Print summary of the combined assets"""
        if not assets:
            logger.info("No assets to summarize")
            return
        
        logger.info(f"\nSUMMARY:")
        logger.info(f"   Total assets processed: {len(assets)}")
        
        # Asset type breakdown
        asset_types = {}
        for asset in assets:
            asset_type = asset.get('asset_type', 'unknown')
            asset_types[asset_type] = asset_types.get(asset_type, 0) + 1
        
        logger.info(f"   Asset breakdown:")
        for asset_type, count in sorted(asset_types.items()):
            logger.info(f"      {asset_type}: {count}")
        
        # Top 10 assets
        logger.info(f"   Top 10 assets by market cap:")
        for i, asset in enumerate(assets[:10]):
            market_cap = asset.get('market_cap', 0)
            if market_cap >= 1e12:
                cap_str = f"${market_cap/1e12:.1f}T"
            else:
                cap_str = f"${market_cap/1e9:.1f}B"
            logger.info(f"     {i+1:2d}. {asset.get('ticker', 'N/A'):<12} | {asset.get('name', 'Unknown'):<30} | {cap_str}")
        
        # Check for major stocks
        major_stocks = ['AAPL', 'MSFT', 'GOOGL', 'AMZN', 'TSLA', 'META', 'NVDA', 'NFLX', 'LVMUY', 'RHHVF']
        found_in_top_50 = []
        for asset in assets[:50]:
            if asset.get('ticker') in major_stocks:
                found_in_top_50.append(asset.get('ticker'))
        
        if len(found_in_top_50) >= 8:
            logger.info(f"   Found {len(found_in_top_50)} major stocks in top 50: {', '.join(found_in_top_50)}")
        else:
            logger.info(f"   Major stocks not found in top 50")
    
    def run(self):
        """Main execution method"""
        logger.info("Starting Global Asset Ranking System")
        
        # Combine all assets
        combined_assets = self.combine_all_assets()
        
        if not combined_assets:
            logger.error("No assets to process")
            return
        
        # Save to JSON
        self.save_to_json(combined_assets)
        
        # Upload to Supabase (check environment variable for clear behavior)
        clear_existing = os.environ.get('CLEAR_EXISTING_DATA', 'false').lower() == 'true'
        if clear_existing:
            today = datetime.now().strftime('%Y-%m-%d')
            logger.warning(f"CLEAR_EXISTING_DATA=true - Will delete existing data for today ({today}) only!")
        self.upload_to_supabase(combined_assets, clear_existing=clear_existing)
        
        # Print summary
        self.print_summary(combined_assets)
        
        logger.info("Global Asset Ranking System completed successfully!")

if __name__ == "__main__":
    combiner = AssetCombiner()
    combiner.run() 