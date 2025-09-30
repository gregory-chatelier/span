package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gregory-chatelier/span/interval"
	flag "github.com/spf13/pflag"
)

// Version will be set during the build process
var Version = "v0.0.1-dev"

// processFunc defines a function signature for processing a single float64 value.
// It's used to pass different interval operations to the stream processor.
type processFunc func(float64) (float64, error)

// processStream reads numbers from stdin, applies a processing function to each,
// and prints the result to stdout.
func processStream(format string, proc processFunc) {
	outputFormat := format + "\n"
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		val, err := strconv.ParseFloat(line, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse input value '%s', skipping: %v\n", line, err)
			continue
		}

		processedVal, err := proc(val)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not process value %f, skipping: %v\n", val, err)
			continue
		}
		fmt.Printf(outputFormat, processedVal)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `NAME:
    span - A Unix-style tool for interval manipulation.

SYNOPSIS:
    span [operation] [flags] [arguments...]
    command | span [operation] [flags] [arguments...]

DESCRIPTION:
    span reads numbers from stdin, performs an interval-based mathematical
    operation, and prints the transformed numbers to stdout. It is designed
    to be a simple, composable tool in the Unix tradition.

OPTIONS:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage

	// --- Global Flags ---
	format := flag.StringP("format", "f", "%g", "Specifies the printf format for floating-point output (e.g., \"%.3f\").")
	versionFlag := flag.Bool("version", false, "Prints version information and exits.")

	// --- Operation Flags ---
	remapFlag := flag.BoolP("remap", "r", false, "Remaps a value from a source interval to a target interval.")
	limitFlag := flag.BoolP("limit", "l", false, "Restricts (clamps) a value to a given interval.")
	encompassFlag := flag.BoolP("encompass", "E", false, "Reads a stream and outputs the min and max values.")
	divideFlag := flag.BoolP("divide", "n", false, "Generates a sequence by dividing an interval.")
	evalFlag := flag.BoolP("eval", "e", false, "Evaluates a parameter 't' (0-1) within an interval.")
	devalFlag := flag.BoolP("deval", "d", false, "De-evaluates a number to a parameter 't' (0-1).")
	randomFlag := flag.BoolP("random", "R", false, "Generates <count> random numbers in an interval.")
	snapFlag := flag.BoolP("snap", "S", false, "Snaps input values to the nearest point on a grid.")
	subintervalsFlag := flag.BoolP("subintervals", "s", false, "Divides an interval into <steps> equal subintervals.")
	sparkFlag := flag.Bool("spark", false, "Generates a sparkline visualization from a stream of numbers.")

	// --- Spark-specific Flags ---
	sparkWidth := flag.Int("spark-width", 0, "For --spark: fixed-width sliding window animation")
	sparkColor := flag.String("spark-color", "", "For --spark: sparkline color (red, green, blue, etc.)")

	flag.Parse()

	if *versionFlag {
		fmt.Println(Version)
		os.Exit(0)
	}

	opCount := 0
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "remap", "limit", "encompass", "divide", "eval", "deval", "random", "snap", "subintervals", "spark":
			opCount++
		}
	})

	args := flag.Args()

	if opCount > 1 {
		fmt.Fprintln(os.Stderr, "Error: Only one operational flag can be used at a time.")
		usage()
		os.Exit(1)
	}

	if opCount == 0 {
		stat, _ := os.Stdin.Stat()
		if len(args) == 0 && (stat.Mode()&os.ModeCharDevice) != 0 {
			usage()
			os.Exit(0)
		}
		if (stat.Mode() & os.ModeNamedPipe) != 0 {
			*sparkFlag = true
		} else {
			fmt.Fprintln(os.Stderr, "Error: An operational flag is required.")
			usage()
			os.Exit(1)
		}
	}

	switch {
	case *sparkFlag:
		config := interval.SparkConfig{
			Width: *sparkWidth,
		}

		var err error
		config.Color, err = interval.ParseColor(*sparkColor)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if len(args) == 2 {
			config.Min, err = strconv.ParseFloat(args[0], 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not parse min value '%s'\n", args[0])
				os.Exit(1)
			}
			config.Max, err = strconv.ParseFloat(args[1], 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not parse max value '%s'\n", args[1])
				os.Exit(1)
			}
			config.HasMin = true
			config.HasMax = true
		} else if len(args) != 0 {
			fmt.Fprintln(os.Stderr, "Error: --spark requires 0 or 2 arguments: [<min> <max>]")
			usage()
			os.Exit(1)
		}

		scanner := bufio.NewScanner(os.Stdin)
		err = interval.GenerateSparkline(scanner, os.Stdout, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating sparkline: %v\n", err)
			os.Exit(1)
		}

		if config.Width == 0 {
			fmt.Println()
		}
	case *remapFlag:
		if len(args) != 4 {
			fmt.Fprintln(os.Stderr, "Error: -r, --remap requires 4 arguments: <src_a> <src_b> <dst_a> <dst_b>")
			usage()
			os.Exit(1)
		}
		srcA, errA := strconv.ParseFloat(args[0], 64)
		srcB, errB := strconv.ParseFloat(args[1], 64)
		dstA, errC := strconv.ParseFloat(args[2], 64)
		dstB, errD := strconv.ParseFloat(args[3], 64)
		if errA != nil || errB != nil || errC != nil || errD != nil {
			fmt.Fprintln(os.Stderr, "Error: could not parse all remap arguments as numbers.")
			os.Exit(1)
		}
		processStream(*format, func(val float64) (float64, error) {
			return interval.Remap(val, srcA, srcB, dstA, dstB)
		})

	case *limitFlag:
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Error: -l, --limit requires 2 arguments: <min> <max>")
			usage()
			os.Exit(1)
		}
		min, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not parse min value '%s': %v\n", args[0], err)
			os.Exit(1)
		}
		max, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not parse max value '%s': %v\n", args[1], err)
			os.Exit(1)
		}
		processStream(*format, func(val float64) (float64, error) {
			return interval.Limit(val, min, max), nil
		})

	case *encompassFlag:
		if len(args) != 0 {
			fmt.Fprintln(os.Stderr, "Error: -E, --encompass takes no arguments.")
			usage()
			os.Exit(1)
		}

		minVal, maxVal, err := interval.Encompass(bufio.NewScanner(os.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputFormat := *format + " " + *format + "\n"
		fmt.Printf(outputFormat, minVal, maxVal)
	case *divideFlag:
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: -n, --divide requires 3 arguments: <steps> <a> <b>")
			usage()
			os.Exit(1)
		}
		steps, errS := strconv.Atoi(args[0])
		a, errA := strconv.ParseFloat(args[1], 64)
		b, errB := strconv.ParseFloat(args[2], 64)
		if errS != nil || errA != nil || errB != nil {
			fmt.Fprintln(os.Stderr, "Error: could not parse all divide arguments.")
			os.Exit(1)
		}

		results, err := interval.Divide(steps, a, b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputFormat := *format + "\n"
		for _, res := range results {
			fmt.Printf(outputFormat, res)
		}
	case *evalFlag:
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Error: -e, --eval requires 2 arguments: <a> <b>")
			usage()
			os.Exit(1)
		}
		a, errA := strconv.ParseFloat(args[0], 64)
		b, errB := strconv.ParseFloat(args[1], 64)
		if errA != nil || errB != nil {
			fmt.Fprintln(os.Stderr, "Error: could not parse all eval arguments as numbers.")
			os.Exit(1)
		}
		processStream(*format, func(val float64) (float64, error) {
			return interval.Eval(val, a, b), nil
		})
	case *devalFlag:
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Error: -d, --deval requires 2 arguments: <a> <b>")
			usage()
			os.Exit(1)
		}
		a, errA := strconv.ParseFloat(args[0], 64)
		b, errB := strconv.ParseFloat(args[1], 64)
		if errA != nil || errB != nil {
			fmt.Fprintln(os.Stderr, "Error: could not parse all deval arguments as numbers.")
			os.Exit(1)
		}
		processStream(*format, func(val float64) (float64, error) {
			return interval.Deval(val, a, b)
		})
	case *randomFlag:
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: -R, --random requires 3 arguments: <count> <a> <b>")
			usage()
			os.Exit(1)
		}
		count, errC := strconv.Atoi(args[0])
		a, errA := strconv.ParseFloat(args[1], 64)
		b, errB := strconv.ParseFloat(args[2], 64)
		if errC != nil || errA != nil || errB != nil {
			fmt.Fprintln(os.Stderr, "Error: could not parse all random arguments.")
			os.Exit(1)
		}

		// Seed the generator for non-deterministic output
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		results, err := interval.Random(r, count, a, b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputFormat := *format + "\n"
		for _, res := range results {
			fmt.Printf(outputFormat, res)
		}
	case *snapFlag:
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: -S, --snap requires 3 arguments: <steps> <a> <b>")
			usage()
			os.Exit(1)
		}
		steps, errS := strconv.Atoi(args[0])
		a, errA := strconv.ParseFloat(args[1], 64)
		b, errB := strconv.ParseFloat(args[2], 64)
		if errS != nil || errA != nil || errB != nil {
			fmt.Fprintln(os.Stderr, "Error: could not parse all snap arguments.")
			os.Exit(1)
		}
		processStream(*format, func(val float64) (float64, error) {
			return interval.Snap(val, steps, a, b)
		})
	case *subintervalsFlag:
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: -s, --subintervals requires 3 arguments: <steps> <a> <b>")
			usage()
			os.Exit(1)
		}
		steps, errS := strconv.Atoi(args[0])
		a, errA := strconv.ParseFloat(args[1], 64)
		b, errB := strconv.ParseFloat(args[2], 64)
		if errS != nil || errA != nil || errB != nil {
			fmt.Fprintln(os.Stderr, "Error: could not parse all subintervals arguments.")
			os.Exit(1)
		}

		results, err := interval.Subintervals(steps, a, b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		outputFormat := *format + " " + *format + "\n"
		for _, res := range results {
			fmt.Printf(outputFormat, res[0], res[1])
		}
	}
}
