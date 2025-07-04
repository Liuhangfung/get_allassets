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

// New structure for batch market cap response
type FMPBatchMarketCap struct {
	Symbol    string  `json:"symbol"`
	Date      string  `json:"date"`
	MarketCap float64 `json:"marketCap"`
}

// FMP API structures
type FMPStockScreener struct {
	Symbol            string  `json:"symbol"`
	CompanyName       string  `json:"companyName"`
	MarketCap         float64 `json:"marketCap"`
	Sector            string  `json:"sector"`
	Industry          string  `json:"industry"`
	Beta              float64 `json:"beta"`
	Price             float64 `json:"price"`
	Volume            float64 `json:"volume"`
	Exchange          string  `json:"exchange"`
	ExchangeShortName string  `json:"exchangeShortName"`
	Country           string  `json:"country"`
	IsEtf             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
}

type FMPQuote struct {
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Price             float64 `json:"price"`
	ChangesPercentage float64 `json:"changesPercentage"`
	Change            float64 `json:"change"`
	MarketCap         float64 `json:"marketCap"`
	Volume            float64 `json:"volume"`
	Open              float64 `json:"open"`
	PreviousClose     float64 `json:"previousClose"`
	Exchange          string  `json:"exchange"`
	SharesOutstanding float64 `json:"sharesOutstanding"`
}

