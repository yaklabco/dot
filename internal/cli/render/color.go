package render

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// Color represents a terminal color.
type Color struct {
	ANSI string
}

// ColorScheme defines a color palette.
type ColorScheme struct {
	Error   Color
	Warning Color
	Success Color
	Info    Color
	Dim     Color
	Accent  Color
}

// Predefined colors using 256-color ANSI codes for consistent muted professional palette.
var (
	// Muted professional colors (256-color palette)
	colorMutedGreen  = Color{ANSI: "\033[38;5;71m"}  // #5F875F - muted green
	colorMutedGold   = Color{ANSI: "\033[38;5;179m"} // #D7AF87 - muted gold
	colorMutedRed    = Color{ANSI: "\033[38;5;167m"} // #D75F5F - muted red
	colorMutedBlue   = Color{ANSI: "\033[38;5;110m"} // #87AFD7 - muted blue
	colorMutedGray   = Color{ANSI: "\033[38;5;245m"} // #808080 - muted gray
	colorMutedPurple = Color{ANSI: "\033[38;5;104m"} // #8787D7 - dark blue/purple
	colorReset       = "\033[0m"
	colorBold        = "\033[1m"

	// Default color scheme with muted professional colors
	DefaultScheme = ColorScheme{
		Success: colorMutedGreen,
		Warning: colorMutedGold,
		Error:   colorMutedRed,
		Info:    colorMutedBlue,
		Dim:     colorMutedGray,
		Accent:  colorMutedPurple,
	}

	// No-color scheme for plain text
	NoColorScheme = ColorScheme{
		Error:   Color{ANSI: ""},
		Warning: Color{ANSI: ""},
		Success: Color{ANSI: ""},
		Info:    Color{ANSI: ""},
		Dim:     Color{ANSI: ""},
		Accent:  Color{ANSI: ""},
	}
)

// Apply applies the color to text.
func (c Color) Apply(text string) string {
	if c.ANSI == "" {
		return text
	}
	return c.ANSI + text + colorReset
}

// ShouldUseColor determines if color output should be enabled.
func ShouldUseColor() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	// Check TERM environment variable
	termEnv := os.Getenv("TERM")
	if termEnv == "" || termEnv == "dumb" {
		return false
	}

	// Check for color support
	if strings.Contains(termEnv, "color") || strings.Contains(termEnv, "256") || strings.Contains(termEnv, "xterm") {
		return true
	}

	return true
}

// GetScheme returns the appropriate color scheme based on environment.
func GetScheme() ColorScheme {
	if ShouldUseColor() {
		return DefaultScheme
	}
	return NoColorScheme
}

// Colorizer provides semantic color functions for consistent CLI output.
type Colorizer struct {
	enabled bool
	scheme  ColorScheme
}

// NewColorizer creates a colorizer with the appropriate scheme.
func NewColorizer(enabled bool) *Colorizer {
	scheme := NoColorScheme
	if enabled {
		scheme = DefaultScheme
	}
	return &Colorizer{
		enabled: enabled,
		scheme:  scheme,
	}
}

// Success formats text with success color (muted green).
func (c *Colorizer) Success(text string) string {
	return c.scheme.Success.Apply(text)
}

// Warning formats text with warning color (muted gold).
func (c *Colorizer) Warning(text string) string {
	return c.scheme.Warning.Apply(text)
}

// Error formats text with error color (muted red).
func (c *Colorizer) Error(text string) string {
	return c.scheme.Error.Apply(text)
}

// Info formats text with info color (muted blue).
func (c *Colorizer) Info(text string) string {
	return c.scheme.Info.Apply(text)
}

// Dim formats text with dim color (muted gray).
func (c *Colorizer) Dim(text string) string {
	return c.scheme.Dim.Apply(text)
}

// Accent formats text with accent color (dark blue/purple).
func (c *Colorizer) Accent(text string) string {
	return c.scheme.Accent.Apply(text)
}

// Bold formats text with bold styling.
func (c *Colorizer) Bold(text string) string {
	if !c.enabled {
		return text
	}
	return colorBold + text + colorReset
}
