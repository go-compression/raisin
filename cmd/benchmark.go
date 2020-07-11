package cmd

import (
	"errors"
	"fmt"
	"os"
	engine "github.com/mrfleap/custom-compression/engine"
	"github.com/spf13/cobra"
)

func init() {
	var maxSearchBuffer int

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
		Run: func (cmd *cobra.Command, args []string) {
			file := args[0] // Args[0] = file as a string
			engine.BenchmarkFile(file, maxSearchBuffer)
		},
	}

	benchmarkCmd.PersistentFlags().IntVarP(&maxSearchBuffer, "max search buffer length", "m", 1024, "maximum length of search buffer")
	rootCmd.AddCommand(benchmarkCmd)
}


