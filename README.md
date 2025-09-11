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

### Flags

*   **`-r, --remap <src_a> <src_b> <dst_a> <dst_b>`**: Remaps a value from a source interval to a target interval.
*   **`-l, --limit <min> <max>`**: Restricts (clamps) a value to a given interval.
*   **`-E, --encompass`**: Reads a stream of numbers and outputs the minimum and maximum values.
*   **`-n, --divide <steps> <a> <b>`**: Generates a sequence of numbers by dividing an interval.
*   **`-e, --eval <a> <b>`**: Evaluates a parameter `t` within an interval.
*   **`-d, --deval <a> <b>`**: De-evaluates a number to its parameter `t`.
*   **`-R, --random <count> <a> <b>`**: Generates `<count>` random numbers within an interval.
*   **`-S, --snap <steps> <a> <b>`**: Snaps input values to the nearest point on a grid defined by `<steps>` within the interval `[a, b]`.
*   **`-s, --subintervals <steps> <a> <b>`**: Divides an interval into `<steps>` equal subintervals, printing each subinterval's start and end on a new line.

### Global Flags

*   **`-f, --format`**: Specifies the `printf` format for floating-point output (e.g., `%.3f`, `%g`).
*   **`--version`**: Prints version information and exits.


## Installation

### From Source (Recommended for Go users)

If you have Go installed (Go 1.18+ is required):

```bash
go install github.com/gregory-chatelier/span@latest
```

This command will download, compile, and install the `span` executable into your `$GOPATH/bin` directory (or `$HOME/go/bin` by default). Ensure this directory is in your system's `PATH`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.