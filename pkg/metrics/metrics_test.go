package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollector_RecordAndSnapshot(t *testing.T) {
	// Initialize with empty address to disable HTTP listener in test environments
	c := NewCollector("")
	defer c.Close()

	c.RecordResult(200, 10*time.Millisecond, nil)
	c.RecordResult(200, 20*time.Millisecond, nil)
	c.RecordResult(500, 30*time.Millisecond, nil)
	c.RecordResult(0, 50*time.Millisecond, errors.New("timeout"))

	c.SetActiveUsers(5)

	stats := c.Snapshot()

	assert.Equal(t, int64(4), stats.TotalRequests)
	assert.Equal(t, int64(2), stats.TotalErrors)
	assert.Equal(t, 50.0, stats.SuccessRate)
	assert.Equal(t, 5, stats.ActiveUsers)

	// Avg: (10 + 20 + 30 + 50) / 4 = 110 / 4 = 27.5ms
	assert.Equal(t, 27500*time.Microsecond, stats.Average)

	// Percentiles (sorted: 10, 20, 30, 50)
	// P50: index 4 * 0.50 = 2 -> 30ms
	// P90: index 4 * 0.90 = 3 -> 50ms
	assert.Equal(t, 30*time.Millisecond, stats.P50)
	assert.Equal(t, 50*time.Millisecond, stats.P90)
	assert.Equal(t, 50*time.Millisecond, stats.P95)
	assert.Equal(t, 50*time.Millisecond, stats.P99)
}
