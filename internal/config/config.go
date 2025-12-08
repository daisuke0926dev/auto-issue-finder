package config

// Config represents the merged configuration from all sources
type Config struct {
	ProjectDir      string
	DefaultTaskFile string
	MaxRetries      int
	LogDir          string
	StartFrom       int
	ClaudeFlags     []string
}

// MergeConfig merges configuration from multiple sources with priority:
// CLI flags > Environment variables > Project settings > Global settings > Defaults
//
// Parameters:
// - cliConfig: Configuration from CLI flags (highest priority)
// - envConfig: Configuration from environment variables
// - defaultConfig: Default configuration (lowest priority)
//
// Returns: Merged configuration
func MergeConfig(cliConfig, envConfig, defaultConfig *Config) *Config {
	merged := &Config{}

	// Project directory
	merged.ProjectDir = selectValue(
		cliConfig.ProjectDir,
		envConfig.ProjectDir,
		defaultConfig.ProjectDir,
	)

	// Default task file
	merged.DefaultTaskFile = selectValue(
		cliConfig.DefaultTaskFile,
		envConfig.DefaultTaskFile,
		defaultConfig.DefaultTaskFile,
	)

	// Max retries (special handling for integers)
	merged.MaxRetries = selectIntValue(
		cliConfig.MaxRetries, cliConfig.MaxRetries >= 0,
		envConfig.MaxRetries, envConfig.MaxRetries >= 0,
		defaultConfig.MaxRetries, true,
	)

	// Log directory
	merged.LogDir = selectValue(
		cliConfig.LogDir,
		envConfig.LogDir,
		defaultConfig.LogDir,
	)

	// Start from (special handling for integers)
	merged.StartFrom = selectIntValue(
		cliConfig.StartFrom, cliConfig.StartFrom >= 1,
		envConfig.StartFrom, envConfig.StartFrom >= 1,
		defaultConfig.StartFrom, true,
	)

	// Claude flags (arrays are merged, not replaced)
	merged.ClaudeFlags = mergeArrays(
		cliConfig.ClaudeFlags,
		envConfig.ClaudeFlags,
		defaultConfig.ClaudeFlags,
	)

	return merged
}

// selectValue selects the first non-empty string value
func selectValue(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// selectIntValue selects the first valid integer value
func selectIntValue(v1 int, hasV1 bool, v2 int, hasV2 bool, v3 int, hasV3 bool) int {
	if hasV1 {
		return v1
	}
	if hasV2 {
		return v2
	}
	if hasV3 {
		return v3
	}
	return 0
}

// mergeArrays merges arrays, preferring the first non-empty array
func mergeArrays(arrays ...[]string) []string {
	for _, arr := range arrays {
		if len(arr) > 0 {
			return arr
		}
	}
	return []string{}
}

// NewDefaultConfig returns the default configuration
func NewDefaultConfig() *Config {
	return &Config{
		ProjectDir:      "",
		DefaultTaskFile: "",
		MaxRetries:      3,
		LogDir:          "logs",
		StartFrom:       1,
		ClaudeFlags:     []string{},
	}
}

// FromEnv creates a Config from EnvConfig
func FromEnv(env *EnvConfig) *Config {
	cfg := &Config{
		MaxRetries: -1,
		StartFrom:  -1,
	}

	if env.HasProjectDir() {
		cfg.ProjectDir = env.ProjectDir
	}
	if env.HasDefaultTaskFile() {
		cfg.DefaultTaskFile = env.DefaultTaskFile
	}
	if env.HasMaxRetries() {
		cfg.MaxRetries = env.MaxRetries
	}
	if env.HasLogDir() {
		cfg.LogDir = env.LogDir
	}
	if env.HasStartFrom() {
		cfg.StartFrom = env.StartFrom
	}
	if env.HasClaudeFlags() {
		cfg.ClaudeFlags = env.ClaudeFlags
	}

	return cfg
}
