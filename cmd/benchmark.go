package cmd

import (
	"errors"
	"fmt"
	"os"
	engine "github.com/mrfleap/custom-compression/engine"
	"github.com/spf13/cobra"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark a file using custom-compressor",
	Long:  `Benchmark a file using custom-compressor`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a file argument to benchmark")
		}
		file := args[0]
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("Could not open file (likely does not exist): %s", args[0])
		} else {
			return nil
		}
	},
	Run: benchmark,
}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
}

func benchmark(cmd *cobra.Command, args []string) {
	file := args[0] // Args[0] = file as a string
	engine.BenchmarkFile(file)
}
