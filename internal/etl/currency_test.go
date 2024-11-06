package etl

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Mock function to replace currency.FetchExchangeRates for testing
func mockFetchExchangeRates(coinID, targetCurrency, from, to string) (map[int64]decimal.Decimal, error) {
	switch coinID {
	case "bitcoin":
		return map[int64]decimal.Decimal{
			1609459200: decimal.NewFromFloat(30000.00),
			1609545600: decimal.NewFromFloat(31000.00),
		}, nil
	case "ethereum":
		return map[int64]decimal.Decimal{
			1609459200: decimal.NewFromFloat(1000.00),
			1609545600: decimal.NewFromFloat(1100.00),
		}, nil
	default:
		return nil, fmt.Errorf("no rates for coin: %s", coinID)
	}
}

func TestGetExchangeRates(t *testing.T) {
	currencyUsageMap := CurrencyUsageMap{
		"bitcoin":  {From: 1609459200, To: 1609545600},
		"ethereum": {From: 1609459200, To: 1609545600},
	}

	// Call GetExchangeRates with the mock fetch function
	exchangeRates, err := GetExchangeRates(currencyUsageMap, "usd", mockFetchExchangeRates)

	assert.NoError(t, err)
	assert.NotNil(t, exchangeRates)
	assert.Equal(t, decimal.NewFromFloat(30000.00), exchangeRates["bitcoin"][1609459200])
	assert.Equal(t, decimal.NewFromFloat(31000.00), exchangeRates["bitcoin"][1609545600])
	assert.Equal(t, decimal.NewFromFloat(1000.00), exchangeRates["ethereum"][1609459200])
	assert.Equal(t, decimal.NewFromFloat(1100.00), exchangeRates["ethereum"][1609545600])
}
