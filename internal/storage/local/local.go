package local

import (
	"bdaggregator/internal/config"
	"io"
	"os"
)

type LocalStorage struct {
	cfg *config.Config
}

func NewLocalStorage(cfg *config.Config) *LocalStorage {
	return &LocalStorage{cfg: cfg}
}

func (l *LocalStorage) Download() (io.Reader, error) {
	file, err := os.Open(l.cfg.LocalStoragePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}
