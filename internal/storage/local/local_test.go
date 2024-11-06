package local

import (
	"bdaggregator/internal/config"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalStorage_Download(t *testing.T) {
	tempFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err, "Expected no error creating temporary file")
	defer os.Remove(tempFile.Name())

	expectedData := "sample data"
	_, err = tempFile.WriteString(expectedData)
	assert.NoError(t, err, "Expected no error writing to temporary file")
	tempFile.Close()

	cfg := &config.Config{
		LocalStoragePath: tempFile.Name(),
	}

	localStorage := NewLocalStorage(cfg)

	reader, err := localStorage.Download()
	assert.NoError(t, err, "Expected no error from Download")

	data, err := io.ReadAll(reader)
	assert.NoError(t, err, "Expected no error reading data")
	assert.Equal(t, expectedData, string(data), "Expected data to match written data")
}

func TestLocalStorage_Download_FileNotFound(t *testing.T) {
	cfg := &config.Config{
		LocalStoragePath: "non_existent_file.txt",
	}

	localStorage := NewLocalStorage(cfg)

	_, err := localStorage.Download()
	assert.Error(t, err, "Expected an error due to non-existent file")
	assert.Contains(t, err.Error(), "no such file or directory", "Expected file not found error")
}
