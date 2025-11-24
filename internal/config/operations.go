package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

// UpgradeConfig upgrades the configuration file at configPath to the latest format.
// It creates a timestamped backup before making changes and merges user values
// with new defaults. Deprecated fields are automatically migrated.
func UpgradeConfig(configPath string, force bool) (string, error) {
	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file does not exist: %s\nRun 'dot config init' to create one", configPath)
	}

	// Load existing config
	loader := NewLoader("dot", configPath)
	oldConfig, err := loader.LoadWithEnv()
	if err != nil {
		return "", fmt.Errorf("failed to load existing config: %w", err)
	}

	// Get backup directory
	backupDir, err := getBackupDir()
	if err != nil {
		return "", fmt.Errorf("failed to get backup directory: %w", err)
	}

	// Create backup
	backupPath, err := createBackup(configPath, backupDir)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Get new defaults
	newConfig := DefaultExtended()

	// Merge configs (preserve user values, add new defaults)
	merged := mergeConfigsForUpgrade(oldConfig, newConfig)

	// Migrate deprecated fields
	migrateDeprecatedFields(merged)

	// Validate merged config
	if err := merged.Validate(); err != nil {
		// Validation failed - restore original
		if restoreErr := os.Rename(backupPath, configPath); restoreErr != nil {
			return "", fmt.Errorf("validation failed and could not restore original: %v (restore error: %v)", err, restoreErr)
		}
		return "", fmt.Errorf("upgraded config failed validation, original restored: %w", err)
	}

	// Write upgraded config with header
	header := generateUpgradeHeader(backupPath, merged)
	if err := WriteConfigWithHeader(configPath, merged, header); err != nil {
		// Write failed - restore original
		if restoreErr := os.Rename(backupPath, configPath); restoreErr != nil {
			return "", fmt.Errorf("failed to write config and could not restore original: %v (restore error: %v)", err, restoreErr)
		}
		return "", fmt.Errorf("failed to write upgraded config, original restored: %w", err)
	}

	// Cleanup old backups (keep last 5)
	if err := cleanupOldBackups(backupDir, 5); err != nil {
		// Non-fatal: log but don't fail the upgrade
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup old backups: %v\n", err)
	}

	return backupPath, nil
}

// mergeConfigsForUpgrade merges old and new configs, preferring values from old when non-zero.
// This is specifically for config upgrades where we want to preserve user customizations
// while adding new fields with their defaults.
func mergeConfigsForUpgrade(old, new *ExtendedConfig) *ExtendedConfig {
	if old == nil {
		return new
	}
	if new == nil {
		return old
	}

	result := &ExtendedConfig{}
	mergeStructs(reflect.ValueOf(old).Elem(), reflect.ValueOf(new).Elem(), reflect.ValueOf(result).Elem())
	return result
}

// mergeStructs recursively merges two structs, preferring values from old when non-zero.
func mergeStructs(old, new, result reflect.Value) {
	if !old.IsValid() || !new.IsValid() || !result.IsValid() {
		return
	}

	t := old.Type()

	for i := 0; i < old.NumField(); i++ {
		oldField := old.Field(i)
		newField := new.Field(i)
		resultField := result.Field(i)

		if !resultField.CanSet() {
			continue
		}

		mergeField(oldField, newField, resultField)
	}

	_ = t // Suppress unused variable warning
}

// mergeField merges a single field based on its kind.
func mergeField(oldField, newField, resultField reflect.Value) {
	switch oldField.Kind() {
	case reflect.Struct:
		mergeStructField(oldField, newField, resultField)
	case reflect.Slice:
		mergeSliceField(oldField, newField, resultField)
	case reflect.Map:
		mergeMapField(oldField, newField, resultField)
	case reflect.String:
		mergeStringField(oldField, newField, resultField)
	case reflect.Bool:
		mergeBoolField(oldField, resultField)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		mergeIntField(oldField, newField, resultField)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mergeUintField(oldField, newField, resultField)
	case reflect.Float32, reflect.Float64:
		mergeFloatField(oldField, newField, resultField)
	default:
		mergeDefaultField(oldField, newField, resultField)
	}
}

func mergeStructField(oldField, newField, resultField reflect.Value) {
	// Check if entire struct is zero - if so, use new defaults
	if isZero(oldField) {
		resultField.Set(newField)
	} else {
		// Recursively merge nested structs
		mergeStructs(oldField, newField, resultField)
	}
}

func mergeSliceField(oldField, newField, resultField reflect.Value) {
	// Use old slice if non-empty, else new default
	if oldField.Len() > 0 {
		resultField.Set(oldField)
	} else {
		resultField.Set(newField)
	}
}

func mergeMapField(oldField, newField, resultField reflect.Value) {
	// Use old map if non-empty, else new default
	if oldField.Len() > 0 {
		resultField.Set(oldField)
	} else {
		resultField.Set(newField)
	}
}

