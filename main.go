package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	RED     = "\033[31m"
	GREEN   = "\033[32m"
	YELLOW  = "\033[33m"
	BLUE    = "\033[34m"
	MAGENTA = "\033[35m"
	CYAN    = "\033[36m"
	RESET   = "\033[0m"
)

func colorize(s, color string, useColor bool) string {
	if !useColor {
		return s
	}
	return color + s + RESET
}

// isTerminal returns true if the file is an interactive terminal.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}


// openInput resolves the input source: -f flag, positional arg, or stdin.
func openInput(file string, args []string) (io.ReadCloser, error) {
	if file != "" {
		return os.Open(file)
	}
	if len(args) > 0 {
		return os.Open(args[0])
	}
	return io.NopCloser(os.Stdin), nil
}

func main() {
	noColor := flag.Bool("nc", false, "disable color")
	file := flag.String("f", "", "input file")
	showLineNum := flag.Bool("l", false, "show line numbers")
	raw := flag.Bool("raw", false, "raw mode")
	only := flag.Bool("only", false, "only control chars")
	spaceChar := flag.String("c", ".", "space char")

	flag.Parse()

	// No args and stdin is a TTY → show help.
	if *file == "" && flag.NArg() == 0 && isTerminal(os.Stdin) {
		flag.Usage() 
		return
	}

	// Auto-enable color only when stdout is a real terminal.
	useColor := isTerminal(os.Stdout) && !*noColor

	f, err := openInput(*file, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	defer f.Close()

	buf := bufio.NewReader(f)

	lineNum := 1
	// atLineStart: print the line-number prefix before the first visible rune
	// on each line, not before the newline that ends the previous line.
	atLineStart := true
	// lastWasNewline: empty input counts as newline-terminated (no warning).
	lastWasNewline := true

	for {
		r, _, err := buf.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "Read error:", err)
			os.Exit(1)
		}

		lastWasNewline = (r == '\n')

		// Raw mode: pass through unchanged, only track line numbers.
		if *raw {
			fmt.Printf("%c", r)
			if r == '\n' {
				lineNum++
				atLineStart = true
			}
			continue
		}

		// Print line-number prefix on the first rune of each line.
		// We skip it for '\n' itself so the prefix appears at the start
		// of content, not glued to the end of the previous line.
		if *showLineNum && atLineStart && r != '\n' {
			fmt.Printf("%4d | ", lineNum)
			atLineStart = false
		}

		switch {
		case r == ' ':
			if !*only {
				fmt.Print(colorize(*spaceChar, RED, useColor))
			}

		case r == '\t':
			fmt.Print(colorize(`\t`, YELLOW, useColor))

		case r == '\r':
			fmt.Print(colorize(`\r`, BLUE, useColor))

		case r == '\n':
			fmt.Print(colorize(`\n`, GREEN, useColor))
			fmt.Print("\n")
			lineNum++
			atLineStart = true

		case r == 0:
			fmt.Print(colorize(`\0`, MAGENTA, useColor))

		case r < 0x20 || r == 0x7f:
			// Other ASCII control chars (SOH, BEL, ESC, DEL, …)
			sym := fmt.Sprintf(`\x%02x`, r)
			fmt.Print(colorize(sym, CYAN, useColor))

		default:
			if !*only {
				fmt.Printf("%c", r)
			}
		}
	}

	// Warn when the file does not end with a newline.
	if !*raw && !lastWasNewline {
		fmt.Print(colorize(" ⏎ (missing final newline)", YELLOW, useColor))
		fmt.Println()
	}
}

