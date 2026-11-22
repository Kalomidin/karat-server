package config

import (
	"karat-server/pkg/provider/google"
	"karat-server/storage"
)

type Config struct {
	Google google.Config  `yaml:"google"`
	DB     Postgres       `yaml:"db"`
	R2     storage.Config `yaml:"r2"`
}
