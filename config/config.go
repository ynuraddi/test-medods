package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		HTTP `yaml:"http"`
		Log  `yaml:"logger"`
		PG
		JWT
		SMTP `yaml:"smtp"`
	}

	JWT struct {
		SecretKey string `env-required:"true" env:"SECRET_KEY"`
	}

	HTTP struct {
		Port string `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	}

	Log struct {
		Level string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	}

	PG struct {
		DSN          string `env-required:"true" env:"PG_DSN"`
		MigrationURL string `env-required:"true" env:"PG_MIGRATION_URL"`
	}

	SMTP struct {
		HOST string `env:"SMTP_HOST" yaml:"host"`
		PORT int    `env-required:"true" env:"SMTP_PORT" yaml:"port"`
		USER string `env:"SMTP_USER" yaml:"user"`
		PASS string `env:"SMTP_PASS"`
		FROM string `env-required:"true" env:"SMTP_FROM" yaml:"from_email"`
	}
)

func MustLoad() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yaml", cfg)
	if err != nil {
		return nil, fmt.Errorf("read config file error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("read env error: %w", err)
	}

	return cfg, nil
}
