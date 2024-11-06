package slogx

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
)

func TestColorize(t *testing.T) {
	fmt.Printf("%s", colorize("hello", Reset))
	fmt.Printf("%s", colorize("hello", Bold))
	fmt.Printf("%s", colorize("hello", Dim))
	fmt.Printf("%s", colorize("hello", Italic))
	fmt.Printf("%s", colorize("hello", Red))
	fmt.Printf("%s", colorize("hello", Green))
	fmt.Printf("%s", colorize("hello", Yellow))
	fmt.Printf("%s", colorize("hello", Blue))
	fmt.Printf("%s", colorize("hello", Magenta))
	fmt.Printf("%s", colorize("hello", Cyan))
	fmt.Printf("%s", colorize("hello", White))
	fmt.Printf("%s", colorize("hello", Gray))
	fmt.Printf("%s", colorize("hello", LightRed))
	fmt.Printf("%s", colorize("hello", LightGreen))
	fmt.Printf("%s", colorize("hello", LightYellow))
	fmt.Printf("%s", colorize("hello", LightBlue))
	fmt.Printf("%s", colorize("hello", LightMagenta))
	fmt.Printf("%s", colorize("hello", LightCyan))
	fmt.Printf("%s", colorize("hello", BgBlack))
	fmt.Printf("%s", colorize("hello", BgRed))
	fmt.Printf("%s", colorize("hello", BgGreen))
	fmt.Printf("%s", colorize("hello", BgYellow))
	fmt.Printf("%s", colorize("hello", BgBlue))
	fmt.Printf("%s", colorize("hello", BgMagenta))
	fmt.Printf("%s", colorize("hello", BgCyan))
	fmt.Printf("%s", colorize("hello", BgWhite))
}

func TestLevel(t *testing.T) {
	logger := slog.New(New(os.Stdout, &Options{Level: slog.LevelDebug}))
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
}

func TestFrontKey(t *testing.T) {
	logger := slog.New(New(os.Stdout, &Options{Level: slog.LevelDebug})).With("front", " [key]")
	logger.Info("This is an info message")
}

func TestDifferentObjects(t *testing.T) {
	logger := slog.New(New(os.Stdout, &Options{Level: slog.LevelDebug}))
	logger.Info("This is an info message", "int", 1, "float", 1.1, "string", "hello")
	// complicated struct
	type Person struct {
		Name string
		Age  int
	}
	logger.Info("This is an info message", "person", Person{"John", 30})
}
