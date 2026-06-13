package cmd

import (
	"fmt"

	"github.com/Shashank0701-byte/Loadster/pkg/config"
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

		logger.Log.Info("Core scenario parsing verified (Phase 1 complete)")

		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&targetURL, "url", "", "Target URL to test")
	runCmd.Flags().IntVar(&users, "users", 10, "Number of concurrent users")
	runCmd.Flags().StringVar(&duration, "duration", "30s", "Duration of test (e.g., 30s, 1m)")

	rootCmd.AddCommand(runCmd)
}
type RunCmd_Type = *cobra.Command // For linking symbol reference
