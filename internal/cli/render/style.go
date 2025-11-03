package render

// StyleFunc applies styling to text.
type StyleFunc func(string) string

// Style defines terminal styling options.
type Style struct {
	color     Color
	bold      bool
	underline bool
}

// NewStyle creates a new style with the given color.
func NewStyle(c Color) Style {
	return Style{color: c}
}

// Bold returns a style with bold enabled.
func (s Style) Bold() Style {
	s.bold = true
	return s
}

// Underline returns a style with underline enabled.
func (s Style) Underline() Style {
	s.underline = true
	return s
}

// Apply applies the style to text.
func (s Style) Apply(text string) string {
	if s.color.ANSI == "" && !s.bold && !s.underline {
		return text
	}

	var prefix string
	if s.bold {
		prefix += colorBold
	}
	if s.underline {
		prefix += "\033[4m"
	}
	if s.color.ANSI != "" {
		prefix += s.color.ANSI
	}

	if prefix == "" {
		return text
	}

	return prefix + text + colorReset
}

// Predefined style functions for common use cases.
var (
	// ErrorStyle for error messages
	ErrorStyle = func(text string) string {
		return NewStyle(colorMutedRed).Bold().Apply(text)
	}

	// WarningStyle for warning messages
	WarningStyle = func(text string) string {
		return NewStyle(colorMutedGold).Apply(text)
	}

	// SuccessStyle for success messages
	SuccessStyle = func(text string) string {
		return NewStyle(colorMutedGreen).Apply(text)
	}

	// InfoStyle for informational messages
	InfoStyle = func(text string) string {
		return NewStyle(colorMutedBlue).Apply(text)
	}

	// EmphasisStyle for emphasized text
	EmphasisStyle = func(text string) string {
		return NewStyle(Color{}).Bold().Apply(text)
	}

	// DimStyle for secondary text
	DimStyle = func(text string) string {
		return NewStyle(colorMutedGray).Apply(text)
	}

	// CodeStyle for code/paths
	CodeStyle = func(text string) string {
		return text // Monospace is terminal-dependent
	}

	// PathStyle for file paths
	PathStyle = func(text string) string {
		return NewStyle(colorMutedBlue).Apply(text)
	}

	// AccentStyle for accent text
	AccentStyle = func(text string) string {
		return NewStyle(colorMutedPurple).Apply(text)
	}
)

// WithColor returns styled text if color is enabled, plain text otherwise.
func WithColor(colorEnabled bool, style StyleFunc, text string) string {
	if !colorEnabled {
		return text
	}
	return style(text)
}
