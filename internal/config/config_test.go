package config_test

import (
	"bdaggregator/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setEnvVars(vars map[string]string) func() {
	// Save the current values and set the new environment variables
	originals := make(map[string]string)
	for key, value := range vars {
		originals[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// Return a function to restore the original values
	return func() {
		for key, value := range originals {
			os.Setenv(key, value)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	// Set the environment variables for testing
	envVars := map[string]string{
		"SEQUENCE_DB_TYPE":                  "BigQuery",
		"SEQUENCE_BIGQUERY_PROJECT":         "test_project",
		"SEQUENCE_BIGQUERY_DATASET":         "test_dataset",
		"SEQUENCE_BIGQUERY_LOCATION":        "US",
		"SEQUENCE_GOOGLE_CLOUD_STORAGE_URL": "https://storage.googleapis.com",
		"SEQUENCE_GCS_BUCKET":               "test_bucket",
		"SEQUENCE_GCS_OBJECT":               "test_object",
		"SEQUENCE_LOCAL_STORAGE_PATH":       "/tmp",
		"SEQUENCE_STORAGE_TYPE":             "GCS",
		"SEQUENCE_COINGECKO_API_KEY":        "test_api_key",
		"SEQUENCE_COINGECKO_API_URL":        "https://api.coingecko.com",
		"SEQUENCE_COINS_FILE_PATH":          "/path/to/coins.json",
		"SEQUENCE_DEFAULT_CURRENCY":         "USD",
	}

	// Set environment variables and defer the restoration of the original values
	defer setEnvVars(envVars)()

	cfg := config.LoadConfig()

	assert.Equal(t, "BigQuery", cfg.DbType)
	assert.Equal(t, "test_project", cfg.BigQueryProject)
	assert.Equal(t, "test_dataset", cfg.BigQueryDataset)
	assert.Equal(t, "US", cfg.BigQueryLocation)
	assert.Equal(t, "https://storage.googleapis.com", cfg.GoogleCloudStorageURL)
	assert.Equal(t, "test_bucket", cfg.GCSBucket)
	assert.Equal(t, "test_object", cfg.GCSObject)
	assert.Equal(t, "/tmp", cfg.LocalStoragePath)
	assert.Equal(t, "GCS", cfg.StorageType)
	assert.Equal(t, "test_api_key", cfg.CoinGeckoAPIKey)
	assert.Equal(t, "https://api.coingecko.com", cfg.CoinGeckoAPIURL)
	assert.Equal(t, "/path/to/coins.json", cfg.CoinListPath)
	assert.Equal(t, "USD", cfg.DefaultCurrency)
}
