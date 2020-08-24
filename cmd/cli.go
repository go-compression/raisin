package cmd

import (
	"flag"
	"fmt"
	engine "github.com/go-compression/raisin/engine"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	// "github.com/pkg/profile" // Profiling package
)

// Commands represents all possible commands that can be used durinv CLI invocation
var Commands = [...]string{"compress", "decompress", "benchmark", "help"}

// MainBehavior represents the main behavior function of the command line. This includes processing of flags and invoking of compression algorithms.
func MainBehavior() []engine.Result {
	// Profiling statement here V
	// defer profile.Start().Stop()
	// ^

	application := os.Args[0]

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	compressCmd := flag.Bool("compress", false, "Compress file")
	decompressCmd := flag.Bool("decompress", false, "Decompress file")
	benchmarkCmd := flag.Bool("benchmark", false, "Benchmark file")
	helpCmd := flag.Bool("help", false, "Help")

	commandArgs := make([]string, len(os.Args))
	copy(commandArgs, os.Args)
	if len(commandArgs) > 1 {
		commandArgs = append(commandArgs[1:2], "")
	}
	if commandArgs[0] == "-compress" || commandArgs[0] == "-decompress" ||
		commandArgs[0] == "-benchmark" || commandArgs[0] == "-help" {
		flag.CommandLine.Parse(commandArgs)
	}

	var generateHTML *bool
	if *benchmarkCmd {
		generateHTML = flag.Bool("generate", false, "Compile benchmark results as an html file")
	}

	commandsSelected := boolsTrue([]bool{*compressCmd, *decompressCmd, *benchmarkCmd, *helpCmd})

	if commandsSelected > 1 {
		errorWithMsg(fmt.Sprintf(
			"Please specify a single command. \n"))
	} else if commandsSelected < 1 {
		True := true
		if strings.HasSuffix(application, "grape") {
			decompressCmd = &True
		} else {
			compressCmd = &True
		}
		// errorWithMsg(fmt.Sprintf(
		// 	"Please specify at least one command. \n"))
	}

	if *helpCmd {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid commands include: \n\t %s\n", strings.Join(Commands[:], ", "))
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		return nil
	}

	// Get flag argument that is not a flag "-algorithm..."
	var file string
	if len(os.Args) > 1 {
		file = os.Args[1]
		for i := 2; len(file) > 0 && file[0] == '-'; i++ {
			file = os.Args[i]
		}
	}

	if file == "" && !strings.Contains(file, ",") {
		if *compressCmd {
			errorWithMsg("Please provide a file to be compressed\n")
		} else if *benchmarkCmd {
			errorWithMsg("Please provide a file to be benchmarked\n")
		} else {
			errorWithMsg("Please provide a file to be decompressed\n")
		}
	} else if strings.Contains(file, ",") {
		for _, filename := range strings.Split(file, ",") {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				errorWithMsg(fmt.Sprintf("Could not open file (likely does not exist): %s\n", filename))
			}
		}
	} else if _, err := os.Stat(file); os.IsNotExist(err) && file != "help" {
		errorWithMsg(fmt.Sprintf("Could not open file (likely does not exist): %s\n", file))
	}

	if *compressCmd {
		algorithm := flag.String("algorithm", "lzss,arithmetic",
			fmt.Sprintf("Which algorithm(s) to use, choices include: \n\t%s", strings.Join(engine.Engines[:], ", ")))

		files := strings.Split(file, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}

		var output, outputExtension *string
		if len(files) == 1 {
			output = flag.String("out", files[0]+".compressed", fmt.Sprintf("File name to output to"))
		} else {
			outputExtension = flag.String("outext", "compressed", fmt.Sprintf("File extension used for the result"))
		}

		deleteAfter := flag.Bool("delete", false, fmt.Sprintf("Delete file after compression"))

		flag.Parse()

		algorithms := strings.Split(*algorithm, ",")
		for i := range files {
			algorithms[i] = strings.TrimSpace(algorithms[i])
		}

		if len(files) > 1 {
			engine.CompressFiles(algorithms, files, "."+*outputExtension)
		} else {
			engine.CompressFile(algorithms, file, *output)
		}

		if *deleteAfter {
			deleteFiles(files)
		}
	} else if *decompressCmd {
		algorithm := flag.String("algorithm", "lzss,arithmetic",
			fmt.Sprintf("Which algorithm(s) to use, choices include: \n\t%s", strings.Join(engine.Engines[:], ", ")))

		files := strings.Split(file, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}

		var output, outputExtension *string
		if len(files) == 1 {
			ext := filepath.Ext(files[0])
			path := strings.TrimSuffix(files[0], ext)
			output = flag.String("out", path, fmt.Sprintf("File name to output to"))
		} else {
			outputExtension = flag.String("outext", "", fmt.Sprintf("File extension used for the result"))
		}

		deleteAfter := flag.Bool("delete", true, fmt.Sprintf("Delete file after compression"))

		flag.Parse()

		algorithms := strings.Split(*algorithm, ",")
		for i := range files {
			algorithms[i] = strings.TrimSpace(algorithms[i])
		}

		if len(files) > 1 {
			engine.DecompressFiles(algorithms, files, "."+*outputExtension)
		} else {
			engine.DecompressFile(algorithms, file, *output)
		}

		if *deleteAfter {
			deleteFiles(files)
		}
	} else if *benchmarkCmd {
		algorithm := flag.String("algorithm", "lzss,arithmetic,huffman,[lzss,arithmetic],gzip",
			fmt.Sprintf("Which algorithm(s) to use, choices include: \n\t%s", strings.Join(engine.Engines[:], ", ")))

		flag.Parse()

		if file == "help" {
			fmt.Fprintf(os.Stderr, "Flags:\n")
			flag.PrintDefaults()
			return nil
		}

		algorithms := parseAlgorithms(*algorithm)

		files := strings.Split(file, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}

		output, results := engine.BenchmarkSuite(files, algorithms, *generateHTML)
		if *generateHTML {
			err := ioutil.WriteFile("index.html", []byte(output), 0644)
			check(err)
			fmt.Println("Wrote table to index.html")
		}
		return results
	} else {
		errorWithMsg(fmt.Sprintf(
			"'%s' is not a valid command, "+
				"please provide a valid command, "+
				"possible commands include: \n\t %s\n", "", strings.Join(Commands[:], ", ")))
	}
	return nil
}

