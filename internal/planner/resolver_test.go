package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

// Task 7.1.1: Test ConflictType enumeration
func TestConflictTypeString(t *testing.T) {
	tests := []struct {
		name string
		ct   ConflictType
		want string
	}{
		{"file exists", ConflictFileExists, "file_exists"},
		{"wrong link", ConflictWrongLink, "wrong_link"},
		{"permission", ConflictPermission, "permission"},
		{"circular", ConflictCircular, "circular"},
		{"dir expected", ConflictDirExpected, "dir_expected"},
		{"file expected", ConflictFileExpected, "file_expected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Task 7.1.2: Test Conflict value object
func TestConflictCreation(t *testing.T) {
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetFilePath,
		"File exists at target location",
	)

	assert.Equal(t, ConflictFileExists, conflict.Type)
	assert.Equal(t, targetFilePath, conflict.Path)
	assert.Equal(t, "File exists at target location", conflict.Details)
	assert.NotNil(t, conflict.Context)
	assert.Empty(t, conflict.Suggestions)
}

func TestConflictWithContext(t *testing.T) {
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetFilePath,
		"File exists",
	)

	conflict = conflict.WithContext("size", "1024")
	conflict = conflict.WithContext("mode", "0644")

	assert.Equal(t, "1024", conflict.Context["size"])
	assert.Equal(t, "0644", conflict.Context["mode"])
}

func TestConflictWithSuggestion(t *testing.T) {
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()

	conflict := NewConflict(
		ConflictFileExists,
		targetFilePath,
		"File exists",
	)

	suggestion := Suggestion{
		Action:      "Use --backup flag",
		Explanation: "Preserves existing file",
	}

	conflict = conflict.WithSuggestion(suggestion)

	assert.Len(t, conflict.Suggestions, 1)
	assert.Equal(t, "Use --backup flag", conflict.Suggestions[0].Action)
}

// Task 7.1.3: Test Resolution Status Types
func TestResolutionStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status ResolutionStatus
		want   string
	}{
		{"ok", ResolveOK, "ok"},
		{"conflict", ResolveConflict, "conflict"},
		{"warning", ResolveWarning, "warning"},
		{"skip", ResolveSkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolutionOutcomeCreation(t *testing.T) {
	t.Run("ok status", func(t *testing.T) {
		outcome := ResolutionOutcome{
			Status:     ResolveOK,
			Operations: []domain.Operation{},
		}
		assert.Equal(t, ResolveOK, outcome.Status)
		assert.NotNil(t, outcome.Operations)
		assert.Nil(t, outcome.Conflict)
		assert.Nil(t, outcome.Warning)
	})

	t.Run("conflict status", func(t *testing.T) {
		targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
		targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()
		conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")

		outcome := ResolutionOutcome{
			Status:   ResolveConflict,
			Conflict: &conflict,
		}
		assert.Equal(t, ResolveConflict, outcome.Status)
		assert.NotNil(t, outcome.Conflict)
		assert.Equal(t, ConflictFileExists, outcome.Conflict.Type)
	})
}

// Task 7.1.4: Test ResolveResult Type
func TestResolveResultConstruction(t *testing.T) {
	t.Run("with operations", func(t *testing.T) {
		ops := []domain.Operation{}
		result := NewResolveResult(ops)
		assert.Len(t, result.Operations, 0)
		assert.Empty(t, result.Conflicts)
		assert.Empty(t, result.Warnings)
	})

	t.Run("with conflicts", func(t *testing.T) {
		targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
		targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()
		conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")

		result := NewResolveResult(nil)
		result = result.WithConflict(conflict)

		assert.Len(t, result.Conflicts, 1)
		assert.Equal(t, ConflictFileExists, result.Conflicts[0].Type)
	})

	t.Run("with warnings", func(t *testing.T) {
		warning := Warning{
			Message:  "File backed up",
			Severity: WarnInfo,
		}

		result := NewResolveResult(nil)
		result = result.WithWarning(warning)

		assert.Len(t, result.Warnings, 1)
		assert.Equal(t, "File backed up", result.Warnings[0].Message)
	})
}

func TestResolveResultQueries(t *testing.T) {
	t.Run("HasConflicts", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.False(t, result.HasConflicts())

		targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
		targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()
		conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")
		result = result.WithConflict(conflict)

		assert.True(t, result.HasConflicts())
	})

	t.Run("ConflictCount", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.Equal(t, 0, result.ConflictCount())

		targetPath1 := domain.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict1 := NewConflict(ConflictFileExists, targetPath1, "File exists")
		result = result.WithConflict(conflict1)

		targetPath2 := domain.NewFilePath("/home/user/.vimrc").Unwrap()
		conflict2 := NewConflict(ConflictWrongLink, targetPath2, "Wrong link")
		result = result.WithConflict(conflict2)

		assert.Equal(t, 2, result.ConflictCount())
	})

	t.Run("WarningCount", func(t *testing.T) {
		result := NewResolveResult(nil)
		assert.Equal(t, 0, result.WarningCount())

		warning := Warning{Message: "Test warning"}
		result = result.WithWarning(warning)

		assert.Equal(t, 1, result.WarningCount())
	})
}

