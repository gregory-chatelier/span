package interval

import (
	"bufio"
	"bytes"
	"math"
	"math/rand"
	"testing"
)

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

// Helper to compare slices of floats
func slicesAlmostEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !almostEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestLimit(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		min     float64
		max     float64
		want    float64
	}{
		{"val inside", 5, 0, 10, 5},
		{"val below", -5, 0, 10, 0},
		{"val above", 15, 0, 10, 10},
		{"val at min", 0, 0, 10, 0},
		{"val at max", 10, 0, 10, 10},
		{"inverted interval, val inside", 5, 10, 0, 5},
		{"inverted interval, val below", -5, 10, 0, 0},
		{"inverted interval, val above", 15, 10, 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Limit(tt.val, tt.min, tt.max); got != tt.want {
				t.Errorf("Limit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		name string
		t    float64
		a    float64
		b    float64
		want float64
	}{
		{"midpoint", 0.5, 0, 100, 50},
		{"start", 0, 0, 100, 0},
		{"end", 1, 0, 100, 100},
		{"outside below", -0.5, 0, 100, -50},
		{"outside above", 1.5, 0, 100, 150},
		{"inverted interval", 0.5, 100, 0, 50},
		{"zero delta", 0.5, 10, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Eval(tt.t, tt.a, tt.b); !almostEqual(got, tt.want) {
				t.Errorf("Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeval(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		a       float64
		b       float64
		want    float64
		wantErr bool
	}{
		{"midpoint", 50, 0, 100, 0.5, false},
		{"start", 0, 0, 100, 0, false},
		{"end", 100, 0, 100, 1, false},
		{"outside below", -50, 0, 100, -0.5, false},
		{"outside above", 150, 0, 100, 1.5, false},
		{"inverted interval", 50, 100, 0, 0.5, false},
		{"zero delta, error", 50, 10, 10, 0, true},
		{"zero delta, no error", 10, 10, 10, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Deval(tt.val, tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !almostEqual(got, tt.want) {
				t.Errorf("Deval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemap(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		srcA    float64
		srcB    float64
		dstA    float64
		dstB    float64
		want    float64
		wantErr bool
	}{
		{"simple remap", 5, 0, 10, 100, 200, 150, false},
		{"to inverted", 5, 0, 10, 200, 100, 150, false},
		{"from inverted", 5, 10, 0, 100, 200, 150, false},
		{"full inverted", 5, 10, 0, 200, 100, 150, false},
		{"zero delta src, error", 5, 10, 10, 100, 200, 0, true},
		{"zero delta src, no error", 10, 10, 10, 100, 200, 100, false},
		{"zero delta dst", 5, 0, 10, 100, 100, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Remap(tt.val, tt.srcA, tt.srcB, tt.dstA, tt.dstB)
			if (err != nil) != tt.wantErr {
				t.Errorf("Remap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !almostEqual(got, tt.want) {
				t.Errorf("Remap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnap(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		steps   int
		a       float64
		b       float64
		want    float64
		wantErr bool
	}{
		{"snap up", 4.8, 10, 0, 10, 5, false},
		{"snap down", 4.2, 10, 0, 10, 4, false},
		{"midpoint snap", 4.5, 10, 0, 10, 5, false},
		{"already on grid", 3, 10, 0, 10, 3, false},
		{"outside above", 12, 10, 0, 10, 10, false},
		{"outside below", -2, 10, 0, 10, 0, false},
		{"inverted interval", 5.2, 10, 10, 0, 5, false},
		{"zero steps", 5, 0, 0, 10, 0, true},
		{"negative steps", 5, -1, 0, 10, 0, true},
		{"zero delta interval", 5, 10, 10, 10, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Snap(tt.val, tt.steps, tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Snap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !almostEqual(got, tt.want) {
				t.Errorf("Snap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		steps   int
		a       float64
		b       float64
		want    []float64
		wantErr bool
	}{
		{"positive steps", 4, 0, 1, []float64{0, 0.25, 0.5, 0.75}, false},
		{"zero steps", 0, 0, 10, []float64{}, false},
		{"one step", 1, 10, 20, []float64{10}, false},
		{"inverted interval", 4, 1, 0, []float64{1, 0.75, 0.5, 0.25}, false},
		{"zero delta", 5, 10, 10, []float64{10, 10, 10, 10, 10}, false},
		{"negative steps", -1, 0, 10, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.steps, tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Divide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !slicesAlmostEqual(got, tt.want) {
				t.Errorf("Divide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandom(t *testing.T) {
	// Use a fixed seed for deterministic output
	src := rand.NewSource(42)
	r := rand.New(src)

	t.Run("correct count and range", func(t *testing.T) {
		count := 100
		a := -10.0
		b := 10.0
		results, err := Random(r, count, a, b)
		if err != nil {
			t.Fatalf("Random() returned an unexpected error: %v", err)
		}
		if len(results) != count {
			t.Fatalf("Random() len = %v, want %v", len(results), count)
		}

		for _, val := range results {
			if val < a || val > b {
				// Handle inverted intervals
				if (b < a) && (val < b || val > a) {
					t.Errorf("Random() value %v is outside inverted interval [%v, %v]", val, b, a)
				}
				if a < b {
					t.Errorf("Random() value %v is outside interval [%v, %v]", val, a, b)
				}
			}
		}
	})

	t.Run("negative count", func(t *testing.T) {
		_, err := Random(r, -1, 0, 1)
		if err == nil {
			t.Error("Random() expected an error for negative count, but got nil")
		}
	})

	t.Run("zero count", func(t *testing.T) {
		results, err := Random(r, 0, 0, 1)
		if err != nil {
			t.Errorf("Random() returned an unexpected error for zero count: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Random() len = %v, want 0 for zero count", len(results))
		}
	})
}

func TestSubintervals(t *testing.T) {
	tests := []struct {
		name    string
		steps   int
		a       float64
		b       float64
		want    [][2]float64
		wantErr bool
	}{
		{"positive steps", 2, 0, 1, [][2]float64{{0, 0.5}, {0.5, 1}}, false},
		{"zero steps", 0, 0, 10, [][2]float64{}, false},
		{"one step", 1, 10, 20, [][2]float64{{10, 20}}, false},
		{"inverted interval", 2, 1, 0, [][2]float64{{1, 0.5}, {0.5, 0}}, false},
		{"zero delta", 3, 10, 10, [][2]float64{{10, 10}, {10, 10}, {10, 10}}, false},
		{"negative steps", -1, 0, 10, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Subintervals(tt.steps, tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subintervals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Fatalf("Subintervals() len = %v, want %v", len(got), len(tt.want))
				}
				for i := range got {
					if !almostEqual(got[i][0], tt.want[i][0]) || !almostEqual(got[i][1], tt.want[i][1]) {
						t.Errorf("Subintervals() got[%d] = %v, want[%d] = %v", i, got[i], i, tt.want[i])
					}
				}
			}
		})
	}
}

func TestEncompass(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMin float64
		wantMax float64
		wantErr bool
	}{
		{"simple case", "1\n2\n3", 1, 3, false},
		{"negative numbers", "-5\n-10\n0", -10, 0, false},
		{"mixed numbers", "-5\n10\n-1\n5", -5, 10, false},
		{"single number", "7", 7, 7, false},
		{"empty input", "", 0, 0, true},
		{"only invalid input", "foo\nbar", 0, 0, true},
		{"mixed valid and invalid", "1\nfoo\n2\nbar\n3", 1, 3, false},
		{"decreasing order", "10\n5\n1", 1, 10, false},
		{"zero delta", "5\n5\n5", 5, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer from the input string
			reader := bytes.NewBufferString(tt.input)
			scanner := bufio.NewScanner(reader)

			gotMin, gotMax, err := Encompass(scanner)

			if (err != nil) != tt.wantErr {
				t.Errorf("Encompass() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !almostEqual(gotMin, tt.wantMin) || !almostEqual(gotMax, tt.wantMax) {
					t.Errorf("Encompass() gotMin = %v, gotMax = %v, wantMin = %v, wantMax = %v", gotMin, gotMax, tt.wantMin, tt.wantMax)
				}
			}
		})
	}
}