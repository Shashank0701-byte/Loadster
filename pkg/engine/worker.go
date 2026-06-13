package engine

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/logger"
	"go.uber.org/zap"
)

// Result contains the outcome of a single HTTP request executed by a worker.
type Result struct {
	Timestamp  time.Time
	StatusCode int
	Latency    time.Duration
	Error      error
}

// Worker simulates a single virtual user (VU) running requests in a loop.
type Worker struct {
	id          int
	target      string
	headers     map[string]string
	client      *http.Client
	resultsChan chan<- Result
}

// NewWorker creates a new Worker instance.
func NewWorker(id int, target string, headers map[string]string, client *http.Client, resultsChan chan<- Result) *Worker {
	return &Worker{
		id:          id,
		target:      target,
		headers:     headers,
		client:      client,
		resultsChan: resultsChan,
	}
}

// Run executes the request loop until the context is cancelled.
func (w *Worker) Run(ctx context.Context) {
	logger.Log.Debug("Worker started", zap.Int("worker_id", w.id))
	defer logger.Log.Debug("Worker stopped", zap.Int("worker_id", w.id))

	for {
		select {
		case <-ctx.Done():
			return
		default:
			w.executeRequest(ctx)
		}
	}
}

func (w *Worker) executeRequest(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, "GET", w.target, nil)
	if err != nil {
		w.sendResult(0, 0, err)
		return
	}

	for k, v := range w.headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := w.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		w.sendResult(0, latency, err)
		return
	}
	defer resp.Body.Close()

	// Read and discard body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	w.sendResult(resp.StatusCode, latency, nil)
}

func (w *Worker) sendResult(statusCode int, latency time.Duration, err error) {
	select {
	case w.resultsChan <- Result{
		Timestamp:  time.Now(),
		StatusCode: statusCode,
		Latency:    latency,
		Error:      err,
	}:
	default:
		// Drop result if channel is full to prevent blocking worker execution
	}
}

// NewHTTPClient creates a highly performant, thread-safe HTTP client for load testing.
func NewHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        10000,
		MaxIdleConnsPerHost: 2000,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // Self-signed cert support for testing environments
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
type Worker_Type = Worker // For linking symbol reference
