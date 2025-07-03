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

// Helper function to check if a string contains a word with proper word boundaries
func containsWord(text, word string) bool {
	// Convert to upper case for case-insensitive comparison
	textUpper := strings.ToUpper(text)
	wordUpper := strings.ToUpper(word)
	
	// Find all occurrences of the word
	index := 0
	for {
		pos := strings.Index(textUpper[index:], wordUpper)
		if pos == -1 {
			break
		}
		
		// Adjust position to absolute index
		pos += index
		
		// Check if it's a complete word (not part of another word)
		isWordStart := pos == 0 || !isAlphaNumeric(textUpper[pos-1])
		isWordEnd := pos+len(wordUpper) == len(textUpper) || !isAlphaNumeric(textUpper[pos+len(wordUpper)])
		
		if isWordStart && isWordEnd {
			return true
		}
		
		// Move to next potential match
		index = pos + 1
	}
	
	return false
}

// Helper function to check if a character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

// Helper function to determine which listing to keep for duplicate companies
func shouldKeepNewListing(newStock, existingStock FMPStockScreener) bool {
	newPriority := getListingPriority(newStock.Symbol, newStock.ExchangeShortName)
	existingPriority := getListingPriority(existingStock.Symbol, existingStock.ExchangeShortName)
	
	// Lower number = higher priority
	if newPriority < existingPriority {
		return true
	} else if newPriority == existingPriority {
		// Same priority, keep the one with higher market cap
		return newStock.MarketCap > existingStock.MarketCap
	}
	return false
}

// Get listing priority (lower number = higher priority)
func getListingPriority(symbol, exchange string) int {
	symbolUpper := strings.ToUpper(symbol)
	exchangeUpper := strings.ToUpper(exchange)
	
	// Hong Kong main listings (highest priority)
	if strings.HasSuffix(symbolUpper, ".HK") || exchangeUpper == "HKSE" {
		return 1
	}
	
	// US ADRs (second priority)
	if (exchangeUpper == "NASDAQ" || exchangeUpper == "NYSE") && 
	   (strings.HasSuffix(symbolUpper, "Y") || strings.HasSuffix(symbolUpper, "H")) {
		return 2
	}
	
	// US main listings (third priority)
	if exchangeUpper == "NASDAQ" || exchangeUpper == "NYSE" {
		return 3
	}
	
	// US OTC (fourth priority)
	if exchangeUpper == "OTC" || strings.HasSuffix(symbolUpper, "F") {
		return 4
	}
	
	// European listings (lower priority)
	if strings.HasSuffix(symbolUpper, ".VI") || strings.HasSuffix(symbolUpper, ".L") || 
	   strings.HasSuffix(symbolUpper, ".PA") || strings.HasSuffix(symbolUpper, ".DE") {
		return 5
	}
	
	// Other exchanges (lowest priority)
	return 6
}

// Get current exchange rate from FMP or use fallback
func (c *FMPClient) getUSDExchangeRate(fromCurrency string) float64 {
	// Try to get real-time exchange rate from FMP
	endpoint := fmt.Sprintf("/v3/fx/%sUSD", fromCurrency)
	
	body, err := c.makeRequest(endpoint)
	if err == nil {
		var fxData []struct {
			Bid float64 `json:"bid"`
			Ask float64 `json:"ask"`
		}
		if json.Unmarshal(body, &fxData) == nil && len(fxData) > 0 {
			// Use mid-price (average of bid and ask)
			return (fxData[0].Bid + fxData[0].Ask) / 2
		}
	}
	
	// Fallback to approximate rates (updated regularly)
	fallbackRates := map[string]float64{
		"HKD": 0.128,     // 1 HKD = ~0.128 USD
		"EUR": 1.08,      // 1 EUR = ~1.08 USD  
		"GBP": 1.26,      // 1 GBP = ~1.26 USD
		"JPY": 0.0067,    // 1 JPY = ~0.0067 USD
		"CAD": 0.74,      // 1 CAD = ~0.74 USD
		"AUD": 0.64,      // 1 AUD = ~0.64 USD
		"CNY": 0.14,      // 1 CNY = ~0.14 USD
		"IDR": 0.000065,  // 1 IDR = ~0.000065 USD (about 15,400 IDR = 1 USD)
		"INR": 0.012,     // 1 INR = ~0.012 USD (about 83 INR = 1 USD)
		"KRW": 0.00075,   // 1 KRW = ~0.00075 USD (about 1,330 KRW = 1 USD)
		"BRL": 0.18,      // 1 BRL = ~0.18 USD (about 5.5 BRL = 1 USD)
		"MXN": 0.058,     // 1 MXN = ~0.058 USD (about 17 MXN = 1 USD)
		"ZAR": 0.055,     // 1 ZAR = ~0.055 USD (about 18 ZAR = 1 USD)
		"THB": 0.029,     // 1 THB = ~0.029 USD (about 34 THB = 1 USD)
		"MYR": 0.22,      // 1 MYR = ~0.22 USD (about 4.5 MYR = 1 USD)
		"PHP": 0.018,     // 1 PHP = ~0.018 USD (about 56 PHP = 1 USD)
		"VND": 0.000040,  // 1 VND = ~0.000040 USD (about 25,000 VND = 1 USD)
	}
	
	if rate, exists := fallbackRates[fromCurrency]; exists {
		return rate
	}
	
	// If unknown currency, assume it's already in USD
	return 1.0
}

