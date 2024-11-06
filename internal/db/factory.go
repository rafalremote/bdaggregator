package db

import (
	"bdaggregator/internal/config"
	"bdaggregator/internal/db/bigquery"
	"context"
	"fmt"
)

func NewDatabase(ctx context.Context, cfg *config.Config) (Database, error) {
	switch cfg.DbType {
	case "BigQuery":
		return bigquery.NewBigQueryDB(ctx, cfg)
	// You can add more cases here for other database types, like ClickHouse
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DbType)
	}
}
