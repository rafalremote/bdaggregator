package storage

/*
 * Altough for this exercise Google Cloud Storage (GCS) is the only requested
 * storage type, for testing and demo purposes, I added the ability to load
 * a file from a local path to showcase storage flexibility.
 * We could add more storage options, such as S3, Azure, etc.
 */
import (
	"fmt"

	"bdaggregator/internal/config"
	"bdaggregator/internal/storage/gcs"
	"bdaggregator/internal/storage/local"
)

func NewStorage(cfg *config.Config) (Storage, error) {
	switch cfg.StorageType {
	case "GCS":
		return gcs.NewGCSStorage(cfg), nil
	case "local":
		return local.NewLocalStorage(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
}
