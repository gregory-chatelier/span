package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gregory-chatelier/span/interval"
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

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [operation] [args...]

**span** is a command line tool for interval manipulation.

Global Flags:
  -f, --format string
    	Output format for numbers (default "%%g")
  --version
    	Print version information and exit

Operational Flags (only one can be used at a time):
  -r, --remap <src_a> <src_b> <dst_a> <dst_b>
  -l, --limit <min> <max>
  -E, --encompass
  -n, --divide <steps> <a> <b>
  -e, --eval <a> <b>
  -d, --deval <a> <b>
  -R, --random <count> <a> <b>
  -S, --snap <steps> <a> <b>
  -s, --subintervals <steps> <a> <b>
`, os.Args[0])
	}

	format := fs.String("f", "%g", "(see usage)")
	versionFlag := fs.Bool("version", false, "(see usage)")

	remapFlag := fs.Bool("r", false, "")
	fs.BoolVar(remapFlag, "remap", false, "")
	limitFlag := fs.Bool("l", false, "")
	fs.BoolVar(limitFlag, "limit", false, "")
	encompassFlag := fs.Bool("E", false, "")
	fs.BoolVar(encompassFlag, "encompass", false, "")
	divideFlag := fs.Bool("n", false, "")
	fs.BoolVar(divideFlag, "divide", false, "")
	evalFlag := fs.Bool("e", false, "")
	fs.BoolVar(evalFlag, "eval", false, "")
	devalFlag := fs.Bool("d", false, "")
	fs.BoolVar(devalFlag, "deval", false, "")
	randomFlag := fs.Bool("R", false, "")
	fs.BoolVar(randomFlag, "random", false, "")
	snapFlag := fs.Bool("S", false, "")
	fs.BoolVar(snapFlag, "snap", false, "")
	subintervalsFlag := fs.Bool("s", false, "")
	fs.BoolVar(subintervalsFlag, "subintervals", false, "")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *versionFlag {
		fmt.Println(Version)
		os.Exit(0)
	}

	opCount := 0
	if *remapFlag {
		opCount++
	}
	if *limitFlag {
		opCount++
	}
	if *encompassFlag {
		opCount++
	}
	if *divideFlag {
		opCount++
	}
	if *evalFlag {
		opCount++
	}
	if *devalFlag {
		opCount++
	}
	if *randomFlag {
		opCount++
	}
	if *snapFlag {
		opCount++
	}
	if *subintervalsFlag {
		opCount++
	}

	if opCount > 1 {
		fmt.Fprintln(os.Stderr, "Error: Only one operational flag can be used at a time.")
		fs.Usage()
		os.Exit(1)
	}

	if opCount == 0 {
		stat, _ := os.Stdin.Stat()
		if len(fs.Args()) == 0 && (stat.Mode()&os.ModeCharDevice) != 0 {
			fs.Usage()
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "Error: An operational flag is required.")
		fs.Usage()
		os.Exit(1)
	}

	args := fs.Args()

	switch {
	case *remapFlag:
		if len(args) != 4 {
			fmt.Fprintln(os.Stderr, "Error: -r, --remap requires 4 arguments: <src_a> <src_b> <dst_a> <dst_b>")
			fs.Usage()
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
			fs.Usage()
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
			fs.Usage()
			os.Exit(1)
		}

		minVal, maxVal, err := interval.Encompass(bufio.NewScanner(os.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Encompass outputs two values, min and max, separated by a space.
		// The format flag applies to each number.
		outputFormat := *format + " " + *format + "\n"
		fmt.Printf(outputFormat, minVal, maxVal)
	case *divideFlag:
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: -n, --divide requires 3 arguments: <steps> <a> <b>")
			fs.Usage()
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
			fs.Usage()
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
			fs.Usage()
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
			fs.Usage()
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
			fs.Usage()
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
			fs.Usage()
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

		// Subintervals outputs two values per line, separated by a space.
		// The format flag applies to each number.
		outputFormat := *format + " " + *format + "\n"
		for _, res := range results {
			fmt.Printf(outputFormat, res[0], res[1])
		}
	}
}
