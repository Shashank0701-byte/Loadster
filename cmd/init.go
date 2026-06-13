package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultScenario = `target: https://api.example.com
headers:
  Content-Type: application/json
timeout: 5s
stages:
  - users: 10
    duration: 30s
  - users: 100
    duration: 1m
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new load testing scenario",
	Long:  `Creates a default test_scenario.yaml file in the current directory if it does not already exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		filename := "test_scenario.yaml"
		
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("❌ File '%s' already exists in the current directory.\n", filename)
			return
		}

		err := os.WriteFile(filename, []byte(defaultScenario), 0644)
		if err != nil {
			fmt.Printf("❌ Failed to create '%s': %v\n", filename, err)
			return
		}

		fmt.Printf("✅ Successfully created '%s'!\n", filename)
		fmt.Printf("👉 Next step: Run your load test with:\n")
		fmt.Println("   loadster run test_scenario.yaml")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