type FMPCommodity struct {
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Price             float64 `json:"price"`
	ChangesPercentage float64 `json:"changesPercentage"`
	Change            float64 `json:"change"`
	PreviousClose     float64 `json:"previousClose"`
	Exchange          string  `json:"exchange"`
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
	Ticker           string  `json:"ticker"`
	Name             string  `json:"name"`
	MarketCap        float64 `json:"market_cap"`
	CurrentPrice     float64 `json:"current_price"`
	PreviousClose    float64 `json:"previous_close"`
	PercentageChange float64 `json:"percentage_change"`
	Volume           float64 `json:"volume"`
	PrimaryExchange  string  `json:"primary_exchange"`
	Country          string  `json:"country"`
	Sector           string  `json:"sector"`
	Industry         string  `json:"industry"`
	AssetType        string  `json:"asset_type"`
	Image            string  `json:"image"`
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

	// Create request with proper headers for UTF-8 encoding
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to properly handle UTF-8 content
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.HTTPClient.Do(req)
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

// New function to get major global stocks using batch endpoint
func (c *FMPClient) GetMajorGlobalStocks() ([]AssetData, error) {
	fmt.Println("🌍 Fetching major global stocks using batch endpoint...")

	// EXPANDED GLOBAL STOCK LIST - targeting 800+ symbols to get top 500 by market cap
	globalStocks := []string{
		// USA - S&P 500 Major Companies (150+ symbols)
		"AAPL", "MSFT", "GOOGL", "GOOG", "AMZN", "NVDA", "META", "TSLA", "AVGO", "ORCL",
		"WMT", "LLY", "V", "JPM", "TSM", "UNH", "XOM", "MA", "PG", "JNJ", "HD", "NFLX",
		"BAC", "ABBV", "CRM", "KO", "COST", "PEP", "TMO", "MRK", "ADBE", "CSCO", "ACN",
		"TMUS", "ABT", "DIS", "WFC", "AMD", "VZ", "CMCSA", "DHR", "INTU", "TXN", "QCOM",
		"PM", "UNP", "IBM", "SPGI", "GS", "HON", "NKE", "AXP", "BLK", "MS", "SYK", "UBER",
		"NOW", "BKNG", "AMAT", "ISRG", "CATV", "GILD", "MDLZ", "LRCX", "ADP", "VRTX",
		"REGN", "KLAC", "PANW", "PYPL", "CDNS", "SNPS", "CRWD", "MRVL", "ORLY", "CTAS",
		"ADSK", "FTNT", "CHTR", "NXPI", "ABNB", "WDAY", "DDOG", "TEAM", "SNOW", "ZS",
		"OKTA", "SPLK", "VEEV", "DOCU", "ZM", "PTON", "ROKU", "TWLO", "SHOP", "SQ",
		"COIN", "RBLX", "PLTR", "HOOD", "RIVN", "LCID", "GRAB", "DIDI", "BNTX", "MRNA",
		"LULU", "SBUX", "CMG", "CTSH", "ANET", "DXCM", "ILMN", "BIIB", "SGEN", "ALGN",
		"IDXX", "FAST", "ODFL", "VRSK", "ANSS", "CDNS", "MCHP", "PAYX", "CSGP", "LOGI",

		// Hong Kong - Major listings (40+ symbols)
		"700.HK", "9988.HK", "3690.HK", "2318.HK", "1299.HK", "939.HK", "1398.HK", "2388.HK",
		"1113.HK", "2628.HK", "1810.HK", "9618.HK", "9888.HK", "3988.HK", "2020.HK", "1093.HK",
		"2382.HK", "2319.HK", "1177.HK", "1928.HK", "1972.HK", "6098.HK", "1024.HK", "9999.HK",
		"1109.HK", "2007.HK", "1211.HK", "2269.HK", "1833.HK", "1336.HK", "2313.HK", "1658.HK",
		"1766.HK", "1919.HK", "1193.HK", "1888.HK", "6060.HK", "1801.HK", "1088.HK", "1361.HK",

		// China - Major A-shares and ADRs (30+ symbols)
		"BABA", "PDD", "JD", "BIDU", "NTES", "TME", "BILI", "NIO", "XPEV", "LI",
		"TCEHY", "WB", "VIPS", "YMM", "BZUN", "DOYU", "HUYA", "IQ", "MOMO", "QTT",
		"TIGR", "TUYA", "CBPO", "FINV", "GOTU", "LABU", "MOXC", "PAGS", "STNE", "YALA",

		// France - CAC 40 + Major Companies (50+ symbols)
		"MC.PA", "OR.PA", "RMS.PA", "TTE.PA", "SAN.PA", "BNP.PA", "AIR.PA", "CAP.PA",
		"BN.PA", "CS.PA", "ML.PA", "EL.PA", "VIV.PA", "KER.PA", "LR.PA", "ATO.PA",
		"BNP.PA", "CA.PA", "GLE.PA", "SU.PA", "UG.PA", "DSY.PA", "DG.PA", "RNO.PA",
		"SGO.PA", "STM.PA", "SW.PA", "TEP.PA", "URW.PA", "VK.PA", "WLN.PA", "EN.PA",
		"ACA.PA", "ADP.PA", "AI.PA", "AIR.PA", "ALO.PA", "BIM.PA", "BOL.PA", "CDI.PA",
		"CNA.PA", "COV.PA", "DIM.PA", "EI.PA", "ERF.PA", "FDJ.PA", "FP.PA", "FTI.PA",

		// Germany - DAX + Major Companies (50+ symbols)
		"SAP", "ASML", "LVMH", "RMS", "TTE", "SAN", "BNP", "AIR", "CAP", "BN",
		"ALV.DE", "BAS.DE", "BAYN.DE", "BMW.DE", "CON.DE", "DAI.DE", "DBK.DE", "DB1.DE",
		"DTE.DE", "EOAN.DE", "FME.DE", "FRE.DE", "HEI.DE", "HEN3.DE", "IFX.DE", "LIN.DE",
		"MRK.DE", "MUV2.DE", "RWE.DE", "SAP.DE", "SIE.DE", "VOW3.DE", "WDI.DE", "ZAL.DE",
		"ADS.DE", "DHER.DE", "MTX.DE", "PUM.DE", "QIA.DE", "SHL.DE", "SY1.DE", "TEG.DE",
		"VNA.DE", "1COV.DE", "AFX.DE", "AIXA.DE", "BC8.DE", "BEI.DE", "BNR.DE", "CWK.DE",

		// UK - FTSE 100 + Major Companies (60+ symbols)
		"AZN", "LSEG", "UL", "RELX", "GSK", "RIO", "BP", "HSBA.L", "VOD.L",
		"BT.L", "LLOY.L", "BARC.L", "GLEN.L", "BHP.L", "RBS.L", "ULVR.L", "SHEL.L",
		"AAL.L", "ADM.L", "AHT.L", "ANTO.L", "AUTO.L", "AVV.L", "AZN.L", "BA.L",
		"BATS.L", "BLND.L", "BNZL.L", "BRB.L", "BREE.L", "BT.A.L", "CCL.L", "CNA.L",
		"CPG.L", "CRDA.L", "DCC.L", "DGE.L", "EXPN.L", "FERG.L", "FLTR.L", "FRAS.L",
		"GLEN.L", "GSK.L", "HALM.L", "HLN.L", "HWDN.L", "ICG.L", "IHG.L", "IMB.L",
		"INF.L", "ITRK.L", "JD.L", "JET.L", "KGF.L", "LAND.L", "LGEN.L", "MNDI.L",

		// Switzerland - Major stocks (30+ symbols)
		"NESN.SW", "RHHBY", "NOVN.SW", "ROG.SW", "UHR.SW", "GIVN.SW", "LONN.SW",
		"UBSG.SW", "CSGN.SW", "SREN.SW", "ABBN.SW", "SLHN.SW", "ZURN.SW", "GEBN.SW",
		"CFR.SW", "SCMN.SW", "BAER.SW", "TEMN.SW", "STMN.SW", "PGHN.SW", "BUCN.SW",
		"DKSH.SW", "GALE.SW", "HELN.SW", "HOLN.SW", "KNIN.SW", "LHGN.SW", "LISN.SW",
		"METD.SW", "METN.SW", "MOBN.SW", "OREA.SW", "PARG.SW", "PSPN.SW", "SGSN.SW",

		// Saudi Arabia - TADAWUL Major stocks (30+ symbols)
		"2222.SR", "2030.SR", "1120.SR", "2380.SR", "1150.SR", "7020.SR", "1180.SR",
		"4030.SR", "2010.SR", "2020.SR", "1210.SR", "2290.SR", "1320.SR", "2350.SR",
		"1050.SR", "2040.SR", "2080.SR", "2090.SR", "2170.SR", "2220.SR", "2230.SR",
		"2250.SR", "2260.SR", "2270.SR", "2280.SR", "2300.SR", "2310.SR", "2320.SR",
		"2330.SR", "2340.SR", "2360.SR", "2370.SR", "4001.SR", "4002.SR", "4003.SR",

		// Japan - Nikkei 225 Major stocks (60+ symbols)
		"7203.T", "6098.T", "6861.T", "8306.T", "9984.T", "4063.T", "6758.T", "6981.T",
		"9432.T", "8035.T", "4568.T", "6594.T", "4751.T", "6902.T", "7974.T", "6762.T",
		"4502.T", "4503.T", "4519.T", "4523.T", "4578.T", "4689.T", "4901.T", "4911.T",
		"4967.T", "5019.T", "5020.T", "5101.T", "5108.T", "5201.T", "5202.T", "5233.T",
		"5401.T", "5411.T", "5541.T", "5631.T", "5703.T", "5711.T", "5713.T", "5714.T",
		"5801.T", "5802.T", "5803.T", "5901.T", "5947.T", "5949.T", "6103.T", "6113.T",
		"6178.T", "6273.T", "6301.T", "6326.T", "6361.T", "6367.T", "6473.T", "6501.T",

		// Australia - ASX 200 Major stocks (40+ symbols)
		"BHP.AX", "CBA.AX", "CSL.AX", "WBC.AX", "ANZ.AX", "NAB.AX", "WES.AX", "TLS.AX",
		"RIO.AX", "TCL.AX", "WDS.AX", "FMG.AX", "STO.AX", "WOW.AX", "MQG.AX", "NCM.AX",
		"APT.AX", "APX.AX", "ASX.AX", "BXB.AX", "CAR.AX", "COL.AX", "CPU.AX", "CWN.AX",
		"DXS.AX", "GMG.AX", "GOZ.AX", "GPT.AX", "IAG.AX", "IEL.AX", "JHX.AX", "LYC.AX",
		"MGR.AX", "MIN.AX", "NST.AX", "NXT.AX", "ORA.AX", "ORG.AX", "PDN.AX", "PLS.AX",

		// Canada - TSX Major stocks (40+ symbols)
		"RY.TO", "TD.TO", "ENB.TO", "CNR.TO", "BNS.TO", "BAM.TO", "CSU.TO", "TRI.TO",
		"BMO.TO", "CM.TO", "CNQ.TO", "L.TO", "ATD.TO", "SU.TO", "WCN.TO", "MFC.TO",
		"ABX.TO", "AEM.TO", "AQN.TO", "AW.UN.TO", "CCO.TO", "CP.TO", "CVE.TO", "DOL.TO",
		"EMA.TO", "ENS.TO", "FFH.TO", "FTS.TO", "GIB.A.TO", "GWO.TO", "H.TO", "IFC.TO",
		"IMO.TO", "K.TO", "KL.TO", "MG.TO", "NA.TO", "NTR.TO", "OTEX.TO", "POW.TO",

		// Netherlands - AEX Major stocks (30+ symbols)
		"ASML", "RDSA", "UNA", "PHIA", "INGA", "HEIA", "MT", "ADYEN", "DSM", "ABN",
		"AKZA.AS", "ASML.AS", "BESI.AS", "DSM.AS", "GLPG.AS", "HEIA.AS", "INGA.AS",
		"KPN.AS", "NN.AS", "PHIA.AS", "PROSUS.AS", "RDSA.AS", "SBMO.AS", "TKWY.AS",
		"UNA.AS", "WKL.AS", "AALB.AS", "ABN.AS", "AD.AS", "AGN.AS", "ALFEN.AS", "APAM.AS",

		// Brazil - Bovespa Major stocks (25+ symbols)
		"VALE", "PETR4.SA", "ITUB4.SA", "BBDC4.SA", "ABEV3.SA", "MGLU3.SA", "WEGE3.SA",
		"B3SA3.SA", "RENT3.SA", "LREN3.SA", "SUZB3.SA", "RAIL3.SA", "VVAR3.SA", "RADL3.SA",
		"HAPV3.SA", "EQTL3.SA", "NTCO3.SA", "CSAN3.SA", "CSNA3.SA", "USIM5.SA", "GOAU4.SA",
		"TIMS3.SA", "KLBN11.SA", "SUZB5.SA", "EMBR3.SA", "CCRO3.SA", "GGBR4.SA", "BEEF3.SA",

		// South Korea - KOSPI Major stocks (25+ symbols)
		"005930.KS", "000660.KS", "035420.KS", "005380.KS", "035720.KS", "051910.KS",
		"006400.KS", "028260.KS", "055550.KS", "105560.KS", "096770.KS", "003670.KS",
		"017670.KS", "030200.KS", "036570.KS", "034020.KS", "018260.KS", "015760.KS",
		"009150.KS", "010950.KS", "011070.KS", "012330.KS", "016360.KS", "021240.KS",
		"024110.KS", "029780.KS", "032640.KS", "047050.KS", "051900.KS", "078930.KS",

		// Taiwan - TWSE Major stocks (25+ symbols)
		"2330.TW", "2317.TW", "2454.TW", "2412.TW", "1303.TW", "2308.TW", "1216.TW",
		"2002.TW", "1301.TW", "2881.TW", "2882.TW", "2891.TW", "2892.TW", "2912.TW",
		"3008.TW", "3034.TW", "3045.TW", "3711.TW", "4904.TW", "4938.TW", "5871.TW",
		"5880.TW", "6505.TW", "6770.TW", "8454.TW", "9904.TW", "9910.TW", "9921.TW",

		// Spain - IBEX 35 Major stocks (25+ symbols)
		"SAN", "IBE", "TEF", "BBVA", "ITX", "REP", "ENG", "AMS", "CABK", "FER",
		"AENA.MC", "ACS.MC", "CLNX.MC", "COL.MC", "ELE.MC", "ENC.MC", "FDR.MC", "GRF.MC",
		"IAG.MC", "IDR.MC", "LOG.MC", "MAP.MC", "MAS.MC", "MEL.MC", "MTS.MC", "MRL.MC",
		"REE.MC", "ROVI.MC", "SCYR.MC", "SLR.MC", "TL5.MC", "TRE.MC", "UNI.MC", "VIS.MC",

		// Italy - FTSE MIB Major stocks (25+ symbols)
		"ENEL.MI", "ENI.MI", "ISP.MI", "UCG.MI", "TIT.MI", "STM.MI", "RACE.MI",
		"AZM.MI", "BAMI.MI", "BMED.MI", "BPSO.MI", "BSRP.MI", "BUZZ.MI", "CNH.MI",
		"CPR.MI", "DIA.MI", "EXO.MI", "FCA.MI", "G.MI", "HER.MI", "IG.MI", "INW.MI",
		"IP.MI", "LDO.MI", "MB.MI", "MONC.MI", "PIRC.MI", "PRY.MI", "PST.MI", "REC.MI",

		// Mexico - BMV Major stocks (20+ symbols)
		"WALMEX.MX", "FEMSA.MX", "GMEXICO.MX", "TLEVISA.MX", "BIMBO.MX", "AMX.MX",
		"ALFA.MX", "ALSEA.MX", "ASUR.MX", "CEMEX.MX", "ELEKTRA.MX", "GFNORTE.MX",
		"GRUMA.MX", "KIMBER.MX", "LALA.MX", "LIVEPOLC.MX", "MEGACPO.MX", "NEMAK.MX",
		"ORBIA.MX", "PINFRA.MX", "SITESB.MX", "VOLARIS.MX", "WALMEX.MX", "AC.MX",

		// South Africa - JSE Major stocks (20+ symbols)
		"NPN.JO", "PRX.JO", "DSY.JO", "SHP.JO", "BHP.JO", "SOL.JO", "MTN.JO",
		"ABG.JO", "AGL.JO", "AMS.JO", "ANG.JO", "APN.JO", "BAW.JO", "BID.JO",
		"BIL.JO", "BVT.JO", "CFR.JO", "CLS.JO", "CPI.JO", "DSY.JO", "EXX.JO",
		"FSR.JO", "GFI.JO", "GRT.JO", "HAR.JO", "IMP.JO", "INL.JO", "INP.JO",

		// Nordic countries - Major stocks (20+ symbols)
		"NOVO-B.CO", "ASML.AS", "RDSA.AS", "UNA.AS", "PHIA.AS", "INGA.AS", "HEIA.AS",
		"VOLV-B.ST", "ERIC-B.ST", "SEB-A.ST", "SWED-A.ST", "TEL2-B.ST", "TELIA.ST",
		"INVESTOR-B.ST", "ATCO-A.ST", "ATCO-B.ST", "KINV-B.ST", "LUNDBERGB.ST",
		"SAND.ST", "SSAB-A.ST", "SSAB-B.ST", "SKF-B.ST", "SECU-B.ST", "SINCH.ST",

		// Additional major ADRs and cross-listings (30+ symbols)
		"TSM", "ASML", "SAP", "TM", "SONY", "NVO", "UL", "DEO", "BCS", "ING",
		"DB", "NVS", "RHHBY", "SNY", "AZN", "GSK", "BP", "RIO", "BHP",
		"WBK", "LYG", "MUFG", "SMFG", "NMR", "E", "CEO", "CX", "BBD", "SID",
		"ITUB", "BBD", "PBR", "VALE", "ABEV", "TIMB", "STNE", "PAGS", "NU", "MELI",
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueStocks := []string{}
	for _, symbol := range globalStocks {
		if !seen[symbol] {
			seen[symbol] = true
			uniqueStocks = append(uniqueStocks, symbol)
		}
	}

	fmt.Printf("📊 Preparing to fetch %d unique global stocks...\n", len(uniqueStocks))

	// Process in batches of 75 (optimized for parallel processing)
	batchSize := 75

	// Pre-calculate exchange rates to avoid repeated API calls
	exchangeRateCache := make(map[string]float64)
	var rateMutex sync.RWMutex
	var allAssets []AssetData
	var totalRequested int
	var totalReceived int
	var totalProcessed int
	var skippedInvalidMarketCap int
	var skippedQuoteFailed int

	// Use mutex for thread-safe operations
	var mu sync.Mutex

	for i := 0; i < len(uniqueStocks); i += batchSize {
		end := i + batchSize
		if end > len(uniqueStocks) {
			end = len(uniqueStocks)
		}

		batch := uniqueStocks[i:end]
		symbols := strings.Join(batch, ",")
		totalRequested += len(batch)

		fmt.Printf("🔄 Processing batch %d/%d (%d symbols)...\n",
			(i/batchSize)+1, (len(uniqueStocks)+batchSize-1)/batchSize, len(batch))

		// Use the batch market cap endpoint (different base URL)
		url := fmt.Sprintf("https://financialmodelingprep.com/stable/market-capitalization-batch?symbols=%s&apikey=%s", symbols, c.APIKey)

		// Create request with proper headers for UTF-8 encoding
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("⚠️  Failed to create batch request %d: %v\n", (i/batchSize)+1, err)
			continue
		}

		// Set headers to properly handle UTF-8 content
		req.Header.Set("Accept", "application/json; charset=utf-8")
		req.Header.Set("Accept-Charset", "utf-8")
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			fmt.Printf("⚠️  Failed to fetch batch %d: %v\n", (i/batchSize)+1, err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Printf("⚠️  Failed to read batch %d response: %v\n", (i/batchSize)+1, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("⚠️  Batch %d API error (status %d): %s\n", (i/batchSize)+1, resp.StatusCode, string(body))
			continue
		}

		var batchData []FMPBatchMarketCap
		if err := json.Unmarshal(body, &batchData); err != nil {
			fmt.Printf("⚠️  Failed to parse batch %d data: %v\n", (i/batchSize)+1, err)
			continue
		}

		fmt.Printf("✅ Batch %d: Received %d stocks\n", (i/batchSize)+1, len(batchData))
		totalReceived += len(batchData)

		// Process each stock in the batch in parallel
		var batchWg sync.WaitGroup
		var batchMu sync.Mutex
		var batchAssets []AssetData

		// Use semaphore to limit concurrent API calls per batch
		quoteSemaphore := make(chan struct{}, 10) // Max 10 concurrent quotes per batch

		for _, stock := range batchData {
			if stock.MarketCap <= 0 {
				skippedInvalidMarketCap++
				continue // Skip invalid market caps
			}

			// Process each stock in parallel
			batchWg.Add(1)
			go func(stock FMPBatchMarketCap) {
				defer batchWg.Done()

				// Acquire semaphore to limit concurrent API calls
				quoteSemaphore <- struct{}{}
				defer func() { <-quoteSemaphore }()

				// Debug: Log original batch data for major stocks
				if stock.MarketCap > 100e9 { // Log stocks > $100B for debugging
					fmt.Printf("🔍 DEBUG Batch Data - %s: Market Cap = %.2fB (from batch endpoint)\n",
						stock.Symbol, stock.MarketCap/1e9)
				}

				// Get real-time quote for additional data
				quote, err := c.GetQuote(stock.Symbol)
				if err != nil {
					// Skip if quote fails but log it
					fmt.Printf("⚠️  Failed to get quote for %s: %v\n", stock.Symbol, err)
					mu.Lock()
					skippedQuoteFailed++
					mu.Unlock()
					return
				}

				// Debug: Log quote data for comparison
				if stock.MarketCap > 100e9 {
					fmt.Printf("🔍 DEBUG Quote Data - %s: Market Cap = %.2fB (from quote endpoint)\n",
						stock.Symbol, quote.MarketCap/1e9)
				}

				// Get company profile for additional details (only for larger companies to save API calls)
				var profile *FMPCompanyProfile
				if stock.MarketCap > 10e9 { // Only fetch profiles for companies > $10B
					profile, err = c.GetCompanyProfile(stock.Symbol)
					if err != nil {
						// Use default values if profile fails
						profile = &FMPCompanyProfile{
							Symbol:      stock.Symbol,
							CompanyName: quote.Name,
							Country:     "Unknown",
							Sector:      "Unknown",
							Industry:    "Unknown",
							Image:       "",
						}
					}
				} else {
					// Use default values for smaller companies to save API calls
					profile = &FMPCompanyProfile{
						Symbol:      stock.Symbol,
						CompanyName: quote.Name,
						Country:     "Unknown",
						Sector:      "Unknown",
						Industry:    "Unknown",
						Image:       "",
					}
				}

				// Detect currency and country
				currencyCode := c.detectCurrency(stock.Symbol, profile.Country)

				// Keep prices in original currency, but determine best market cap source
				currentPrice := quote.Price
				previousClose := quote.PreviousClose

				// Choose more reliable market cap source
				var sourceMarketCap float64
				var marketCapSource string

				// For non-USD stocks, prefer quote endpoint if significantly different from batch
				if currencyCode != "USD" && quote.MarketCap > 0 {
					// Compare batch vs quote market caps
					batchCap := stock.MarketCap
					quoteCap := quote.MarketCap

					// If quote cap is much larger than batch cap, quote might already be in USD
					ratio := quoteCap / batchCap
					if ratio > 10 && quoteCap > 100e9 { // Quote is >10x larger and >$100B
						sourceMarketCap = quoteCap
						marketCapSource = "quote (likely USD)"
						fmt.Printf("🔄 Using quote endpoint for %s: batch=%.1fB vs quote=%.1fB (ratio=%.1fx)\n",
							stock.Symbol, batchCap/1e9, quoteCap/1e9, ratio)
					} else {
						sourceMarketCap = batchCap
						marketCapSource = "batch (local currency)"
					}
				} else {
					sourceMarketCap = stock.MarketCap
					marketCapSource = "batch"
				}

				marketCapUSD := sourceMarketCap
				originalMarketCap := sourceMarketCap

				// Convert market cap to USD if needed and source is in local currency
				if currencyCode != "USD" && marketCapSource != "quote (likely USD)" {
					// Use cached exchange rate if available
					rateMutex.RLock()
					exchangeRate, exists := exchangeRateCache[currencyCode]
					rateMutex.RUnlock()

					if !exists {
						// Fetch and cache the exchange rate
						exchangeRate = c.getUSDExchangeRate(currencyCode)
						rateMutex.Lock()
						exchangeRateCache[currencyCode] = exchangeRate
						rateMutex.Unlock()
					}

					marketCapUSD = sourceMarketCap * exchangeRate

					// Log ALL currency conversions for debugging
					fmt.Printf("💱 Converting %s (%s): %.1fB %s × %.6f = $%.1fB USD [source: %s]\n",
						stock.Symbol, profile.CompanyName, originalMarketCap/1e9, currencyCode, exchangeRate, marketCapUSD/1e9, marketCapSource)
				} else {
					// Log USD/already converted stocks
					if sourceMarketCap > 50e9 { // Only log major stocks
						fmt.Printf("💵 %s (%s): $%.1fB USD [source: %s, no conversion needed]\n",
							stock.Symbol, profile.CompanyName, marketCapUSD/1e9, marketCapSource)
					}
				}

				// Determine country from symbol if not in profile
				country := profile.Country
				if country == "Unknown" || country == "" {
					country = c.detectCountryFromSymbol(stock.Symbol)
				}

				asset := AssetData{
					Ticker:           stock.Symbol,
					Name:             profile.CompanyName,
					MarketCap:        marketCapUSD,  // Already in USD from batch endpoint
					CurrentPrice:     currentPrice,  // Keep in original currency
					PreviousClose:    previousClose, // Keep in original currency
					PercentageChange: quote.ChangesPercentage,
					Volume:           quote.Volume,
					PrimaryExchange:  quote.Exchange,
					Country:          country,
					Sector:           profile.Sector,
					Industry:         profile.Industry,
					AssetType:        "stock",
					Image:            profile.Image,
				}

				// Add to batch results with mutex protection
				batchMu.Lock()
				batchAssets = append(batchAssets, asset)
				batchMu.Unlock()

				// Log major additions
				if marketCapUSD > 50e9 { // Log stocks > $50B
					fmt.Printf("🌍 Added: %s (%s) - %s | Market Cap: $%.1fB USD | Country: %s\n",
						stock.Symbol, profile.CompanyName, currencyCode, marketCapUSD/1e9, country)
				}
			}(stock)
		}

		// Wait for all stocks in this batch to complete
		batchWg.Wait()

		// Add batch results to global results with mutex protection
		mu.Lock()
		allAssets = append(allAssets, batchAssets...)
		totalProcessed += len(batchAssets)
		mu.Unlock()

		// Rate limiting between batches (reduced since we're using parallel processing)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("✅ Successfully processed %d global stocks from batch endpoint\n", len(allAssets))

	// Print detailed statistics
	fmt.Printf("\n📊 BATCH PROCESSING STATISTICS:\n")
	fmt.Printf("   🎯 Total symbols requested: %d\n", totalRequested)
	fmt.Printf("   📥 Total symbols received from API: %d\n", totalReceived)
	fmt.Printf("   ✅ Total symbols successfully processed: %d\n", totalProcessed)
	fmt.Printf("   ❌ Skipped due to invalid market cap: %d\n", skippedInvalidMarketCap)
	fmt.Printf("   ❌ Skipped due to quote API failure: %d\n", skippedQuoteFailed)
	fmt.Printf("   📉 Missing symbols (not returned by batch API): %d\n", totalRequested-totalReceived)

	return allAssets, nil
}

// Helper function to detect country from symbol
func (c *FMPClient) detectCountryFromSymbol(symbol string) string {
	symbolUpper := strings.ToUpper(symbol)

	if strings.HasSuffix(symbolUpper, ".HK") {
		return "HK"
	} else if strings.HasSuffix(symbolUpper, ".PA") {
		return "FR"
	} else if strings.HasSuffix(symbolUpper, ".L") {
		return "GB"
	} else if strings.HasSuffix(symbolUpper, ".DE") {
		return "DE"
	} else if strings.HasSuffix(symbolUpper, ".SW") {
		return "CH"
	} else if strings.HasSuffix(symbolUpper, ".SR") {
		return "SA"
	} else if strings.HasSuffix(symbolUpper, ".T") {
		return "JP"
	} else if strings.HasSuffix(symbolUpper, ".AX") {
		return "AU"
	} else if strings.HasSuffix(symbolUpper, ".TO") {
		return "CA"
	} else if strings.HasSuffix(symbolUpper, ".SA") {
		return "BR"
	} else if strings.HasSuffix(symbolUpper, ".KS") {
		return "KR"
	} else if strings.HasSuffix(symbolUpper, ".TW") {
		return "TW"
	} else if strings.HasSuffix(symbolUpper, ".MI") {
		return "IT"
	} else if strings.HasSuffix(symbolUpper, ".MX") {
		return "MX"
	} else if strings.HasSuffix(symbolUpper, ".ME") {
		return "RU"
	} else if strings.HasSuffix(symbolUpper, ".JO") {
		return "ZA"
	} else if strings.HasSuffix(symbolUpper, ".CO") {
		return "DK"
	} else if strings.HasSuffix(symbolUpper, ".AS") {
		return "NL"
	} else {
		return "US" // Default to US for symbols without country suffix
	}
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

	// London primary listings (highest priority for UK companies)
	if strings.HasSuffix(symbolUpper, ".L") || exchangeUpper == "LSE" {
		return 1
	}

	// European main listings (high priority for EU companies)
	if strings.HasSuffix(symbolUpper, ".PA") || strings.HasSuffix(symbolUpper, ".DE") ||
		strings.HasSuffix(symbolUpper, ".VI") || strings.HasSuffix(symbolUpper, ".SW") {
		return 2
	}

	// US main listings (high priority for US companies)
	if (exchangeUpper == "NASDAQ" || exchangeUpper == "NYSE") &&
		!strings.HasSuffix(symbolUpper, "Y") && !strings.HasSuffix(symbolUpper, "H") {
		return 3
	}

	// US ADRs (lower priority than primary listings)
	if (exchangeUpper == "NASDAQ" || exchangeUpper == "NYSE") &&
		(strings.HasSuffix(symbolUpper, "Y") || strings.HasSuffix(symbolUpper, "H")) {
		return 4
	}

	// US OTC (lower priority)
	if exchangeUpper == "OTC" || strings.HasSuffix(symbolUpper, "F") {
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
		"HKD": 0.128,    // 1 HKD = ~0.128 USD
		"EUR": 1.08,     // 1 EUR = ~1.08 USD
		"GBP": 1.26,     // 1 GBP = ~1.26 USD
		"JPY": 0.0067,   // 1 JPY = ~0.0067 USD
		"CAD": 0.74,     // 1 CAD = ~0.74 USD
		"AUD": 0.64,     // 1 AUD = ~0.64 USD
		"CNY": 0.14,     // 1 CNY = ~0.14 USD
		"DKK": 0.145,    // 1 DKK = ~0.145 USD (about 6.9 DKK = 1 USD)
		"IDR": 0.000065, // 1 IDR = ~0.000065 USD (about 15,400 IDR = 1 USD)
		"INR": 0.012,    // 1 INR = ~0.012 USD (about 83 INR = 1 USD)
		"KRW": 0.00075,  // 1 KRW = ~0.00075 USD (about 1,330 KRW = 1 USD)
		"BRL": 0.18,     // 1 BRL = ~0.18 USD (about 5.5 BRL = 1 USD)
		"MXN": 0.058,    // 1 MXN = ~0.058 USD (about 17 MXN = 1 USD)
		"ZAR": 0.055,    // 1 ZAR = ~0.055 USD (about 18 ZAR = 1 USD)
		"THB": 0.029,    // 1 THB = ~0.029 USD (about 34 THB = 1 USD)
		"MYR": 0.22,     // 1 MYR = ~0.22 USD (about 4.5 MYR = 1 USD)
		"PHP": 0.018,    // 1 PHP = ~0.018 USD (about 56 PHP = 1 USD)
		"VND": 0.000040, // 1 VND = ~0.000040 USD (about 25,000 VND = 1 USD)
		"SGD": 0.74,     // 1 SGD = ~0.74 USD
		"TWD": 0.031,    // 1 TWD = ~0.031 USD (about 32 TWD = 1 USD)
		"CLP": 0.0010,   // 1 CLP = ~0.0010 USD (about 1,000 CLP = 1 USD)
		"SAR": 0.267,    // 1 SAR = ~0.267 USD (about 3.75 SAR = 1 USD)
		"ILS": 0.27,     // 1 ILS = ~0.27 USD (about 3.7 ILS = 1 USD)
		"COP": 0.00025,  // 1 COP = ~0.00025 USD (about 4,000 COP = 1 USD)
		"PEN": 0.27,     // 1 PEN = ~0.27 USD (about 3.7 PEN = 1 USD)
		"EGP": 0.020,    // 1 EGP = ~0.020 USD (about 50 EGP = 1 USD)
		"TRY": 0.029,    // 1 TRY = ~0.029 USD (about 34 TRY = 1 USD)
		"RUB": 0.010,    // 1 RUB = ~0.010 USD (about 100 RUB = 1 USD)
	}

	if rate, exists := fallbackRates[fromCurrency]; exists {
		return rate
	}

	// If unknown currency, assume it's already in USD
	return 1.0
}

func (c *FMPClient) GetGlobalStocks() ([]AssetData, error) {
	fmt.Println("🌍 Fetching top 500 stocks globally with USD conversion...")

	// Get all stocks globally and sort by market cap
	endpoint := "/v3/stock-screener?marketCapMoreThan=100000000&limit=5000&order=desc&sortBy=marketcap&isActivelyTrading=true"

	fmt.Printf("📡 Fetching global stocks from FMP...\n")

	body, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch global stocks: %w", err)
	}

	var allStocks []FMPStockScreener
	if err := json.Unmarshal(body, &allStocks); err != nil {
		return nil, fmt.Errorf("failed to parse global stocks: %w", err)
	}

	fmt.Printf("✅ Received %d stocks globally\n", len(allStocks))

	// Remove duplicates and ETFs with smarter deduplication
	var validStocks []FMPStockScreener
	seenSymbols := make(map[string]bool)
	companyListings := make(map[string]FMPStockScreener) // Track best listing per company

	for _, stock := range allStocks {
		// Skip ETFs and index funds
		if stock.IsEtf {
			continue
		}

		nameUpper := strings.ToUpper(stock.CompanyName)
		if containsWord(nameUpper, "ETF") ||
			containsWord(nameUpper, "INDEX") ||
			containsWord(nameUpper, "FUND") ||
			containsWord(nameUpper, "SPDR") ||
			containsWord(nameUpper, "ISHARES") ||
			containsWord(nameUpper, "VANGUARD") {
			continue
		}

		// Skip if already seen this exact symbol
		if seenSymbols[stock.Symbol] {
			continue
		}
		seenSymbols[stock.Symbol] = true

		if stock.IsActivelyTrading && stock.MarketCap > 0 {
			// Check if we already have a listing for this company
			if existingStock, exists := companyListings[stock.CompanyName]; exists {
				// Keep the better listing based on priority
				if shouldKeepNewListing(stock, existingStock) {
					companyListings[stock.CompanyName] = stock
				}
			} else {
				// First time seeing this company
				companyListings[stock.CompanyName] = stock
			}
		}
	}

	// Convert map to slice
	for _, stock := range companyListings {
		validStocks = append(validStocks, stock)
	}

	fmt.Printf("🔄 Filtered to %d valid stocks (removed ETFs and duplicates)\n", len(validStocks))

	// Process stocks and convert to USD
	var assets []AssetData
	maxStocks := 490 // Target 490 stocks + 10 crypto = 500 total

	fmt.Printf("💱 Converting market caps to USD (keeping original prices) and getting real-time data...\n")

	for i, stock := range validStocks {
		if len(assets) >= maxStocks {
			break
		}

		// Detect currency from symbol and country
		currencyCode := c.detectCurrency(stock.Symbol, stock.Country)

		// Keep original price in local currency, only convert market cap to USD
		currentPrice := stock.Price
		marketCapUSD := stock.MarketCap

		if currencyCode != "USD" {
			exchangeRate := c.getUSDExchangeRate(currencyCode)
			// Only convert market cap to USD for comparison (not price)
			marketCapUSD = stock.MarketCap * exchangeRate

			// Log major conversions for transparency
			if marketCapUSD > 10e9 { // Log conversions for assets > $10B
				fmt.Printf("💱 %s: %.2f %s (original) | Market Cap: $%.1fB USD\n",
					stock.Symbol, stock.Price, currencyCode, marketCapUSD/1e9)
			}
		}

		// Get real-time quote for current prices
		quote, err := c.GetQuote(stock.Symbol)
		var percentageChange float64
		var previousClose float64
		var volume float64

		if err == nil && quote != nil {
			// Use real-time data in original currency
			currentPrice = quote.Price
			previousClose = quote.PreviousClose
			percentageChange = quote.ChangesPercentage
			volume = quote.Volume

			// Only convert market cap to USD (not prices)
			if currencyCode != "USD" {
				exchangeRate := c.getUSDExchangeRate(currencyCode)
				// Recalculate market cap in USD with real-time price
				if quote.SharesOutstanding > 0 {
					marketCapUSD = (quote.Price * exchangeRate) * quote.SharesOutstanding
				}
			}
		} else {
			// Fallback to screener data
			previousClose = currentPrice * 0.99 // Approximate
			percentageChange = 1.0              // Approximate
			volume = stock.Volume
		}

		// Determine asset type
		assetType := "stock"
		nameUpper := strings.ToUpper(stock.CompanyName)
		if containsWord(nameUpper, "REIT") {
			assetType = "reit"
		}

		// Get company profile for image
		imageURL := ""
		profile, err := c.GetCompanyProfile(stock.Symbol)
		if err == nil && profile != nil {
			imageURL = profile.Image
		}

		asset := AssetData{
			Ticker:           stock.Symbol,
			Name:             stock.CompanyName,
			MarketCap:        marketCapUSD,  // USD for comparison
			CurrentPrice:     currentPrice,  // Original currency
			PreviousClose:    previousClose, // Original currency
			PercentageChange: percentageChange,
			Volume:           volume,
			PrimaryExchange:  stock.ExchangeShortName,
			Country:          stock.Country,
			Sector:           stock.Sector,
			Industry:         stock.Industry,
			AssetType:        assetType,
			Image:            imageURL,
		}

		assets = append(assets, asset)

		// Progress update
		if (i+1)%50 == 0 {
			fmt.Printf("📊 Processed %d/%d stocks...\n", i+1, len(validStocks))
		}

		// Rate limiting
		time.Sleep(20 * time.Millisecond)
	}

	// Re-rank by USD market cap (most important step!)
	fmt.Printf("🏆 Re-ranking %d assets by USD market cap...\n", len(assets))
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].MarketCap > assets[j].MarketCap
	})

	// Take top 490 after USD conversion and ranking
	if len(assets) > maxStocks {
		assets = assets[:maxStocks]
	}

	fmt.Printf("✅ Final result: Top %d stocks ranked by USD market cap\n", len(assets))

	return assets, nil
}

