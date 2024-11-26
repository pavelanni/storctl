package config

const (
	// ToolName is the name of the tool, used for directory names and logging
	ToolName = "labshop"

	// DefaultConfigDir is the default directory for all tool-related files
	DefaultConfigDir = ".labshop"

	// KeysDir is the subdirectory name for storing SSH keys
	KeysDir = "keys"

	// ConfigFileName is the name of the configuration file
	ConfigFileName = "config.yaml"
)

// Time related constants
const (
	// DefaultTimeout is the default timeout for operations
	DefaultTimeout = "30s"

	// DefaultKeyTTL is the default time-to-live for SSH keys
	DefaultTTL = "1h"
)
