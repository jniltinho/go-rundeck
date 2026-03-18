package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config is the root configuration struct.
type Config struct {
	Server    ServerConfig    `toml:"server"`
	Database  DatabaseConfig  `toml:"database"`
	SSH       SSHConfig       `toml:"ssh"`
	Scheduler SchedulerConfig `toml:"scheduler"`
	Log       LogConfig       `toml:"log"`
}

// ServerConfig holds application-level and TLS settings.
type ServerConfig struct {
	Name           string `toml:"name"`
	Env            string `toml:"env"`
	Port           int    `toml:"port"`
	SessionSecret  string `toml:"session_secret"`
	SessionTimeout int    `toml:"session_timeout"` // inactivity timeout in minutes
	Debug          bool   `toml:"debug"`
	SSLEnabled     bool   `toml:"ssl_enable"`
	SSLCert        string `toml:"ssl_cert"`
	SSLKey         string `toml:"ssl_key"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
	Charset  string `toml:"charset"`
}

// SSHConfig holds SSH-related settings.
type SSHConfig struct {
	DefaultUser     string `toml:"default_user"`
	DefaultPort     int    `toml:"default_port"`
	ConnectTimeout  int    `toml:"connect_timeout"`
	KeyStoragePath  string `toml:"key_storage_path"`
}

// SchedulerConfig holds scheduler settings.
type SchedulerConfig struct {
	Enabled       bool `toml:"enabled"`
	CheckInterval int  `toml:"check_interval"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
	Output string `toml:"output"`
}

// Load reads and parses the TOML config file at the given path.
func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// DSN returns the MySQL DSN string for GORM.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name, d.Charset,
	)
}

// Addr returns the listen address for the Echo server.
func (s *ServerConfig) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

func validate(cfg *Config) error {
	if cfg.Server.SessionSecret == "" {
		return fmt.Errorf("server.session_secret must not be empty")
	}
	if len(cfg.Server.SessionSecret) < 32 {
		return fmt.Errorf("server.session_secret must be at least 32 characters long")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host must not be empty")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database.name must not be empty")
	}
	return nil
}
