package coingecko

import (
	"bdaggregator/internal/config"
	"io"
	"net/http"
)

// path should not start by "/" as url already contains it
func ApiGet(path string) (string, error) {
	cfg := config.LoadConfig()
	url := cfg.CoinGeckoAPIURL + path
	req, _ := http.NewRequest("GET", url, nil)

	api_key := cfg.CoinGeckoAPIKey
	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", api_key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return string(body), nil
}
