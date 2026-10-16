package openbymadata_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/carvalab/openbymadata"
)

// newExampleServer stands in for the BYMA API so the examples run offline.
// Drop BaseURL from ClientOptions to talk to the real API.
func newExampleServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "leading-equity"):
			fmt.Fprint(w, `{"data":[{"symbol":"GGAL","settlementType":"2","settlementPrice":5825,"imbalance":-1.19,"volume":203162}]}`)
		case strings.HasSuffix(r.URL.Path, "cotizacion"):
			fmt.Fprint(w, `{"data":[{"symbol":"GGAL","settlementType":"2","bidPrice":5825,"offerPrice":5835,"settlementPrice":5830,"previousClosingPrice":5895}]}`)
		case strings.HasSuffix(r.URL.Path, "market-time"):
			fmt.Fprint(w, `{"isWorkingDay":true,"marketOpeningTime":"10:30:00","marketClosingTime":"17:00:00","timezone":"America/Argentina/Buenos_Aires"}`)
		case strings.HasSuffix(r.URL.Path, "history"):
			fmt.Fprint(w, `{"s":"ok","t":[1760313600,1760400000],"o":[5800,5860],"h":[5900,5925],"l":[5790,5810],"c":[5895,5825],"v":[180000,203162]}`)
		default:
			http.NotFound(w, r)
		}
	}))
}

// ExampleClient_GetBluechips lists the Merval leading equities (panel líder).
func ExampleClient_GetBluechips() {
	server := newExampleServer()
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{BaseURL: server.URL})
	defer client.Close()

	bluechips, err := client.GetBluechips(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range bluechips {
		fmt.Printf("%s: $%.2f (%.2f%%)\n", s.Symbol, s.Last, s.Change)
	}

	// Output:
	// GGAL: $5825.00 (-1.19%)
}

// ExampleClient_GetCurrentQuote fetches the live bid/ask for a ticker at
// T+1 (24HS) settlement.
func ExampleClient_GetCurrentQuote() {
	server := newExampleServer()
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{BaseURL: server.URL})
	defer client.Close()

	quote, err := client.GetCurrentQuote(context.Background(), "GGAL", openbymadata.Settlement24HS)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s bid $%.2f / ask $%.2f, last $%.2f\n", quote.Symbol, quote.Bid, quote.Ask, quote.Last)

	// Output:
	// GGAL bid $5825.00 / ask $5835.00, last $5830.00
}

// ExampleClient_GetHistory gets daily OHLCV bars. History symbols need a
// settlement suffix such as "GGAL 24HS"; a bare "GGAL" returns no data.
func ExampleClient_GetHistory() {
	server := newExampleServer()
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{BaseURL: server.URL})
	defer client.Close()

	to := time.Now()
	from := to.AddDate(0, 0, -30)
	bars, err := client.GetHistory(context.Background(), "GGAL 24HS", openbymadata.ResolutionDaily, from, to)
	if err != nil {
		log.Fatal(err)
	}
	for i := range bars.Time {
		fmt.Printf("%s close $%.2f volume %d\n", bars.Time[i].UTC().Format("2006-01-02"), bars.Close[i], bars.Volume[i])
	}

	// Output:
	// 2025-10-13 close $5895.00 volume 180000
	// 2025-10-14 close $5825.00 volume 203162
}

// ExampleClient_GetMarketTime checks whether BYMA trades today and when.
func ExampleClient_GetMarketTime() {
	server := newExampleServer()
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{BaseURL: server.URL})
	defer client.Close()

	mt, err := client.GetMarketTime(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("working day: %v, %s to %s\n", mt.IsWorkingDay, mt.OpeningText, mt.ClosingText)

	// Output:
	// working day: true, 10:30:00 to 17:00:00
}
