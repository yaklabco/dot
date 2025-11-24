package adopt

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestNewArrowSelector(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()

	selector := NewArrowSelector(input, output, fs, "/tmp/test-config")

	assert.NotNil(t, selector)
	assert.Equal(t, input, selector.input)
	assert.Equal(t, output, selector.output)
}

func TestArrowSelector_EmptyItems(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}
	fs := adapters.NewMemFS()
	selector := NewArrowSelector(input, output, fs, "/tmp/test-config")

	indices, err := selector.SelectMultiple([]string{}, []DotfileCandidate{})

	assert.NoError(t, err)
	assert.Empty(t, indices)
}

// Test the Bubble Tea model directly
func TestBubbleModel_Init(t *testing.T) {
	m := bubbleModel{
		items:      []string{"item1", "item2"},
		selected:   make(map[int]bool),
		ignoring:   make(map[int]bool),
		ignoreTime: make(map[int]time.Time),
	}

	cmd := m.Init()
	assert.NotNil(t, cmd) // Should return tick command
}

func TestBubbleModel_Update_Navigation(t *testing.T) {
	m := bubbleModel{
		items:    []string{"item1", "item2", "item3"},
		selected: make(map[int]bool),
		cursor:   0,
		height:   24,
	}

	// Test down arrow
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'↓'}})
	m = newModel.(bubbleModel)
	assert.Equal(t, 0, m.cursor) // Rune doesn't match, stays at 0

	// Use string-based key
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(bubbleModel)
	assert.Equal(t, 1, m.cursor)

	// Test up arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(bubbleModel)
	assert.Equal(t, 0, m.cursor)

	// Test down at boundary
	m.cursor = 2
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(bubbleModel)
	assert.Equal(t, 2, m.cursor) // Should stay at 2
}

func TestBubbleModel_Update_Selection(t *testing.T) {
	m := bubbleModel{
		items:    []string{"item1", "item2", "item3"},
		selected: make(map[int]bool),
		cursor:   0,
		height:   24,
	}

	// Test space to select
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = newModel.(bubbleModel)
	assert.True(t, m.selected[0])

	// Test space to deselect
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = newModel.(bubbleModel)
	assert.False(t, m.selected[0])

	// Test select all
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = newModel.(bubbleModel)
	assert.True(t, m.selected[0])
	assert.True(t, m.selected[1])
	assert.True(t, m.selected[2])

	// Test select none
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(bubbleModel)
	assert.False(t, m.selected[0])
	assert.False(t, m.selected[1])
	assert.False(t, m.selected[2])
}

func TestBubbleModel_Update_Quit(t *testing.T) {
	m := bubbleModel{
		items:    []string{"item1", "item2"},
		selected: make(map[int]bool),
		height:   24,
	}

	// Test quit with 'q'
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = newModel.(bubbleModel)
	assert.True(t, m.quitting)
	assert.False(t, m.confirmed)
	assert.NotNil(t, cmd) // Should return tea.Quit

	// Test confirm with Enter
	m = bubbleModel{
		items:    []string{"item1", "item2"},
		selected: make(map[int]bool),
		height:   24,
	}
	newModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(bubbleModel)
	assert.True(t, m.quitting)
	assert.True(t, m.confirmed)
	assert.NotNil(t, cmd)
}

func TestBubbleModel_View(t *testing.T) {
	m := bubbleModel{
		items:    []string{"item1", "item2", "item3"},
		selected: map[int]bool{1: true},
		cursor:   1,
		height:   24,
		width:    80,
	}

	view := m.View()

	// Should contain the title
	assert.Contains(t, view, "Select Dotfiles")
	assert.Contains(t, view, "1/3 selected")

	// Should contain items
	assert.Contains(t, view, "item1")
	assert.Contains(t, view, "item2")
	assert.Contains(t, view, "item3")

	// Should contain instructions
	assert.Contains(t, view, "navigate")
	assert.Contains(t, view, "toggle")
}

func TestBubbleModel_View_Quitting(t *testing.T) {
	m := bubbleModel{
		items:    []string{"item1", "item2"},
		selected: make(map[int]bool),
		quitting: true,
	}

	view := m.View()
	assert.Empty(t, view) // View should be empty when quitting
}

func TestBubbleModel_UpdateViewport(t *testing.T) {
	m := bubbleModel{
		items:       []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		selected:    make(map[int]bool),
		cursor:      0,
		viewportTop: 0,
		height:      10, // Small height to test scrolling
	}

	// Scroll down beyond viewport
	m.cursor = 5
	m.updateViewport()
	assert.True(t, m.viewportTop > 0) // Should have scrolled

	// Scroll up
	m.cursor = 0
	m.updateViewport()
	assert.Equal(t, 0, m.viewportTop) // Should be at top
}
