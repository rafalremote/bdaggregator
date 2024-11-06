package gcs

import (
	"bdaggregator/internal/config"
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type GCSStorage struct {
	cfg *config.Config
}

func NewGCSStorage(cfg *config.Config) *GCSStorage {
	return &GCSStorage{cfg: cfg}
}

func (g *GCSStorage) Download() (io.Reader, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(g.cfg.GCSBucket)
	obj := bucket.Object(g.cfg.GCSObject)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader, nil
}
