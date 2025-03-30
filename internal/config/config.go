package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Example struct {
	Diff    string `json:"diff"`
	Message string `json:"message"`
}

type Config struct {
	Examples map[string][]Example `json:"examples"`
}

const (
	configFileName = "config.json"
	appConfigDir   = "diffgpt"
)

var (
	osUserConfigDir = os.UserConfigDir
	osMkdirAll      = os.MkdirAll
	osWriteFile     = os.WriteFile
	osReadFile      = os.ReadFile
	osRename        = os.Rename
	osRemove        = os.Remove // Need this for SaveConfig cleanup check
)

func GetConfigPath() (string, error) {
	configDir, err := osUserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	appDir := filepath.Join(configDir, appConfigDir)
	return filepath.Join(appDir, configFileName), nil
}

func ensureConfigDir() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	appDir := filepath.Dir(configPath)

	if err := osMkdirAll(appDir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", appDir, err)
	}
	return nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{Examples: make(map[string][]Example)}

	data, err := osReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// no config file, return default empty
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	if cfg.Examples == nil {
		cfg.Examples = make(map[string][]Example)
	}

	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	tempFile := configPath + ".tmp"
	err = osWriteFile(tempFile, data, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write temporary config file %s: %w", tempFile, err)
	}

	if err := osRename(tempFile, configPath); err != nil {
		// Handle cleanup failure explicitly rather than ignoring it
		if removeErr := osRemove(tempFile); removeErr != nil {
			return fmt.Errorf("failed to rename temp file to %s: %w (and failed to remove temp file: %v)",
				configPath, err, removeErr)
		}
		return fmt.Errorf("failed to rename temporary config file to %s: %w", configPath, err)
	}

	return nil
}
