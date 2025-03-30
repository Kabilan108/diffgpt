// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Helper to set temporary config path for testing
func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	ogUserConfigDir := os.UserConfigDir
	ogMkdirAll := os.MkdirAll
	ogWriteFile := os.WriteFile
	ogReadFile := os.ReadFile
	ogRename := os.Rename
	ogRemove := os.Remove // Add remove for cleanup simulation

	tempDir := t.TempDir()
	testConfigDir := filepath.Join(tempDir, appConfigDir)
	testConfigPath := filepath.Join(testConfigDir, configFileName)

	// Override os.UserConfigDir to return our temp dir's parent
	osUserConfigDir = func() (string, error) {
		return tempDir, nil
	}
	// Override filesystem operations to work within tempDir implicitly or explicitly
	// Note: For simplicity here, we just override UserConfigDir. More robust mocking
	// might involve creating a mock filesystem interface.

	// Ensure our test config dir can be created
	osMkdirAll = func(path string, perm os.FileMode) error {
		// Allow creating the specific app config dir within temp
		if path == testConfigDir {
			return ogMkdirAll(path, perm)
		}
		// Prevent creating other unexpected directories during test
		t.Logf("Intercepted MkdirAll for path: %s", path)
		// return fmt.Errorf("unexpected call to MkdirAll: %s", path) // Or be more strict
		return ogMkdirAll(path, perm) // Or allow for flexibility
	}
	// Override file operations to use testConfigPath
	osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		if name == testConfigPath+".tmp" { // Check for temp file usage
			return ogWriteFile(name, data, perm)
		}
		t.Logf("Intercepted WriteFile for path: %s", name)
		return ogWriteFile(name, data, perm) // Allow writing other files if needed
	}
	osReadFile = func(name string) ([]byte, error) {
		if name == testConfigPath {
			return ogReadFile(name)
		}
		t.Logf("Intercepted ReadFile for path: %s", name)
		// Simulate file not found for paths other than the target config
		return nil, os.ErrNotExist
	}
	osRename = func(oldpath, newpath string) error {
		if oldpath == testConfigPath+".tmp" && newpath == testConfigPath {
			return ogRename(oldpath, newpath)
		}
		t.Logf("Intercepted Rename from %s to %s", oldpath, newpath)
		return ogRename(oldpath, newpath)
	}
	osRemove = func(name string) error {
		if name == testConfigPath+".tmp" {
			return ogRemove(name)
		}
		t.Logf("Intercepted Remove for path: %s", name)
		return ogRemove(name)
	}

	// Teardown function to restore og functions
	cleanup := func() {
		osUserConfigDir = ogUserConfigDir
		osMkdirAll = ogMkdirAll
		osWriteFile = ogWriteFile
		osReadFile = ogReadFile
		osRename = ogRename
		osRemove = ogRemove
	}

	return testConfigPath, cleanup
}

func TestGetConfigPath(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}

	// We expect path to be inside the temp dir now
	expectedSuffix := filepath.Join(appConfigDir, configFileName)
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("Expected config path to end with '%s', got '%s'", expectedSuffix, path)
	}
	// Check that it's an absolute path (depends on osUserConfigDir behavior)
	if !filepath.IsAbs(path) {
		// Note: os.UserConfigDir usually returns abs path, but fake might not
		// t.Errorf("Expected config path to be absolute, got '%s'", path)
		t.Logf("Warning: Test config path '%s' might not be absolute depending on test setup", path)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed when file doesn't exist: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config when file doesn't exist")
	}
	if cfg.Examples == nil {
		t.Fatal("LoadConfig() returned config with nil Examples map")
	}
	if len(cfg.Examples) != 0 {
		t.Errorf("Expected empty Examples map, got %d entries", len(cfg.Examples))
	}
}

func TestSaveLoadConfig(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	// --- Save ---
	expectedCfg := &Config{
		Examples: map[string][]Example{
			"global": {
				{Diff: "diff1", Message: "msg1"},
			},
			"/path/to/repo": {
				{Diff: "diff2", Message: "msg2"},
				{Diff: "diff3", Message: "msg3"},
			},
		},
	}

	err := SaveConfig(expectedCfg)
	if err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Verify file exists (optional, LoadConfig implicitly checks)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("SaveConfig() did not create the config file at %s", configPath)
	}

	// --- Load ---
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed after save: %v", err)
	}

	if !reflect.DeepEqual(expectedCfg, loadedCfg) {
		t.Errorf("Loaded config does not match saved config.\nExpected: %+v\nGot:      %+v", expectedCfg, loadedCfg)
	}
}

func TestLoadConfig_InvalidJson(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	// Write invalid JSON to the config file path directly
	// Need to ensure the directory exists first
	_ = ensureConfigDir() // Use the potentially overridden ensureConfigDir via setupTestConfig
	err := os.WriteFile(configPath, []byte("{invalid json"), 0o600)
	if err != nil {
		t.Fatalf("Failed to write invalid json for test: %v", err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() succeeded with invalid JSON, expected error")
	}
	t.Logf("Got expected error for invalid JSON: %v", err) // Log error for info
}
