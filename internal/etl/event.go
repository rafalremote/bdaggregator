package etl

import (
	"bdaggregator/internal/currency"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Props struct {
	ChainID         string `json:"chainId"`
	TxnHash         string `json:"txnHash"`
	CurrencySymbol  string `json:"currencySymbol"`
	CurrencyAddress string `json:"currencyAddress"`
}

type Nums struct {
	CurrencyValueDecimal string `json:"currencyValueDecimal"`
}

type Event struct {
	Ts                   time.Time
	TsUnix               int64
	Event                string
	ProjectID            int
	CurrencySymbol       string
	CoinID               string
	CurrencyExchangeRate decimal.Decimal
	CurrencyValueDecimal decimal.Decimal
}

func NewEvent(ts time.Time, coinID, event, currencySymbol string, projectID int, currencyExchangeRate, currencyValueDecimal decimal.Decimal) Event {
	return Event{
		Ts:                   ts,
		TsUnix:               ts.Unix(),
		Event:                event,
		ProjectID:            projectID,
		CurrencySymbol:       currencySymbol,
		CoinID:               coinID,
		CurrencyExchangeRate: currencyExchangeRate,
		CurrencyValueDecimal: currencyValueDecimal,
	}
}

func ReadCSVRows(reader io.Reader, rowChan chan<- []string) {
	defer close(rowChan)
	csvReader := csv.NewReader(reader)

	// Skip the header row
	if _, err := csvReader.Read(); err != nil {
		log.Printf("failed to read or skip header row: %v", err)
		return
	}

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("failed to read CSV row: %v", err)
			continue
		}
		rowChan <- row
	}
}

func ParseRowToEvent(row []string, currencyUsageMap CurrencyUsageMap, mu *sync.Mutex, coins []currency.Coin) (Event, error) {
	// Parse the timestamp
	ts, err := time.Parse("2006-01-02 15:04:05.000", row[1])
	if err != nil {
		return Event{}, err
	}

	// Parse project_id
	projectID, err := strconv.Atoi(row[3])
	if err != nil {
		return Event{}, err
	}

	// Parse event type
	eventType := row[2]

	// Parse props (JSON)
	var props Props
	if err := json.Unmarshal([]byte(row[14]), &props); err != nil {
		return Event{}, err
	}
	currencySymbol := props.CurrencySymbol
	currencyAddress := props.CurrencyAddress
	chainID := props.ChainID

	// Parse nums (JSON)
	var nums Nums
	if err := json.Unmarshal([]byte(row[15]), &nums); err != nil {
		return Event{}, err
	}

	currencyValueDecimal, err := decimal.NewFromString(nums.CurrencyValueDecimal)
	if err != nil {
		return Event{}, err
	}

	// Map currency symbol and currency address to coin ID
	coinID, err := currency.MapCurrencyToCoinID(coins, currencySymbol, currencyAddress, chainID)
	if err != nil {
		return Event{}, err
	}

	// initialize with 0 before we update events with exchange rate
	currencyExchangeRate := decimal.NewFromInt(0)

	UpdateCurrencyUsageMap(currencyUsageMap, mu, coinID, ts.Unix())

	return NewEvent(ts, coinID, eventType, currencySymbol, projectID, currencyExchangeRate, currencyValueDecimal), nil
}

func CollectEvents(eventChan <-chan Event) []Event {
	var events []Event
	for event := range eventChan {
		events = append(events, event)
	}
	return events
}

func SortEventsByTimestamp(events []Event) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].TsUnix < events[j].TsUnix
	})
}
