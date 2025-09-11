package interval

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
)

// Deval returns the parameter 't' of a value within an interval [a, b].
// It returns an error if the interval has a delta of zero (a == b).
func Deval(val, a, b float64) (float64, error) {
	// Handle NaN and Inf inputs
	if math.IsNaN(val) || math.IsNaN(a) || math.IsNaN(b) {
		return 0, fmt.Errorf("cannot de-evaluate: NaN values are not supported")
	}
	if math.IsInf(val, 0) || math.IsInf(a, 0) || math.IsInf(b, 0) {
		return 0, fmt.Errorf("cannot de-evaluate: infinite values are not supported")
	}

	delta := b - a
	const epsilon = 1e-15

	if math.Abs(delta) < epsilon {
		if math.Abs(val-a) < epsilon {
			return 0, nil
		}
		return 0, fmt.Errorf("cannot de-evaluate in an interval with near-zero delta")
	}

	t := (val - a) / delta
	return t, nil
}

// Eval evaluates a parameter 't' within the interval [a, b].
func Eval(t, a, b float64) float64 {
	if math.IsNaN(t) || math.IsNaN(a) || math.IsNaN(b) {
		return math.NaN()
	}
	return a + (b-a)*t
}

// Remap translates a value from a source interval [srcA, srcB] to a target interval [dstA, dstB].
// It returns an error if the source interval has a delta of zero.
func Remap(val, srcA, srcB, dstA, dstB float64) (float64, error) {
	if math.IsNaN(val) || math.IsNaN(srcA) || math.IsNaN(srcB) ||
		math.IsNaN(dstA) || math.IsNaN(dstB) {
		return 0, fmt.Errorf("cannot remap: NaN values are not supported")
	}

	if math.IsInf(val, 0) || math.IsInf(srcA, 0) || math.IsInf(srcB, 0) ||
		math.IsInf(dstA, 0) || math.IsInf(dstB, 0) {
		return 0, fmt.Errorf("cannot remap: infinite values are not supported")
	}

	t, err := Deval(val, srcA, srcB)
	if err != nil {
		return 0, fmt.Errorf("cannot remap from a source interval with zero delta")
	}
	return Eval(t, dstA, dstB), nil
}

// Limit restricts (clamps) a value to be within the interval [min, max].
// It correctly handles cases where min > max by ordering them first.
func Limit(val, min, max float64) float64 {
	if math.IsNaN(val) || math.IsNaN(min) || math.IsNaN(max) {
		return math.NaN()
	}

	if math.IsInf(val, 0) {
		if math.IsInf(val, 1) {
			return max
		}
		return min
	}

	if min > max {
		min, max = max, min // Ensure min is less than or equal to max
	}
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// Snap snaps a value to the nearest point on a grid defined by an interval and a number of steps.
func Snap(val float64, steps int, a, b float64) (float64, error) {
	if math.IsNaN(val) || math.IsNaN(a) || math.IsNaN(b) {
		return 0, fmt.Errorf("cannot snap: NaN values are not supported")
	}
	if math.IsInf(val, 0) || math.IsInf(a, 0) || math.IsInf(b, 0) {
		return 0, fmt.Errorf("cannot snap: infinite values are not supported")
	}

	if steps <= 0 {
		return 0, fmt.Errorf("steps must be a positive integer")
	}

	// Ensure the interval is ordered for clamping
	min, max := a, b
	if min > max {
		min, max = max, min
	}
	if val <= min {
		return min, nil
	}
	if val >= max {
		return max, nil
	}

	// If the interval is zero-width, all points snap to 'a'
	if a == b {
		return a, nil
	}

	t, err := Deval(val, a, b)
	if err != nil {
		return 0, fmt.Errorf("bin error: %v", err)
	}

	stepIndex := math.Round(t * float64(steps))
	snappedT := stepIndex / float64(steps)

	return Eval(snappedT, a, b), nil
}

// Divide generates a sequence of numbers by dividing an interval into a number of steps.
// It does not include the end point (b) in the sequence.
func Divide(steps int, a, b float64) ([]float64, error) {
	if math.IsNaN(a) || math.IsNaN(b) {
		return nil, fmt.Errorf("cannot divide: NaN values are not supported")
	}
	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		return nil, fmt.Errorf("cannot divide: infinite values are not supported")
	}
	if steps < 0 {
		return nil, fmt.Errorf("steps cannot be negative")
	}
	if steps == 0 {
		return []float64{}, nil
	}

	results := make([]float64, steps)
	if a == b {
		for i := 0; i < steps; i++ {
			results[i] = a
		}
		return results, nil
	}

	stepSize := (b - a) / float64(steps)
	for i := 0; i < steps; i++ {
		results[i] = a + (float64(i) * stepSize)
	}

	return results, nil
}

// Random generates a sequence of random numbers within an interval [a, b].
// It uses the provided rand.Rand source for testability.
func Random(r *rand.Rand, count int, a, b float64) ([]float64, error) {
	if math.IsNaN(a) || math.IsNaN(b) {
		return nil, fmt.Errorf("cannot generate random values: NaN bounds")
	}
	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		return nil, fmt.Errorf("cannot generate random values: infinite bounds")
	}
	if count < 0 {
		return nil, fmt.Errorf("count cannot be negative")
	}
	if count == 0 {
		return []float64{}, nil
	}

	start, end := a, b
	if start > end {
		start, end = end, start
	}

	results := make([]float64, count)
	for i := range results {
		t := r.Float64()
		results[i] = Eval(t, start, end)
	}

	return results, nil
}

// Subintervals generates a sequence of interval pairs.
func Subintervals(steps int, a, b float64) ([][2]float64, error) {
	if math.IsNaN(a) || math.IsNaN(b) {
		return nil, fmt.Errorf("cannot create subintervals: NaN bounds")
	}
	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		return nil, fmt.Errorf("cannot create subintervals: infinite bounds")
	}
	if steps < 0 {
		return nil, fmt.Errorf("steps cannot be negative")
	}
	if steps == 0 {
		return [][2]float64{}, nil
	}

	results := make([][2]float64, steps)
	if a == b {
		for i := 0; i < steps; i++ {
			results[i] = [2]float64{a, a}
		}
		return results, nil
	}

	stepSize := (b - a) / float64(steps)
	for i := 0; i < steps; i++ {
		start := a + (float64(i) * stepSize)
		end := a + (float64(i+1) * stepSize)
		results[i] = [2]float64{start, end}
	}

	return results, nil
}

// Encompass reads a stream of numbers and returns the minimum and maximum values.
// It returns an error if no valid numbers are found in the input.
func Encompass(scanner *bufio.Scanner) (float64, float64, error) {
	minVal := math.Inf(1)  // Positive infinity
	maxVal := math.Inf(-1) // Negative infinity
	foundNumber := false

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

		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
		foundNumber = true
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("error reading from input: %v", err)
	}

	if !foundNumber {
		return 0, 0, fmt.Errorf("no numbers found in input")
	}

	return minVal, maxVal, nil
}
