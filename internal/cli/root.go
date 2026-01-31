package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pbom",
	Short: "Pipeline Bill of Materials — track how artifacts reach production",
	Long: `PBOM captures the lineage of code as it transforms into a deployed artifact.

An SBOM tells you what is inside the artifact.
A PBOM tells you how it got there.

Use pbom to generate, validate, inspect, and push pipeline metadata
across your GitHub Actions and Kargo environments.`,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(webhookCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
