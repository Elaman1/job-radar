package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

const defaultConfigPath = "./config/config.yaml"

type Config struct {
	App      App      `yaml:"app"`
	Server   Server   `yaml:"server"`
	Logger   Logger   `yaml:"logger"`
	Postgres Postgres `yaml:"postgres"`
	Sources  Sources  `yaml:"sources"`
}

type App struct {
	ServiceName string `yaml:"service_name"`
}

type Server struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type Logger struct {
	Level string `yaml:"level"`
	Env   string `yaml:"env"`
}

type Postgres struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	MaxConns        int32  `yaml:"max_conns"`
	MinConns        int32  `yaml:"min_conns"`
	MaxConnLifeTime int    `yaml:"max_conn_life_time"`
	MaxConnIdleTime int    `yaml:"max_conn_idle_time"`
	SSLMode         string `yaml:"sslmode"`
}

type Sources struct {
	Adzuna Adzuna `yaml:"adzuna"`
}

type Adzuna struct {
	AppId  string `yaml:"app_id"`
	AppKey string `yaml:"app_key"`
}

func InitConf() (*Config, error) {
	data, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read file error %w", err)
	}

	conf := Config{}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	if err = validateConfig(&conf); err != nil {
		return nil, fmt.Errorf("validate config error: %w", err)
	}

	return &conf, nil
}

func validateConfig(conf *Config) error {
	if conf.App.ServiceName == "" {
		return fmt.Errorf("not filled app.service_name")
	}

	if conf.Server.Address == "" {
		return fmt.Errorf("not filled server.address")
	}

	if conf.Logger.Level == "" {
		return fmt.Errorf("not filled logger.level")
	}

	if conf.Logger.Env == "" {
		return fmt.Errorf("not filled logger.env")
	}

	if conf.Postgres.Host == "" {
		return fmt.Errorf("not filled postgres.host")
	}

	if conf.Postgres.Port == 0 {
		return fmt.Errorf("not filled postgres.port")
	}

	if conf.Postgres.User == "" {
		return fmt.Errorf("not filled postgres.user")
	}

	if conf.Postgres.Database == "" {
		return fmt.Errorf("not filled postgres.database")
	}

	return nil
}
