package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/config"
	"github.com/Shashank0701-byte/Loadster/pkg/engine"
	"github.com/Shashank0701-byte/Loadster/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	targetURL string
	users     int
	duration  string
)

var runCmd = &cobra.Command{
	Use:   "run [scenario.yaml]",
	Short: "Run a load test locally",
	Long:  `Executes a load test locally, either targeting a single URL or using a scenario YAML file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfg *config.Config
		var err error

		if len(args) > 0 {
			// Scenario YAML mode
			scenarioFile := args[0]
			logger.Log.Info("Loading test scenario from YAML", zap.String("file", scenarioFile))
			cfg, err = config.ParseFile(scenarioFile)
			if err != nil {
				return fmt.Errorf("failed to load scenario: %w", err)
			}
		} else {
			// CLI flag mode
			if targetURL == "" {
				return fmt.Errorf("target URL is required (specify via --url or use scenario.yaml)")
			}
			logger.Log.Info("Configuring test from CLI flags", zap.String("url", targetURL), zap.Int("users", users), zap.String("duration", duration))
			cfg = &config.Config{
				Target: targetURL,
				Stages: []config.Stage{
					{
						Users:       users,
						RawDuration: duration,
					},
				},
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid parameters: %w", err)
			}
		}

		logger.Log.Info("Scenario configuration validated successfully",
			zap.String("target", cfg.Target),
			zap.Int("stages_count", len(cfg.Stages)),
		)

		// Setup signal notification context for graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// Initialize execution engine
		eng := engine.NewEngine(cfg)

		// Background worker to consume results and prevent blocking the engine
		var wg sync.WaitGroup
		var totalRequests int64
		var successfulRequests int64
		var failedRequests int64
		var metricsMutex sync.Mutex

		wg.Add(1)
		go func() {
			defer wg.Done()
			for res := range eng.Results() {
				metricsMutex.Lock()
				totalRequests++
				if res.Error != nil || res.StatusCode >= 400 {
					failedRequests++
				} else {
					successfulRequests++
				}
				metricsMutex.Unlock()
			}
		}()

		// Periodically log progress in the background
		progressCtx, progressCancel := context.WithCancel(ctx)
		defer progressCancel()
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-progressCtx.Done():
					return
				case <-ticker.C:
					metricsMutex.Lock()
					logger.Log.Info("Running load test...",
						zap.Int("active_vus", eng.ActiveVUs()),
						zap.Int64("requests_sent", totalRequests),
						zap.Int64("success_count", successfulRequests),
						zap.Int64("failure_count", failedRequests),
					)
					metricsMutex.Unlock()
				}
			}
		}()

		// Start execution
		err = eng.Run(ctx)
		progressCancel() // Stop the progress logger immediately

		// Wait for results consumer to finish draining the channel
		wg.Wait()

		if err != nil && err != context.Canceled {
			logger.Log.Error("Engine run completed with error", zap.Error(err))
		}

		logger.Log.Info("Load test execution completed",
			zap.Int64("total_requests", totalRequests),
			zap.Int64("successful_requests", successfulRequests),
			zap.Int64("failed_requests", failedRequests),
		)

		return err
	},
}

func init() {
	runCmd.Flags().StringVar(&targetURL, "url", "", "Target URL to test")
	runCmd.Flags().IntVar(&users, "users", 10, "Number of concurrent users")
	runCmd.Flags().StringVar(&duration, "duration", "30s", "Duration of test (e.g., 30s, 1m)")

	rootCmd.AddCommand(runCmd)
}
type RunCmd_Type = *cobra.Command // For linking symbol reference
