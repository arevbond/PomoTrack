package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Timer TimerConfig `yaml:"timer"`
}

type TimerConfig struct {
	FocusDuration time.Duration `yaml:"focus_duration"`
	BreakDuration time.Duration `yaml:"break_duration"`
}

const (
	configName           = "config/config.yaml"
	defaultFocusDuration = 25 * time.Minute
	defaultBreakDuration = 5 * time.Minute
)

func Init() (*Config, error) {
	var focusDurationFlag time.Duration
	var breakDurationFlag time.Duration

	flag.DurationVar(&focusDurationFlag, "focus-duration", 0, "edit focus timer duration")
	flag.DurationVar(&breakDurationFlag, "break-duration", 0, "edit break timer duration")

	flag.Parse()

	config, err := readConfig()
	if err != nil {
		return nil, fmt.Errorf("can't read config: %w", err)
	}

	if focusDurationFlag > 0 || breakDurationFlag > 0 {
		if focusDurationFlag > 0 {
			config.Timer.FocusDuration = focusDurationFlag
		}
		if breakDurationFlag > 0 {
			config.Timer.BreakDuration = breakDurationFlag
		}

		if err = writeConfig(config); err != nil {
			return nil, fmt.Errorf("can't write config: %w", err)
		}
	}

	if config.Timer.FocusDuration == 0 || config.Timer.BreakDuration == 0 {
		if config.Timer.FocusDuration == 0 {
			config.Timer.FocusDuration = defaultFocusDuration
		}

		if config.Timer.BreakDuration == 0 {
			config.Timer.BreakDuration = defaultBreakDuration
		}

		if err = writeConfig(config); err != nil {
			return nil, fmt.Errorf("can't write config: %w", err)
		}
	}

	return config, nil
}

func readConfig() (*Config, error) {
	var config Config

	file, err := os.OpenFile(configName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("can't open file: %w", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("can't read config file: %w", err)
	}

	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshall file into struct: %w", err)
	}
	return &config, nil
}

func writeConfig(config *Config) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("can't marshal yaml file: %w", err)
	}

	if err = os.WriteFile(configName, yamlData, 0600); err != nil {
		return fmt.Errorf("can't update config file: %w", err)
	}
	return nil
}
