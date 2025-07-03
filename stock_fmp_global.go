package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// FMP API structures
type FMPStockScreener struct {
	Symbol             string  `json:"symbol"`
	CompanyName        string  `json:"companyName"`
	MarketCap          float64 `json:"marketCap"`
	Sector             string  `json:"sector"`
	Industry           string  `json:"industry"`
	Beta               float64 `json:"beta"`
	Price              float64 `json:"price"`
	Volume             float64 `json:"volume"`
	Exchange           string  `json:"exchange"`
	ExchangeShortName  string  `json:"exchangeShortName"`
	Country            string  `json:"country"`
	IsEtf              bool    `json:"isEtf"`
	IsActivelyTrading  bool    `json:"isActivelyTrading"`
}

type FMPQuote struct {
	Symbol             string  `json:"symbol"`
	Name               string  `json:"name"`
	Price              float64 `json:"price"`
	ChangesPercentage  float64 `json:"changesPercentage"`
	Change             float64 `json:"change"`
	MarketCap          float64 `json:"marketCap"`
	Volume             float64 `json:"volume"`
	Open               float64 `json:"open"`
	PreviousClose      float64 `json:"previousClose"`
	Exchange           string  `json:"exchange"`
	SharesOutstanding  float64 `json:"sharesOutstanding"`
}

type FMPCommodity struct {
	Symbol             string  `json:"symbol"`
	Name               string  `json:"name"`
	Price              float64 `json:"price"`
	ChangesPercentage  float64 `json:"changesPercentage"`
	Change             float64 `json:"change"`
	PreviousClose      float64 `json:"previousClose"`
	Exchange           string  `json:"exchange"`
}

