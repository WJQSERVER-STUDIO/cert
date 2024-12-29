package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Log     LogConfig
	Account AccountConfig
	Path    PathConfig
	Domain  DomainConfig
}

type LogConfig struct {
	LogFilePath string `toml:"logfilepath"`
	MaxLogSize  int    `toml:"maxlogsize"`
}

type AccountConfig struct {
	Email string
	Token string
}
type PathConfig struct {
	Cert   string
	Key    string
	CaCert string
	Json   string
}

type DomainConfig struct {
	Name string
}

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
