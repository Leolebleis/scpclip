package main

import (
	"fmt"
	"os"
)

var colorEnabled = func() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	o, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return o.Mode()&os.ModeCharDevice != 0
}()

func colorize(code, s string) string {
	if !colorEnabled {
		return s
	}
	return code + s + "\033[0m"
}

func green(s string) string { return colorize("\033[32m", s) }
func red(s string) string   { return colorize("\033[31m", s) }
func bold(s string) string  { return colorize("\033[1m", s) }
func faint(s string) string { return colorize("\033[2m", s) }

func success(format string, a ...any) {
	fmt.Fprintln(os.Stderr, green("✓")+" "+fmt.Sprintf(format, a...))
}

func fail(format string, a ...any) {
	fmt.Fprintln(os.Stderr, red("✗")+" "+fmt.Sprintf(format, a...))
}

func hint(format string, a ...any) {
	fmt.Fprintln(os.Stderr, "  "+faint(fmt.Sprintf(format, a...)))
}