func parseAlgorithms(algorithmString string) (algorithms [][]string) {
	var buffer []byte
	var inLayer bool
	var layer []string
	for _, char := range []byte(algorithmString) {
		if char == ',' {
			if inLayer && len(buffer) > 0 {
				layer = append(layer, string(buffer))
			} else if len(buffer) > 0 {
				algorithms = append(algorithms, []string{string(buffer)})
			}
			buffer = make([]byte, 0)
		} else if char == '[' {
			inLayer = true
		} else if char == ']' {
			layer = append(layer, string(buffer))
			buffer = make([]byte, 0)
			inLayer = false
			algorithms = append(algorithms, layer)
			layer = make([]string, 0)
		} else {
			buffer = append(buffer, char)
		}
	}
	if len(buffer) > 0 {
		algorithms = append(algorithms, []string{string(buffer)})
	}
	return algorithms
}

func deleteFiles(files []string) {
	for _, file := range files {
		err := os.Remove(file)
		check(err)
	}
}

func boolsTrue(bools []bool) int {
	found := 0
	for _, boolean := range bools {
		if boolean {
			found++
		}
	}
	return found

}

func getPosAfterCommand(command string, args []string) int {
	for i, s := range args {
		if s == command {
			return i + 1
		}
	}
	return -1
}

func errorWithMsg(msg string) {
	fmt.Print(msg)
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