// Helper function to detect currency from symbol and country
func (c *FMPClient) detectCurrency(symbol, country string) string {
	symbolUpper := strings.ToUpper(symbol)
	countryUpper := strings.ToUpper(country)

	// Symbol-based detection (most reliable)
	if strings.HasSuffix(symbolUpper, ".JK") || countryUpper == "ID" {
		return "IDR"
	} else if strings.HasSuffix(symbolUpper, ".L") || countryUpper == "GB" {
		return "GBP"
	} else if strings.HasSuffix(symbolUpper, ".PA") || strings.HasSuffix(symbolUpper, ".DE") ||
		strings.HasSuffix(symbolUpper, ".MI") || strings.HasSuffix(symbolUpper, ".AS") ||
		countryUpper == "FR" || countryUpper == "DE" || countryUpper == "IT" || countryUpper == "ES" || countryUpper == "NL" {
		return "EUR"
	} else if strings.HasSuffix(symbolUpper, ".T") || countryUpper == "JP" {
		return "JPY"
	} else if strings.HasSuffix(symbolUpper, ".HK") || countryUpper == "HK" {
		return "HKD"
	} else if strings.HasSuffix(symbolUpper, ".TO") || countryUpper == "CA" {
		return "CAD"
	} else if strings.HasSuffix(symbolUpper, ".AX") || countryUpper == "AU" {
		return "AUD"
	} else if strings.HasSuffix(symbolUpper, ".SS") || strings.HasSuffix(symbolUpper, ".SZ") || countryUpper == "CN" {
		return "CNY"
	} else if strings.HasSuffix(symbolUpper, ".TW") || countryUpper == "TW" {
		return "TWD"
	} else if strings.HasSuffix(symbolUpper, ".KS") || strings.HasSuffix(symbolUpper, ".KQ") || countryUpper == "KR" {
		return "KRW"
	} else if strings.HasSuffix(symbolUpper, ".SA") || countryUpper == "BR" {
		return "BRL"
	} else if strings.HasSuffix(symbolUpper, ".MX") || countryUpper == "MX" {
		return "MXN"
	} else if strings.HasSuffix(symbolUpper, ".TA") || countryUpper == "IL" {
		return "ILS"
	} else if strings.HasSuffix(symbolUpper, ".SR") || countryUpper == "SA" {
		return "SAR"
	} else if strings.HasSuffix(symbolUpper, ".BA") || countryUpper == "AR" {
		return "ARS"
	} else if strings.HasSuffix(symbolUpper, ".CO") || countryUpper == "DK" {
		return "DKK"
	} else if countryUpper == "IN" {
		return "INR"
	} else if countryUpper == "ZA" {
		return "ZAR"
	}

	// Default to USD
	return "USD"
}