// Task 7.1.5-7: Test Conflict Detection
func TestDetectFileExistsConflict(t *testing.T) {
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)

	current := CurrentState{
		Files: map[string]FileInfo{
			targetPath.String(): {Size: 100},
		},
		Links: make(map[string]LinkTarget),
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveConflict, outcome.Status)
	assert.NotNil(t, outcome.Conflict)
	assert.Equal(t, ConflictFileExists, outcome.Conflict.Type)
	assert.Contains(t, outcome.Conflict.Details, "File exists")
}

func TestDetectWrongLinkConflict(t *testing.T) {
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	wrongPath := domain.NewFilePath("/packages/other/dot-bashrc").Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)

	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: map[string]LinkTarget{
			targetPath.String(): {Target: wrongPath.String()},
		},
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveConflict, outcome.Status)
	assert.NotNil(t, outcome.Conflict)
	assert.Equal(t, ConflictWrongLink, outcome.Conflict.Type)
}

func TestDetectNoConflict(t *testing.T) {
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)

	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: make(map[string]LinkTarget),
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveOK, outcome.Status)
	assert.Nil(t, outcome.Conflict)
	assert.Len(t, outcome.Operations, 1)
}

func TestDetectLinkAlreadyCorrect(t *testing.T) {
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()

	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)

	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: map[string]LinkTarget{
			targetPath.String(): {Target: sourcePath.String()},
		},
	}

	outcome := detectLinkCreateConflicts(op, current)

	assert.Equal(t, ResolveSkip, outcome.Status)
	assert.Nil(t, outcome.Conflict)
}

