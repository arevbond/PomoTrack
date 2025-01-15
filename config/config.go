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
	FocusDuration   time.Duration `yaml:"focus_duration"`
	BreakDuration   time.Duration `yaml:"break_duration"`
	HiddenFocusTime bool          `yaml:"hidden_focus_time"`
}

type flags struct {
	FocusDuration   time.Duration `yaml:"focus_duration"`
	BreakDuration   time.Duration `yaml:"break_duration"`
	HiddenFocusTime bool          `yaml:"hidden_focus_time"`
}

const (
	configName           = ".pomotrack-config.yaml"
	defaultFocusDuration = 25 * time.Minute
	defaultBreakDuration = 5 * time.Minute
)

func parseFlags() flags {
	var f flags

	flag.DurationVar(&f.FocusDuration, "focus-duration", 0, "edit focus timer duration")
	flag.DurationVar(&f.BreakDuration, "break-duration", 0, "edit break timer duration")
	flag.BoolVar(&f.HiddenFocusTime, "hidden-focus-time", false, "show or hide clock on focus page")
	flag.Parse()

	return f
}

// Init create application config.
func Init() (*Config, error) {
	f := parseFlags()

	configPath := getConfigPath()
	config, err := readConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("can't read config: %w", err)
	}

	changed := applyFlagsToConfig(f, config)
	if changed {
		err = writeConfig(config, configPath)
		if err != nil {
			return nil, fmt.Errorf("can't save config to file: %w", err)
		}
	}

	if config.Timer.FocusDuration == 0 {
		config.Timer.FocusDuration = defaultFocusDuration
	}
	if config.Timer.BreakDuration == 0 {
		config.Timer.BreakDuration = defaultBreakDuration
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

func applyFlagsToConfig(f flags, c *Config) bool {
	var changed bool

	if f.FocusDuration > 0 && f.FocusDuration != c.Timer.FocusDuration {
		c.Timer.FocusDuration = f.FocusDuration
		changed = true
	}

	if f.BreakDuration > 0 && f.BreakDuration != c.Timer.BreakDuration {
		c.Timer.BreakDuration = f.BreakDuration
		changed = true
	}

	if isFlagPassed("hidden-focus-time") && f.HiddenFocusTime != c.Timer.HiddenFocusTime {
		c.Timer.HiddenFocusTime = f.HiddenFocusTime
		changed = true
	}
	return changed
}

func isFlagPassed(name string) bool {
	passed := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			passed = true
		}
	})
	return passed
}
