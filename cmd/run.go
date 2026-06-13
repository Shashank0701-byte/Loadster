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
	"github.com/Shashank0701-byte/Loadster/pkg/metrics"
	"github.com/Shashank0701-byte/Loadster/pkg/report"
	"github.com/Shashank0701-byte/Loadster/pkg/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	targetURL string
	users     int
	duration  string
	noUI      bool
	promAddr  string
	jsonPath  string
	htmlPath  string
)

var runCmd = &cobra.Command{
	Use:   "run [scenario.yaml]",
	Short: "Run a load test locally",
	Long:  `Executes a load test locally, either targeting a single URL or using a scenario YAML file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()
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

		// Redirection of logs when using the CLI dashboard (prevents output corruption)
		useUI := !noUI && !debug
		if useUI {
			if err := logger.InitFileLogger("loadster.log"); err != nil {
				return fmt.Errorf("failed to initialize file logger: %w", err)
			}
		}

		// Initialize metrics collector (starts Prometheus scraper server)
		collector := metrics.NewCollector(promAddr)
		defer collector.Close()

		// Initialize execution engine
		eng := engine.NewEngine(cfg)

		// Background worker to consume results and feed them to the metrics collector
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for res := range eng.Results() {
				collector.RecordResult(res.StatusCode, res.Latency, res.Error)
			}
		}()

		// Track active virtual users in the background
		vuCtx, vuCancel := context.WithCancel(ctx)
		defer vuCancel()
		go func() {
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-vuCtx.Done():
					return
				case <-ticker.C:
					collector.SetActiveUsers(eng.ActiveVUs())
				}
			}
		}()

		var p *tea.Program
		var dash *ui.Dashboard
		if useUI {
			dash = ui.NewDashboard(collector, len(cfg.Stages))
			eng.OnStageChange = func(ev engine.StageChangeEvent) {
				dash.UpdateStage(ev.StageNum, ev.TargetUsers, ev.Duration)
			}
			p = tea.NewProgram(dash, tea.WithAltScreen())
		} else {
			// Print progress log messages in no-UI mode
			go func() {
				ticker := time.NewTicker(2 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-vuCtx.Done():
						return
					case <-ticker.C:
						stats := collector.Snapshot()
						logger.Log.Info("Running load test...",
							zap.Int("active_vus", stats.ActiveUsers),
							zap.Float64("current_rps", stats.CurrentRPS),
							zap.Int64("total_requests", stats.TotalRequests),
							zap.Int64("errors_count", stats.TotalErrors),
						)
					}
				}
			}()
		}

		// Start execution
		var runErr error
		if useUI {
			engineCtx, engineCancel := context.WithCancel(ctx)
			defer engineCancel()

			go func() {
				runErr = eng.Run(engineCtx)
				p.Send(tea.Quit())
			}()

			// Run Bubble Tea program (blocks until tea.Quit is called)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("dashboard execution failed: %w", err)
			}
			engineCancel()
		} else {
			runErr = eng.Run(ctx)
		}

		vuCancel()
		wg.Wait()

		testDuration := time.Since(startTime)
		stats := collector.Snapshot()
		if useUI {
			// Print a beautiful text summary since logs were redirected to loadster.log
			fmt.Println("\n==================================================")
			fmt.Println("             LOAD TEST RUN SUMMARY")
			fmt.Println("==================================================")
			fmt.Printf("Target URL:         %s\n", cfg.Target)
			fmt.Printf("Total Requests:     %d\n", stats.TotalRequests)
			fmt.Printf("Successful:         %d\n", stats.TotalRequests-stats.TotalErrors)
			fmt.Printf("Failed/Errors:      %d\n", stats.TotalErrors)
			fmt.Printf("Success Rate:       %.2f%%\n", stats.SuccessRate)
			fmt.Printf("Average Latency:    %s\n", stats.Average.Round(time.Millisecond))
			fmt.Printf("P50 Latency:        %s\n", stats.P50.Round(time.Millisecond))
			fmt.Printf("P95 Latency:        %s\n", stats.P95.Round(time.Millisecond))
			fmt.Printf("P99 Latency:        %s\n", stats.P99.Round(time.Millisecond))
			fmt.Println("==================================================")
		} else {
			logger.Log.Info("Load test execution completed",
				zap.Int64("total_requests", stats.TotalRequests),
				zap.Int64("successful_requests", stats.TotalRequests-stats.TotalErrors),
				zap.Int64("failed_requests", stats.TotalErrors),
				zap.Float64("success_rate", stats.SuccessRate),
				zap.Duration("avg_latency", stats.Average),
			)
		}

		// Export reports if paths are specified
		reportData := report.BuildReportData(cfg.Target, testDuration, collector)
		if jsonPath != "" {
			if err := report.WriteJSON(reportData, jsonPath); err != nil {
				logger.Log.Error("Failed to write JSON report", zap.String("path", jsonPath), zap.Error(err))
			} else {
				logger.Log.Info("JSON report exported successfully", zap.String("path", jsonPath))
			}
		}

		if htmlPath != "" {
			if err := report.WriteHTML(reportData, htmlPath); err != nil {
				logger.Log.Error("Failed to write HTML report", zap.String("path", htmlPath), zap.Error(err))
			} else {
				logger.Log.Info("HTML report exported successfully", zap.String("path", htmlPath))
			}
		}

		if runErr != nil && runErr != context.Canceled {
			return runErr
		}
		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&targetURL, "url", "", "Target URL to test")
	runCmd.Flags().IntVar(&users, "users", 10, "Number of concurrent users")
	runCmd.Flags().StringVar(&duration, "duration", "30s", "Duration of test (e.g., 30s, 1m)")
	runCmd.Flags().BoolVar(&noUI, "no-ui", false, "Disable live terminal UI dashboard")
	runCmd.Flags().StringVar(&promAddr, "prom-addr", "localhost:9090", "Prometheus metrics server address")
	runCmd.Flags().StringVar(&jsonPath, "json-out", "", "File path to save the JSON report")
	runCmd.Flags().StringVar(&htmlPath, "html-out", "", "File path to save the HTML report")

	rootCmd.AddCommand(runCmd)
}
type RunCmd_Type = *cobra.Command // For linking symbol reference
