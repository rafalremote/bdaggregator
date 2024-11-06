package currency

import (
	"bdaggregator/internal/currency/coingecko"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shopspring/decimal"
)

// Coin represents a single coin entry in coins.json
type Coin struct {
	ID        string            `json:"id"`
	Symbol    string            `json:"symbol"`
	Name      string            `json:"name"`
	Platforms map[string]string `json:"platforms"`
}

type PriceData struct {
	Prices [][]interface{} `json:"prices"`
}

// MapCurrencyToCoinID maps a currency symbol and address to a coin ID using the loaded coins data
func MapCurrencyToCoinID(coins []Coin, currencySymbol, currencyAddress, chainID string) (string, error) {

	// Search for a match by currencySymbol and confirm it has the same currencyAddress
	for _, coin := range coins {
		if strings.EqualFold(coin.Symbol, currencySymbol) {
			for _, address := range coin.Platforms {
				if strings.EqualFold(address, currencyAddress) {
					return coin.ID, nil
				}
			}
		}
	}

	if currencySymbol == "MATIC" && chainID == "137" {
		return "matic-network", nil
	}

	return "", fmt.Errorf("no matching coin found for symbol '%s' and address '%s'", currencySymbol, currencyAddress)
}

func LoadCoins(filePath string) ([]Coin, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var coins []Coin
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&coins); err != nil {
		return nil, err
	}

	return coins, nil
}

func FetchExchangeRates(coinID, targetCurrency, from, to string) (map[int64]decimal.Decimal, error) {
	jsonData, err := coingecko.ApiGet("coins/" + coinID + "/market_chart/range?vs_currency=" + targetCurrency + "&from=" + from + "&to=" + to + "&precision=full")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rates: %v", err)
	}

	log.Println("Fetched exchange rates for", coinID, "from", from, "to", to, "with target currency", targetCurrency)

	var priceData PriceData
	if err := json.Unmarshal([]byte(jsonData), &priceData); err != nil {
		log.Fatalf("failed to parse JSON: %v", err)
	}

	// Convert the exchange rates to a map
	exchangeRateMap := transformToExchangeRateMap(priceData.Prices)
	return exchangeRateMap, nil
}

func transformToExchangeRateMap(prices [][]interface{}) map[int64]decimal.Decimal {
	exchangeRateMap := make(map[int64]decimal.Decimal)

	for _, priceEntry := range prices {
		timestamp := int64(priceEntry[0].(float64))
		exchangeRate := decimal.NewFromFloat(priceEntry[1].(float64))

		exchangeRateMap[timestamp] = exchangeRate
	}

	return exchangeRateMap
}
