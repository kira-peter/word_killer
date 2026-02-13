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

	// Sentence mode dictionary path
	SentenceDictPath string `json:"sentence_dict_path"`

	// Time-challenge mode settings
	CountdownDuration int `json:"countdown_duration"` // 倒计时模式时长（秒），默认60

	// Speed Run mode settings
	SpeedRunWordCount int `json:"speedrun_word_count"` // 极速模式单词数量

	// Rhythm Master mode settings
	RhythmInitialTimeLimit float64 `json:"rhythm_initial_time_limit"` // 节奏大师初始时间限制（秒）
	RhythmMinTimeLimit     float64 `json:"rhythm_min_time_limit"`     // 节奏大师最小时间限制（秒）
	RhythmDifficultyStep   float64 `json:"rhythm_difficulty_step"`    // 节奏大师难度递增步长（秒）
	RhythmWordsPerLevel    int     `json:"rhythm_words_per_level"`    // 节奏大师每级所需单词数

	// Rhythm Dance mode settings
	RhythmDanceDuration      int     `json:"rhythm_dance_duration"`        // 节奏舞蹈模式时长（秒）
	RhythmDanceInitialSpeed  float64 `json:"rhythm_dance_initial_speed"`   // 节奏舞蹈指针初始速度
	RhythmDanceSpeedIncrement float64 `json:"rhythm_dance_speed_increment"` // 每完成一个单词的速度增量
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		WordCount:        20, // default 20 words
		ShortDictPath:    "data/google-10000-short.txt",
		MediumDictPath:   "data/google-10000-medium.txt",
		LongDictPath:     "data/google-10000-long.txt",
		ShortRatio:       30,
		MediumRatio:      50,
		LongRatio:        20,
		SentenceDictPath: "data/sentences.txt",
		// Time-challenge mode defaults
		CountdownDuration: 60, // 默认60秒
		// Speed Run mode defaults
		SpeedRunWordCount: 25, // 25个单词
		// Rhythm Master mode defaults
		RhythmInitialTimeLimit: 2.0,  // 初始2秒/词
		RhythmMinTimeLimit:     0.5,  // 最小0.5秒/词
		RhythmDifficultyStep:   0.1,  // 每级减少0.1秒
		RhythmWordsPerLevel:    10,   // 每10个词升级
		// Rhythm Dance mode defaults
		RhythmDanceDuration:       60,    // 默认60秒
		RhythmDanceInitialSpeed:   0.05,  // 初始速度0.05
		RhythmDanceSpeedIncrement: 0.005, // 每完成一个单词增加0.005
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
