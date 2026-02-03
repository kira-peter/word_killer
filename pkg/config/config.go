package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 游戏配置
type Config struct {
	// WordDictPath 单词词库文件路径
	WordDictPath string `json:"word_dict_path"`
	// WordCount 每场游戏的单词数量（0 表示不限制）
	WordCount int `json:"word_count"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		WordDictPath: "data/words.txt",
		WordCount:    20, // 默认 20 个单词
	}
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	// 如果文件不存在，使用默认配置
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// Save 保存配置到文件
func Save(cfg *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}
