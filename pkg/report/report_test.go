package report

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportGeneration(t *testing.T) {
	c := metrics.NewCollector("")
	defer c.Close()

	c.RecordResult(200, 100*time.Millisecond, nil)
	c.RecordResult(200, 200*time.Millisecond, nil)
	c.RecordResult(404, 300*time.Millisecond, nil)

	data := BuildReportData("http://localhost:8080", 5*time.Second, c)

	assert.Equal(t, "http://localhost:8080", data.Target)
	assert.Equal(t, int64(3), data.TotalRequests)
	assert.Equal(t, int64(1), data.TotalErrors)
	assert.InDelta(t, 66.66, data.SuccessRate, 0.1)
	assert.Equal(t, 200.0, data.AverageLatencyMs)
	assert.Equal(t, 200.0, data.P50LatencyMs)
	assert.Equal(t, 300.0, data.P95LatencyMs)
	assert.Equal(t, 5.0, data.TestDurationSec)
	assert.InDelta(t, 0.6, data.ThroughputRps, 0.01)

	// Test writing JSON
	jsonFile := "test_report.json"
	err := WriteJSON(data, jsonFile)
	require.NoError(t, err)
	defer os.Remove(jsonFile)

	// Verify JSON file content
	readBytes, err := os.ReadFile(jsonFile)
	require.NoError(t, err)

	var readData ReportData
	err = json.Unmarshal(readBytes, &readData)
	require.NoError(t, err)
	assert.Equal(t, data.Target, readData.Target)

	// Test writing HTML
	htmlFile := "test_report.html"
	err = WriteHTML(data, htmlFile)
	require.NoError(t, err)
	defer os.Remove(htmlFile)

	// Verify HTML file was created and is non-empty
	info, err := os.Stat(htmlFile)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}
