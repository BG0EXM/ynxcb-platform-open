package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用配置
type Config struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
	Database struct {
		Path string `json:"path"`
	} `json:"database"`
	JWT struct {
		Secret string `json:"secret"`
	} `json:"jwt"`
	Upload struct {
		Dir   string `json:"dir"`
		MaxMB int64  `json:"max_mb"`
	} `json:"upload"`
	Admin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"admin"`
}

// Default 返回默认配置
func Default() *Config {
	c := &Config{}
	c.Server.Port = "8080"
	c.Database.Path = "data/ynxcb.db"
	c.JWT.Secret = "change-me-to-a-random-secret"
	c.Upload.Dir = "data/uploads"
	c.Upload.MaxMB = 50
	c.Admin.Username = "admin"
	c.Admin.Password = "admin123"
	return c
}

// Load 从 JSON 文件加载配置，文件不存在时使用默认值
func Load(path string) (*Config, error) {
	c := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
