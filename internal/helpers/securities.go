package helpers

import (
	"fmt"
	"strings"

	"github.com/carvalab/openbymadata/internal/api"
)

// FindSecurityBySymbol searches for a security in multiple collections
func FindSecurityBySymbol(symbol string, bluechips, cedears, galpones []api.Security) (*api.Security, error) {
	for i := range bluechips {
		if bluechips[i].Symbol == symbol {
			return &bluechips[i], nil
		}
	}

	for i := range cedears {
		if cedears[i].Symbol == symbol {
			return &cedears[i], nil
		}
	}

	for i := range galpones {
		if galpones[i].Symbol == symbol {
			return &galpones[i], nil
		}
	}

	return nil, fmt.Errorf("security %s not found", symbol)
}

// FindSecurityInCollection searches for a security in a specific collection
func FindSecurityInCollection(symbol string, securities []api.Security) (*api.Security, error) {
	for i := range securities {
		if securities[i].Symbol == symbol {
			return &securities[i], nil
		}
	}
	return nil, fmt.Errorf("security %s not found", symbol)
}

// FindBondBySymbol searches for a bond in multiple bond collections
func FindBondBySymbol(symbol string, bonds, shortBonds, corporateBonds []api.Bond) (*api.Bond, error) {
	for i := range bonds {
		if bonds[i].Symbol == symbol {
			return &bonds[i], nil
		}
	}

	for i := range shortBonds {
		if shortBonds[i].Symbol == symbol {
			return &shortBonds[i], nil
		}
	}

	for i := range corporateBonds {
		if corporateBonds[i].Symbol == symbol {
			return &corporateBonds[i], nil
		}
	}

	return nil, fmt.Errorf("bond %s not found", symbol)
}

// FindOptionBySymbol searches for an option by symbol
func FindOptionBySymbol(symbol string, options []api.Option) (*api.Option, error) {
	for i := range options {
		if options[i].Symbol == symbol {
			return &options[i], nil
		}
	}
	return nil, fmt.Errorf("option %s not found", symbol)
}

// FindFutureBySymbol searches for a future by symbol
func FindFutureBySymbol(symbol string, futures []api.Future) (*api.Future, error) {
	for i := range futures {
		if futures[i].Symbol == symbol {
			return &futures[i], nil
		}
	}
	return nil, fmt.Errorf("future %s not found", symbol)
}

// GetMultipleSecurities creates a lookup map for multiple securities
func GetMultipleSecurities(symbols []string, bluechips, cedears, galpones []api.Security) map[string]*api.Security {
	results := make(map[string]*api.Security)

	bluechipMap := make(map[string]*api.Security)
	for i := range bluechips {
		bluechipMap[bluechips[i].Symbol] = &bluechips[i]
	}

	cedearMap := make(map[string]*api.Security)
	for i := range cedears {
		cedearMap[cedears[i].Symbol] = &cedears[i]
	}

	galponeMap := make(map[string]*api.Security)
	for i := range galpones {
		galponeMap[galpones[i].Symbol] = &galpones[i]
	}

	for _, symbol := range symbols {
		if security, exists := bluechipMap[symbol]; exists {
			results[symbol] = security
		} else if security, exists := cedearMap[symbol]; exists {
			results[symbol] = security
		} else if security, exists := galponeMap[symbol]; exists {
			results[symbol] = security
		}
	}

	return results
}

// SearchSecurities searches for securities containing the given text
func SearchSecurities(searchText string, bluechips, cedears, galpones []api.Security) []api.Security {
	needle := strings.ToLower(searchText)
	var results []api.Security

	for _, security := range bluechips {
		if strings.Contains(strings.ToLower(security.Symbol), needle) {
			results = append(results, security)
		}
	}

	for _, security := range cedears {
		if strings.Contains(strings.ToLower(security.Symbol), needle) {
			results = append(results, security)
		}
	}

	for _, security := range galpones {
		if strings.Contains(strings.ToLower(security.Symbol), needle) {
			results = append(results, security)
		}
	}

	return results
}
