

<pre>
╔═══╗╔═══╗╔═══╗╔═╗ ╔╗
║╔═╗║║╔═╗║║╔═╗║║║╚╗║║
║╚══╗║╚═╝║║║ ║║║╔╗╚╝║
╚══╗║║╔══╝║╚═╝║║║╚╗║║
║╚═╝║║║   ║╔═╗║║║ ║║║
╚═══╝╚╝   ╚╝ ╚╝╚╝ ╚═╝
</pre>

# SPAN - Command line Interval Tool

Map, normalize, and generate values across numeric intervals

## Why span?

Working with numeric ranges is a common task in scripting, data analysis, and system monitoring. While powerful tools like `awk` can perform these operations, their syntax is often verbose and hard to remember for common tasks like remapping, clamping, or generating sequences.

`span` simplifies these operations, providing a clean, intuitive, and highly composable command-line interface. It adheres to the Unix philosophy: do one thing well, and work seamlessly with other tools via pipes.

## Command Reference

`span` uses flags to determine its mode of operation. Only one operational flag can be used at a time.

### Global Flags

*   **`-f, --format`**: Specifies the `printf` format for floating-point output (e.g., `%.3f`). To format as an integer, use `%.0f`.
*   **`--version`**: Prints version information and exits.

### Operational Flags

*   **`-r, --remap <src_a> <src_b> <dst_a> <dst_b>`**: Remaps a value from a source interval to a target interval.
    *   *Ex.:* `echo 5 | span -r 0 10 100 200` -> `150`
*   **`-l, --limit <min> <max>`**: Restricts (clamps) a value to a given interval.
    *   *Ex.:* `echo 150 | span -l 0 100` -> `100`
*   **`-E, --encompass`**: Reads a stream of numbers and outputs the minimum and maximum values.
    *   *Ex.:* `printf "10\n5\n20" | span -E` -> `5 20`
*   **`-n, --divide <steps> <a> <b>`**: Generates a sequence of numbers by dividing an interval.
    *   *Ex.:* `span -n 4 0 1` -> `0.0\n0.25\n0.5\n0.75`
*   **`-e, --eval <a> <b>`**: Evaluates a parameter `t` within an interval.
    *   *Ex.:* `echo 0.5 | span -e 100 200` -> `150`
*   **`-d, --deval <a> <b>`**: De-evaluates a number to its parameter `t`.
    *   *Ex.:* `echo 150 | span -d 100 200` -> `0.5`
*   **`-R, --random <count> <a> <b>`**: Generates `<count>` random numbers within an interval.
    *   *Ex.:* `span -R 3 0 10` -> (Three random numbers between 0 and 10)
*   **`-S, --snap <steps> <a> <b>`**: Snaps input values to the nearest point on a grid.
    *   *Ex.:* `echo 4.78 | span -S 10 0 10` -> `5`
*   **`-s, --subintervals <steps> <a> <b>`**: Divides an interval into `<steps>` equal subintervals.
    *   *Ex.:* `span -s 2 0 1` -> `0 0.5\n0.5 1`




## Installation

`span` provides flexible installation options.

### Quick Install (Recommended)

This single command will download and install `span` to a sensible default location for your system.

**User-level Installation (Recommended for most users):**
Installs `span` to `$HOME/.local/bin` (Linux/macOS) or a user-specific `bin` directory (Windows).

```bash
curl -sSfL https://raw.githubusercontent.com/gregory-chatelier/span/main/install.sh | sh
```

**System-wide Installation (Requires `sudo`):**
Installs `span` to `/usr/local/bin` (Linux/macOS).

```bash
sudo curl -sSfL https://raw.githubusercontent.com/gregory-chatelier/span/main/install.sh | sh
```

### Custom Installation Directory

You can specify a custom installation directory using the `INSTALL_DIR` environment variable:

```bash
INSTALL_DIR=$HOME/my-tools curl -sSfL https://raw.githubusercontent.com/gregory-chatelier/span/main/install.sh | sh
```

### From Source

If you have Go installed (Go 1.18+ is required):

```bash
go install github.com/gregory-chatelier/span@latest
```



## Common Usage

### Adjusting a Value (Remap)
Convert a 4-star rating to a percentage.
```bash
echo "3.5" | span -r 0 4 0 100
# Result: 87.5
```

### Basic Input Clamping (Limit)
Ensure a user-supplied count is not negative before using it.
```bash
count="-5"
safe_count=$(echo "$count" | span -l 0 1000)
# safe_count is now "0"
```

### Getting Steps for a Loop (Divide)
Create 5 evenly spaced points between -1 and 1.
```bash
span -n -f "%.1f" 5 -1 1
# Result:
# -1.0
# -0.6
# -0.2
# 0.2
# 0.6
```

### Quick Data Range Check (Encompass)
Quickly see the range of values in a log file.
```bash
cat api_response_times.log | span -E
# Result: 52 1250
```

### Generating Points on a Circle
Generate 8 evenly spaced angles (in radians) around a full circle (`0` to `2*PI`).
```bash
# PI is approx 3.1415926535
span -n -f "%.3f" 8 0 6.283185307
# Result:
# 0.000
# 0.785
# 1.571
# ...
# 5.498
```



## Advanced Examples

Here are some more powerful ways to use `span` by composing it with other tools:

### Terminal Bar Charts

Create simple, dynamic visualizations directly in the terminal by remapping numbers to a fixed width.

