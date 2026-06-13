package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Run_Integration(t *testing.T) {
	// Spin up a mock HTTP test server
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Configure a short test scenario: 1 stage, target 5 VUs, duration 1s
	cfg := &config.Config{
		Target: server.URL,
		Stages: []config.Stage{
			{
				Users:       5,
				Duration:    1 * time.Second,
				RawDuration: "1s",
			},
		},
	}

	eng := NewEngine(cfg)

	// Consume results from engine in a background goroutine so they don't block
	resultsDone := make(chan struct{})
	var receivedResultsCount int64
	go func() {
		defer close(resultsDone)
		for range eng.Results() {
			atomic.AddInt64(&receivedResultsCount, 1)
		}
	}()

	// Run the engine
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := eng.Run(ctx)
	require.NoError(t, err)

	// Wait for results to drain
	<-resultsDone

	// Assertions
	reqs := atomic.LoadInt64(&requestCount)
	assert.Greater(t, reqs, int64(0), "Should execute some HTTP requests")
	assert.Equal(t, reqs, atomic.LoadInt64(&receivedResultsCount), "Received results should match server requests")
}

func TestEngine_GracefulShutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Scenario with longer duration
	cfg := &config.Config{
		Target: server.URL,
		Stages: []config.Stage{
			{
				Users:       10,
				Duration:    10 * time.Second,
				RawDuration: "10s",
			},
		},
	}

	eng := NewEngine(cfg)

	resultsDone := make(chan struct{})
	go func() {
		defer close(resultsDone)
		for range eng.Results() {
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	
	// Start engine in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- eng.Run(ctx)
	}()

	// Let it run for 500ms
	time.Sleep(500 * time.Millisecond)
	
	// Cancel it
	cancel()

	// It should return context.Canceled error
	err := <-errChan
	assert.ErrorIs(t, err, context.Canceled)

	// Results channel should be closed and drained
	select {
	case <-resultsDone:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for results channel to close")
	}
}
