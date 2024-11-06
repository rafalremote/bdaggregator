package etl

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCollectEvents(t *testing.T) {
	event1 := NewEvent(time.Now(), "sfl-token", "BUY_ITEMS", "SFL", 4974, decimal.NewFromInt(0), decimal.RequireFromString("0.6136203411678249"))
	event2 := NewEvent(time.Now(), "sfl-token", "BUY_ITEMS", "SFL", 4974, decimal.NewFromInt(0), decimal.RequireFromString("2.361412166673735"))

	eventChan := make(chan Event, 2)
	eventChan <- event1
	eventChan <- event2
	close(eventChan)

	events := CollectEvents(eventChan)
	assert.Equal(t, 2, len(events), "Expected 2 events to be collected")
}

func TestSortEventsByTimestamp(t *testing.T) {
	event1 := NewEvent(time.Now().Add(-time.Hour), "sfl-token", "BUY_ITEMS", "SFL", 4974, decimal.NewFromInt(0), decimal.RequireFromString("0.6136203411678249"))
	event2 := NewEvent(time.Now(), "sfl-token", "BUY_ITEMS", "SFL", 4974, decimal.NewFromInt(0), decimal.RequireFromString("2.361412166673735"))

	events := []Event{event2, event1}

	SortEventsByTimestamp(events)

	assert.True(t, events[0].TsUnix < events[1].TsUnix, "Events should be sorted by timestamp in ascending order")
}
