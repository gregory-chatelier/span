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
	Min, Max float64
	HasMin   bool
	HasMax   bool
	Width    int
	Color    SparkColor
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

// GenerateSparkline is a dispatcher that chooses the correct sparkline generation method.
func GenerateSparkline(scanner *bufio.Scanner, writer io.Writer, config SparkConfig) error {
	// Use streaming for fixed-width or fixed-interval modes.
	if config.Width > 0 || config.HasMin {
		return generateSparklineStream(scanner, writer, config)
	}

	// For auto-scaled, growing sparklines, we must buffer.
	numbers, err := readAllNumbers(scanner)
	if err != nil {
		return err
	}
	return generateSparklineFromSlice(numbers, writer, config)
}

// generateSparklineFromSlice renders a sparkline from a slice of numbers already in memory.
func generateSparklineFromSlice(numbers []float64, writer io.Writer, config SparkConfig) error {
	if len(numbers) == 0 {
		return nil
	}

	min, max := math.Inf(1), math.Inf(-1)
	if config.HasMin {
		min = config.Min
	} else {
		// Calculate min from slice
		for _, n := range numbers {
			if n < min {
				min = n
			}
		}
	}
	if config.HasMax {
		max = config.Max
	} else {
		// Calculate max from slice
		for _, n := range numbers {
			if n > max {
				max = n
			}
		}
	}

	var output strings.Builder

	for _, num := range numbers {
		charIndex := 0.0
		if max > min {
			charIndex, _ = Remap(num, min, max, 0, float64(len(SparkCharacters)-1))
		}
		clampedIndex := int(Limit(charIndex, 0, float64(len(SparkCharacters)-1)))
		output.WriteRune(SparkCharacters[clampedIndex])
	}

	fmt.Fprint(writer, applyColor(output.String(), config.Color))
	return nil
}

// generateSparklineStream renders a sparkline by processing the input stream number by number.
func generateSparklineStream(scanner *bufio.Scanner, writer io.Writer, config SparkConfig) error {
	var buffer *circularBuffer
	if config.Width > 0 {
		buffer = newCircularBuffer(config.Width)
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		for _, field := range fields {
			val, err := strconv.ParseFloat(field, 64)
			if err != nil {
				continue // Skip non-numeric fields
			}

			if config.Width > 0 {
				buffer.Add(val)
				min, max := config.Min, config.Max
				if !config.HasMin { // If no fixed interval, calculate from buffer
					min, max = buffer.MinMax()
				}
				renderSlidingWindow(writer, buffer, min, max, config.Color)
			} else { // Growing sparkline with fixed interval
				var charBuilder strings.Builder
				renderGrowingCharacter(&charBuilder, val, config.Min, config.Max)
				fmt.Fprint(writer, applyColor(charBuilder.String(), config.Color))
			}
		}
	}
	return scanner.Err()
}

func renderSlidingWindow(writer io.Writer, buffer *circularBuffer, min, max float64, color SparkColor) {
	var output strings.Builder
	numbers := buffer.GetAll()

	for _, num := range numbers {
		charIndex := 0.0
		if max > min {
			charIndex, _ = Remap(num, min, max, 0, float64(len(SparkCharacters)-1))
		}
		clampedIndex := int(Limit(charIndex, 0, float64(len(SparkCharacters)-1)))
		output.WriteRune(SparkCharacters[clampedIndex])
	}

	fmt.Fprintf(writer, "\r%s", applyColor(output.String(), color))
}

func renderGrowingCharacter(builder *strings.Builder, val, min, max float64) {
	charIndex := 0.0
	if max > min {
		charIndex, _ = Remap(val, min, max, 0, float64(len(SparkCharacters)-1))
	}
	clampedIndex := int(Limit(charIndex, 0, float64(len(SparkCharacters)-1)))
	builder.WriteRune(SparkCharacters[clampedIndex])
}

func readAllNumbers(scanner *bufio.Scanner) ([]float64, error) {
	var numbers []float64
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		for _, field := range fields {
			val, err := strconv.ParseFloat(field, 64)
			if err != nil {
				continue
			}
			numbers = append(numbers, val)
		}
	}
	return numbers, scanner.Err()
}

func applyColor(s string, color SparkColor) string {
	if color == ColorNone {
		return s
	}
	return string(color) + s + string(ColorReset)
}

type circularBuffer struct {
	data []float64
	head int
	full bool
}

func newCircularBuffer(size int) *circularBuffer {
	if size <= 0 {
		size = 1
	}
	return &circularBuffer{
		data: make([]float64, size),
	}
}

func (cb *circularBuffer) Add(val float64) {
	cb.data[cb.head] = val
	cb.head = (cb.head + 1) % len(cb.data)
	if cb.head == 0 && !cb.full {
		cb.full = true
	}
}

func (cb *circularBuffer) GetAll() []float64 {
	if !cb.full {
		return cb.data[:cb.head]
	}
	// The buffer is full and wrapped, so we need to reorder it
	reordered := make([]float64, len(cb.data))
	copy(reordered, cb.data[cb.head:])
	copy(reordered[len(cb.data)-cb.head:], cb.data[:cb.head])
	return reordered
}

func (cb *circularBuffer) MinMax() (float64, float64) {
	data := cb.GetAll()
	if len(data) == 0 {
		return 0, 0
	}
	min, max := data[0], data[0]
	for _, v := range data[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}
