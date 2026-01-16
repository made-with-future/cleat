package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v0.1.0-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Cleat",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Cleat %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
