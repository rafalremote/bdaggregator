package db

import (
	"bdaggregator/internal/config"
	"bdaggregator/internal/db/bigquery"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabase(t *testing.T) {
	ctx := context.Background()

	// Test case: BigQuery database type
	cfg := &config.Config{
		DbType:          "BigQuery",
		BigQueryProject: "test-project",
		BigQueryDataset: "test-dataset",
	}
	database, err := NewDatabase(ctx, cfg)

	assert.NoError(t, err, "Expected no error for supported DbType 'BigQuery'")
	assert.IsType(t, &bigquery.BigQueryDB{}, database, "Expected BigQueryDB type for DbType 'BigQuery'")

	// Test case: Unsupported database type
	cfg.DbType = "UnsupportedDB"
	database, err = NewDatabase(ctx, cfg)

	assert.Nil(t, database, "Expected nil database for unsupported DbType")
	assert.Error(t, err, "Expected an error for unsupported DbType")
	assert.EqualError(t, err, "unsupported database type: UnsupportedDB", "Expected specific error message for unsupported DbType")
}
