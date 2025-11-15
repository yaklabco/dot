package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/domain"
)

// Writer handles writing configuration to files.
type Writer struct {
	path string
}

// NewWriter creates a configuration writer.
func NewWriter(path string) *Writer {
	return &Writer{
		path: path,
	}
}

// Write writes configuration to file.
func (w *Writer) Write(cfg *ExtendedConfig, opts WriteOptions) error {
	// Ensure directory exists
	dir := filepath.Dir(w.path)
	if err := os.MkdirAll(dir, domain.PermUserRWX); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Marshal config based on format
	data, err := w.marshal(cfg, opts)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(w.path, data, domain.PermUserRW); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// WriteDefault writes default configuration with comments.
func (w *Writer) WriteDefault(opts WriteOptions) error {
	cfg := DefaultExtended()
	opts.IncludeComments = opts.IncludeComments || opts.Format == "yaml"
	return w.Write(cfg, opts)
}

// Update updates specific value in configuration file.
func (w *Writer) Update(key string, value interface{}) error {
	// Load existing config
	var cfg *ExtendedConfig
	var err error

	if fileExists(w.path) {
		cfg, err = LoadExtendedFromFile(w.path)
		if err != nil {
			return fmt.Errorf("load existing config: %w", err)
		}
	} else {
		// File doesn't exist, create with default
		cfg = DefaultExtended()
	}

	// Update value
	if err := w.setValue(cfg, key, value); err != nil {
		return fmt.Errorf("set value: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Write back
	opts := WriteOptions{
		Format:          w.DetectFormat(),
		IncludeComments: false,
	}
	return w.Write(cfg, opts)
}

// WriteOptions controls configuration file output.
type WriteOptions struct {
	Format          string // yaml, json, toml
	IncludeComments bool
	Indent          int
}

// marshal converts config to bytes in specified format using strategy pattern.
func (w *Writer) marshal(cfg *ExtendedConfig, opts WriteOptions) ([]byte, error) {
	format := opts.Format
	if format == "" {
		format = w.DetectFormat()
	}

	strategy, err := GetStrategy(format)
	if err != nil {
		return nil, err
	}

	marshalOpts := MarshalOptions{
		IncludeComments: opts.IncludeComments,
		Indent:          opts.Indent,
	}

	return strategy.Marshal(cfg, marshalOpts)
}

// DetectFormat detects format from file extension.
func (w *Writer) DetectFormat() string {
	ext := filepath.Ext(w.path)
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	default:
		return "yaml"
	}
}

// setValue sets a configuration value by dotted key path.
func (w *Writer) setValue(cfg *ExtendedConfig, key string, value interface{}) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key: %s (must be section.field)", key)
	}

	section := parts[0]
	field := parts[1]

	switch section {
	case "directories":
		return setDirectoriesValue(&cfg.Directories, field, value)
	case "logging":
		return setLoggingValue(&cfg.Logging, field, value)
	case "symlinks":
		return setSymlinksValue(&cfg.Symlinks, field, value)
	case "ignore":
		return setIgnoreValue(&cfg.Ignore, field, value)
	case "dotfile":
		return setDotfileValue(&cfg.Dotfile, field, value)
	case "output":
		return setOutputValue(&cfg.Output, field, value)
	case "operations":
		return setOperationsValue(&cfg.Operations, field, value)
	case "packages":
		return setPackagesValue(&cfg.Packages, field, value)
	case "doctor":
		return setDoctorValue(&cfg.Doctor, field, value)
	case "experimental":
		return setExperimentalValue(&cfg.Experimental, field, value)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
}

func setDirectoriesValue(cfg *DirectoriesConfig, field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("directories.%s: value must be string", field)
	}

	switch field {
	case "package":
		cfg.Package = str
	case "target":
		cfg.Target = str
	case "manifest":
		cfg.Manifest = str
	default:
		return fmt.Errorf("unknown field: directories.%s", field)
	}

	return nil
}

func setLoggingValue(cfg *LoggingConfig, field string, value interface{}) error {
	switch field {
	case "level", "format", "destination", "file":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("logging.%s: value must be string", field)
		}

		switch field {
		case "level":
			cfg.Level = str
		case "format":
			cfg.Format = str
		case "destination":
			cfg.Destination = str
		case "file":
			cfg.File = str
		}
	default:
		return fmt.Errorf("unknown field: logging.%s", field)
	}

	return nil
}

func setSymlinksValue(cfg *SymlinksConfig, field string, value interface{}) error {
	switch field {
	case "mode", "backup_suffix":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("symlinks.%s: value must be string", field)
		}

		switch field {
		case "mode":
			cfg.Mode = str
		case "backup_suffix":
			cfg.BackupSuffix = str
		}

	case "folding", "overwrite", "backup":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("symlinks.%s: value must be bool", field)
		}

		switch field {
		case "folding":
			cfg.Folding = b
		case "overwrite":
			cfg.Overwrite = b
		case "backup":
			cfg.Backup = b
		}

	default:
		return fmt.Errorf("unknown field: symlinks.%s", field)
	}

	return nil
}

