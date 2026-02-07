package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConnectionInfo_ValidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rd-engine.json")

	config := ConnectionInfo{
		User:     "testuser",
		Password: "testpass",
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
	assert.Equal(t, "testuser", result.User)
	assert.Equal(t, "testpass", result.Password)
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
		User:     "fileuser",
		Password: "filepass",
		Host:     "filehost",
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
		User:     "cliuser",
		Password: "clipass",
		Host:     "clihost",
		Port:     1234,
	}

	result, err := GetConnectionInfo(false)
	require.NoError(t, err)
	assert.Equal(t, "cliuser", result.User)
	assert.Equal(t, "clipass", result.Password)
	assert.Equal(t, "clihost", result.Host)
	assert.Equal(t, 1234, result.Port)
}

func TestGetConnectionInfo_DefaultHost(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rd-engine.json")

	config := ConnectionInfo{
		User:     "testuser",
		Password: "testpass",
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
		Host: "testhost",
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