type FMPCompanyProfile struct {
	Symbol      string  `json:"symbol"`
	CompanyName string  `json:"companyName"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
	Beta        float64 `json:"beta"`
	VolAvg      float64 `json:"volAvg"`
	MktCap      float64 `json:"mktCap"`
	Industry    string  `json:"industry"`
	Sector      string  `json:"sector"`
	Country     string  `json:"country"`
	Exchange    string  `json:"exchange"`
	Website     string  `json:"website"`
	Description string  `json:"description"`
}

type AssetData struct {
	Ticker            string  `json:"ticker"`
	Name              string  `json:"name"`
	MarketCap         float64 `json:"market_cap"`
	CurrentPrice      float64 `json:"current_price"`
	PreviousClose     float64 `json:"previous_close"`
	PercentageChange  float64 `json:"percentage_change"`
	Volume            float64 `json:"volume"`
	PrimaryExchange   string  `json:"primary_exchange"`
	Country           string  `json:"country"`
	Sector            string  `json:"sector"`
	Industry          string  `json:"industry"`
	AssetType         string  `json:"asset_type"`
	Image             string  `json:"image"`
}

type FMPClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewFMPClient(apiKey string) *FMPClient {
	return &FMPClient{
		APIKey:  apiKey,
		BaseURL: "https://financialmodelingprep.com/api",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *FMPClient) makeRequest(endpoint string) ([]byte, error) {
	separator := "?"
	if strings.Contains(endpoint, "?") {
		separator = "&"
	}
	url := fmt.Sprintf("%s%s%sapikey=%s", c.BaseURL, endpoint, separator, c.APIKey)
	
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("FMP API Error Response: %s\n", string(body))
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return body, nil
}

func (c *FMPClient) GetQuote(symbol string) (*FMPQuote, error) {
	endpoint := fmt.Sprintf("/v3/quote/%s", symbol)
	
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote for %s: %w", symbol, err)
	}

	var quotes []FMPQuote
	if err := json.Unmarshal(body, &quotes); err != nil {
		return nil, fmt.Errorf("failed to parse quote data for %s: %w", symbol, err)
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote data found for %s", symbol)
	}

	return &quotes[0], nil
}

func (c *FMPClient) GetCompanyProfile(symbol string) (*FMPCompanyProfile, error) {
	endpoint := fmt.Sprintf("/v3/profile/%s", symbol)
	
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get company profile for %s: %w", symbol, err)
	}

	var profiles []FMPCompanyProfile
	if err := json.Unmarshal(body, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse company profile data for %s: %w", symbol, err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no company profile data found for %s", symbol)
	}

	return &profiles[0], nil
}



func (c *FMPClient) GetGlobalStocks() ([]AssetData, error) {
	fmt.Println("🌍 Fetching ALL stocks from ALL countries...")
	
	// Get maximum stocks from all countries, all exchanges
	endpoint := "/v3/stock-screener?marketCapMoreThan=1000000&limit=50000&order=desc&sortBy=marketcap&isActivelyTrading=true"
	
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock screener data: %w", err)
	}

	var screenerData []FMPStockScreener
	if err := json.Unmarshal(body, &screenerData); err != nil {
		return nil, fmt.Errorf("failed to parse screener data: %w", err)
	}

	fmt.Printf("✅ Received %d securities from stock screener\n", len(screenerData))

	// Sort by market cap descending to ensure we get the largest companies first
	sort.Slice(screenerData, func(i, j int) bool {
		return screenerData[i].MarketCap > screenerData[j].MarketCap
	})
	fmt.Printf("📊 Sorted by market cap (largest first)\n")

	var assets []AssetData
	stockCount := 0
	maxStocks := 490 // Increased to get ~500 total assets (490 stocks + 10 crypto)
	
	for _, stock := range screenerData {
		if stockCount >= maxStocks {
			break
		}
		if !stock.IsActivelyTrading {
			continue
		}

		// Skip ETFs and index funds
		if stock.IsEtf {
			continue
		}
		
		// Skip index funds, ETFs in company name
		nameUpper := strings.ToUpper(stock.CompanyName)
		if strings.Contains(nameUpper, "ETF") ||
		   strings.Contains(nameUpper, "INDEX") ||
		   strings.Contains(nameUpper, "FUND") ||
		   strings.Contains(nameUpper, "SPDR") ||
		   strings.Contains(nameUpper, "ISHARES") ||
		   strings.Contains(nameUpper, "VANGUARD") ||
		   strings.Contains(nameUpper, "INVESCO") {
			continue
		}

		assetType := "stock"
		if strings.Contains(nameUpper, "REIT") {
			assetType = "reit"
		}

		// Get real-time quote data for this stock
		quote, err := c.GetQuote(stock.Symbol)
		var percentageChange float64
		var previousClose float64
		
		if err == nil && quote != nil {
			percentageChange = quote.ChangesPercentage
			previousClose = quote.PreviousClose
		} else {
			// Fallback to screener data if quote fails
			percentageChange = 0.0
			previousClose = stock.Price
		}
		
		// Get company profile for image
		var imageURL string
		profile, err := c.GetCompanyProfile(stock.Symbol)
		if err == nil && profile != nil {
			imageURL = profile.Image
		}
		
		// Small delay to avoid hitting API rate limits
		time.Sleep(100 * time.Millisecond) // Increased delay due to more API calls

		asset := AssetData{
			Ticker:           stock.Symbol,
			Name:             stock.CompanyName,
			MarketCap:        stock.MarketCap,
			CurrentPrice:     stock.Price,
			PreviousClose:    previousClose,
			PercentageChange: percentageChange,
			Volume:           stock.Volume,
			PrimaryExchange:  stock.ExchangeShortName,
			Country:          stock.Country,
			Sector:           stock.Sector,
			Industry:         stock.Industry,
			AssetType:        assetType,
			Image:            imageURL,
		}

		assets = append(assets, asset)
		stockCount++
		
		// Progress reporting
		if stockCount%50 == 0 {
			fmt.Printf("📊 Processed %d/%d top stocks by market cap...\n", stockCount, maxStocks)
		}
	}

	fmt.Printf("✅ Processed %d global securities\n", len(assets))
	return assets, nil
}

// Helper function to identify real physical commodities (only essential ones)
func isRealCommodity(name, symbol string) bool {
	nameUpper := strings.ToUpper(name)
	symbolUpper := strings.ToUpper(symbol)
	
	// Essential commodities we want (matching FMP names and symbols exactly)
	essentialCommodities := map[string]bool{
		// Metals only (name contains)
		"GOLD":       true,
		"SILVER":     true,
		"PLATINUM":   true,
		"PALLADIUM":  true,
		"COPPER":     true,
	}
	
	// Check if name contains any essential commodity
	for commodity := range essentialCommodities {
		if strings.Contains(nameUpper, commodity) {
			return true
		}
	}
	
	// Essential symbols we want (exact matches from FMP) - Main contracts only
	essentialSymbols := map[string]bool{
		"GCUSD":   true,  // Gold Futures (main contract)
		"SIUSD":   true,  // Silver Futures (main contract)
		"PLUSD":   true,  // Platinum
		"PAUSD":   true,  // Palladium
		"HGUSD":   true,  // Copper
		// Excluding micro contracts to avoid duplicates:
		// "MGCUSD":  Micro Gold Futures (duplicate of GCUSD)
		// "SILUSD":  Micro Silver Futures (duplicate of SIUSD)
	}
	
	// Check for exact symbol match
	if essentialSymbols[symbolUpper] {
		return true
	}
	
	return false
}

func (c *FMPClient) GetCommodities() ([]AssetData, error) {
	fmt.Println("🥇 Fetching commodities (Gold, Silver, Oil, etc.) from FMP...")
	
	endpoint := "/v3/quotes/commodity"
	
	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get commodity data: %w", err)
	}

	var commodities []FMPCommodity
	if err := json.Unmarshal(body, &commodities); err != nil {
		return nil, fmt.Errorf("failed to parse commodity data: %w", err)
	}

	fmt.Printf("✅ Received %d commodities\n", len(commodities))

	var assets []AssetData
	for _, commodity := range commodities {
		// Check if it's a commodity we want
		if !isRealCommodity(commodity.Name, commodity.Symbol) {
			continue // Skip non-essential commodities silently
		}
		


		// Calculate actual market cap for commodities based on estimated total supply
		var marketCap float64
		symbolUpper := strings.ToUpper(commodity.Symbol)
		
		switch symbolUpper {
		case "GCUSD": // Gold (main contract only)
			// Estimated ~200,000 tonnes of gold mined (6.4B ounces)
			marketCap = commodity.Price * 6400000000 // 6.4B ounces
		case "SIUSD": // Silver (main contract only)
			// Estimated ~1.7M tonnes of silver mined (54.6B ounces)
			marketCap = commodity.Price * 54600000000 // 54.6B ounces
		case "PLUSD": // Platinum
			// Estimated ~8,000 tonnes of platinum mined (257M ounces)
			marketCap = commodity.Price * 257000000 // 257M ounces
		case "PAUSD": // Palladium
			// Estimated ~175M ounces of palladium mined
			marketCap = commodity.Price * 175000000 // 175M ounces
		case "HGUSD": // Copper
			// Estimated ~700M tonnes of copper mined (price per tonne)
			marketCap = commodity.Price * 700000000 // 700M tonnes
		default:
			// Fallback for unknown commodities
			marketCap = commodity.Price * 1000000000 // 1B units
		}

		asset := AssetData{
			Ticker:           commodity.Symbol,
			Name:             commodity.Name,
			MarketCap:        marketCap,
			CurrentPrice:     commodity.Price,
			PreviousClose:    commodity.PreviousClose,
			PercentageChange: commodity.ChangesPercentage,
			Volume:           0,
			PrimaryExchange:  commodity.Exchange,
			Country:          "Global",
			Sector:           "Commodities",
			Industry:         "Commodities",
			AssetType:        "commodity",
			Image:            "", // No images available for commodities
		}

		assets = append(assets, asset)
	}

	fmt.Printf("✅ Processed %d commodities\n", len(assets))
	return assets, nil
}

func saveToJSON(data []AssetData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func printSummary(data []AssetData) {
	if len(data) == 0 {
		fmt.Println("No data to display")
		return
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].MarketCap > data[j].MarketCap
	})

	fmt.Printf("\n=== TOP %d GLOBAL ASSETS BY MARKET CAP (FMP DATA) ===\n", len(data))
	fmt.Printf("%-8s %-25s %-8s %-12s %-10s %-15s %-12s %-8s\n", 
		"Ticker", "Name", "Country", "Price", "Change%", "Market Cap", "Type", "Exchange")
	fmt.Println(strings.Repeat("-", 100))

	for i, asset := range data {
		if i >= 30 {
			break
		}
		
		name := asset.Name
		if len(name) > 23 {
			name = name[:23] + ".."
		}
		
		country := asset.Country
		if len(country) > 6 {
			country = country[:6]
		}
		
		marketCapStr := formatLargeNumber(asset.MarketCap)
		
		typeDisplay := asset.AssetType
		if asset.AssetType == "commodity" {
			typeDisplay = "🥇 " + asset.AssetType
		}
		
		fmt.Printf("%-8s %-25s %-8s $%-11.2f %-9.2f%% %-15s %-12s %-8s\n",
			asset.Ticker, name, country, asset.CurrentPrice, asset.PercentageChange, 
			marketCapStr, typeDisplay, asset.PrimaryExchange)
	}
	
	fmt.Printf("\nTotal global assets processed: %d\n", len(data))
	
	assetTypeCounts := make(map[string]int)
	countryCounts := make(map[string]int)
	
	for _, asset := range data {
		assetTypeCounts[asset.AssetType]++
		if asset.Country != "" {
			countryCounts[asset.Country]++
		}
	}
	
	fmt.Printf("\n📊 Asset Type Breakdown:\n")
	for assetType, count := range assetTypeCounts {
		if assetType == "commodity" {
			fmt.Printf("  🥇 %d commodities (Gold, Silver, Oil, etc.)\n", count)
		} else if assetType == "stock" {
			fmt.Printf("  📈 %d individual stocks\n", count)
		} else {
			fmt.Printf("  🏢 %d %ss\n", count, assetType)
		}
	}
}

func formatLargeNumber(num float64) string {
	if num >= 1e12 {
		return fmt.Sprintf("%.1fT", num/1e12)
	} else if num >= 1e9 {
		return fmt.Sprintf("%.1fB", num/1e9)
	} else if num >= 1e6 {
		return fmt.Sprintf("%.1fM", num/1e6)
	} else if num >= 1e3 {
		return fmt.Sprintf("%.1fK", num/1e3)
	}
	return fmt.Sprintf("%.0f", num)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found, using environment variables")
	}

	apiKey := os.Getenv("FMP_API_KEY")
	if apiKey == "" {
		log.Fatal("FMP_API_KEY environment variable is required")
	}

	client := NewFMPClient(apiKey)

	fmt.Println("🌟 GLOBAL MARKET ANALYSIS WITH FMP API")
	fmt.Println("Fetching top 500 individual stocks by market cap globally:")
	fmt.Println("🌍 ALL Countries (US, EU, Asia, Hong Kong, etc.)")
	fmt.Println("🏢 ALL Exchanges (NYSE, NASDAQ, LSE, SEHK, etc.)")
	fmt.Println("🥇 Plus Essential Commodities (Gold, Silver, etc.)")
	fmt.Println("⚠️  Excluding: ETFs, Index Funds, Mutual Funds")
	fmt.Println()
	
	startTime := time.Now()
	var allAssets []AssetData
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	wg.Add(1)
	go func() {
		defer wg.Done()
		stocks, err := client.GetGlobalStocks()
		if err != nil {
			fmt.Printf("❌ Failed to fetch global stocks: %v\n", err)
			return
		}
		mu.Lock()
		allAssets = append(allAssets, stocks...)
		mu.Unlock()
	}()
	
	wg.Add(1)
	go func() {
		defer wg.Done()
		commodities, err := client.GetCommodities()
		if err != nil {
			fmt.Printf("❌ Failed to fetch commodities: %v\n", err)
			return
		}
		mu.Lock()
		allAssets = append(allAssets, commodities...)
		mu.Unlock()
	}()
	

	
	wg.Wait()
	
	if len(allAssets) == 0 {
		log.Fatal("❌ No assets fetched successfully!")
	}
	
	// Separate stocks from commodities
	var stocks []AssetData
	var commodities []AssetData
	
	for _, asset := range allAssets {
		if asset.AssetType == "commodity" {
			commodities = append(commodities, asset)
		} else {
			stocks = append(stocks, asset)
		}
	}
	
	fmt.Printf("\n📊 Received %d stocks and %d commodities\n", len(stocks), len(commodities))
	
	// Sort stocks by market cap and take top 500
	sort.Slice(stocks, func(i, j int) bool {
		return stocks[i].MarketCap > stocks[j].MarketCap
	})
	
	if len(stocks) > 500 {
		stocks = stocks[:500]
		fmt.Printf("✂️  Limited to top 500 stocks by market cap\n")
	}
	
	// Combine top 500 stocks with commodities
	allAssets = append(stocks, commodities...)
	
	fmt.Printf("🔗 Final dataset: %d stocks + %d commodities = %d total assets\n", 
		len(stocks), len(commodities), len(allAssets))
	
	// Final sort by market cap
	sort.Slice(allAssets, func(i, j int) bool {
		return allAssets[i].MarketCap > allAssets[j].MarketCap
	})

	filename := "global_assets_fmp.json"
	if err := saveToJSON(allAssets, filename); err != nil {
		log.Printf("Failed to save to file: %v", err)
	} else {
		fmt.Printf("💾 Data saved to %s\n", filename)
	}

	printSummary(allAssets)
	
	duration := time.Since(startTime)
	fmt.Printf("\n🎉 Total processing time: %v\n", duration)
	fmt.Printf("🌟 Retrieved data from worldwide markets using FMP Ultimate API!\n")
} 