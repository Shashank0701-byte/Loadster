package engine

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/config"
	"github.com/Shashank0701-byte/Loadster/pkg/logger"
	"go.uber.org/zap"
)

// Engine orchestrates the load test stages and worker pool lifecycle.
type Engine struct {
	cfg         *config.Config
	client      *http.Client
	resultsChan chan Result
	activeVUs   int
	vuMutex     sync.Mutex
	workerWg    sync.WaitGroup
}

// NewEngine creates a new Engine instance.
func NewEngine(cfg *config.Config) *Engine {
	timeout := 10 * time.Second
	if cfg.Timeout != "" {
		if t, err := time.ParseDuration(cfg.Timeout); err == nil {
			timeout = t
		}
	}

	return &Engine{
		cfg:         cfg,
		client:      NewHTTPClient(timeout),
		resultsChan: make(chan Result, 10000), // Buffered channel to support high throughput
	}
}

// Results returns the channel where request results are sent.
func (e *Engine) Results() <-chan Result {
	return e.resultsChan
}

// ActiveVUs returns the current count of active virtual users.
func (e *Engine) ActiveVUs() int {
	e.vuMutex.Lock()
	defer e.vuMutex.Unlock()
	return e.activeVUs
}

// Run executes the load test scenario. It blocks until the test is completed or the context is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	logger.Log.Info("Starting load test execution",
		zap.String("target", e.cfg.Target),
		zap.Int("stages", len(e.cfg.Stages)),
	)

	var workerCancels []context.CancelFunc
	resultsCloseDone := make(chan struct{})

	defer func() {
		logger.Log.Info("Cleaning up running workers")
		for _, cancel := range workerCancels {
			cancel()
		}

		// Start the background routine to wait for all workers to exit and close the results channel safely.
		// We start this only during cleanup to prevent it from executing prematurely when the worker list is empty.
		go func() {
			e.workerWg.Wait()
			close(e.resultsChan)
			logger.Log.Info("Engine results channel closed")
			close(resultsCloseDone)
		}()

		// Block returning until all worker goroutines have fully exited and resultsChan is closed.
		<-resultsCloseDone
	}()

	startUsers := 0
	workerIDSeq := 0

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for i, stage := range e.cfg.Stages {
		logger.Log.Info("Starting stage",
			zap.Int("stage_number", i+1),
			zap.Int("target_users", stage.Users),
			zap.String("duration", stage.RawDuration),
		)

		stageStart := time.Now()
		stageDuration := stage.Duration
		endUsers := stage.Users

		stageDone := false
		for !stageDone {
			select {
			case <-ctx.Done():
				logger.Log.Info("Test execution cancelled")
				return ctx.Err()
			case <-ticker.C:
				elapsed := time.Since(stageStart)
				if elapsed >= stageDuration {
					stageDone = true
					break
				}

				// Linear ramping logic
				ratio := float64(elapsed) / float64(stageDuration)
				targetUsers := startUsers + int(float64(endUsers-startUsers)*ratio)

				workerCancels = e.adjustWorkerPool(ctx, targetUsers, &workerIDSeq, workerCancels)
			}
		}

		// Ensure we hit the exact target users at the end of the stage
		workerCancels = e.adjustWorkerPool(ctx, endUsers, &workerIDSeq, workerCancels)
		startUsers = endUsers
	}

	logger.Log.Info("All stages completed successfully")
	return nil
}

func (e *Engine) adjustWorkerPool(ctx context.Context, targetUsers int, workerIDSeq *int, cancels []context.CancelFunc) []context.CancelFunc {
	e.vuMutex.Lock()
	defer e.vuMutex.Unlock()

	current := len(cancels)
	if targetUsers > current {
		diff := targetUsers - current
		logger.Log.Debug("Scaling up workers", zap.Int("current", current), zap.Int("target", targetUsers), zap.Int("diff", diff))
		for i := 0; i < diff; i++ {
			*workerIDSeq++
			wCtx, wCancel := context.WithCancel(ctx)
			cancels = append(cancels, wCancel)
			w := NewWorker(*workerIDSeq, e.cfg.Target, e.cfg.Headers, e.client, e.resultsChan)
			
			e.workerWg.Add(1)
			go func(worker *Worker, wctx context.Context) {
				defer e.workerWg.Done()
				worker.Run(wctx)
			}(w, wCtx)
		}
	} else if targetUsers < current {
		diff := current - targetUsers
		logger.Log.Debug("Scaling down workers", zap.Int("current", current), zap.Int("target", targetUsers), zap.Int("diff", diff))
		for i := 0; i < diff; i++ {
			lastIdx := len(cancels) - 1
			cancels[lastIdx]()
			cancels = cancels[:lastIdx]
		}
	}
	e.activeVUs = len(cancels)
	return cancels
}
type Engine_Type = Engine // For linking symbol reference
