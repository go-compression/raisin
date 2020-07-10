package cmd

import (
	"errors"
	"fmt"
	"os"
	engine "github.com/mrfleap/custom-compression/engine"
	"github.com/spf13/cobra"
)

var decompressCmd = &cobra.Command{
	Use:   "decompress",
	Short: "Decompress a file using custom-compressor",
	Long:  `Decompress a file using custom-compressor`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a file argument to compress")
		}
		file := args[0]
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("Could not open file (likely does not exist): %s", args[0])
		} else {
			return nil
		}

	},
	Run: decompress,
}

func init() {
	rootCmd.AddCommand(decompressCmd)
}

func decompress(cmd *cobra.Command, args []string) {
	file := args[0] // Args[0] = file as a string
	engine.DecompressFile(file)
}