```bash
# Set random data
span -f "%.0f" -R 6 0 10 > data.txt

# Set the desired width for the bar chart
chart_width=40

# Get the data's range using encompass
data_range=$(span --encompass < data.txt)

# Remap the data to the chart width and print bars
span -r $data_range 1 $chart_width < data.txt | awk '{ for (i=0; i<$1; i++) printf "█"; print "" }'
```

### Color Space Conversion (Lerp & Inverse Lerp)

Convert a color from one range to another. This example takes an 8-bit RGB color value (0-255) and converts it to a floating-point grayscale value (0.0-1.0).

```bash
# The RGB value we want to convert (e.g., 192, a light gray)
rgb_value=192

# Use 'deval' to perform an Inverse Lerp: find the parameter 't'
# of the color in the 8-bit RGB space [0, 255].
t=$(echo "$rgb_value" | span -d 0 255)

# Use 'eval' to perform a Lerp: evaluate 't' in the target
# floating-point grayscale space [0.0, 1.0].
grayscale_value=$(echo "$t" | span -e 0.0 1.0)

echo "The RGB value $rgb_value is equivalent to $grayscale_value in grayscale."
```

### Calculating Progress Through a Time Period

Use `span --deval` to normalize the current time within a start and end timestamp to calculate the percentage of a time window that has elapsed.

```bash
# Calculate progress of a time window
start_time="2023-10-15 09:00:00"
end_time="2023-10-15 17:00:00"

# Convert to timestamps
start_ts=$(date -d "$start_time" +%s)
end_ts=$(date -d "$end_time" +%s)
current_ts=$(date +%s)

# Normalize current time to a percentage [0, 1]
progress=$(echo "$current_ts" | span -d "$start_ts" "$end_ts")

# Format as a percentage and print
echo "Progress through the day: $(echo "$progress * 100" | bc -l | awk '{printf "%.1f%%", $1}')"

# Use it to draw a simple bar (50 chars wide)
bar_length=$(echo "$progress * 50" | bc -l | awk '{printf "%.0f", $1}')
printf "|[%-50s]|
" "$(printf '#%.0s' $(seq 1 $bar_length))"
```
Output (at 1:00 PM):

```text
Progress through the day: 50.0%
|[#########################                         ]|
```

### Controlling Script Parameters Safely

Use `span --limit` to clamp a user-provided or calculated value to a safe, known range before using it in a script. This prevents errors from out-of-bounds numbers.

```bash
# Safely set a system parameter, like screen brightness.

# Get the desired brightness from the first script argument, default to 50.
desired_brightness=${1:-50}

# Use 'span -l' to clamp the value to a safe range [5, 100].
# This prevents setting brightness to 0 (off) or an invalid negative value.
safe_brightness=$(echo "$desired_brightness" | span -l 5 100)

echo "Desired brightness: $desired_brightness"
echo "Setting safe brightness to: $safe_brightness"

# Now, use the sanitized value to call the system command.
# set_screen_brightness --level "$safe_brightness"
```

### Generating a Color Gradient

Use `span -n` to generate a smooth gradient between two colors. A gradient is simply a series of evenly spaced steps between two endpoints.

```bash
# Generate a 10-step gradient from Red (255,0,0) to Yellow (255,255,0).
# We only need to generate the steps for the Green channel, from 0 to 255.

echo "CSS Gradient Steps:"
span -n 10 0 255 | awk '{ printf "rgb(255, %.0f, 0)\n", $1 }'
```

### Generating Sample Data for Testing

Use `span -R` to generate a stream of random numbers within an expected range to test scripts or programs.

```bash
# Generate 10 random temperature readings between 18.5 and 21.5 degrees.
span -R 10 18.5 21.5 > mock_sensor_data.txt

# You can now use this file to test your processing script.
# ./process_temperatures.sh < mock_sensor_data.txt
```

### Audio Volume Control (Snapping)

Use `span -S` to quantize a value to the nearest valid step. This is useful for things like volume or brightness controls, which often use discrete levels.

```bash
# A system volume control uses 20 steps from 0 to 100 (0, 5, 10, ...).
# A script calculates a desired volume of 47.8%.
desired_volume=47.8

# Use 'span -S' to snap this value to the nearest of the 20 steps.
snapped_volume=$(echo "$desired_volume" | span -S 20 0 100)

echo "Desired volume: $desired_volume%"
echo "Snapping to nearest valid level: $snapped_volume%"

# Now, use the clean, snapped value to set the system volume.
# set_system_volume --percentage "$snapped_volume"
```

### Parallel Data Processing

Use `span -s` to divide a large task into smaller, independent chunks for parallel processing. This is a powerful way to speed up scripts that handle batch operations.

```bash
#!/bin/bash
# Process a large number of items (e.g., 1 to 1000) in parallel batches.

# A function that does the work for one chunk.
# (Here, we just simulate work with 'sleep'.)
do_work() {
  start=$1
  end=$2
  echo "Processing batch from item $start to $end..."
  sleep 2
  echo "Finished batch $start-$end."
}
export -f do_work # Export the function so xargs can use it

# Use 'span -s' to divide 1000 items into 10 chunks.
# Then, use xargs to run up to 4 'do_work' processes in parallel.
span -s 10 1 1001 | xargs -n 2 -P 4 bash -c 'do_work "$@"' _

echo "All batches processed."
```



## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.