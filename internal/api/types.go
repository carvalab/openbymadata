package api

import "time"

// Security represents a stock or equity security
type Security struct {
	Symbol        string    `json:"symbol"`
	Settlement    string    `json:"settlement"`
	BidSize       int64     `json:"bid_size"`
	Bid           float64   `json:"bid"`
	Ask           float64   `json:"ask"`
	AskSize       int64     `json:"ask_size"`
	Last          float64   `json:"last"`
	Close         float64   `json:"close"`
	Change        float64   `json:"change"`
	Open          float64   `json:"open"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	PreviousClose float64   `json:"previous_close"`
	Turnover      float64   `json:"turnover"`
	Volume        int64     `json:"volume"`
	Operations    int64     `json:"operations"`
	DateTime      time.Time `json:"datetime"`
	Group         string    `json:"group"`
}

// Bond represents a fixed income security
type Bond struct {
	Symbol        string    `json:"symbol"`
	Settlement    string    `json:"settlement"`
	BidSize       int64     `json:"bid_size"`
	Bid           float64   `json:"bid"`
	Ask           float64   `json:"ask"`
	AskSize       int64     `json:"ask_size"`
	Last          float64   `json:"last"`
	Close         float64   `json:"close"`
	Change        float64   `json:"change"`
	Open          float64   `json:"open"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	PreviousClose float64   `json:"previous_close"`
	Turnover      float64   `json:"turnover"`
	Volume        int64     `json:"volume"`
	Operations    int64     `json:"operations"`
	DateTime      time.Time `json:"datetime"`
	Group         string    `json:"group"`
	Expiration    time.Time `json:"expiration"`
}

// Option represents an options contract
type Option struct {
	Symbol          string    `json:"symbol"`
	BidSize         int64     `json:"bid_size"`
	Bid             float64   `json:"bid"`
	Ask             float64   `json:"ask"`
	AskSize         int64     `json:"ask_size"`
	Last            float64   `json:"last"`
	Close           float64   `json:"close"`
	Change          float64   `json:"change"`
	Open            float64   `json:"open"`
	High            float64   `json:"high"`
	Low             float64   `json:"low"`
	PreviousClose   float64   `json:"previous_close"`
	Turnover        float64   `json:"turnover"`
	Volume          int64     `json:"volume"`
	Operations      int64     `json:"operations"`
	DateTime        time.Time `json:"datetime"`
	UnderlyingAsset string    `json:"underlying_asset"`
	Expiration      time.Time `json:"expiration"`
}

// Future represents a futures contract
type Future struct {
	Symbol        string    `json:"symbol"`
	BidSize       int64     `json:"bid_size"`
	Bid           float64   `json:"bid"`
	Ask           float64   `json:"ask"`
	AskSize       int64     `json:"ask_size"`
	Last          float64   `json:"last"`
	Close         float64   `json:"close"`
	Change        float64   `json:"change"`
	Open          float64   `json:"open"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	PreviousClose float64   `json:"previous_close"`
	Turnover      float64   `json:"turnover"`
	Volume        int64     `json:"volume"`
	Operations    int64     `json:"operations"`
	DateTime      time.Time `json:"datetime"`
	Expiration    time.Time `json:"expiration"`
	OpenInterest  int64     `json:"open_interest"`
}

// Index represents a market index
type Index struct {
	Description   string  `json:"description"`
	Symbol        string  `json:"symbol"`
	Last          float64 `json:"last"`
	Change        float64 `json:"change"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	PreviousClose float64 `json:"previous_close"`
}

// MarketSummary represents market summary data
type MarketSummary struct {
	Symbol          string  `json:"symbol"`
	AssetType       string  `json:"assetType"`
	ParentKey       string  `json:"parentKey"`
	TotalNegotiated float64 `json:"totalNegotiated"`
	Volume          int64   `json:"volume"`
	Operations      int64   `json:"operations"`
}

// News represents market news
type News struct {
	Fecha       time.Time `json:"fecha"`
	Titulo      string    `json:"titulo"`
	Descripcion string    `json:"descripcion"`
	Descarga    string    `json:"descarga"`
}

// IncomeStatement represents financial statement data
type IncomeStatement struct {
	Symbol          string `json:"symbol"`
	Periodo         string `json:"periodo"`
	TipoPeriodo     string `json:"tipoPeriodo"`
	FechaCierre     string `json:"fechaCierre"`
	BalancesArchivo string `json:"balancesArchivo"`
}

// MarketTimeResponse represents the market time API response
type MarketTimeResponse struct {
	IsWorkingDay       bool   `json:"isWorkingDay"`
	MarketOpeningTime  string `json:"marketOpeningTime"`
	MarketClosingTime  string `json:"marketClosingTime"`
	Timezone           string `json:"timezone"`
}

// MarketTime exposes the parsed market-time response, with the opening
// and closing instants derived from the textual fields when present.
type MarketTime struct {
	IsWorkingDay bool
	Opens        time.Time
	Closes       time.Time
	Timezone     string
	OpeningText  string
	ClosingText  string
}

// CurrentQuote represents a per-symbol quote returned by the
// /bnown/fichatecnica/especies/cotizacion endpoint.
type CurrentQuote struct {
	Symbol     string  `json:"symbol"`
	Settlement string  `json:"settlementType"`
	BidSize    int64   `json:"quantityBid"`
	Bid        float64 `json:"bidPrice"`
	AskSize    int64   `json:"quantityOffer"`
	Ask        float64 `json:"offerPrice"`
	Last       float64 `json:"settlementPrice"`
	Open       float64 `json:"openingPrice"`
	High       float64 `json:"tradingHighPrice"`
	Low        float64 `json:"tradingLowPrice"`
	Close      float64 `json:"closingPrice"`
	Previous   float64 `json:"previousClosingPrice"`
	Volume     int64   `json:"volume"`
	Turnover   float64 `json:"volumeAmount"`
	Operations int64   `json:"numberOfOrders"`
	TradeHour  string  `json:"tradeHour"`
}

// CompanyInfo is the public profile of a listed society.
type CompanyInfo struct {
	Symbol string `json:"symbol"`
	Fields map[string]string
}

// EquityProfile is the technical profile of an instrument/species.
type EquityProfile struct {
	Symbol string `json:"symbol"`
	Fields map[string]string
}

// CompanyDirector represents a row in the company management endpoint.
type CompanyDirector struct {
	Name     string `json:"nombre"`
	Position string `json:"cargo"`
	Since    string `json:"desde"`
	Until    string `json:"hasta"`
	Type     string `json:"tipo"`
}

// CompanyBalance represents a single balance-sheet account row.
type CompanyBalance struct {
	Account string
	Values  map[string]string
}

// HistoryResponse represents the response structure for historical data
type HistoryResponse struct {
	Status string    `json:"s"` // Status: "ok" or "no_data"
	Time   []int64   `json:"t"` // Array of timestamps
	Close  []float64 `json:"c"` // Array of closing prices
	Open   []float64 `json:"o"` // Array of opening prices
	High   []float64 `json:"h"` // Array of high prices
	Low    []float64 `json:"l"` // Array of low prices
	Volume []int64   `json:"v"` // Array of volumes
}
