package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".mysqlctl_config.json"

// getConfigPath 获取配置文件路径
func getConfigPath() (string, error) {
	// 首先尝试在当前目录创建
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, configFileName), nil
}

// SaveConfigToFile 存储连接配置到文件
func SaveConfigToFile(cfg *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

// LoadConfigFromFile 从文件加载连接配置
func LoadConfigFromFile() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// ClearConfigFile 清除配置文件
func ClearConfigFile() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// AutoConnect 尝试从配置文件自动连接
func AutoConnect() error {
	cfg, err := LoadConfigFromFile()
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("no saved connection found")
	}

	return Connect(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
}
