package main

import (
	"os"

	"golang.org/x/term"
)

// Muted color scheme for subtle, quality CLI UX
var (
	// Muted colors - subtle and professional
	mutedGreen  = "\033[38;5;71m"  // Muted green
	mutedYellow = "\033[38;5;179m" // Muted yellow/gold
	mutedRed    = "\033[38;5;167m" // Muted red/rose
	mutedBlue   = "\033[38;5;110m" // Muted blue
	mutedGray   = "\033[38;5;245m" // Muted gray
	mutedCyan   = "\033[38;5;109m" // Muted cyan
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

// colorize applies color if colors should be used
func colorize(color, text string) string {
	if !shouldUseColor() {
		return text
	}
	return color + text + colorReset
}

// shouldUseColor determines if color output should be enabled
// Precedence: --no-color flag > NO_COLOR env > terminal detection
func shouldUseColor() bool {
	// Check --no-color flag first (highest precedence)
	if globalCfg.noColor {
		return false
	}

	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	return true
}

// Color helper functions for consistent styling
func success(text string) string {
	return colorize(mutedGreen, text)
}

func warning(text string) string {
	return colorize(mutedYellow, text)
}

func errorText(text string) string {
	return colorize(mutedRed, text)
}

func info(text string) string {
	return colorize(mutedBlue, text)
}

func dim(text string) string {
	return colorize(mutedGray, text)
}

func accent(text string) string {
	return colorize(mutedCyan, text)
}

func bold(text string) string {
	if !shouldUseColor() {
		return text
	}
	return colorBold + text + colorReset
}
