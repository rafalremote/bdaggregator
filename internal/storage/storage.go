package storage

import "io"

// Define methods for downloading data
type Storage interface {
	Download() (io.Reader, error)
}
