package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStyle(t *testing.T) {
	color := Color{ANSI: "\033[31m"}
	style := NewStyle(color)
	assert.Equal(t, color, style.color)
	assert.False(t, style.bold)
	assert.False(t, style.underline)
}

func TestStyle_Bold(t *testing.T) {
	style := NewStyle(Color{}).Bold()
	assert.True(t, style.bold)
}

func TestStyle_Underline(t *testing.T) {
	style := NewStyle(Color{}).Underline()
	assert.True(t, style.underline)
}

func TestStyle_Apply_Plain(t *testing.T) {
	style := NewStyle(Color{})
	result := style.Apply("test")
	assert.Equal(t, "test", result)
}

func TestStyle_Apply_Color(t *testing.T) {
	color := Color{ANSI: "\033[31m"}
	style := NewStyle(color)
	result := style.Apply("test")
	assert.Contains(t, result, "\033[31m")
	assert.Contains(t, result, "test")
}

func TestStyle_Apply_Bold(t *testing.T) {
	style := NewStyle(Color{}).Bold()
	result := style.Apply("test")
	assert.Contains(t, result, colorBold)
	assert.Contains(t, result, "test")
}

func TestStyle_Apply_Underline(t *testing.T) {
	style := NewStyle(Color{}).Underline()
	result := style.Apply("test")
	assert.Contains(t, result, "\033[4m")
	assert.Contains(t, result, "test")
}

func TestStyle_Apply_Combined(t *testing.T) {
	color := Color{ANSI: "\033[31m"}
	style := NewStyle(color).Bold().Underline()
	result := style.Apply("test")
	assert.Contains(t, result, "\033[31m")
	assert.Contains(t, result, colorBold)
	assert.Contains(t, result, "\033[4m")
	assert.Contains(t, result, "test")
}

func TestErrorStyle(t *testing.T) {
	result := ErrorStyle("error")
	assert.Contains(t, result, "error")
	assert.Contains(t, result, "\033[")
}

func TestWarningStyle(t *testing.T) {
	result := WarningStyle("warning")
	assert.Contains(t, result, "warning")
	assert.Contains(t, result, "\033[")
}

func TestSuccessStyle(t *testing.T) {
	result := SuccessStyle("success")
	assert.Contains(t, result, "success")
	assert.Contains(t, result, "\033[")
}

func TestInfoStyle(t *testing.T) {
	result := InfoStyle("info")
	assert.Contains(t, result, "info")
	assert.Contains(t, result, "\033[")
}

func TestEmphasisStyle(t *testing.T) {
	result := EmphasisStyle("emphasis")
	assert.Contains(t, result, "emphasis")
}

func TestDimStyle(t *testing.T) {
	result := DimStyle("dim")
	assert.Contains(t, result, "dim")
	assert.Contains(t, result, "\033[")
}

func TestCodeStyle(t *testing.T) {
	result := CodeStyle("code")
	assert.Equal(t, "code", result)
}

func TestPathStyle(t *testing.T) {
	result := PathStyle("/path/to/file")
	assert.Contains(t, result, "/path/to/file")
}

func TestWithColor_Enabled(t *testing.T) {
	result := WithColor(true, ErrorStyle, "test")
	assert.Contains(t, result, "\033[")
	assert.Contains(t, result, "test")
}

func TestWithColor_Disabled(t *testing.T) {
	result := WithColor(false, ErrorStyle, "test")
	assert.Equal(t, "test", result)
	assert.NotContains(t, result, "\033[")
}

func TestStyle_Chaining(t *testing.T) {
	// Test that chaining methods works correctly
	style := NewStyle(colorMutedRed).Bold().Underline()
	result := style.Apply("chained")

	assert.Contains(t, result, "chained")
	assert.Contains(t, result, colorBold)
	assert.Contains(t, result, "\033[4m")
}

func TestAllPredefinedStyles(t *testing.T) {
	styles := map[string]StyleFunc{
		"error":    ErrorStyle,
		"warning":  WarningStyle,
		"success":  SuccessStyle,
		"info":     InfoStyle,
		"emphasis": EmphasisStyle,
		"dim":      DimStyle,
		"code":     CodeStyle,
		"path":     PathStyle,
	}

	for name, style := range styles {
		t.Run(name, func(t *testing.T) {
			result := style("test")
			assert.Contains(t, result, "test")
		})
	}
}
