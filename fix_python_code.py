# Fix for combine_all_assets.py - Remove on_conflict parameter
# Replace line 379 with:

# OLD CODE:
# result = self.supabase.table('assets').upsert(batch, on_conflict='ticker,snapshot_date').execute()

# NEW CODE:
result = self.supabase.table('assets').upsert(batch).execute()

# OR use insert with manual duplicate handling:
# result = self.supabase.table('assets').insert(batch).execute()

# This will work without requiring the unique constraint 