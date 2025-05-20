package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/noorbala7418/grafana-oncall-notifier/internal/models"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*models.Config, error) {
	// Create config structure
	config := &models.Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("function NewConfig: Error in open config file. err: %w", err)
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, fmt.Errorf("function NewConfig: Error in decode config file. err: %w", err)
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func validateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		logrus.Error("function ValidateConfigPath: error in config address validation. err: ", err)
		return fmt.Errorf("function ValidateConfigPath: error in config address validation. err: %w", err)
	}
	if s.IsDir() {
		logrus.Errorf("function ValidateConfigPath: '%s' is a directory, not a normal file", path)
		return fmt.Errorf("function ValidateConfigPath: '%s' is a directory, not a normal file", path)
	}
	return nil
}

// ParseFlags will create and parse the CLI flags and return the path to be used elsewhere
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := validateConfigPath(configPath); err != nil {
		logrus.Error("function ParseFlags: Error in validate config path. err: ", err)
		return "", fmt.Errorf("function ParseFlags: Error in validate config path. err: %w", err)
	}

	// Return the configuration path
	return configPath, nil
}