func TestDetectDirCreateConflicts(t *testing.T) {
	t.Run("file exists where directory expected", func(t *testing.T) {
		dirPath := domain.NewFilePath("/home/user/.config").Unwrap()

		op := domain.NewDirCreate("dir-auto", dirPath)

		current := CurrentState{
			Files: map[string]FileInfo{
				dirPath.String(): {Size: 100},
			},
			Links: make(map[string]LinkTarget),
		}

		outcome := detectDirCreateConflicts(op, current)

		assert.Equal(t, ResolveConflict, outcome.Status)
		assert.NotNil(t, outcome.Conflict)
		assert.Equal(t, ConflictFileExpected, outcome.Conflict.Type)
	})

	t.Run("directory already exists", func(t *testing.T) {
		dirPath := domain.NewFilePath("/home/user/.config").Unwrap()

		op := domain.NewDirCreate("dir-auto", dirPath)

		current := CurrentState{
			Files: make(map[string]FileInfo),
			Links: make(map[string]LinkTarget),
			Dirs:  map[string]bool{dirPath.String(): true},
		}

		outcome := detectDirCreateConflicts(op, current)

		assert.Equal(t, ResolveSkip, outcome.Status)
	})

	t.Run("no conflict", func(t *testing.T) {
		dirPath := domain.NewFilePath("/home/user/.config").Unwrap()

		op := domain.NewDirCreate("dir-auto", dirPath)

		current := CurrentState{
			Files: make(map[string]FileInfo),
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		outcome := detectDirCreateConflicts(op, current)

		assert.Equal(t, ResolveOK, outcome.Status)
	})
}

// Task 7.4.1: Test Main Resolve Function
func TestResolveFunction(t *testing.T) {
	t.Run("no conflicts", func(t *testing.T) {
		sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
		targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

		ops := []domain.Operation{
			domain.NewLinkCreate("link-auto", sourcePath, targetPath),
		}

		current := CurrentState{
			Files: make(map[string]FileInfo),
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		policies := DefaultPolicies()

		result := Resolve(ops, current, policies, "/backup")

		assert.False(t, result.HasConflicts())
		assert.Len(t, result.Operations, 1)
		assert.Empty(t, result.Warnings)
	})

	t.Run("with conflict using fail policy", func(t *testing.T) {
		sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
		targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

		ops := []domain.Operation{
			domain.NewLinkCreate("link-auto", sourcePath, targetPath),
		}

		current := CurrentState{
			Files: map[string]FileInfo{
				targetPath.String(): {Size: 100},
			},
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		policies := DefaultPolicies() // Defaults to PolicyFail

		result := Resolve(ops, current, policies, "/backup")

		assert.True(t, result.HasConflicts())
		assert.Len(t, result.Conflicts, 1)
		assert.Equal(t, ConflictFileExists, result.Conflicts[0].Type)

		// Should have suggestions
		assert.NotEmpty(t, result.Conflicts[0].Suggestions)
	})

	t.Run("with conflict using skip policy", func(t *testing.T) {
		sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
		targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

		ops := []domain.Operation{
			domain.NewLinkCreate("link-auto", sourcePath, targetPath),
		}

		current := CurrentState{
			Files: map[string]FileInfo{
				targetPath.String(): {Size: 100},
			},
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		policies := DefaultPolicies()
		policies.OnFileExists = PolicySkip

		result := Resolve(ops, current, policies, "/backup")

		assert.False(t, result.HasConflicts())
		assert.Empty(t, result.Operations) // Operation was skipped
		assert.Len(t, result.Warnings, 1)
	})
}

// Task 7.4.2: Test Conflict Aggregation
func TestConflictAggregation(t *testing.T) {
	source1 := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	target1 := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

	source2 := domain.NewFilePath("/packages/vim/dot-vimrc").Unwrap()
	target2 := domain.NewTargetPath("/home/user/.vimrc").Unwrap()

	ops := []domain.Operation{
		domain.NewLinkCreate("link-auto", source1, target1),
		domain.NewLinkCreate("link-auto", source2, target2),
	}

	current := CurrentState{
		Files: map[string]FileInfo{
			target1.String(): {Size: 100},
			target2.String(): {Size: 200},
		},
		Links: make(map[string]LinkTarget),
		Dirs:  make(map[string]bool),
	}

	policies := DefaultPolicies()

	result := Resolve(ops, current, policies, "/backup")

	// Both operations should have conflicts
	assert.True(t, result.HasConflicts())
	assert.Equal(t, 2, result.ConflictCount())

	// Both conflicts should have suggestions
	for _, c := range result.Conflicts {
		assert.NotEmpty(t, c.Suggestions)
	}
}

func TestMixedOperations(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	dirPath := domain.NewFilePath("/home/user/.config").Unwrap()

	ops := []domain.Operation{
		domain.NewLinkCreate("link-auto", sourcePath, targetPath),
		domain.NewDirCreate("dir-auto", dirPath),
	}

	current := CurrentState{
		Files: map[string]FileInfo{
			targetPath.String(): {Size: 100}, // Conflict for link
		},
		Links: make(map[string]LinkTarget),
		Dirs:  make(map[string]bool), // No conflict for dir
	}

	policies := DefaultPolicies()
	policies.OnFileExists = PolicySkip

	result := Resolve(ops, current, policies, "/backup")

	// One operation skipped (link), one succeeded (dir)
	assert.False(t, result.HasConflicts())
	assert.Len(t, result.Operations, 1) // Only dir create
	assert.Len(t, result.Warnings, 1)   // Warning for skipped link
}

// Additional coverage tests
func TestWarningSeverityString(t *testing.T) {
	tests := []struct {
		name     string
		severity WarningSeverity
		want     string
	}{
		{"info", WarnInfo, "info"},
		{"caution", WarnCaution, "caution"},
		{"danger", WarnDanger, "danger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.severity.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveOperationWithAllTypes(t *testing.T) {
	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: make(map[string]LinkTarget),
		Dirs:  make(map[string]bool),
	}
	policies := DefaultPolicies()

	t.Run("LinkDelete passes through", func(t *testing.T) {
		linkPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
		op := domain.NewLinkDelete("link-del-auto", linkPath)

		outcome := resolveOperation(op, current, policies, "")

		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 1)
	})

	t.Run("DirDelete passes through", func(t *testing.T) {
		dirPath := domain.NewFilePath("/home/user/.config").Unwrap()
		op := domain.NewDirDelete("dir-del-auto", dirPath)

		outcome := resolveOperation(op, current, policies, "")

		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 1)
	})

	t.Run("FileMove passes through", func(t *testing.T) {
		source := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
		dest := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
		op := domain.NewFileMove("move-auto", source, dest)

		outcome := resolveOperation(op, current, policies, "")

		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 1)
	})

	t.Run("FileBackup passes through", func(t *testing.T) {
		source := domain.NewFilePath("/home/user/.bashrc").Unwrap()
		backup := domain.NewFilePath("/backup/.bashrc").Unwrap()
		op := domain.NewFileBackup("backup-auto", source, backup)

		outcome := resolveOperation(op, current, policies, "")

		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 1)
	})
}

func TestApplyPolicyToDirCreate(t *testing.T) {
	dirPath := domain.NewFilePath("/home/user/.config").Unwrap()
	op := domain.NewDirCreate("dir-auto", dirPath)
	conflict := NewConflict(ConflictFileExpected, dirPath, "File exists")

	t.Run("fail policy", func(t *testing.T) {
		outcome := applyPolicyToDirCreate(op, conflict, PolicyFail)
		assert.Equal(t, ResolveConflict, outcome.Status)
		assert.NotNil(t, outcome.Conflict)
	})

	t.Run("skip policy", func(t *testing.T) {
		outcome := applyPolicyToDirCreate(op, conflict, PolicySkip)
		assert.Equal(t, ResolveSkip, outcome.Status)
		assert.NotNil(t, outcome.Warning)
		assert.Contains(t, outcome.Warning.Message, "Skipping")
	})

	t.Run("unknown policy defaults to fail", func(t *testing.T) {
		outcome := applyPolicyToDirCreate(op, conflict, ResolutionPolicy(999))
		assert.Equal(t, ResolveConflict, outcome.Status)
	})
}

func TestApplyPolicyToLinkCreateEdgeCases(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	targetFilePath := domain.NewFilePath(targetPath.String()).Unwrap()
	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)
	conflict := NewConflict(ConflictFileExists, targetFilePath, "File exists")

	t.Run("backup policy creates backup and delete operations", func(t *testing.T) {
		outcome := applyPolicyToLinkCreate(op, conflict, PolicyBackup, "/backup")
		// Should create FileBackup, FileDelete, and LinkCreate operations
		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 3)
	})

	t.Run("overwrite policy creates delete operation", func(t *testing.T) {
		outcome := applyPolicyToLinkCreate(op, conflict, PolicyOverwrite, "/backup")
		// Should create FileDelete and LinkCreate operations
		assert.Equal(t, ResolveOK, outcome.Status)
		assert.Len(t, outcome.Operations, 2)
	})

	t.Run("unknown policy defaults to fail", func(t *testing.T) {
		outcome := applyPolicyToLinkCreate(op, conflict, ResolutionPolicy(999), "/backup")
		assert.Equal(t, ResolveConflict, outcome.Status)
	})
}

