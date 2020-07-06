package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "custom-compress",
		Short: "Custom compressor",
		Long:  `CLI tool to compress files using the custom compressor algorithm`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {

}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}
