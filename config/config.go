package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
	configName           = ".pomotrack-config.yaml"
	defaultFocusDuration = 25 * time.Minute
	defaultBreakDuration = 5 * time.Minute
)

// Init create application config.
func Init() (*Config, error) {
	var focusDurationFlag time.Duration
	var breakDurationFlag time.Duration

	flag.DurationVar(&focusDurationFlag, "focus-duration", 0, "edit focus timer duration")
	flag.DurationVar(&breakDurationFlag, "break-duration", 0, "edit break timer duration")

	flag.Parse()

	configPath := getConfigPath()
	config, err := readConfig(configPath)
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

		if err = writeConfig(config, configPath); err != nil {
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

		if err = writeConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("can't write config: %w", err)
		}
	}

	return config, nil
}

func readConfig(configPath string) (*Config, error) {
	var config Config

	file, err := os.OpenFile(configPath, os.O_RDONLY|os.O_CREATE, 0o600)
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

func GetConfigDir() string {
	// create config in user config directory
	// example for unix: $HOME/.config/pomotrack/
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	pomotrackCfgDir := filepath.Join(userCfgDir, "pomotrack")
	dirErr := os.MkdirAll(pomotrackCfgDir, 0o750)
	if dirErr != nil {
		log.Println("[WARN] can't create config directory:", dirErr)
		return ""
	}
	return pomotrackCfgDir
}

func getConfigPath() string {
	// first try find locally config
	if _, err := os.Stat(configName); !errors.Is(err, os.ErrNotExist) {
		return configName
	}
	return filepath.Join(GetConfigDir(), configName)
}

func writeConfig(config *Config, configPath string) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("can't marshal yaml file: %w", err)
	}

	if err = os.WriteFile(configPath, yamlData, 0o600); err != nil {
		return fmt.Errorf("can't update config file: %w", err)
	}
	return nil
}
