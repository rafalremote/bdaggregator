package storage

import (
	"bdaggregator/internal/config"
	"bdaggregator/internal/storage/gcs"
	"bdaggregator/internal/storage/local"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name         string
		storageType  string
		expectedType interface{}
		expectError  bool
	}{
		{
			name:         "GCS storage type",
			storageType:  "GCS",
			expectedType: &gcs.GCSStorage{},
			expectError:  false,
		},
		{
			name:         "Local storage type",
			storageType:  "local",
			expectedType: &local.LocalStorage{},
			expectError:  false,
		},
		{
			name:         "Unsupported storage type",
			storageType:  "unsupported",
			expectedType: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				StorageType: tt.storageType,
			}

			storage, err := NewStorage(cfg)

			if tt.expectError {
				assert.Error(t, err, "Expected an error for unsupported storage type")
				assert.Nil(t, storage, "Expected storage to be nil for unsupported type")
			} else {
				assert.NoError(t, err, "Expected no error for supported storage type")
				assert.IsType(t, tt.expectedType, storage, "Expected storage type to match")
			}
		})
	}
}
