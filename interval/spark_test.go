package interval

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestGenerateSparkline(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		config SparkConfig
		want   string
	}{
		{
			name:  "Basic Generation",
			input: "10\n20\n30\n40\n50\n60\n70\n80",
			config: SparkConfig{},
			want:   " ▂▃▄▅▆▇█",
		},
		{
			name:  "Fixed Interval",
			input: "10\n50\n90",
			config: SparkConfig{Min: 0, Max: 100, HasMin: true, HasMax: true},
			want:   " ▄▇",
		},
		{
			name:  "Clamping Values",
			input: "-10\n50\n110",
			config: SparkConfig{Min: 0, Max: 100, HasMin: true, HasMax: true},
			want:   " ▄█",
		},
		{
			name:  "Empty Input",
			input: "",
			config: SparkConfig{},
			want:   "",
		},
		{
			name:  "Non-numeric Input",
			input: "10\nhello\n80\nworld",
			config: SparkConfig{},
			want:   " █",
		},
		{
			name:  "With Color",
			input: "10\n80",
			config: SparkConfig{Color: ColorBlue},
			want:   "\033[34m █\033[0m",
		},
		{
			name:  "Sliding Window (Width)",
			input: "10\n20\n30\n40\n50",
			config: SparkConfig{Width: 3},
			want:   "\r▄▆█", // Renders based on 30, 40, 50; scales based on 10-50
		},
		{
			name:  "Sliding Window with Color",
			input: "10\n20\n30\n40\n50",
			config: SparkConfig{Width: 3, Color: ColorRed},
			want:   "\r\033[31m▄▆█\033[0m",
		},
		{
			name:  "Single Value",
			input: "42",
			config: SparkConfig{},
			want:   " ", // With a single value, min == max, so it should pick the lowest char
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tc.input))
			var writer bytes.Buffer

			err := GenerateSparkline(scanner, &writer, tc.config)

			if err != nil {
				t.Fatalf("GenerateSparkline() returned an unexpected error: %v", err)
			}

			if got := writer.String(); got != tc.want {
				t.Errorf("GenerateSparkline()\n  got: %q\n want: %q", got, tc.want)
			}
		})
	}
}

func TestParseColor(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    SparkColor
		wantErr bool
	}{
		{"Empty gives None", "", ColorNone, false},
		{"red", "red", ColorRed, false},
		{"GREEN", "GREEN", ColorGreen, false}, // Case-insensitivity
		{"Unknown color", "orange", ColorNone, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseColor(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseColor() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("ParseColor() = %v, want %v", got, tc.want)
			}
		})
	}
}
