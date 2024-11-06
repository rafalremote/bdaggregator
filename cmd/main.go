package main

import (
	"context"
	"log"
	"time"

	"bdaggregator/internal/config"
	"bdaggregator/internal/currency"
	"bdaggregator/internal/db"
	"bdaggregator/internal/etl"
	"bdaggregator/internal/storage"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	// ------------------------ SETUP ------------------------------------------

	storageClient, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	// Download the CSV file using the selected storage type
	reader, err := storageClient.Download()
	if err != nil {
		log.Fatalf("failed to download file: %v", err)
	}
	log.Println("File downloaded successfully")

	dbClient, err := db.NewDatabase(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer dbClient.Close()

	if err := dbClient.SetupDatabase(ctx); err != nil {
		log.Fatalf("failed to setup database: %v", err)
	}

	if err := dbClient.SetupTable(ctx, "aggregation"); err != nil {
		log.Fatalf("failed to setup table: %v", err)
	}

	supported_coins, err := currency.LoadCoins(cfg.CoinListPath)
	if err != nil {
		log.Fatalf("failed to load coins: %v", err)
	}

	// ------------------------ PROCESS DATA -----------------------------------
	startTime := time.Now()

	// Extract events and collect currency usage
	events, currencyUsageMap := etl.ExtractEvents(reader, supported_coins)

	// Get exchange rates
	exchangeRates, err := etl.GetExchangeRates(currencyUsageMap, cfg.DefaultCurrency, currency.FetchExchangeRates)

	if err != nil {
		log.Fatalf("failed to get exchange rates: %v", err)
	}

	// Update exchange rates in each event
	etl.UpdateExchangeRates(events, exchangeRates)

	// Aggregate events using concurrency
	aggregatedEvents := etl.AggregateEvents(events, cfg.DefaultCurrency)

	// Upserts aggregated events into BigQuery
	if err := dbClient.Upsert(ctx, "aggregation", aggregatedEvents); err != nil {
		log.Fatalf("failed to merge records into BigQuery: %v", err)
	}

	duration := time.Since(startTime).Seconds()
	log.Printf("Processed %d events in %v sec", len(events), duration)
}
