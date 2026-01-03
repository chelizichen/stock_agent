package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	AI AIConfig `yaml:"ai"`
}

// AIConfig AI相关配置
type AIConfig struct {
	Provider  string `yaml:"provider"`
	APIKey    string `yaml:"api_key"`
	BaseURL   string `yaml:"base_url"`
	ModelName string `yaml:"model_name"`
}

// LoadConfig 从配置文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 如果未指定配置文件路径，使用默认路径
	if configPath == "" {
		configPath = "config.yaml"
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证必需字段
	if config.AI.Provider == "" {
		return nil, fmt.Errorf("配置文件中缺少 ai.provider")
	}
	if config.AI.APIKey == "" {
		return nil, fmt.Errorf("配置文件中缺少 ai.api_key")
	}
	if config.AI.BaseURL == "" {
		return nil, fmt.Errorf("配置文件中缺少 ai.base_url")
	}
	if config.AI.ModelName == "" {
		return nil, fmt.Errorf("配置文件中缺少 ai.model_name")
	}

	return &config, nil
}
