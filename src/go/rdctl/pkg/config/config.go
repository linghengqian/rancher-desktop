/*
Copyright © 2022 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package config handles all the config-related parts of rdctl
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/paths"
)

// ConnectionInfo stores the parameters needed to connect to an HTTP server
type ConnectionInfo struct {
	User     string
	Password string
	Host     string
	Port     int
}

var (
	connectionSettings ConnectionInfo
	verbose            bool

	configPath string
	// DefaultConfigPath - used to differentiate not being able to find a user-specified config file from the default
	DefaultConfigPath string
)

// DefineGlobalFlags sets up the global flags, available for all sub-commands
func DefineGlobalFlags(rootCmd *cobra.Command) {
	appPaths, err := paths.GetPaths()
	if err != nil {
		log.Fatalf("failed to get paths: %s", err)
	}
	DefaultConfigPath = filepath.Join(appPaths.AppHome, "rd-engine.json")
	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", "", fmt.Sprintf("config file (default %s)", DefaultConfigPath))
	rootCmd.PersistentFlags().StringVar(&connectionSettings.User, "user", "", "overrides the user setting in the config file")
	rootCmd.PersistentFlags().StringVar(&connectionSettings.Host, "host", "", "default is 127.0.0.1; most useful for WSL")
	rootCmd.PersistentFlags().IntVar(&connectionSettings.Port, "port", 0, "overrides the port setting in the config file")
	rootCmd.PersistentFlags().StringVar(&connectionSettings.Password, "password", "", "overrides the password setting in the config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Be verbose")
}

// GetConnectionInfo returns the connection details of the application API server.
// As a special case this function may return a nil *ConnectionInfo and nil error
// when the config file has not been specified explicitly, the default config file
// does not exist, and the mayBeMissing parameter is true.
func GetConnectionInfo(mayBeMissing bool) (*ConnectionInfo, error) {
	var settings ConnectionInfo

	if configPath == "" {
		configPath = DefaultConfigPath
	}
	content, readFileError := os.ReadFile(configPath)
	if readFileError != nil {
		// It is ok if the default config path doesn't exist; the user may have specified the required settings on the commandline.
		// But it is an error if the file specified via --config-path cannot be read.
		if configPath != DefaultConfigPath || !errors.Is(readFileError, os.ErrNotExist) {
			return nil, readFileError
		}
	} else if err := json.Unmarshal(content, &settings); err != nil {
		return nil, fmt.Errorf("error parsing config file %q: %w", configPath, err)
	}

	// CLI options override file settings
	if connectionSettings.Host != "" {
		settings.Host = connectionSettings.Host
	}
	if settings.Host == "" {
		settings.Host = "127.0.0.1"
	}
	if connectionSettings.User != "" {
		settings.User = connectionSettings.User
	}
	if connectionSettings.Password != "" {
		settings.Password = connectionSettings.Password
	}
	if connectionSettings.Port != 0 {
		settings.Port = connectionSettings.Port
	}
	if settings.Port == 0 || settings.User == "" || settings.Password == "" {
		// Missing the default config file may or may not be considered an error
		if readFileError != nil {
			if mayBeMissing {
				readFileError = nil
			}
			return nil, readFileError
		}
		return nil, errors.New("insufficient connection settings (missing one or more of: port, user, and password)")
	}

	return &settings, nil
}

// PersistentPreRunE is meant to be executed as the cobra hook
func PersistentPreRunE(cmd *cobra.Command, args []string) error {
	if verbose {
		logrus.SetLevel(logrus.TraceLevel)
	}
	return nil
}