// Helper function to identify real physical commodities (only essential ones)
func isRealCommodity(name, symbol string) bool {
	nameUpper := strings.ToUpper(name)
	symbolUpper := strings.ToUpper(symbol)

	// FIRST: Exclude micro contracts explicitly (to prevent duplicates)
	excludedSymbols := map[string]bool{
		"MGCUSD": true, // Micro Gold (duplicate of GCUSD)
		"SILUSD": true, // Micro Silver (duplicate of SIUSD)
	}

	// Check exclusion FIRST before anything else
	if excludedSymbols[symbolUpper] {
		return false
	}

	// Essential commodities we want (matching FMP names and symbols exactly)
	essentialCommodities := map[string]bool{
		// Metals only (name contains)
		"GOLD":      true,
		"SILVER":    true,
		"PLATINUM":  true,
		"PALLADIUM": true,
		"COPPER":    true,
	}

	// Check if name contains any essential commodity
	for commodity := range essentialCommodities {
		if strings.Contains(nameUpper, commodity) {
			return true
		}
	}

	// Essential symbols we want (exact matches from FMP) - Main contracts only
	essentialSymbols := map[string]bool{
		"GCUSD": true, // Gold Futures (main contract)
		"SIUSD": true, // Silver Futures (main contract)
		"PLUSD": true, // Platinum
		"PAUSD": true, // Palladium
		"HGUSD": true, // Copper
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
	// Disable HTML escaping to preserve special characters like é, ë, ã, ç
	encoder.SetEscapeHTML(false)

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

	fmt.Println("🌟 COMPREHENSIVE GLOBAL MARKET ANALYSIS WITH FMP API")
	fmt.Println("📈 STRATEGY: Fetch 800+ global stocks → Rank by market cap → Select TOP 500")
	fmt.Println("Fetching expanded global stock list from all major markets:")
	fmt.Println("🇺🇸 USA: 150+ stocks (AAPL, MSFT, GOOGL, NVDA, META, TSLA, UBER, SNOW...)")
	fmt.Println("🇭🇰 Hong Kong: 40+ stocks (700.HK Tencent, 9988.HK Alibaba, 3690.HK Meituan...)")
	fmt.Println("🇫🇷 France: 50+ stocks (MC.PA LVMH, RMS.PA Hermes, TTE.PA TotalEnergies...)")
	fmt.Println("🇨🇭 Switzerland: 30+ stocks (NESN.SW Nestle, NOVN.SW Novartis, ROG.SW Roche...)")
	fmt.Println("🇸🇦 Saudi Arabia: 30+ stocks (2222.SR Saudi Aramco, 1120.SR Al Rajhi Bank...)")
	fmt.Println("🇬🇧 UK: 60+ stocks (SHEL.L Shell, AZN AstraZeneca, BP, ULVR.L Unilever...)")
	fmt.Println("🇯🇵 Japan: 60+ stocks (7203.T Toyota, 6861.T Keyence, 6098.T Recruit...)")
	fmt.Println("🇨🇳 China: 30+ stocks (BABA, PDD, JD, BIDU, TCEHY, NIO, XPEV...)")
	fmt.Println("🌍 Plus: Germany, Australia, Canada, Brazil, Korea, Taiwan, Spain, Italy...")
	fmt.Println("🥇 Plus Essential Commodities (Gold, Silver, Oil, etc.)")
	fmt.Println("📊 RANKING: All assets ranked by USD market cap → TOP 500 selected")
	fmt.Println("💵 Market caps converted to USD for ranking (prices kept in original currency)")
	fmt.Println("⚠️  Excluding: Indian stocks (as requested), ETFs, Index Funds")
	fmt.Println()

	startTime := time.Now()
	var allAssets []AssetData

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Remove the old GetGlobalStocks() to avoid duplicates
	// Using only GetMajorGlobalStocks() which is more comprehensive

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		majorGlobalStocks, err := client.GetMajorGlobalStocks()
		if err != nil {
			fmt.Printf("❌ Failed to fetch major global stocks: %v\n", err)
			return
		}
		mu.Lock()
		allAssets = append(allAssets, majorGlobalStocks...)
		mu.Unlock()
	}()

	wg.Wait()

	if len(allAssets) == 0 {
		log.Fatal("❌ No assets fetched successfully!")
	}

	// Separate stocks from commodities and count by country
	var stocks []AssetData
	var commodities []AssetData
	countryCounts := make(map[string]int)

	for _, asset := range allAssets {
		if asset.AssetType == "commodity" {
			commodities = append(commodities, asset)
		} else {
			stocks = append(stocks, asset)
			countryCounts[asset.Country]++
		}
	}

	fmt.Printf("\n📊 Retrieved %d stocks from %d countries and %d commodities\n", len(stocks), len(countryCounts), len(commodities))

	// Sort ALL stocks by market cap (no limit yet)
	sort.Slice(stocks, func(i, j int) bool {
		return stocks[i].MarketCap > stocks[j].MarketCap
	})

	// Combine ALL stocks with commodities first
	allAssets = append(stocks, commodities...)

	// Sort ALL assets by market cap to get true top 500
	sort.Slice(allAssets, func(i, j int) bool {
		return allAssets[i].MarketCap > allAssets[j].MarketCap
	})

	// NOW limit to top 500 by market cap across all asset types
	if len(allAssets) > 500 {
		allAssets = allAssets[:500]
		fmt.Printf("✂️  Selected top 500 assets by market cap from %d total assets\n", len(stocks)+len(commodities))
	}

	// Recount final asset types
	finalStocks := 0
	finalCommodities := 0
	finalCountryCounts := make(map[string]int)
	for _, asset := range allAssets {
		if asset.AssetType == "commodity" {
			finalCommodities++
		} else {
			finalStocks++
			finalCountryCounts[asset.Country]++
		}
	}

	fmt.Printf("🔗 Final top 500 dataset: %d stocks from %d countries + %d commodities = %d total assets\n",
		finalStocks, len(finalCountryCounts), finalCommodities, len(allAssets))

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
	fmt.Printf("🌍 Comprehensive global coverage using batch market cap endpoint!\n")
}
