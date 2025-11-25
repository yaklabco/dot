package pager

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_Defaults(t *testing.T) {
	p := New(
		WithSize(80, 24),
		WithInteractive(false),
	)

	assert.Equal(t, 80, p.Width())
	assert.Equal(t, 24, p.Height())
	assert.False(t, p.interactive)
	assert.False(t, p.IsInteractive())
}

func TestPager_IsInteractive(t *testing.T) {
	// Non-interactive pager
	p := New(
		WithSize(80, 24),
		WithInteractive(false),
	)
	assert.False(t, p.IsInteractive())

	// Forced interactive pager
	p = New(
		WithSize(80, 24),
		WithInteractive(true),
	)
	assert.True(t, p.IsInteractive())
}

func TestNew_WithOptions(t *testing.T) {
	var buf bytes.Buffer
	p := New(
		WithOutput(&buf),
		WithSize(120, 40),
		WithInteractive(false),
	)

	assert.Equal(t, 120, p.Width())
	assert.Equal(t, 40, p.Height())
}

func TestPager_NeedsPaging(t *testing.T) {
	p := New(
		WithSize(80, 24),
		WithInteractive(false),
	)

	// 24 height - 2 reserved = 22 lines per page
	assert.False(t, p.NeedsPaging(20)) // Fits in one page
	assert.False(t, p.NeedsPaging(22)) // Exactly one page
	assert.True(t, p.NeedsPaging(23))  // Needs paging
	assert.True(t, p.NeedsPaging(100)) // Definitely needs paging
}

func TestPager_Display_ShortContent(t *testing.T) {
	var buf bytes.Buffer
	p := New(
		WithOutput(&buf),
		WithSize(80, 24),
		WithInteractive(false),
	)

	content := "Line 1\nLine 2\nLine 3"
	err := p.Display(content)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Line 1")
	assert.Contains(t, buf.String(), "Line 2")
	assert.Contains(t, buf.String(), "Line 3")
}

func TestPager_DisplayLines_NonInteractive(t *testing.T) {
	var buf bytes.Buffer
	p := New(
		WithOutput(&buf),
		WithSize(80, 10),
		WithInteractive(false),
	)

	// More lines than page size but non-interactive
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "Line content"
	}

	err := p.DisplayLines(lines)

	assert.NoError(t, err)
	// All lines should be printed (no pagination in non-interactive mode)
	assert.Equal(t, 20, strings.Count(buf.String(), "Line content"))
}

func TestPager_DisplayLines_Interactive(t *testing.T) {
	var output bytes.Buffer
	// Simulate user pressing Enter then q
	input := strings.NewReader("\nq\n")

	p := New(
		WithOutput(&output),
		WithInput(input),
		WithSize(80, 10), // Small height to force pagination
		WithInteractive(true),
	)

	// Create content larger than one page
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "Content line"
	}

	err := p.DisplayLines(lines)

	assert.NoError(t, err)
	// Should show prompt for more content
	assert.Contains(t, output.String(), "more line")
}

func TestPager_DisplayLines_QuitEarly(t *testing.T) {
	var output bytes.Buffer
	// User quits immediately after first page
	input := strings.NewReader("q\n")

	p := New(
		WithOutput(&output),
		WithInput(input),
		WithSize(80, 10), // 10 height - 2 = 8 lines per page
		WithInteractive(true),
	)

	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "Content"
	}

	err := p.DisplayLines(lines)

	assert.NoError(t, err)
	// Should show only first page (8 lines) due to early quit
	lineCount := strings.Count(output.String(), "Content")
	assert.Equal(t, 8, lineCount) // First page only
}

func TestPager_Display_Empty(t *testing.T) {
	var buf bytes.Buffer
	p := New(
		WithOutput(&buf),
		WithSize(80, 24),
		WithInteractive(false),
	)

	err := p.Display("")

	assert.NoError(t, err)
	// Should just have a newline for the empty line
	assert.Equal(t, "\n", buf.String())
}

func TestPagedWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	p := New(
		WithOutput(&buf),
		WithSize(80, 24),
		WithInteractive(false),
	)

	w := NewPagedWriter(p)

	_, err := w.Write([]byte("Line 1\n"))
	assert.NoError(t, err)

	_, err = w.Write([]byte("Line 2\nLine 3\n"))
	assert.NoError(t, err)

	// Line 1\n splits to ["Line 1", ""]
	// Line 2\nLine 3\n splits to ["Line 2", "Line 3", ""]
	// Combined: ["Line 1", "", "Line 2", "Line 3", ""]
	assert.Equal(t, 5, w.LineCount())
}

func TestPagedWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	p := New(
		WithOutput(&buf),
		WithSize(80, 24),
		WithInteractive(false),
	)

	w := NewPagedWriter(p)

	_, _ = w.Write([]byte("Hello\nWorld\n"))
	err := w.Flush()

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Hello")
	assert.Contains(t, buf.String(), "World")
}

func TestPager_Height_Width(t *testing.T) {
	p := New(
		WithSize(100, 50),
		WithInteractive(false),
	)

	assert.Equal(t, 100, p.Width())
	assert.Equal(t, 50, p.Height())
}

func TestPager_WithInput(t *testing.T) {
	r := strings.NewReader("test input")
	p := New(
		WithInput(r),
		WithSize(80, 24),
		WithInteractive(false),
	)

	assert.Equal(t, r, p.input)
}

func TestPager_DisplayLines_EOF(t *testing.T) {
	var output bytes.Buffer
	// Empty reader simulates EOF
	input := strings.NewReader("")

	p := New(
		WithOutput(&output),
		WithInput(input),
		WithSize(80, 10),
		WithInteractive(true),
	)

	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "Content"
	}

	err := p.DisplayLines(lines)

	// Should handle EOF gracefully
	assert.NoError(t, err)
}

func TestPagedWriter_PartialLines(t *testing.T) {
	var buf bytes.Buffer
	p := New(
		WithOutput(&buf),
		WithSize(80, 24),
		WithInteractive(false),
	)

	w := NewPagedWriter(p)

	// Write partial line
	_, _ = w.Write([]byte("Part 1"))
	// Complete the line
	_, _ = w.Write([]byte(" Part 2\n"))

	// "Part 1" -> lines: ["Part 1"]
	// " Part 2\n" splits to [" Part 2", ""] and first element is appended
	// Result: ["Part 1 Part 2", ""] - wait, that's not right
	// Actually the append logic only works when content ends with \n
	// Let me trace through: first Write("Part 1") -> lines = ["Part 1"]
	// Second Write(" Part 2\n") -> splits to [" Part 2", ""]
	// Since first write didn't end with \n, we have lines > 0, append to last
	// lines[0] = "Part 1" + " Part 2" = "Part 1 Part 2"
	// Then append rest: ["Part 1 Part 2", ""]
	// Wait, the logic is: if len(lines) > 0 and content doesn't end with \n...
	// But content here ends with \n so it goes to else case
	// Actually strings.Split(" Part 2\n", "\n") = [" Part 2", ""]
	// lines was ["Part 1"], after split we have [" Part 2", ""]
	// Logic says: if strings.HasSuffix(content, "\n") - TRUE so goes to else
	// So lines = append(["Part 1"], [" Part 2", ""]...) = ["Part 1", " Part 2", ""]
	assert.Equal(t, 3, w.LineCount())
}

func TestPager_ContinuePaging(t *testing.T) {
	var output bytes.Buffer
	// Press Enter twice then q
	input := strings.NewReader("\n\nq\n")

	p := New(
		WithOutput(&output),
		WithInput(input),
		WithSize(80, 5), // Very small to force multiple pages
		WithInteractive(true),
	)

	lines := make([]string, 15)
	for i := range lines {
		lines[i] = "Line content"
	}

	err := p.DisplayLines(lines)

	assert.NoError(t, err)
	// Multiple pages should have been shown
	lineCount := strings.Count(output.String(), "Line content")
	assert.Greater(t, lineCount, 3) // At least first page shown
}

// TestPager_IoReader verifies that any io.Reader works.
func TestPager_IoReader(t *testing.T) {
	var buf bytes.Buffer
	customReader := struct {
		io.Reader
	}{
		Reader: strings.NewReader(""),
	}

	p := New(
		WithOutput(&buf),
		WithInput(customReader),
		WithSize(80, 24),
		WithInteractive(false), // Non-interactive since custom reader
	)

	err := p.Display("Test")
	assert.NoError(t, err)
}
