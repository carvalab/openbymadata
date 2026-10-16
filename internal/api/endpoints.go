package api

import "strings"

// Path segments for BYMA Open Data endpoints, excluding the host and the
// shared /vanoms-be-core/rest/api/bymadata/free/ prefix that the client adds.
const (
	pathMarketTime      = "market-time"
	pathIndexPrice      = "index-price"
	pathLeadingEquity   = "leading-equity"
	pathGeneralEquity   = "general-equity"
	pathCedears         = "cedears"
	pathTotalNegotiated = "total-negotiated"
	pathPublicBonds     = "public-bonds"
	pathLetters         = "lebacs"
	pathCorpBonds       = "negociable-obligations"
	pathOptions         = "options"
	pathIndexFuture     = "index-future"
	pathCurrentQuote    = "bnown/fichatecnica/especies/cotizacion"
	pathCompanyGeneral  = "bnown/fichatecnica/sociedades/general"
	pathCompanyProfile  = "bnown/fichatecnica/especies/general"
	pathCompanyMgmt     = "bnown/fichatecnica/sociedades/administracion"
	pathCompanyBalance  = "bnown/bolsar/balances"
	pathHistorical      = "chart/historical-series/history"
	pathNews            = "bnown/byma-ads"
	pathBalances        = "bnown/seriesHistoricas/balances"
	pathETF             = "etf"
	pathMEP             = "getBymaIndexsMep"
	pathTopMovers       = "main-ups-downs"
	pathBondIndices     = "index-bonds"
	pathGreenBonds      = "green-bond"
	pathSocialBonds     = "social-bond"
	pathSustainBonds    = "sustainability-bond"
	pathLinkedBonds     = "linked-bond"
	pathSenebiBonds     = "senebi-bonds"
	pathSenebiLetters   = "senebi-letras"
	pathSenebiCorpBonds = "senebi-obligaciones-negociables"
	pathCoupons         = "bnown/seriesHistoricas/coupons"
	pathResponsibles    = "bnown/fichatecnica/sociedades/responsables"
	pathIndexHistory    = "chart/index-historical-series/history"
	pathSearch          = "instruments-for-header-search"
)

// sectionHeader is the value BYMA expects in the "Options" request header
// for each endpoint, mirroring the open.bymadata.com.ar web app. Requests
// without it may be rejected with HTTP 400.
var sectionHeader = map[string]string{
	pathMarketTime:      "dashboard",
	pathTotalNegotiated: "dashboard",
	pathIndexPrice:      "renta-variable",
	pathLeadingEquity:   "renta-variable",
	pathGeneralEquity:   "renta-variable",
	pathCedears:         "renta-variable",
	pathHistorical:      "renta-variable",
	pathPublicBonds:     "renta-fija",
	pathLetters:         "renta-fija",
	pathCorpBonds:       "renta-fija",
	pathOptions:         "derivados",
	pathIndexFuture:     "derivados",
	pathNews:            "notices-and-events",
	pathCurrentQuote:    "technical-details",
	pathCompanyGeneral:  "technical-details",
	pathCompanyProfile:  "technical-details",
	pathCompanyMgmt:     "technical-details",
	pathCompanyBalance:  "technical-details",
	pathBalances:        "technical-details",
	pathETF:             "renta-variable",
	pathMEP:             "dashboard",
	pathTopMovers:       "dashboard",
	pathSearch:          "dashboard",
	pathIndexHistory:    "renta-variable",
	pathBondIndices:     "renta-fija",
	pathGreenBonds:      "renta-fija",
	pathSocialBonds:     "renta-fija",
	pathSustainBonds:    "renta-fija",
	pathLinkedBonds:     "renta-fija",
	pathSenebiBonds:     "renta-fija",
	pathSenebiLetters:   "renta-fija",
	pathSenebiCorpBonds: "renta-fija",
	pathCoupons:         "technical-details",
	pathResponsibles:    "technical-details",
}

// Settlement identifies the settlement window used by BYMA quote endpoints.
// The numeric codes are what the API expects; the textual forms are accepted
// on input for ergonomics and serialized to the canonical numeric code.
type Settlement string

const (
	SettlementCI   Settlement = "1"
	Settlement24HS Settlement = "2"
	Settlement48HS Settlement = "3"
)

// ParseSettlement accepts either the numeric code ("1"/"2"/"3") or the
// textual label ("CI"/"24HS"/"48HS") and returns the canonical Settlement.
func ParseSettlement(s string) (Settlement, bool) {
	switch Settlement(strings.ToUpper(s)) {
	case SettlementCI, "CI":
		return SettlementCI, true
	case Settlement24HS, "24HS":
		return Settlement24HS, true
	case Settlement48HS, "48HS":
		return Settlement48HS, true
	}
	return "", false
}

// Resolution is the timeframe for historical data points.
type Resolution string

const (
	Resolution1Min    Resolution = "1"
	Resolution5Min    Resolution = "5"
	Resolution60Min   Resolution = "60"
	ResolutionDaily   Resolution = "D"
	ResolutionWeekly  Resolution = "W"
	ResolutionMonthly Resolution = "M"
)
