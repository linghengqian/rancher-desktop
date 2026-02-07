package config

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeFileInfo struct {
	mode os.FileMode
}

func (info fakeFileInfo) Name() string       { return "wslpath" }
func (info fakeFileInfo) Size() int64        { return 0 }
func (info fakeFileInfo) Mode() os.FileMode  { return info.mode }
func (info fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (info fakeFileInfo) IsDir() bool        { return info.mode.IsDir() }
func (info fakeFileInfo) Sys() any           { return nil }

func saveWSLEnvs(t *testing.T) {
	originalEnvs := map[string]string{}
	for _, envName := range wslDistroEnvs {
		if value, ok := os.LookupEnv(envName); ok {
			originalEnvs[envName] = value
		}
	}
	t.Cleanup(func() {
		for _, envName := range wslDistroEnvs {
			if value, present := originalEnvs[envName]; present {
				os.Setenv(envName, value)
			} else {
				os.Unsetenv(envName)
			}
		}
	})
}

func TestIsWSLDistro(t *testing.T) {
	testCases := []struct {
		name     string
		mode     os.FileMode
		hasEnvs  bool
		expected bool
	}{
		{"with wslpath symlink and WSL envs", os.ModeSymlink, true, true},
		{"with wslpath symlink without WSL envs", os.ModeSymlink, false, false},
		{"with wslpath executable and WSL envs", 0755, true, true},
		{"with wslpath executable without WSL envs", 0755, false, false},
		{"with non-executable wslpath and WSL envs", 0644, true, false},
		{"with non-executable wslpath without WSL envs", 0644, false, false},
		{"without wslpath and WSL envs", 0, true, false},
		{"without wslpath and without WSL envs", 0, false, false},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("returns %t %s", tc.expected, tc.name), func(t *testing.T) {
			saveWSLEnvs(t)
			for _, envName := range wslDistroEnvs {
				os.Unsetenv(envName)
			}
			originalLstat := lstatFunc
			t.Cleanup(func() { lstatFunc = originalLstat })
			lstatFunc = func(_ string) (os.FileInfo, error) {
				return fakeFileInfo{mode: tc.mode}, nil
			}
			if tc.hasEnvs {
				os.Setenv(wslDistroEnvs[0], "Ubuntu")
			}
			if tc.expected {
				assert.True(t, isWSLDistro(), "expected isWSLDistro to be true")
			} else {
				assert.False(t, isWSLDistro(), "expected isWSLDistro to be false")
			}
		})
	}
}

func TestHasWSLEnvs(t *testing.T) {
	t.Run("returns false when none set", func(t *testing.T) {
		saveWSLEnvs(t)
		for _, envName := range wslDistroEnvs {
			os.Unsetenv(envName)
		}
		assert.False(t, hasWSLEnvs(), "expected hasWSLEnvs to be false without WSL envs")
	})

	t.Run("returns true when any set", func(t *testing.T) {
		saveWSLEnvs(t)
		for _, envName := range wslDistroEnvs {
			os.Unsetenv(envName)
		}
		os.Setenv(wslDistroEnvs[0], "Ubuntu")
		assert.True(t, hasWSLEnvs(), "expected hasWSLEnvs to be true with WSL envs")
	})
}

func TestIsWSLDistroLstatError(t *testing.T) {
	saveWSLEnvs(t)
	originalLstat := lstatFunc
	t.Cleanup(func() { lstatFunc = originalLstat })
	lstatFunc = func(_ string) (os.FileInfo, error) {
		return nil, errors.New("lstat failed")
	}
	os.Setenv(wslDistroEnvs[0], "Ubuntu")
	assert.False(t, isWSLDistro(), "expected isWSLDistro to be false when lstat fails")
}
