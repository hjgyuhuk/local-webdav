package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const defaultTemplate = `# local-webdav 配置文件

[server]
# 监听地址（仅本机访问）
address = "127.0.0.1:43621"

# 每个 [[shares]] 定义一个 WebDAV 共享
# 客户端通过 http://<address>/<name>/ 访问，不会暴露真实路径
#
# [[shares]]
# name = "documents"
# path = "~/Documents"
# username = "admin"
# password = "changeme"
# 不设置 username/password 则无需认证
`

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".config", "local-webdav", "config.toml"), nil
}

func EnsureConfig(path string) (created bool, err error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(defaultTemplate), 0644); err != nil {
		return false, fmt.Errorf("write default config: %w", err)
	}

	return true, nil
}

type Config struct {
	Server Server  `toml:"server"`
	Shares []Share `toml:"shares"`
}

type Server struct {
	Address string `toml:"address"`
}

type Share struct {
	Name     string `toml:"name"`
	Path     string `toml:"path"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Address == "" {
		c.Server.Address = "127.0.0.1:43621"
	}

	if len(c.Shares) == 0 {
		return fmt.Errorf("no shares defined; please edit the config file to add at least one [[shares]] section")
	}

	names := make(map[string]bool)
	for i, s := range c.Shares {
		if s.Name == "" {
			return fmt.Errorf("share %d: name is required", i)
		}
		if names[s.Name] {
			return fmt.Errorf("share %d: duplicate name %q", i, s.Name)
		}
		names[s.Name] = true

		if s.Path == "" {
			return fmt.Errorf("share %q: path is required", s.Name)
		}

		expanded, err := expandHome(s.Path)
		if err != nil {
			return fmt.Errorf("share %q: %w", s.Name, err)
		}
		absPath, err := filepath.Abs(expanded)
		if err != nil {
			return fmt.Errorf("share %q: invalid path %q: %w", s.Name, s.Path, err)
		}
		c.Shares[i].Path = absPath

		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("share %q: path %q: %w", s.Name, absPath, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("share %q: path %q is not a directory", s.Name, absPath)
		}

	}

	return nil
}

func expandHome(p string) (string, error) {
	if len(p) < 2 || p[0] != '~' || (p[1] != '/' && p[1] != '\\') {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand ~: %w", err)
	}
	return filepath.Join(home, p[1:]), nil
}
