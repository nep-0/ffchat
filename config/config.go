package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	LLM    LLMConfig    `json:"llm"`
	FFmpeg FFmpegConfig `json:"ffmpeg"`
}

type LLMConfig struct {
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature,omitempty"`
}

type FFmpegConfig struct {
	Path string `json:"path,omitempty"`
}

func (c *Config) Validate() error {
	if c.LLM.BaseURL == "" {
		return fmt.Errorf("LLM base URL is required")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("LLM model is required")
	}
	return nil
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	configPath := getConfigPath()

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	config.mergeEnvVars()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".ffchat.json")
}

func (c *Config) mergeEnvVars() {
	if c.LLM.BaseURL == "" {
		if url := os.Getenv("FFCHAT_LLM_BASE_URL"); url != "" {
			c.LLM.BaseURL = url
		}
	}

	if c.LLM.APIKey == "" {
		if key := os.Getenv("FFCHAT_LLM_API_KEY"); key != "" {
			c.LLM.APIKey = key
		}
	}

	if c.LLM.Model == "" {
		if model := os.Getenv("FFCHAT_LLM_MODEL"); model != "" {
			c.LLM.Model = model
		}
	}

	if c.LLM.Temperature == 0 {
		if temp := os.Getenv("FFCHAT_LLM_TEMPERATURE"); temp != "" {
			var t float64
			fmt.Sscanf(temp, "%f", &t)
			c.LLM.Temperature = t
		}
	}

	if c.FFmpeg.Path == "" {
		if path := os.Getenv("FFCHAT_FFMPEG_PATH"); path != "" {
			c.FFmpeg.Path = path
		}
	}
}

func SaveConfig(config *Config) error {
	configPath := getConfigPath()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