func (c *FMPClient) GetGlobalStocks() ([]AssetData, error) {
	fmt.Println("🌍 Fetching ALL stocks from ALL countries systematically...")
	
	// Comprehensive country-by-country approach (Ultimate plan = 3000 requests/minute)
	endpoints := []struct {
		url string
		desc string
	}{
		// Major Markets (High Priority)
		{"/v3/stock-screener?marketCapMoreThan=1000000&limit=10000&order=desc&sortBy=marketcap&isActivelyTrading=true", "Global (Default)"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=US", "United States"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=CN", "China"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=HK", "Hong Kong"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=JP", "Japan"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=IN", "India"},
		
		// European Markets (Netflix, LVMH, Hermès should be here)
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=GB", "United Kingdom"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=FR", "France"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=DE", "Germany"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=CH", "Switzerland"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=NL", "Netherlands"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=IT", "Italy"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=ES", "Spain"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=SE", "Sweden"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=DK", "Denmark"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=NO", "Norway"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=FI", "Finland"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=BE", "Belgium"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=AT", "Austria"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=PL", "Poland"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=IE", "Ireland"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=PT", "Portugal"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=GR", "Greece"},
		
		// Asia-Pacific Markets
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=KR", "South Korea"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=TW", "Taiwan"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=SG", "Singapore"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=AU", "Australia"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=NZ", "New Zealand"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=TH", "Thailand"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=MY", "Malaysia"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=ID", "Indonesia"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=PH", "Philippines"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=VN", "Vietnam"},
		
		// Americas
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=CA", "Canada"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=BR", "Brazil"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=MX", "Mexico"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=AR", "Argentina"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=CL", "Chile"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=CO", "Colombia"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=PE", "Peru"},
		
		// Middle East & Africa
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=IL", "Israel"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=SA", "Saudi Arabia"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=AE", "UAE"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=ZA", "South Africa"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=EG", "Egypt"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=TR", "Turkey"},
		{"/v3/stock-screener?marketCapMoreThan=500000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true&country=RU", "Russia"},
	}
	
	var allScreenerData []FMPStockScreener
	totalRequests := len(endpoints)
	successfulRequests := 0
	
	fmt.Printf("📊 Making %d API calls to fetch global stocks...\n", totalRequests)
	
	for i, endpoint := range endpoints {
		fmt.Printf("📡 [%d/%d] Fetching %s stocks...\n", i+1, totalRequests, endpoint.desc)
		
		body, err := c.makeRequest(endpoint.url)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to fetch %s stocks: %v\n", endpoint.desc, err)
			continue
		}

		var screenerData []FMPStockScreener
		if err := json.Unmarshal(body, &screenerData); err != nil {
			fmt.Printf("⚠️  Warning: Failed to parse %s stocks: %v\n", endpoint.desc, err)
			continue
		}

		fmt.Printf("✅ %s: %d stocks\n", endpoint.desc, len(screenerData))
		allScreenerData = append(allScreenerData, screenerData...)
		successfulRequests++
		
		// Rate limiting: Ultimate plan = 3000 requests/minute = 50 requests/second
		// We'll be conservative and limit to 20 requests/second to be safe
		if i < len(endpoints)-1 { // Don't sleep after the last request
			time.Sleep(50 * time.Millisecond) // 20 requests per second
		}
		
		// Progress update every 10 requests
		if (i+1)%10 == 0 {
			fmt.Printf("📈 Progress: %d/%d markets completed (%d total stocks so far)\n", 
				i+1, totalRequests, len(allScreenerData))
		}
	}
	
	fmt.Printf("🎯 API Summary: %d/%d successful requests\n", successfulRequests, totalRequests)

	fmt.Printf("✅ Combined total: %d stocks from all regions\n", len(allScreenerData))

	// If stock screener returned very few results, warn but continue
	if len(allScreenerData) < 1000 {
		fmt.Printf("⚠️  Warning: Only got %d stocks from screener. Expected more with global coverage.\n", len(allScreenerData))
	}

	// Remove duplicates based on company name AND symbol
	// Priority: HK main listing > US ADR > US OTC > European
	companyNames := make(map[string]FMPStockScreener)
	var uniqueScreenerData []FMPStockScreener
	
	for _, stock := range allScreenerData {
		companyKey := strings.ToLower(stock.CompanyName)
		
		// Check if we've seen this company before
		if existingStock, exists := companyNames[companyKey]; exists {
			// Keep the better listing (HK > US ADR > US OTC > European)
			if shouldKeepNewListing(stock, existingStock) {
				companyNames[companyKey] = stock
			}
		} else {
			companyNames[companyKey] = stock
		}
	}
	
	// Convert map back to slice
	for _, stock := range companyNames {
		uniqueScreenerData = append(uniqueScreenerData, stock)
	}
	
	fmt.Printf("🔄 Removed duplicates: %d unique companies\n", len(uniqueScreenerData))

	// Sort by market cap descending to ensure we get the largest companies first
	sort.Slice(uniqueScreenerData, func(i, j int) bool {
		return uniqueScreenerData[i].MarketCap > uniqueScreenerData[j].MarketCap
	})
	fmt.Printf("📊 Sorted by market cap (largest first)\n")

	var assets []AssetData
	stockCount := 0
	maxStocks := 490 // Target ~500 total assets (490 stocks + 10 crypto)
	
	fmt.Printf("🎯 Processing top %d stocks by market cap from %d unique companies...\n", maxStocks, len(uniqueScreenerData))
	
	for _, stock := range uniqueScreenerData {
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
		
		// Skip index funds, ETFs in company name (using word boundaries)
		nameUpper := strings.ToUpper(stock.CompanyName)
		if containsWord(nameUpper, "ETF") ||
		   containsWord(nameUpper, "INDEX") ||
		   containsWord(nameUpper, "FUND") ||
		   containsWord(nameUpper, "SPDR") ||
		   containsWord(nameUpper, "ISHARES") ||
		   containsWord(nameUpper, "VANGUARD") ||
		   containsWord(nameUpper, "INVESCO") {
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
		
		// Convert all prices to USD for proper ranking
		currentPrice := stock.Price
		previousCloseUSD := previousClose
		
		// Determine currency and convert to USD
		var currencyCode string
		if strings.HasSuffix(strings.ToUpper(stock.Symbol), ".HK") || 
		   strings.ToUpper(stock.ExchangeShortName) == "HKSE" ||
		   strings.ToUpper(stock.Country) == "HK" {
			currencyCode = "HKD"
		} else if strings.HasSuffix(strings.ToUpper(stock.Symbol), ".L") ||
		          strings.ToUpper(stock.ExchangeShortName) == "LSE" {
			currencyCode = "GBP"
		} else if strings.HasSuffix(strings.ToUpper(stock.Symbol), ".PA") ||
		          strings.HasSuffix(strings.ToUpper(stock.Symbol), ".DE") ||
		          strings.Contains(strings.ToUpper(stock.ExchangeShortName), "EUR") {
			currencyCode = "EUR"
		} else if strings.HasSuffix(strings.ToUpper(stock.Symbol), ".T") ||
		          strings.ToUpper(stock.Country) == "JP" {
			currencyCode = "JPY"
		} else if strings.HasSuffix(strings.ToUpper(stock.Symbol), ".TO") ||
		          strings.ToUpper(stock.Country) == "CA" {
			currencyCode = "CAD"
		} else if strings.ToUpper(stock.Country) == "CN" {
			currencyCode = "CNY"
		} else if strings.HasSuffix(strings.ToUpper(stock.Symbol), ".JK") ||
		          strings.ToUpper(stock.Country) == "ID" {
			currencyCode = "IDR"
		} else if strings.ToUpper(stock.Country) == "IN" {
			currencyCode = "INR"
		} else if strings.ToUpper(stock.Country) == "KR" {
			currencyCode = "KRW"
		} else if strings.ToUpper(stock.Country) == "BR" {
			currencyCode = "BRL"
		} else if strings.ToUpper(stock.Country) == "MX" {
			currencyCode = "MXN"
		} else if strings.ToUpper(stock.Country) == "ZA" {
			currencyCode = "ZAR"
		} else if strings.ToUpper(stock.Country) == "TH" {
			currencyCode = "THB"
		} else if strings.ToUpper(stock.Country) == "MY" {
			currencyCode = "MYR"
		} else if strings.ToUpper(stock.Country) == "PH" {
			currencyCode = "PHP"
		} else if strings.ToUpper(stock.Country) == "VN" {
			currencyCode = "VND"
		} else {
			currencyCode = "USD" // Default to USD
		}
		
		// Convert to USD if not already in USD
		marketCapUSD := stock.MarketCap
		if currencyCode != "USD" {
			exchangeRate := c.getUSDExchangeRate(currencyCode)
			currentPrice = stock.Price * exchangeRate
			previousCloseUSD = previousClose * exchangeRate
			marketCapUSD = stock.MarketCap * exchangeRate // Convert market cap too!
			
			// Only show conversion for major stocks (top 10 by market cap) to avoid spam
			if stockCount < 10 {
				fmt.Printf("💱 %s: %.2f %s → $%.2f USD | Market Cap: %.1fT %s → $%.1fB USD\n", 
					stock.Symbol, stock.Price, currencyCode, currentPrice,
					stock.MarketCap/1e12, currencyCode, marketCapUSD/1e9)
			}
		}
		
		// Sanity check: Skip if market cap is unrealistically large (> $10 trillion)
		// This prevents currency conversion errors from corrupting the ranking
		if marketCapUSD > 10e12 {
			fmt.Printf("⚠️  Skipping %s: Market cap too large ($%.1fT) - likely currency conversion error\n", 
				stock.Symbol, marketCapUSD/1e12)
			continue
		}
		
		// Additional filter for emerging market data quality issues
		// Indonesian stocks often have inflated market cap data in FMP
		if strings.HasSuffix(strings.ToUpper(stock.Symbol), ".JK") && marketCapUSD > 100e9 {
			fmt.Printf("⚠️  Skipping %s: Indonesian stock with suspicious market cap ($%.1fB)\n", 
				stock.Symbol, marketCapUSD/1e9)
			continue
		}
		
		// Small delay to avoid hitting API rate limits
		time.Sleep(100 * time.Millisecond) // Increased delay due to more API calls

		asset := AssetData{
			Ticker:           stock.Symbol,
			Name:             stock.CompanyName,
			MarketCap:        marketCapUSD,     // Now converted to USD
			CurrentPrice:     currentPrice,     // Now converted to USD
			PreviousClose:    previousCloseUSD, // Now converted to USD
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
	
	// FIRST: Exclude micro contracts explicitly (to prevent duplicates)
	excludedSymbols := map[string]bool{
		"MGCUSD":  true,  // Micro Gold (duplicate of GCUSD)
		"SILUSD":  true,  // Micro Silver (duplicate of SIUSD)
	}
	
	// Check exclusion FIRST before anything else
	if excludedSymbols[symbolUpper] {
		return false
	}
	
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
	fmt.Println("🔄 Smart Deduplication (HK main > US ADR > US OTC > EU)")
	fmt.Println("💵 All prices standardized to USD for accurate ranking")
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