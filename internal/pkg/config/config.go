package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Log      LogConfig      `mapstructure:"log"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type AppConfig struct {
	Env        string `mapstructure:"env"`
	Port       string `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	Debug      bool   `mapstructure:"debug"`
	ApiVersion string `mapstructure:"api_version"`
	ApiPrefix  string `mapstructure:"api_prefix"`
}

type LogConfig struct {
	LogLevel string `mapstructure:"log_level"`
}

type PostgresConfig struct {
	Master PostgresNodeConfig `mapstructure:"master"`
	Slaves PostgresNodeConfig `mapstructure:"slaves"`
}

type PostgresNodeConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	AccessTokenSecret string `mapstructure:"access_token_secret"`
	AccessTokenTTL    int    `mapstructure:"access_token_ttl"`
	RefreshTokenTTL   int    `mapstructure:"refresh_token_ttl"`
}

func LoadConfig() (*Config, error) {
	return loadConfig([]string{"."})
}

func loadConfig(paths []string) (*Config, error) {
	var config Config

	_ = godotenv.Load()
	for _, path := range paths {
		f := filepath.Join(path, ".env")
		if _, err := os.Stat(f); err == nil {
			_ = godotenv.Load(f)
		}
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	for _, path := range paths {
		v.AddConfigPath(path)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}

func (c PostgresNodeConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   c.Database,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return u.String()
}
