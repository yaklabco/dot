package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand_Homebrew(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "simple formula",
			pkgName:  "dot",
			wantArgs: []string{"upgrade", "dot"},
			wantErr:  false,
		},
		{
			name:     "tap-qualified formula",
			pkgName:  "yaklabco/dot/dot",
			wantArgs: []string{"upgrade", "yaklabco/dot/dot"},
			wantErr:  false,
		},
		{
			name:    "shell injection attempt semicolon",
			pkgName: "dot; rm -rf /",
			wantErr: true,
		},
		{
			name:    "shell injection attempt pipe",
			pkgName: "dot | cat /etc/passwd",
			wantErr: true,
		},
		{
			name:    "shell injection attempt backtick",
			pkgName: "dot`whoami`",
			wantErr: true,
		},
		{
			name:    "shell injection attempt dollar paren",
			pkgName: "dot$(whoami)",
			wantErr: true,
		},
		{
			name:    "empty package name",
			pkgName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := NewCommand(SourceHomebrew, tt.pkgName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "brew", cmd.Name())
			assert.Equal(t, tt.wantArgs, cmd.Args())
			assert.Equal(t, SourceHomebrew, cmd.Source())
		})
	}
}

func TestNewCommand_Apt(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "valid package",
			pkgName:  "dot",
			wantArgs: []string{"apt-get", "install", "--only-upgrade", "-y", "dot"},
			wantErr:  false,
		},
		{
			name:     "package with numbers",
			pkgName:  "python3",
			wantArgs: []string{"apt-get", "install", "--only-upgrade", "-y", "python3"},
			wantErr:  false,
		},
		{
			name:    "invalid package uppercase",
			pkgName: "Dot",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := NewCommand(SourceApt, tt.pkgName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "sudo", cmd.Name())
			assert.Equal(t, tt.wantArgs, cmd.Args())
		})
	}
}

func TestNewCommand_Pacman(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "valid package",
			pkgName:  "dot",
			wantArgs: []string{"pacman", "-S", "--noconfirm", "dot"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := NewCommand(SourcePacman, tt.pkgName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "sudo", cmd.Name())
			assert.Equal(t, tt.wantArgs, cmd.Args())
		})
	}
}

func TestNewCommand_GoInstall(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		wantArgs []string
		wantErr  bool
	}{
		{
			name:     "valid module@latest",
			pkgName:  "github.com/yaklabco/dot/cmd/dot@latest",
			wantArgs: []string{"install", "github.com/yaklabco/dot/cmd/dot@latest"},
			wantErr:  false,
		},
		{
			name:     "valid module@version",
			pkgName:  "github.com/yaklabco/dot/cmd/dot@v1.0.0",
			wantArgs: []string{"install", "github.com/yaklabco/dot/cmd/dot@v1.0.0"},
			wantErr:  false,
		},
		{
			name:    "missing version",
			pkgName: "github.com/yaklabco/dot/cmd/dot",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := NewCommand(SourceGoInstall, tt.pkgName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "go", cmd.Name())
			assert.Equal(t, tt.wantArgs, cmd.Args())
		})
	}
}

func TestNewCommand_UnsupportedSource(t *testing.T) {
	_, err := NewCommand(SourceManual, "anything")
	assert.Error(t, err)
}

func TestCommand_String(t *testing.T) {
	cmd, err := NewCommand(SourceHomebrew, "dot")
	require.NoError(t, err)
	assert.Equal(t, "brew upgrade dot", cmd.String())
}

func TestContainsShellMetachars(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"dot", false},
		{"yaklabco/dot/dot", false},
		{"dot;echo", true},
		{"dot|cat", true},
		{"dot&echo", true},
		{"dot$HOME", true},
		{"dot`id`", true},
		{"dot\"test", true},
		{"dot'test", true},
		{"dot\\test", true},
		{"dot<file", true},
		{"dot>file", true},
		{"dot()", true},
		{"dot{}", true},
		{"dot[]", true},
		{"dot!", true},
		{"dot*", true},
		{"dot?", true},
		{"dot~", true},
		{"dot#comment", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, containsShellMetachars(tt.input))
		})
	}
}

func TestValidateArgument(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		pattern string
		wantErr bool
	}{
		{"empty argument", "", "", true},
		{"null byte", "dot\x00inject", "", true},
		{"shell semicolon", "dot;echo", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgument(tt.arg, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
