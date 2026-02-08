package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConnectionInfo_ValidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rd-engine.json")

	config := ConnectionInfo{
		User:     "example_user",
		Password: "example_password",
		Host:     "192.168.1.1",
		Port:     8080,
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(t, err)

	originalConfigPath := configPath
	originalDefaultConfigPath := DefaultConfigPath
	t.Cleanup(func() {
		configPath = originalConfigPath
		DefaultConfigPath = originalDefaultConfigPath
	})

	configPath = configFile
	DefaultConfigPath = configFile

	result, err := GetConnectionInfo(false)
	require.NoError(t, err)
	assert.Equal(t, "example_user", result.User)
	assert.Equal(t, "example_password", result.Password)
	assert.Equal(t, "192.168.1.1", result.Host)
	assert.Equal(t, 8080, result.Port)
}

func TestGetConnectionInfo_MissingConfigFile_MayBeMissing(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.json")

	originalConfigPath := configPath
	originalDefaultConfigPath := DefaultConfigPath
	t.Cleanup(func() {
		configPath = originalConfigPath
		DefaultConfigPath = originalDefaultConfigPath
	})

	configPath = ""
	DefaultConfigPath = nonExistentFile

	result, err := GetConnectionInfo(true)
	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestGetConnectionInfo_MissingConfigFile_Required(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.json")

	originalConfigPath := configPath
	originalDefaultConfigPath := DefaultConfigPath
	t.Cleanup(func() {
		configPath = originalConfigPath
		DefaultConfigPath = originalDefaultConfigPath
	})

	configPath = nonExistentFile
	DefaultConfigPath = nonExistentFile

	result, err := GetConnectionInfo(false)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestGetConnectionInfo_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rd-engine.json")

	err := os.WriteFile(configFile, []byte("not valid json"), 0600)
	require.NoError(t, err)

	originalConfigPath := configPath
	originalDefaultConfigPath := DefaultConfigPath
	t.Cleanup(func() {
		configPath = originalConfigPath
		DefaultConfigPath = originalDefaultConfigPath
	})

	configPath = configFile
	DefaultConfigPath = configFile

	result, err := GetConnectionInfo(false)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing config file")
}

func TestGetConnectionInfo_CLIOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rd-engine.json")

	config := ConnectionInfo{
		User:     "config_user",
		Password: "config_password",
		Host:     "config_host",
		Port:     9999,
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(t, err)

	originalConfigPath := configPath
	originalDefaultConfigPath := DefaultConfigPath
	originalConnectionSettings := connectionSettings
	t.Cleanup(func() {
		configPath = originalConfigPath
		DefaultConfigPath = originalDefaultConfigPath
		connectionSettings = originalConnectionSettings
	})

	configPath = configFile
	DefaultConfigPath = configFile
	connectionSettings = ConnectionInfo{
		User:     "override_user",
		Password: "override_password",
		Host:     "override_host",
		Port:     1234,
	}

	result, err := GetConnectionInfo(false)
	require.NoError(t, err)
	assert.Equal(t, "override_user", result.User)
	assert.Equal(t, "override_password", result.Password)
	assert.Equal(t, "override_host", result.Host)
	assert.Equal(t, 1234, result.Port)
}

func TestGetConnectionInfo_DefaultHost(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rd-engine.json")

	config := ConnectionInfo{
		User:     "example_user",
		Password: "example_password",
		Port:     8080,
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(t, err)

	originalConfigPath := configPath
	originalDefaultConfigPath := DefaultConfigPath
	originalConnectionSettings := connectionSettings
	t.Cleanup(func() {
		configPath = originalConfigPath
		DefaultConfigPath = originalDefaultConfigPath
		connectionSettings = originalConnectionSettings
	})

	configPath = configFile
	DefaultConfigPath = configFile
	connectionSettings = ConnectionInfo{}

	result, err := GetConnectionInfo(false)
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", result.Host)
}

func TestGetConnectionInfo_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rd-engine.json")

	config := ConnectionInfo{
		Host: "example_host",
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(t, err)

	originalConfigPath := configPath
	originalDefaultConfigPath := DefaultConfigPath
	originalConnectionSettings := connectionSettings
	t.Cleanup(func() {
		configPath = originalConfigPath
		DefaultConfigPath = originalDefaultConfigPath
		connectionSettings = originalConnectionSettings
	})

	configPath = configFile
	DefaultConfigPath = configFile
	connectionSettings = ConnectionInfo{}

	result, err := GetConnectionInfo(false)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient connection settings")
}

func TestPersistentPreRunE_Verbose(t *testing.T) {
	originalLevel := logrus.GetLevel()
	t.Cleanup(func() {
		logrus.SetLevel(originalLevel)
	})

	originalVerbose := verbose
	t.Cleanup(func() {
		verbose = originalVerbose
	})

	verbose = true
	err := PersistentPreRunE(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, logrus.TraceLevel, logrus.GetLevel())
}

func TestPersistentPreRunE_NotVerbose(t *testing.T) {
	originalLevel := logrus.GetLevel()
	t.Cleanup(func() {
		logrus.SetLevel(originalLevel)
	})

	logrus.SetLevel(logrus.InfoLevel)

	originalVerbose := verbose
	t.Cleanup(func() {
		verbose = originalVerbose
	})

	verbose = false
	err := PersistentPreRunE(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
}

func TestIsWSLDistro_NotWSL(t *testing.T) {
	originalLstatFunc := lstatFunc
	t.Cleanup(func() {
		lstatFunc = originalLstatFunc
	})

	lstatFunc = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	assert.False(t, isWSLDistro())
}

func TestIsWSLDistro_NoWSLEnvs(t *testing.T) {
	originalLstatFunc := lstatFunc
	t.Cleanup(func() {
		lstatFunc = originalLstatFunc
	})

	lstatFunc = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{mode: os.ModeSymlink | 0o755}, nil
	}

	// Clear WSL env vars if they exist
	for _, envName := range wslDistroEnvs {
		if val, ok := os.LookupEnv(envName); ok {
			t.Setenv(envName, "")
			os.Unsetenv(envName)
			t.Cleanup(func() {
				os.Setenv(envName, val)
			})
		}
	}

	assert.False(t, isWSLDistro())
}

func TestIsWSLDistro_WithSymlink(t *testing.T) {
	originalLstatFunc := lstatFunc
	t.Cleanup(func() {
		lstatFunc = originalLstatFunc
	})

	lstatFunc = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{mode: os.ModeSymlink | 0o755}, nil
	}

	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
	assert.True(t, isWSLDistro())
}

func TestIsWSLDistro_WithExecutable(t *testing.T) {
	originalLstatFunc := lstatFunc
	t.Cleanup(func() {
		lstatFunc = originalLstatFunc
	})

	lstatFunc = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{mode: 0o755}, nil
	}

	t.Setenv("WSL_INTEROP", "/run/WSL/123_interop")
	assert.True(t, isWSLDistro())
}

func TestHasWSLEnvs(t *testing.T) {
	// Clear all WSL env vars first
	for _, envName := range wslDistroEnvs {
		if val, ok := os.LookupEnv(envName); ok {
			os.Unsetenv(envName)
			t.Cleanup(func() {
				os.Setenv(envName, val)
			})
		}
	}

	assert.False(t, hasWSLEnvs())

	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
	assert.True(t, hasWSLEnvs())
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	mode os.FileMode
}

func (m *mockFileInfo) Name() string       { return "wslpath" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }
