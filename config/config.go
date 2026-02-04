// Package config provides functionality to load and validate configuration
// from JSON configuration files for the Redis data processing tool.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InputConfig represents the input file configuration for different data types.
type InputConfig struct {
	Daily   []string `json:"daily"`   // Daily rank data files
	Total   []string `json:"total"`   // Total rank data files
	Reward  []string `json:"reward"`  // Reward record files
	Club    []string `json:"club"`    // Club data files
	Mail    []string `json:"mail"`    // Mail log files
	Month   []string `json:"month"`   // Month rank data files
	ClubPid []string `json:"clubpid"` // Club pid data files
}

// Config represents the complete configuration structure.
type Config struct {
	Input   InputConfig `json:"input"`   // Input file configuration
	Output  string      `json:"output"`  // Output Excel file path
	Mode    string      `json:"mode"`    // Processing mode: daily, total, reward, or all
	Archive bool        `json:"archive"` // Whether to archive files after processing
}

// LoadConfig loads and parses a configuration file from the specified path.
// It returns the parsed configuration and any error encountered.
func LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON configuration: %w", err)
	}

	// Resolve relative paths relative to config file directory
	configDir := filepath.Dir(configPath)
	if err := cfg.resolvePaths(configDir); err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// resolvePaths resolves relative paths in the configuration relative to the config file directory.
func (c *Config) resolvePaths(configDir string) error {
	// Resolve input file paths
	if err := c.Input.resolvePaths(configDir); err != nil {
		return err
	}

	// Resolve output path if relative
	if c.Output != "" && !filepath.IsAbs(c.Output) {
		c.Output = filepath.Join(configDir, c.Output)
	}

	return nil
}

// resolvePaths resolves relative paths in InputConfig.
func (ic *InputConfig) resolvePaths(configDir string) error {
	resolveFileList := func(files []string) ([]string, error) {
		resolved := make([]string, len(files))
		for i, file := range files {
			if filepath.IsAbs(file) {
				resolved[i] = file
			} else {
				resolved[i] = filepath.Join(configDir, file)
			}
		}
		return resolved, nil
	}

	var err error
	if ic.Daily, err = resolveFileList(ic.Daily); err != nil {
		return err
	}
	if ic.Total, err = resolveFileList(ic.Total); err != nil {
		return err
	}
	if ic.Reward, err = resolveFileList(ic.Reward); err != nil {
		return err
	}
	if ic.Club, err = resolveFileList(ic.Club); err != nil {
		return err
	}
	if ic.Mail, err = resolveFileList(ic.Mail); err != nil {
		return err
	}
	if ic.Month, err = resolveFileList(ic.Month); err != nil {
		return err
	}
	if ic.ClubPid, err = resolveFileList(ic.ClubPid); err != nil {
		return err
	}

	return nil
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	// Validate output path is provided
	if strings.TrimSpace(c.Output) == "" {
		return fmt.Errorf("output path is required")
	}

	// Validate processing mode
	validModes := map[string]bool{
		"daily":   true,
		"total":   true,
		"reward":  true,
		"mail":    true,
		"month":   true,
		"clubpid": true,
		"all":     true,
	}
	if !validModes[c.Mode] {
		return fmt.Errorf("invalid processing mode: %s (must be: daily, total, reward, mail, month, clubpid, or all)", c.Mode)
	}

	// Validate input files exist (if specified)
	if err := c.Input.validateFiles(); err != nil {
		return err
	}

	return nil
}

// validateFiles validates that all specified input files exist.
func (ic *InputConfig) validateFiles() error {
	validateFileList := func(files []string, category string) error {
		for _, file := range files {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return fmt.Errorf("%s file does not exist: %s", category, file)
			}
		}
		return nil
	}

	if err := validateFileList(ic.Daily, "daily"); err != nil {
		return err
	}
	if err := validateFileList(ic.Total, "total"); err != nil {
		return err
	}
	if err := validateFileList(ic.Reward, "reward"); err != nil {
		return err
	}
	if err := validateFileList(ic.Club, "club"); err != nil {
		return err
	}
	if err := validateFileList(ic.Mail, "mail"); err != nil {
		return err
	}
	if err := validateFileList(ic.Month, "month"); err != nil {
		return err
	}
	if err := validateFileList(ic.ClubPid, "clubpid"); err != nil {
		return err
	}

	return nil
}

// GetProcessedFiles returns a list of all input files that will be processed.
func (c *Config) GetProcessedFiles() []string {
	var files []string
	files = append(files, c.Input.Daily...)
	files = append(files, c.Input.Total...)
	files = append(files, c.Input.Reward...)
	files = append(files, c.Input.Club...)
	files = append(files, c.Input.Mail...)
	files = append(files, c.Input.Month...)
	files = append(files, c.Input.ClubPid...)
	return files
}
