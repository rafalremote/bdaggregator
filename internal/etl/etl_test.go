package etl

import (
	"bdaggregator/internal/currency"
	"bytes"
	"encoding/csv"
	"io"
	"log"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestExtractEvents(t *testing.T) {
	// Sample CSV data
	csvData := `"app","ts","event","project_id","source","ident","user_id","session_id","country","device_type","device_os","device_os_ver","device_browser","device_browser_ver","props","nums"
"seq-market","2024-04-15 02:15:07.167","BUY_ITEMS","4974","","1","0896ae95dcaeee38e83fa5c43bef99780d7b2be23bcab36214","5d8afd8fec2fbf3e","DE","desktop","linux","x86_64","chrome","122.0.0.0","{""tokenId"":""215"",""txnHash"":""0xd919290e80df271e77d1cbca61f350d2727531e0334266671ec20d626b2104a2"",""chainId"":""137"",""collectionAddress"":""0x22d5f9b75c524fec1d6619787e582644cd4d7422"",""currencyAddress"":""0xd1f9c58e33933a993a3891f8acfe05a68e1afc05"",""currencySymbol"":""SFL"",""marketplaceType"":""amm"",""requestId"":""""}","{""currencyValueDecimal"":""0.6136203411678249"",""currencyValueRaw"":""613620341167824900""}"
"seq-market","2024-04-15 02:26:37.134","BUY_ITEMS","4974","","1","0896ae95dcaeee38e83fa5c43bef99780d7b2be23bcab36214","5d8afd8fec2fbf3e","DE","desktop","linux","x86_64","chrome","122.0.0.0","{""currencyAddress"":""0xd1f9c58e33933a993a3891f8acfe05a68e1afc05"",""currencySymbol"":""SFL"",""marketplaceType"":""amm"",""requestId"":"""",""tokenId"":""602"",""txnHash"":""0x1133d2837267e0de2eddf3655a3df99e055d172cb53c4e8e108e70322438e994"",""chainId"":""137"",""collectionAddress"":""0x22d5f9b75c524fec1d6619787e582644cd4d7422""}","{""currencyValueDecimal"":""2.361412166673735"",""currencyValueRaw"":""2361412166673735000""}"
`

	mockCoins := []currency.Coin{
		{
			ID:        "SFL",
			Symbol:    "SFL",
			Platforms: map[string]string{"137": "0xd1f9c58e33933a993a3891f8acfe05a68e1afc05"},
		},
	}

	// Initialize CSV reader
	reader := csv.NewReader(bytes.NewReader([]byte(csvData)))
	// Read header
	_, err := reader.Read()
	assert.NoError(t, err, "Expected no error reading CSV header")

	// Prepare channels and map for testing
	rowChan := make(chan []string, 100)
	eventChan := make(chan Event, 100)
	currencyUsageMap := make(CurrencyUsageMap)
	var mu sync.Mutex

	// Launch CSV reading in background
	go func() {
		for {
			row, err := reader.Read()
			if err == io.EOF {
				close(rowChan)
				break
			}
			if err != nil {
				log.Fatalf("Error reading CSV: %v", err)
			}
			rowChan <- row
		}
	}()

	// Start workers to process rows
	numWorkers := 2
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for row := range rowChan {
				event, err := ParseRowToEvent(row, currencyUsageMap, &mu, mockCoins)
				if err == nil {
					eventChan <- event
				} else {
					log.Printf("Worker %d: failed to parse row: %v", workerID, err)
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(eventChan)
	}()

	// Collect events
	var events []Event
	for event := range eventChan {
		events = append(events, event)
	}

	assert.Len(t, events, 2, "Expected 2 events")
	assert.Contains(t, currencyUsageMap, "SFL", "CurrencyUsageMap should contain 'SFL'")
	assert.Equal(t, "SFL", events[0].CurrencySymbol, "Expected CurrencySymbol to be 'SFL'")
	assert.Equal(t, 4974, events[0].ProjectID, "Expected ProjectID to be 4974")

	expectedCurrencyValue := decimal.RequireFromString("0.6136203411678249")
	assert.Equal(t, expectedCurrencyValue, events[0].CurrencyValueDecimal, "Expected parsed CurrencyValueDecimal")
}
