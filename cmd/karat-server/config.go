package main

import (
	"karat-server/pkg/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig  `yaml:"server"`
	Config config.Config `yaml:"config"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"PORT" env-default:"8080"`
}

func (cfg *Config) Load() error {
	_ = godotenv.Load()
	if err := cleanenv.ReadConfig("./configs/dev.yaml", cfg); err != nil {
		return err
	}
	return cleanenv.ReadEnv(cfg)
}
