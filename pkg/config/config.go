package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config game configuration
type Config struct {
	// WordDictPath word dictionary file path
	WordDictPath string `json:"word_dict_path"`
	// WordCount number of words per game (0 means unlimited)
	WordCount int `json:"word_count"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		WordDictPath: "data/words.txt",
		WordCount:    20, // default 20 words
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
