package etl

import (
	"bdaggregator/internal/currency"
	"io"
	"log"
	"sync"
)

// Process the CSV and extracts events and currency usage information
func ExtractEvents(reader io.Reader, coins []currency.Coin) ([]Event, CurrencyUsageMap) {
	rowChan := make(chan []string, 100)
	eventChan := make(chan Event, 100)

	var events []Event
	currencyUsageMap := make(CurrencyUsageMap)
	var mu sync.Mutex

	var wg sync.WaitGroup

	// Read CSV rows and send them to rowChan
	go ReadCSVRows(reader, rowChan)

	// Start workers to process rows, build CurrencyUsageMap, and send events
	numWorkers := 4
	startWorkers(rowChan, eventChan, &wg, numWorkers, currencyUsageMap, &mu, coins)

	events = CollectEvents(eventChan)
	SortEventsByTimestamp(events)

	// Print currency usage for debugging purposes
	PrintCurrencyUsage(currencyUsageMap)

	return events, currencyUsageMap
}

func startWorkers(
	rowChan <-chan []string,
	eventChan chan<- Event,
	wg *sync.WaitGroup,
	numWorkers int,
	currencyUsageMap CurrencyUsageMap,
	mu *sync.Mutex,
	coins []currency.Coin,
) {
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for row := range rowChan {
				event, err := ParseRowToEvent(row, currencyUsageMap, mu, coins)
				if err != nil {
					log.Printf("Worker %d: failed to parse row: %v", workerID, err)
					continue
				}
				eventChan <- event
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(eventChan)
	}()
}
