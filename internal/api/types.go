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
	Currency      string    `json:"currency"` // as sent by BYMA: "ARS", "USD", or "EXT" (seen on cable-dollar instruments)
	Market        string    `json:"market"`   // "BYMA", or "SENEBI" for the non-guaranteed segment
}

// Bond represents a fixed income security
type Bond struct {
	Symbol         string    `json:"symbol"`
	Settlement     string    `json:"settlement"`
	BidSize        int64     `json:"bid_size"`
	Bid            float64   `json:"bid"`
	Ask            float64   `json:"ask"`
	AskSize        int64     `json:"ask_size"`
	Last           float64   `json:"last"`
	Close          float64   `json:"close"`
	Change         float64   `json:"change"`
	Open           float64   `json:"open"`
	High           float64   `json:"high"`
	Low            float64   `json:"low"`
	PreviousClose  float64   `json:"previous_close"`
	Turnover       float64   `json:"turnover"`
	Volume         int64     `json:"volume"`
	Operations     int64     `json:"operations"`
	DateTime       time.Time `json:"datetime"`
	Group          string    `json:"group"`
	Expiration     time.Time `json:"expiration"`
	DaysToMaturity int64     `json:"days_to_maturity"`
	Underlying     string    `json:"underlying"` // repos ("PESOS", "DOLAR", "CABLE") and securities lending (the lent ticker)
	Currency       string    `json:"currency"`   // as sent by BYMA: "ARS", "USD", or "EXT" (seen on cable-dollar instruments)
	Market         string    `json:"market"`     // "BYMA", or "SENEBI" for the non-guaranteed segment
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
	OptionType      string    `json:"option_type"` // "CALL" or "PUT"
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

// IncomeStatement is one filed financial statement. FechaCierre is the
// period end date and FechaPresentacion the filing date (both YYYY-MM-DD);
// BalancesArchivo links to the PDF. BYMA no longer sends Periodo or
// TipoPeriodo, so those are usually empty.
type IncomeStatement struct {
	Symbol            string `json:"symbol"`
	Periodo           string `json:"periodo"`
	TipoPeriodo       string `json:"tipoPeriodo"`
	FechaCierre       string `json:"fechaCierre"`
	FechaPresentacion string `json:"fechaPresentacion"`
	BalancesArchivo   string `json:"balancesArchivo"`
}

// MarketTimeResponse represents the market time API response
type MarketTimeResponse struct {
	IsWorkingDay      bool   `json:"isWorkingDay"`
	MarketOpeningTime string `json:"marketOpeningTime"`
	MarketClosingTime string `json:"marketClosingTime"`
	Timezone          string `json:"timezone"`
}

// MarketTime exposes the parsed market-time response. Opens and Closes are
// today's opening and closing instants in the market's timezone (UTC-3).
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
	Status string    `json:"s"`      // Status: "ok", "no_data" or "error"
	ErrMsg string    `json:"errmsg"` // set when Status is "error"
	Time   []int64   `json:"t"`      // Array of timestamps
	Close  []float64 `json:"c"`      // Array of closing prices
	Open   []float64 `json:"o"`      // Array of opening prices
	High   []float64 `json:"h"`      // Array of high prices
	Low    []float64 `json:"l"`      // Array of low prices
	Volume []int64   `json:"v"`      // Array of volumes
}

// BondIndex is a BYMA fixed-income index, such as short-term dollar bonds.
type BondIndex struct {
	Symbol      string    `json:"symbol"`
	Description string    `json:"description"`
	Currency    string    `json:"currency"`
	Value       float64   `json:"value"`
	Varianza    float64   `json:"varianza"` // as sent by BYMA, unscaled
	Date        time.Time `json:"date"`
}

// BondCoupon is one interest and/or amortization payment of a bond.
// Percentages are as published by BYMA (64 means 64%).
type BondCoupon struct {
	Symbol                    string    `json:"symbol"`
	NoticeDate                time.Time `json:"notice_date"`
	InterestPaymentDate       time.Time `json:"interest_payment_date"`
	AmortizationPaymentDate   time.Time `json:"amortization_payment_date"`
	InterestServiceNumber     int64     `json:"interest_service_number"`
	AmortizationServiceNumber int64     `json:"amortization_service_number"`
	AmortizationPercent       float64   `json:"amortization_percent"`
	ResidualValuePercent      float64   `json:"residual_value_percent"`
	NoticeURL                 string    `json:"notice_url"`
}

// TopMovers lists stocks, government bonds and options ordered by BYMA's
// "main ups and downs" panel.
type TopMovers struct {
	Stocks  []Security `json:"stocks"`
	Bonds   []Bond     `json:"bonds"`
	Options []Option   `json:"options"`
}

// MarketResponsible is a company's designated market representative.
type MarketResponsible struct {
	Name            string    `json:"name"`
	Position        string    `json:"position"`
	CompanyPosition string    `json:"company_position"`
	Since           time.Time `json:"since"`
}

// CatalogInstrument is one entry of BYMA's full instrument list.
type CatalogInstrument struct {
	ID      int64  `json:"id"`
	Code    string `json:"code"` // ticker, e.g. "GGAL" or "AL30D.SB"
	Kind    string `json:"kind"` // "BONDINSTRUMENT" or "EQUITYINSTRUMENT"
	Deleted bool   `json:"deleted"`
}

// InstrumentMatch is a symbol search result.
type InstrumentMatch struct {
	Symbol       string `json:"symbol"`
	Description  string `json:"description"`
	Type         string `json:"type"`       // e.g. "Accion", "Bono", "Cedear"
	Market       string `json:"market"`     // e.g. "BYMA"
	Settlement   string `json:"settlement"` // e.g. "24hs"
	Presentation string `json:"presentation"`
}
