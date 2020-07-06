package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of the custom-compressor",
	Long:  `Prints version number and build info`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Custom Compressor v0.1")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
