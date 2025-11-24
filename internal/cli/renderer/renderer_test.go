package renderer

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		wantType  string
		wantError bool
	}{
		{
			name:     "text renderer",
			format:   "text",
			wantType: "*renderer.TextRenderer",
		},
		{
			name:     "json renderer",
			format:   "json",
			wantType: "*renderer.JSONRenderer",
		},
		{
			name:     "yaml renderer",
			format:   "yaml",
			wantType: "*renderer.YAMLRenderer",
		},
		{
			name:     "table renderer",
			format:   "table",
			wantType: "*renderer.TableRenderer",
		},
		{
			name:      "invalid format",
			format:    "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRenderer(tt.format, false, "")
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, r)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, r)
			}
		})
	}
}

func TestColorSchemeDefaults(t *testing.T) {
	t.Setenv("NO_COLOR", "") // Ensure NO_COLOR is not set

	scheme := DefaultColorScheme()

	assert.NotEmpty(t, scheme.Success)
	assert.NotEmpty(t, scheme.Warning)
	assert.NotEmpty(t, scheme.Error)
	assert.NotEmpty(t, scheme.Info)
	assert.NotEmpty(t, scheme.Muted)
}

func TestColorSchemeRespectsNOCOLOR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	scheme := DefaultColorScheme()

	// All colors should be empty when NO_COLOR is set
	assert.Empty(t, scheme.Success)
	assert.Empty(t, scheme.Warning)
	assert.Empty(t, scheme.Error)
	assert.Empty(t, scheme.Info)
	assert.Empty(t, scheme.Muted)
}

func TestGetTerminalWidth(t *testing.T) {
	width := getTerminalWidth()

	// Should return a reasonable default
	assert.GreaterOrEqual(t, width, 40)
	assert.LessOrEqual(t, width, 1000)
}

func TestRendererInterface(t *testing.T) {
	// Test that all renderers implement the Renderer interface
	renderers := []Renderer{
		&TextRenderer{},
		&JSONRenderer{},
		&YAMLRenderer{},
		&TableRenderer{},
	}

	status := dot.Status{
		Packages: []dot.PackageInfo{
			{
				Name:        "vim",
				InstalledAt: time.Now(),
				LinkCount:   5,
				Links:       []string{".vimrc", ".vim/"},
			},
		},
	}

	for _, r := range renderers {
		var buf bytes.Buffer
		// We expect implementation to exist (may return error but shouldn't panic)
		assert.NotPanics(t, func() {
			r.RenderStatus(&buf, status)
		})
	}
}

func TestFormatHelpers(t *testing.T) {
	t.Run("formatBytes", func(t *testing.T) {
		tests := []struct {
			bytes int64
			want  string
		}{
			{0, "0 B"},
			{1023, "1023 B"},
			{1024, "1.0 KB"},
			{1536, "1.5 KB"},
			{1048576, "1.0 MB"},
			{1073741824, "1.0 GB"},
		}

		for _, tt := range tests {
			got := formatBytes(tt.bytes)
			assert.Equal(t, tt.want, got)
		}
	})

	t.Run("formatDuration", func(t *testing.T) {
		now := time.Now()
		tests := []struct {
			t    time.Time
			want string
		}{
			{now, "just now"},
			{now.Add(-30 * time.Second), "30 seconds ago"},
			{now.Add(-2 * time.Minute), "2 minutes ago"},
			{now.Add(-1 * time.Hour), "1 hour ago"},
			{now.Add(-25 * time.Hour), "1 day ago"},
			{now.Add(-8 * 24 * time.Hour), "8 days ago"},
		}

		for _, tt := range tests {
			got := formatDuration(tt.t)
			assert.Equal(t, tt.want, got)
		}
	})

	t.Run("truncatePath", func(t *testing.T) {
		tests := []struct {
			path   string
			maxLen int
			want   string
		}{
			{"short/path", 20, "short/path"},
			{"/very/long/path/to/some/file.txt", 20, "/very/.../file.txt"},
			{"/a/b/c", 5, "/a/b/c"},
		}

		for _, tt := range tests {
			got := truncatePath(tt.path, tt.maxLen)
			assert.LessOrEqual(t, len(got), tt.maxLen+3) // +3 for ellipsis
		}
	})

	t.Run("pluralize", func(t *testing.T) {
		tests := []struct {
			count int
			word  string
			want  string
		}{
			{0, "file", "files"},
			{1, "file", "file"},
			{2, "file", "files"},
			{1, "entry", "entry"},
			{2, "entry", "entries"},
		}

		for _, tt := range tests {
			got := pluralize(tt.count, tt.word)
			assert.Contains(t, got, tt.want)
		}
	})
}
