package etl

import (
	"sync"

	"github.com/shopspring/decimal"
)

type AggregatePerProject struct {
	Day                            string
	ProjectID                      int
	NumberOfTransactionsPerProject int
	TotalVolumePerProject          float64
	Currency                       string
}

// Aggregates events by day and project, calculating total volume in the specified currency.
func AggregateEvents(events []Event, defaultCurrency string) []AggregatePerProject {
	numWorkers := 4
	chunkSize := (len(events) + numWorkers - 1) / numWorkers
	aggregatedData := make(map[string]map[int]*AggregatePerProject)
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(events) {
			end = len(events)
		}

		wg.Add(1)
		go func(eventsChunk []Event) {
			defer wg.Done()
			chunkAggregate := aggregateChunk(eventsChunk, defaultCurrency)
			mergeChunkIntoMainData(chunkAggregate, aggregatedData, &mutex)
		}(events[start:end])
	}

	wg.Wait()
	return convertToSlice(aggregatedData, defaultCurrency)
}

// Process a chunk of events, computing daily totals per project.
func aggregateChunk(eventsChunk []Event, defaultCurrency string) map[string]map[int]*AggregatePerProject {
	chunkAggregate := make(map[string]map[int]*AggregatePerProject)
	for _, event := range eventsChunk {
		day := event.Ts.Format("2006-01-02")
		volume := calculateVolume(event)

		if _, exists := chunkAggregate[day]; !exists {
			chunkAggregate[day] = make(map[int]*AggregatePerProject)
		}

		initializeAggregateEntry(chunkAggregate[day], event.ProjectID, day, defaultCurrency)
		updateAggregateEntry(chunkAggregate[day][event.ProjectID], volume)
	}
	return chunkAggregate
}

func calculateVolume(event Event) float64 {
	currencyValue := event.CurrencyValueDecimal
	if event.CoinID == "matic-network" {
		currencyValue = currencyValue.Div(decimal.NewFromInt(1e18))
	}
	return currencyValue.Mul(event.CurrencyExchangeRate).InexactFloat64()
}

func initializeAggregateEntry(projectData map[int]*AggregatePerProject, projectID int, day, defaultCurrency string) {
	if _, exists := projectData[projectID]; !exists {
		projectData[projectID] = &AggregatePerProject{
			Day:                            day,
			ProjectID:                      projectID,
			NumberOfTransactionsPerProject: 0,
			TotalVolumePerProject:          0,
			Currency:                       defaultCurrency,
		}
	}
}

func updateAggregateEntry(entry *AggregatePerProject, volume float64) {
	entry.NumberOfTransactionsPerProject++
	entry.TotalVolumePerProject += volume
}

// Merge a chunk’s data into the main aggregation map.
func mergeChunkIntoMainData(chunkData, mainData map[string]map[int]*AggregatePerProject, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()

	for day, projectData := range chunkData {
		mergeProjectData(mainData, day, projectData)
	}
}

// Merge a single day's project data into the main map.
func mergeProjectData(mainData map[string]map[int]*AggregatePerProject, day string, projectData map[int]*AggregatePerProject) {
	if _, exists := mainData[day]; !exists {
		mainData[day] = make(map[int]*AggregatePerProject)
	}
	for projectID, aggregate := range projectData {
		mergeAggregateEntry(mainData[day], projectID, aggregate)
	}
}

// Merges an individual aggregate entry into the main data map.
func mergeAggregateEntry(dayData map[int]*AggregatePerProject, projectID int, aggregate *AggregatePerProject) {
	if existing, exists := dayData[projectID]; !exists {
		dayData[projectID] = aggregate
	} else {
		existing.NumberOfTransactionsPerProject += aggregate.NumberOfTransactionsPerProject
		existing.TotalVolumePerProject += aggregate.TotalVolumePerProject
	}
}

func convertToSlice(aggregatedData map[string]map[int]*AggregatePerProject, defaultCurrency string) []AggregatePerProject {
	var result []AggregatePerProject
	for _, projectData := range aggregatedData {
		for _, aggregate := range projectData {
			// Round to 2 decimal places for BigQuery compatibility
			aggregate.TotalVolumePerProject = decimal.NewFromFloat(aggregate.TotalVolumePerProject).Round(2).InexactFloat64()
			aggregate.Currency = defaultCurrency // Set to specified currency
			result = append(result, *aggregate)
		}
	}
	return result
}
