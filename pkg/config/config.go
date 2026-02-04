package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config game configuration
type Config struct {
	// WordCount number of words per game (0 means unlimited)
	WordCount int `json:"word_count"`

	// Difficulty word dictionary paths
	ShortDictPath  string `json:"short_dict_path"`
	MediumDictPath string `json:"medium_dict_path"`
	LongDictPath   string `json:"long_dict_path"`

	// Difficulty ratios (will be normalized to percentages)
	ShortRatio  float64 `json:"short_ratio"`
	MediumRatio float64 `json:"medium_ratio"`
	LongRatio   float64 `json:"long_ratio"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		WordCount:      20, // default 20 words
		ShortDictPath:  "data/google-10000-short.txt",
		MediumDictPath: "data/google-10000-medium.txt",
		LongDictPath:   "data/google-10000-long.txt",
		ShortRatio:     30,
		MediumRatio:    50,
		LongRatio:      20,
	}
}

// Load loads configuration file
func Load(path string) (*Config, error) {
	// Use default config if file doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// NormalizeRatios normalizes difficulty ratios to percentages (0-1 range)
// Returns normalized short, medium, long ratios
func (c *Config) NormalizeRatios() (float64, float64, float64, error) {
	// Validate all ratios are non-negative
	if c.ShortRatio < 0 || c.MediumRatio < 0 || c.LongRatio < 0 {
		return 0, 0, 0, fmt.Errorf("all ratios must be >= 0")
	}

	// Calculate total
	total := c.ShortRatio + c.MediumRatio + c.LongRatio

	// At least one ratio must be positive
	if total <= 0 {
		return 0, 0, 0, fmt.Errorf("at least one ratio must be > 0")
	}

	// Normalize to percentages (0-1 range)
	return c.ShortRatio / total, c.MediumRatio / total, c.LongRatio / total, nil
}
