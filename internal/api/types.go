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
	IsWorkingDay bool `json:"isWorkingDay"`
}
