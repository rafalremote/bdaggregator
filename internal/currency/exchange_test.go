package currency

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

// Test for MapCurrencyToCoinID
func TestMapCurrencyToCoinID(t *testing.T) {
	coins := []Coin{
		{
			ID:     "bitcoin",
			Symbol: "BTC",
			Name:   "Bitcoin",
			Platforms: map[string]string{
				"platform1": "address1",
			},
		},
		{
			ID:     "matic-network",
			Symbol: "MATIC",
			Name:   "Polygon",
			Platforms: map[string]string{
				"platform2": "address2",
			},
		},
	}

	tests := []struct {
		currencySymbol   string
		currencyAddress  string
		chainID          string
		expectedCoinID   string
		expectedErrorMsg string
	}{
		{"BTC", "address1", "", "bitcoin", ""},
		{"MATIC", "address2", "137", "matic-network", ""},
		{"ETH", "address3", "", "", "no matching coin found for symbol 'ETH' and address 'address3'"},
	}

	for _, test := range tests {
		coinID, err := MapCurrencyToCoinID(coins, test.currencySymbol, test.currencyAddress, test.chainID)
		if test.expectedErrorMsg == "" && err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if test.expectedErrorMsg != "" {
			if err == nil || !strings.Contains(err.Error(), test.expectedErrorMsg) {
				t.Errorf("expected error containing: %v, got: %v", test.expectedErrorMsg, err)
			}
		} else if coinID != test.expectedCoinID {
			t.Errorf("expected coin ID: %s, got: %s", test.expectedCoinID, coinID)
		}
	}
}

// Test for LoadCoins
func TestLoadCoins(t *testing.T) {
	fileContent := `[{"id": "bitcoin", "symbol": "BTC", "name": "Bitcoin", "platforms": {"platform1": "address1"}}]`
	filePath := "test_coins.json"
	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to create mock coins file: %v", err)
	}
	defer os.Remove(filePath)

	coins, err := LoadCoins(filePath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(coins) != 1 || coins[0].ID != "bitcoin" || coins[0].Symbol != "BTC" {
		t.Errorf("unexpected coin data: %+v", coins[0])
	}
}

// Test for FetchExchangeRates
func TestFetchExchangeRates(t *testing.T) {
	mockResponse := `{"prices": [[1609459200000, 29000.0], [1609545600000, 29500.0]]}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	os.Setenv("SEQUENCE_COINGECKO_API_URL", mockServer.URL+"/")

	coinID := "bitcoin"
	targetCurrency := "usd"
	from := "1609459200"
	to := "1609545600"

	exchangeRates, err := FetchExchangeRates(coinID, targetCurrency, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedRates := map[int64]decimal.Decimal{
		1609459200000: decimal.NewFromFloat(29000.0),
		1609545600000: decimal.NewFromFloat(29500.0),
	}

	for timestamp, expectedRate := range expectedRates {
		if rate, exists := exchangeRates[timestamp]; !exists || !rate.Equal(expectedRate) {
			t.Errorf("expected exchange rate for timestamp %d: %s, got: %s", timestamp, expectedRate, rate)
		}
	}
}