func TestResolveLinkCreateWithDifferentConflicts(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	op := domain.NewLinkCreate("link-auto", sourcePath, targetPath)

	policies := DefaultPolicies()
	policies.OnWrongLink = PolicySkip
	policies.OnPermissionErr = PolicySkip

	t.Run("wrong link with skip policy", func(t *testing.T) {
		wrongPath := domain.NewFilePath("/packages/other/file").Unwrap()
		current := CurrentState{
			Files: make(map[string]FileInfo),
			Links: map[string]LinkTarget{
				targetPath.String(): {Target: wrongPath.String()},
			},
			Dirs: make(map[string]bool),
		}

		outcome := resolveLinkCreate(op, current, policies, "")
		assert.Equal(t, ResolveSkip, outcome.Status)
	})
}

func TestResolveDirCreateEdgeCases(t *testing.T) {
	dirPath := domain.NewFilePath("/home/user/.config").Unwrap()
	op := domain.NewDirCreate("dir-auto", dirPath)

	policies := DefaultPolicies()
	policies.OnTypeMismatch = PolicySkip

	t.Run("type mismatch with skip policy", func(t *testing.T) {
		current := CurrentState{
			Files: map[string]FileInfo{
				dirPath.String(): {Size: 100},
			},
			Links: make(map[string]LinkTarget),
			Dirs:  make(map[string]bool),
		}

		outcome := resolveDirCreate(op, current, policies)
		assert.Equal(t, ResolveSkip, outcome.Status)
		assert.NotNil(t, outcome.Warning)
	})
}

func TestResolveWithWarnings(t *testing.T) {
	sourcePath := domain.NewFilePath("/packages/bash/dot-bashrc").Unwrap()
	targetPath := domain.NewTargetPath("/home/user/.bashrc").Unwrap()

	ops := []domain.Operation{
		domain.NewLinkCreate("link-auto", sourcePath, targetPath),
	}

	current := CurrentState{
		Files: make(map[string]FileInfo),
		Links: map[string]LinkTarget{
			targetPath.String(): {Target: sourcePath.String()},
		},
		Dirs: make(map[string]bool),
	}

	policies := DefaultPolicies()

	result := Resolve(ops, current, policies, "/backup")

	// Link already correct, should skip
	assert.False(t, result.HasConflicts())
	assert.Empty(t, result.Operations)
	assert.Empty(t, result.Warnings)
}
