package helpers

import (
	"fmt"

	"github.com/carvalab/openbymadata/internal/api"
)

// FindSecurityBySymbol searches for a security in multiple collections
func FindSecurityBySymbol(symbol string, bluechips, cedears, galpones []api.Security) (*api.Security, error) {
	// Search in blue chips first
	for i := range bluechips {
		if bluechips[i].Symbol == symbol {
			return &bluechips[i], nil
		}
	}

	// Search in CEDEARs
	for i := range cedears {
		if cedears[i].Symbol == symbol {
			return &cedears[i], nil
		}
	}

	// Search in galpones
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
	// Search in regular bonds
	for i := range bonds {
		if bonds[i].Symbol == symbol {
			return &bonds[i], nil
		}
	}

	// Search in short-term bonds
	for i := range shortBonds {
		if shortBonds[i].Symbol == symbol {
			return &shortBonds[i], nil
		}
	}

	// Search in corporate bonds
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

	// Create lookup maps for efficient searching
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

	// Find each requested symbol
	for _, symbol := range symbols {
		if security, exists := bluechipMap[symbol]; exists {
			results[symbol] = security
		} else if security, exists := cedearMap[symbol]; exists {
			results[symbol] = security
		} else if security, exists := galponeMap[symbol]; exists {
			results[symbol] = security
		}
		// If not found, it's simply not included in results
	}

	return results
}

// SearchSecurities searches for securities containing the given text
func SearchSecurities(searchText string, bluechips, cedears, galpones []api.Security) []api.Security {
	var results []api.Security

	// Search in blue chips
	for _, security := range bluechips {
		if contains(security.Symbol, searchText) {
			results = append(results, security)
		}
	}

	// Search in CEDEARs
	for _, security := range cedears {
		if contains(security.Symbol, searchText) {
			results = append(results, security)
		}
	}

	// Search in galpones
	for _, security := range galpones {
		if contains(security.Symbol, searchText) {
			results = append(results, security)
		}
	}

	return results
}

// contains checks if a string contains another string (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			stringContains(s, substr))
}

// stringContains performs case-insensitive substring search
func stringContains(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return simpleContains(sLower, substrLower)
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + ('a' - 'A')
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// simpleContains checks if s contains substr
func simpleContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
