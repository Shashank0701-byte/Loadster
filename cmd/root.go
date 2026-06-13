package cmd

import (
	"fmt"
	"os"

	"github.com/Shashank0701-byte/Loadster/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "loadster",
	Short: "Loadster is a modern distributed load testing platform",
	Long:  `Loadster is a cloud-native load testing engine built in Go. It can run local performance tests or orchestrate distributed execution across Kubernetes clusters.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := logger.InitLogger(debug); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		logger.Log.Debug("Logger initialized successfully")
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.loadster.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigType("yaml")
			viper.SetConfigName(".loadster")
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if logger.Log != nil {
			logger.Log.Info("Using config file", zap.String("file", viper.ConfigFileUsed()))
		}
	}
}
type RootCmd_Type = *cobra.Command // For linking symbol reference