func mergeStringField(oldField, newField, resultField reflect.Value) {
	// Use old string if non-empty, else new default
	if oldField.String() != "" {
		resultField.SetString(oldField.String())
	} else {
		resultField.SetString(newField.String())
	}
}

func mergeBoolField(oldField, resultField reflect.Value) {
	// For booleans, always use old value (even if false)
	// This preserves explicit false settings by the user
	resultField.SetBool(oldField.Bool())
}

func mergeIntField(oldField, newField, resultField reflect.Value) {
	// Use old int if non-zero, else new default
	if oldField.Int() != 0 {
		resultField.SetInt(oldField.Int())
	} else {
		resultField.SetInt(newField.Int())
	}
}

func mergeUintField(oldField, newField, resultField reflect.Value) {
	// Use old uint if non-zero, else new default
	if oldField.Uint() != 0 {
		resultField.SetUint(oldField.Uint())
	} else {
		resultField.SetUint(newField.Uint())
	}
}

func mergeFloatField(oldField, newField, resultField reflect.Value) {
	// Use old float if non-zero, else new default
	if oldField.Float() != 0 {
		resultField.SetFloat(oldField.Float())
	} else {
		resultField.SetFloat(newField.Float())
	}
}

func mergeDefaultField(oldField, newField, resultField reflect.Value) {
	// For other types, prefer old value
	if !isZero(oldField) {
		resultField.Set(oldField)
	} else {
		resultField.Set(newField)
	}
}

// isZero checks if a value is the zero value for its type.
func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// For structs, check if all fields are zero
		for i := 0; i < v.NumField(); i++ {
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

// migrateDeprecatedFields migrates deprecated fields to their new format.
func migrateDeprecatedFields(cfg *ExtendedConfig) {
	// Migrate ignore.overrides to negation patterns in ignore.patterns
	if len(cfg.Ignore.Overrides) > 0 {
		// Convert each override pattern to a negation pattern
		for _, pattern := range cfg.Ignore.Overrides {
			negationPattern := "!" + pattern
			// Add to patterns if not already there
			found := false
			for _, p := range cfg.Ignore.Patterns {
				if p == negationPattern {
					found = true
					break
				}
			}
			if !found {
				cfg.Ignore.Patterns = append(cfg.Ignore.Patterns, negationPattern)
			}
		}
	}
}

// getBackupDir returns the backup directory path, creating it if needed.
func getBackupDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	backupDir := filepath.Join(configHome, "dot", "backups")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return "", fmt.Errorf("cannot create backup directory: %w", err)
	}

	return backupDir, nil
}

// createBackup creates a timestamped backup of the config file.
func createBackup(configPath, backupDir string) (string, error) {
	// Read original config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("cannot read config file: %w", err)
	}

	// Generate timestamp-based filename
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("%s-config.bak", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	// Write backup
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return "", fmt.Errorf("cannot write backup file: %w", err)
	}

	return backupPath, nil
}

// cleanupOldBackups keeps only the most recent 'keep' backups, deleting older ones.
func cleanupOldBackups(backupDir string, keep int) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("cannot read backup directory: %w", err)
	}

	// Filter for backup files and collect their info
	type backupInfo struct {
		name string
		path string
		time time.Time
	}

	backups := make([]backupInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Match pattern: YYYYMMDD-HHMMSS-config.bak
		if !strings.HasSuffix(name, "-config.bak") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, backupInfo{
			name: name,
			path: filepath.Join(backupDir, name),
			time: info.ModTime(),
		})
	}

	// If we have fewer backups than the keep limit, nothing to delete
	if len(backups) <= keep {
		return nil
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].time.After(backups[j].time)
	})

	// Delete old backups beyond the keep limit
	for i := keep; i < len(backups); i++ {
		if err := os.Remove(backups[i].path); err != nil {
			return fmt.Errorf("cannot remove old backup %s: %w", backups[i].name, err)
		}
	}

	return nil
}

// generateUpgradeHeader generates a YAML header comment for upgraded configs.
func generateUpgradeHeader(backupPath string, cfg *ExtendedConfig) string {
	var header strings.Builder

	header.WriteString("# Dot Configuration\n")
	header.WriteString(fmt.Sprintf("# Upgraded on %s\n", time.Now().Format("2006-01-02 15:04:05")))
	header.WriteString(fmt.Sprintf("# Backup saved to: %s\n", backupPath))
	header.WriteString("#\n")

	// Add migration notes if overrides were migrated
	if len(cfg.Ignore.Overrides) > 0 {
		header.WriteString("# Deprecated fields migrated:\n")
		header.WriteString("#   - ignore.overrides → ignore.patterns (with ! prefix for negation)\n")
		header.WriteString("#\n")
	}

	header.WriteString("# See https://github.com/yaklabco/dot for documentation\n")
	header.WriteString("\n")

	return header.String()
}
