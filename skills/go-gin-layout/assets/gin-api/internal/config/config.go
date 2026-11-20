package config

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultPath = "configs/config.yaml"

// Config 与 configs/config.yaml 的字段一一对应，是启动时加载的只读数据。
// 整体放入 svc.ServiceContext.Config 供 Service 读取；模块配置增多或需要安全隔离时，
// 再按业务拆分子结构（如 OrderConfig），不要新增汇总的 BusinessConfig。
type Config struct {
	HTTP     HTTPConfig  `yaml:"http"`
	Infra    InfraConfig `yaml:"infra"`
	LogLevel LogLevel    `yaml:"log_level"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

// InfraConfig 汇集外部基础设施配置；MySQL 之外的能力按需增加子段。
type InfraConfig struct {
	MySQL MySQLConfig `yaml:"mysql"`
}

// MySQLConfig 保存连接参数和连接池配置；DSN 由 infra/database 统一构造。
type MySQLConfig struct {
	Host     string            `yaml:"host"`
	Port     int               `yaml:"port"`
	User     string            `yaml:"user"`
	Password string            `yaml:"password"`
	Database string            `yaml:"database"`
	Params   map[string]string `yaml:"params"`
	Pool     MySQLPool         `yaml:"pool"`
}

type MySQLPool struct {
	MaxOpenConns    int      `yaml:"max_open_conns"`
	MaxIdleConns    int      `yaml:"max_idle_conns"`
	ConnMaxLifetime Duration `yaml:"conn_max_lifetime"`
	PingTimeout     Duration `yaml:"ping_timeout"`
}

// LogLevel 支持 YAML 中的 debug、info、warn、error。
type LogLevel slog.Level

func (l *LogLevel) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		*l = LogLevel(slog.LevelInfo)
		return nil
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return err
	}
	*l = LogLevel(level)
	return nil
}

// Duration 支持 YAML 中的 Go duration 写法，例如 3m、3s。
type Duration time.Duration

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	if strings.TrimSpace(raw) == "" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

func Load() (Config, error) {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = defaultPath
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %s: %w", path, err)
	}
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = "127.0.0.1:8080"
	}
	if err := validateHTTPAddr(cfg.HTTP.Addr); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", path, err)
	}
	return cfg, nil
}

func validateHTTPAddr(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("http.addr: %w", err)
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("http.addr port must be between 1 and 65535")
	}
	return nil
}
