package coingecko

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Mock Config for tests
func mockConfig() {
	os.Setenv("SEQUENCE_COINGECKO_API_KEY", "test_api_key")
	os.Setenv("SEQUENCE_COINGECKO_API_URL", "http://mockapi.com/")
}

func TestApiGet(t *testing.T) {
	mockConfig()

	// Create a mock server to simulate the CoinGecko API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the correct headers are being set
		if r.Header.Get("accept") != "application/json" {
			t.Errorf("expected Accept header to be application/json, got %s", r.Header.Get("accept"))
		}
		if r.Header.Get("x-cg-demo-api-key") != "test_api_key" {
			t.Errorf("expected x-cg-demo-api-key header to be test_api_key, got %s", r.Header.Get("x-cg-demo-api-key"))
		}

		// Respond with a mock JSON response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"mock": "data"}`))
	}))
	defer mockServer.Close()

	// Override CoinGecko API URL to point to the mock server
	os.Setenv("SEQUENCE_COINGECKO_API_URL", mockServer.URL+"/")

	response, err := ApiGet("test_path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if the response matches expected mock response
	expectedResponse := `{"mock": "data"}`
	if response != expectedResponse {
		t.Errorf("expected response to be %s, got %s", expectedResponse, response)
	}
}