func setIgnoreValue(cfg *IgnoreConfig, field string, value interface{}) error {
	switch field {
	case "use_defaults":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ignore.%s: value must be bool", field)
		}
		cfg.UseDefaults = b

	case "patterns", "overrides":
		// Accept both []string and string
		var arr []string
		switch v := value.(type) {
		case []string:
			arr = v
		case string:
			// Split comma-separated string
			arr = strings.Split(v, ",")
			for i := range arr {
				arr[i] = strings.TrimSpace(arr[i])
			}
		default:
			return fmt.Errorf("ignore.%s: value must be []string or string", field)
		}

		switch field {
		case "patterns":
			cfg.Patterns = arr
		case "overrides":
			cfg.Overrides = arr
		}

	default:
		return fmt.Errorf("unknown field: ignore.%s", field)
	}

	return nil
}

func setDotfileValue(cfg *DotfileConfig, field string, value interface{}) error {
	switch field {
	case "translate":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("dotfile.%s: value must be bool", field)
		}
		cfg.Translate = b

	case "prefix":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("dotfile.%s: value must be string", field)
		}
		cfg.Prefix = str

	default:
		return fmt.Errorf("unknown field: dotfile.%s", field)
	}

	return nil
}

func setOutputValue(cfg *OutputConfig, field string, value interface{}) error {
	switch field {
	case "format", "color":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("output.%s: value must be string", field)
		}

		switch field {
		case "format":
			cfg.Format = str
		case "color":
			cfg.Color = str
		}

	case "progress":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("output.%s: value must be bool", field)
		}
		cfg.Progress = b

	case "verbosity", "width":
		var i int
		switch v := value.(type) {
		case int:
			i = v
		case float64:
			i = int(v)
		default:
			return fmt.Errorf("output.%s: value must be int", field)
		}

		switch field {
		case "verbosity":
			cfg.Verbosity = i
		case "width":
			cfg.Width = i
		}

	default:
		return fmt.Errorf("unknown field: output.%s", field)
	}

	return nil
}

func setOperationsValue(cfg *OperationsConfig, field string, value interface{}) error {
	switch field {
	case "dry_run", "atomic":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("operations.%s: value must be bool", field)
		}

		switch field {
		case "dry_run":
			cfg.DryRun = b
		case "atomic":
			cfg.Atomic = b
		}

	case "max_parallel":
		var i int
		switch v := value.(type) {
		case int:
			i = v
		case float64:
			i = int(v)
		default:
			return fmt.Errorf("operations.%s: value must be int", field)
		}
		cfg.MaxParallel = i

	default:
		return fmt.Errorf("unknown field: operations.%s", field)
	}

	return nil
}

func setPackagesValue(cfg *PackagesConfig, field string, value interface{}) error {
	switch field {
	case "sort_by":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("packages.%s: value must be string", field)
		}
		cfg.SortBy = str

	case "auto_discover", "validate_names":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("packages.%s: value must be bool", field)
		}

		switch field {
		case "auto_discover":
			cfg.AutoDiscover = b
		case "validate_names":
			cfg.ValidateNames = b
		}

	default:
		return fmt.Errorf("unknown field: packages.%s", field)
	}

	return nil
}

func setDoctorValue(cfg *DoctorConfig, field string, value interface{}) error {
	b, ok := value.(bool)
	if !ok {
		return fmt.Errorf("doctor.%s: value must be bool", field)
	}

	switch field {
	case "auto_fix":
		cfg.AutoFix = b
	case "check_manifest":
		cfg.CheckManifest = b
	case "check_broken_links":
		cfg.CheckBrokenLinks = b
	case "check_orphaned":
		cfg.CheckOrphaned = b
	case "check_permissions":
		cfg.CheckPermissions = b
	default:
		return fmt.Errorf("unknown field: doctor.%s", field)
	}

	return nil
}

func setExperimentalValue(cfg *ExperimentalConfig, field string, value interface{}) error {
	b, ok := value.(bool)
	if !ok {
		return fmt.Errorf("experimental.%s: value must be bool", field)
	}

	switch field {
	case "parallel":
		cfg.Parallel = b
	case "profiling":
		cfg.Profiling = b
	default:
		return fmt.Errorf("unknown field: experimental.%s", field)
	}

	return nil
}

// WriteConfigWithHeader writes configuration with a custom header comment.
func WriteConfigWithHeader(path string, cfg *ExtendedConfig, header string) error {
	writer := NewWriter(path)

	// Marshal config
	opts := WriteOptions{
		Format:          writer.DetectFormat(),
		IncludeComments: false,
		Indent:          2,
	}

	data, err := writer.marshal(cfg, opts)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Prepend header
	finalData := []byte(header)
	finalData = append(finalData, data...)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, domain.PermUserRWX); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(path, finalData, domain.PermUserRW); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
