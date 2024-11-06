package etl

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/shopspring/decimal"
)

// Holds exchange rates mapped by coin ID and timestamp
type FetchedExchangeRates map[string]map[int64]decimal.Decimal

// Stores information about each currency symbol, the earliest and latest Unix timestamps
type CurrencyUsage struct {
	From int64
	To   int64
}

type CurrencyUsageMap map[string]CurrencyUsage

// Update the usage range for a currency
func UpdateCurrencyUsageMap(currencyUsageMap CurrencyUsageMap, mu *sync.Mutex, coinID string, tsUnix int64) {
	mu.Lock()
	defer mu.Unlock()

	usage, exists := currencyUsageMap[coinID]
	if !exists {
		usage = CurrencyUsage{From: tsUnix, To: tsUnix}
	} else {
		if tsUnix < usage.From {
			usage.From = tsUnix
		}
		if tsUnix > usage.To {
			usage.To = tsUnix
		}
	}

	currencyUsageMap[coinID] = usage
}

// Retrieve exchange rates for each currency in the usage map, using a provided fetch function.
func GetExchangeRates(currencyUsageMap CurrencyUsageMap, targetCurrency string, fetchFunc func(string, string, string, string) (map[int64]decimal.Decimal, error)) (FetchedExchangeRates, error) {
	allExchangeRates := make(FetchedExchangeRates)
	var wg sync.WaitGroup
	mu := &sync.Mutex{}
	errChan := make(chan error, len(currencyUsageMap))

	for coinID, timeRange := range currencyUsageMap {
		wg.Add(1)
		go func(coinID string, timeRange CurrencyUsage) {
			defer wg.Done()
			from := strconv.FormatInt(timeRange.From, 10)
			to := strconv.FormatInt(timeRange.To, 10)

			exchangeRates, err := fetchFunc(coinID, targetCurrency, from, to)
			if err != nil {
				errChan <- fmt.Errorf("failed to get exchange rates for %s: %w", coinID, err)
				return
			}

			mu.Lock()
			allExchangeRates[coinID] = exchangeRates
			mu.Unlock()
		}(coinID, timeRange)
	}

	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return allExchangeRates, nil
}

// Update events with the closest exchange rate
func UpdateExchangeRates(events []Event, exchangeRates FetchedExchangeRates) {
	var wg sync.WaitGroup
	for i := range events {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			closestRate := findClosestExchangeRate(events[i].CoinID, events[i].TsUnix, exchangeRates)
			events[i].CurrencyExchangeRate = closestRate
		}(i)
	}
	wg.Wait()
}

// Find the closest exchange rate in time for a given coin ID and timestamp
func findClosestExchangeRate(coinID string, tsUnix int64, exchangeRates FetchedExchangeRates) decimal.Decimal {
	rates, exists := exchangeRates[coinID]
	if !exists {
		log.Println("No exchange rates found for coin ID:", coinID)
		return decimal.Zero
	}

	var closestRate decimal.Decimal
	var minDiff int64 = 1<<63 - 1 // max int64 value

	for rateTs, rate := range rates {
		diff := abs(tsUnix - rateTs)
		if diff < minDiff {
			minDiff = diff
			closestRate = rate
		}
	}
	return closestRate
}

func PrintCurrencyUsage(currencyUsageMap CurrencyUsageMap) {
	for coinID, usage := range currencyUsageMap {
		log.Printf("Currency: %s, From: %d, To: %d", coinID, usage.From, usage.To)
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
