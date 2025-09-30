package interval

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// SparkCharacters are the default characters used to render the sparkline.
var SparkCharacters = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// SparkColor represents an ANSI color code for sparklines.
type SparkColor string

// Predefined colors for the --color flag.
const (
	ColorNone    SparkColor = ""
	ColorRed     SparkColor = "\033[31m"
	ColorGreen   SparkColor = "\033[32m"
	ColorYellow  SparkColor = "\033[33m"
	ColorBlue    SparkColor = "\033[34m"
	ColorMagenta SparkColor = "\033[35m"
	ColorCyan    SparkColor = "\033[36m"
	ColorReset   SparkColor = "\033[0m"
)

// SparkConfig holds the configuration for generating a sparkline.
type SparkConfig struct {
	// For fixed-interval mode
	Min, Max float64
	HasMin   bool
	HasMax   bool

	// For sliding window animation
	Width int

	// For color output
	Color SparkColor
}

// ParseColor translates a string name into a SparkColor.
func ParseColor(s string) (SparkColor, error) {
	switch strings.ToLower(s) {
	case "":
		return ColorNone, nil
	case "red":
		return ColorRed, nil
	case "green":
		return ColorGreen, nil
	case "yellow":
		return ColorYellow, nil
	case "blue":
		return ColorBlue, nil
	case "magenta":
		return ColorMagenta, nil
	case "cyan":
		return ColorCyan, nil
	default:
		return ColorNone, fmt.Errorf("unknown color: %s", s)
	}
}

// GenerateSparkline reads numbers from the scanner and writes a sparkline to the writer.
func GenerateSparkline(scanner *bufio.Scanner, writer io.Writer, config SparkConfig) error {
	var numbers []float64
	min, max := math.Inf(1), math.Inf(-1)

	// If interval is fixed, use it.
	if config.HasMin {
		min = config.Min
	}
	if config.HasMax {
		max = config.Max
	}

	// Read all numbers from scanner
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// Split the line by whitespace to handle space-separated numbers
		fields := strings.Fields(line)
		for _, field := range fields {
			val, err := strconv.ParseFloat(field, 64)
			if err != nil {
				// Silently skip non-numeric fields
				continue
			}
			numbers = append(numbers, val)

			// If interval is not fixed, update min/max.
			if !config.HasMin && val < min {
				min = val
			}
			if !config.HasMax && val > max {
				max = val
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from input: %w", err)
	}

	if len(numbers) == 0 {
		return nil // Nothing to do
	}

	var output strings.Builder
	if config.Color != ColorNone {
		output.WriteString(string(config.Color))
	}

	// Determine the subset of numbers to render for the sliding window
	startIndex := 0
	if config.Width > 0 && len(numbers) > config.Width {
		startIndex = len(numbers) - config.Width
	}
	renderableNumbers := numbers[startIndex:]

	for _, num := range renderableNumbers {
		var charIndex float64
		// If the range is zero, all numbers are the same.
		// Default to the lowest character index.
		if max == min {
			charIndex = 0
		} else {
			// Remap the number to the index of the spark character
			var err error
			charIndex, err = Remap(num, min, max, 0, float64(len(SparkCharacters)-1))
			if err != nil {
				// This should not happen if max != min, but as a safeguard:
				charIndex = 0
			}
		}

		// Clamp the index to be within bounds
		clampedIndex := int(Limit(charIndex, 0, float64(len(SparkCharacters)-1)))
		output.WriteRune(SparkCharacters[clampedIndex])
	}

	if config.Color != ColorNone {
		output.WriteString(string(ColorReset))
	}

	// For sliding window, add carriage return
	if config.Width > 0 {
		fmt.Fprintf(writer, "\r%s", output.String())
	} else {
		fmt.Fprint(writer, output.String())
	}

	return nil
}
