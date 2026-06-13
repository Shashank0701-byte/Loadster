package metrics

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loadster_requests_total",
		Help: "Total number of HTTP requests sent",
	})
	errorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loadster_errors_total",
		Help: "Total number of HTTP request errors",
	})
	latencyHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "loadster_latency_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	})
	activeUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "loadster_active_users",
		Help: "Current active virtual users count",
	})
	currentRPS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "loadster_rps",
		Help: "Current requests per second",
	})
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(errorsTotal)
	prometheus.MustRegister(latencyHistogram)
	prometheus.MustRegister(activeUsers)
	prometheus.MustRegister(currentRPS)
}

// Stats holds aggregated real-time statistics.
type Stats struct {
	TotalRequests int64
	TotalErrors   int64
	SuccessRate   float64
	Average       time.Duration
	P50           time.Duration
	P90           time.Duration
	P95           time.Duration
	P99           time.Duration
	CurrentRPS    float64
	ActiveUsers   int
}

// Collector collects raw results and computes aggregated metrics.
type Collector struct {
	mu          sync.RWMutex
	latencies   []time.Duration
	totalReqs   int64
	totalErrors int64
	activeVUs   int
	windowReqs  int64
	windowStart time.Time
	currentRPS  float64
	metricsSrv  *http.Server
	cancelRps   context.CancelFunc
}

// NewCollector creates a new metrics collector and starts the Prometheus metrics HTTP server if enabled.
func NewCollector(listenAddress string) *Collector {
	c := &Collector{
		windowStart: time.Now(),
	}

	if listenAddress != "" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		c.metricsSrv = &http.Server{
			Addr:    listenAddress,
			Handler: mux,
		}
		go func() {
			_ = c.metricsSrv.ListenAndServe()
		}()
	}

	rpsCtx, rpsCancel := context.WithCancel(context.Background())
	c.cancelRps = rpsCancel

	// Update RPS every second in the background
	go c.rpsLoop(rpsCtx)

	return c
}

// Close shuts down the Prometheus HTTP server and stops background loops.
func (c *Collector) Close() {
	if c.cancelRps != nil {
		c.cancelRps()
	}
	if c.metricsSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = c.metricsSrv.Shutdown(ctx)
	}
}

// RecordResult records a request outcome.
func (c *Collector) RecordResult(statusCode int, latency time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalReqs++
	c.windowReqs++
	c.latencies = append(c.latencies, latency)

	requestsTotal.Inc()
	latencyHistogram.Observe(latency.Seconds())

	if err != nil || statusCode >= 400 {
		c.totalErrors++
		errorsTotal.Inc()
	}
}

// SetActiveUsers updates the gauge for active virtual users.
func (c *Collector) SetActiveUsers(count int) {
	c.mu.Lock()
	c.activeVUs = count
	c.mu.Unlock()

	activeUsers.Set(float64(count))
}

func (c *Collector) rpsLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			dur := now.Sub(c.windowStart).Seconds()
			if dur > 0 {
				c.currentRPS = float64(c.windowReqs) / dur
				currentRPS.Set(c.currentRPS)
			}
			c.windowReqs = 0
			c.windowStart = now
			c.mu.Unlock()
		}
	}
}

// Snapshot computes and returns a copy of the current metrics.
func (c *Collector) Snapshot() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := Stats{
		TotalRequests: c.totalReqs,
		TotalErrors:   c.totalErrors,
		CurrentRPS:    c.currentRPS,
		ActiveUsers:   c.activeVUs,
	}

	if c.totalReqs > 0 {
		stats.SuccessRate = float64(c.totalReqs-c.totalErrors) / float64(c.totalReqs) * 100
	}

	n := len(c.latencies)
	if n == 0 {
		return stats
	}

	var total time.Duration
	for _, l := range c.latencies {
		total += l
	}
	stats.Average = total / time.Duration(n)

	sorted := make([]time.Duration, n)
	copy(sorted, c.latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	stats.P50 = sorted[int(float64(n)*0.50)]
	stats.P90 = sorted[int(float64(n)*0.90)]
	stats.P95 = sorted[int(float64(n)*0.95)]
	stats.P99 = sorted[int(float64(n)*0.99)]

	return stats
}

// GetLatencies returns a copy of raw latencies collected.
func (c *Collector) GetLatencies() []time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	copied := make([]time.Duration, len(c.latencies))
	copy(copied, c.latencies)
	return copied
}
type Collector_Type = Collector // For linking symbol reference
type Stats_Type = Stats // For linking symbol reference
