package main

import (
	"flag"
	"fmt"
	"strings"
	"os"
	"io/ioutil"
	engine "github.com/mrfleap/custom-compression/engine"
)

// https://github.com/spf13/cobra#getting-started

var Commands = [...]string{"compress", "decompress", "benchmark", "help"}

var algorithm string

func main() {
	flag.StringVar(&algorithm, "algorithm", "default", 
		fmt.Sprintf("Which algorithm to use, choices include: \n\t%s", strings.Join(engine.Engines[:], ", ")))
	
	_ = flag.NewFlagSet("compress", flag.ExitOnError)
	_ = flag.NewFlagSet("decompress", flag.ExitOnError)
	benchmarkCmd := flag.NewFlagSet("benchmark", flag.ExitOnError)

	generateHTML := benchmarkCmd.Bool("generate", false, "Compile benchmark results as an html file")

	flag.Parse()
	command := flag.Arg(0)
	if command == "" {
		errorMsg(fmt.Sprintf(
			"Please provide a valid command, possible commands include: \n\t %s\n", strings.Join(Commands[:], ", ")))
	}

	// Non compression commands
	switch command {
	case "help":
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid commands include: \n\t %s\n", strings.Join(Commands[:], ", "))
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		return
	}

	file := flag.Arg(1)
	for i := 2; len(file) > 0 && file[0] == '-'; i++ {
		file = flag.Arg(i)
	}
	if file == "" && !strings.Contains(file, ",") {
		errorMsg("Please provide a file to be compressed/decompressed")
	} else if _, err := os.Stat(file); os.IsNotExist(err) && file != "help" && !strings.Contains(file, ",") {
		errorMsg(fmt.Sprintf("Could not open file (likely does not exist): %s", file))
	}

	switch command {
	case "compress", "c":
		if algorithm == "default" { algorithm = "lzss" }
		engine.CompressFile(algorithm, file)
	case "decompress", "d":
		if algorithm == "default" { algorithm = "lzss" }
		engine.DecompressFile(algorithm, file)
	case "benchmark":
		benchmarkCmd.Parse(os.Args[2:])

		if file == "help" {
			fmt.Fprintf(os.Stderr, "Flags:\n")
			flag.PrintDefaults()
			return
		}

		if algorithm == "default" { algorithm = "suite" }

		files := strings.Split(file, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}

		if algorithm == "all" || algorithm == "suite" {			
			output := engine.BenchmarkSuite(files, engine.Suites[algorithm], *generateHTML)
			if *generateHTML {
				err := ioutil.WriteFile("index.html", []byte(output), 0644)
				check(err)
				fmt.Println("Wrote table to index.html")
			}
		} else {
			if len(files) > 0 {
				errorMsg("Cannot benchmark more than one file without using multiple algorithms currently")
			}
			engine.BenchmarkFile(algorithm, files[0], false)
		}
	default:
		errorMsg(fmt.Sprintf(
			"'&s' is not a valid command, " +
			"please provide a valid command, " +
			"possible commands include: \n\t %s\n", command, strings.Join(Commands[:], ", ")))
	}
}

func errorMsg(msg string) {
	fmt.Print(msg)
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}