package config

import (
	"karat-server/pkg/provider/google"
	"karat-server/storage"
)

type Config struct {
	Google google.Config  `yaml:"google"`
	Shop   Shop           `yaml:"shop"`
	DB     Postgres       `yaml:"db"`
	R2     storage.Config `yaml:"r2"`
}

type Shop struct {
	Timezone string `yaml:"timezone" env:"SHOP_TIMEZONE" env-default:"UTC"`
}
