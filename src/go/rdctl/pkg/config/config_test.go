package config

import (
	"os"
	"testing"
)

func TestIsWSLDistro(t *testing.T) {
	t.Run("returns false without WSL envs", func(t *testing.T) {
		for _, envName := range wslDistroEnvs {
			t.Setenv(envName, "")
		}
		if isWSLDistro() {
			t.Fatalf("expected isWSLDistro to be false without WSL envs")
		}
	})

	t.Run("returns true with WSL envs", func(t *testing.T) {
		requiresWsl := false
		for _, envName := range wslDistroEnvs {
			if _, ok := os.LookupEnv(envName); ok {
				requiresWsl = true
			}
		}
		if !requiresWsl {
			t.Skip("WSL environment variables not set in this environment")
		}
		if !isWSLDistro() {
			t.Fatalf("expected isWSLDistro to be true with WSL envs")
		}
	})
}
