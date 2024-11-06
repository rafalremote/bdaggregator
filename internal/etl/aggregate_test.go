package etl

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAggregateEvents(t *testing.T) {
	defaultCurrency := "USD"
	events := []Event{
		{
			Ts:                   time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
			ProjectID:            1,
			CoinID:               "matic-network",
			CurrencyExchangeRate: decimal.NewFromFloat(1.5),
			CurrencyValueDecimal: decimal.NewFromInt(1000000000000000000), // 1 MATIC
		},
		{
			Ts:                   time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
			ProjectID:            1,
			CoinID:               "other-coin",
			CurrencyExchangeRate: decimal.NewFromFloat(2.0),
			CurrencyValueDecimal: decimal.NewFromFloat(100.0),
		},
		{
			Ts:                   time.Date(2024, 4, 16, 0, 0, 0, 0, time.UTC),
			ProjectID:            2,
			CoinID:               "other-coin",
			CurrencyExchangeRate: decimal.NewFromFloat(1.0),
			CurrencyValueDecimal: decimal.NewFromFloat(50.0),
		},
	}

	expected := []AggregatePerProject{
		{
			Day:                            "2024-04-15",
			ProjectID:                      1,
			NumberOfTransactionsPerProject: 2,
			TotalVolumePerProject:          201.5,
			Currency:                       defaultCurrency,
		},
		{
			Day:                            "2024-04-16",
			ProjectID:                      2,
			NumberOfTransactionsPerProject: 1,
			TotalVolumePerProject:          50.0,
			Currency:                       defaultCurrency,
		},
	}

	aggregated := AggregateEvents(events, defaultCurrency)

	assert.ElementsMatch(t, expected, aggregated)
}

func TestCalculateVolume(t *testing.T) {
	event := Event{
		CoinID:               "matic-network",
		CurrencyExchangeRate: decimal.NewFromFloat(1.5),
		CurrencyValueDecimal: decimal.NewFromInt(1000000000000000000), // 1 MATIC
	}
	volume := calculateVolume(event)
	assert.Equal(t, 1.5, volume)
}
