package config

import (
	"os"
)

type Config struct {
	DbType                string
	BigQueryProject       string
	BigQueryDataset       string
	BigQueryLocation      string
	GoogleCloudStorageURL string
	GCSBucket             string
	GCSObject             string
	LocalStoragePath      string
	StorageType           string
	CoinGeckoAPIKey       string
	CoinGeckoAPIURL       string
	CoinListPath          string
	DefaultCurrency       string
}

// SEQUENCE_ prefix allows to grepping the env variables and
// avoid potencial conflicts with other variables
func LoadConfig() *Config {
	return &Config{
		DbType:                os.Getenv("SEQUENCE_DB_TYPE"),
		BigQueryProject:       os.Getenv("SEQUENCE_BIGQUERY_PROJECT"),
		BigQueryDataset:       os.Getenv("SEQUENCE_BIGQUERY_DATASET"),
		BigQueryLocation:      os.Getenv("SEQUENCE_BIGQUERY_LOCATION"),
		GoogleCloudStorageURL: os.Getenv("SEQUENCE_GOOGLE_CLOUD_STORAGE_URL"),
		GCSBucket:             os.Getenv("SEQUENCE_GCS_BUCKET"),
		GCSObject:             os.Getenv("SEQUENCE_GCS_OBJECT"),
		LocalStoragePath:      os.Getenv("SEQUENCE_LOCAL_STORAGE_PATH"),
		StorageType:           os.Getenv("SEQUENCE_STORAGE_TYPE"),
		CoinGeckoAPIKey:       os.Getenv("SEQUENCE_COINGECKO_API_KEY"),
		CoinGeckoAPIURL:       os.Getenv("SEQUENCE_COINGECKO_API_URL"),
		CoinListPath:          os.Getenv("SEQUENCE_COINS_FILE_PATH"),
		DefaultCurrency:       os.Getenv("SEQUENCE_DEFAULT_CURRENCY"),
	}
}
