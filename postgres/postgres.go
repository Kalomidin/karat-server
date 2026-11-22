package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Config interface {
	GetDatabase() string
	GetSchema() string
	GetUserName() string
	GetPassword() string
	GetHost() string
	GetPort() string
	GetMaxWaitForConnection() time.Duration
	GetMaxConns() int
	GetSSLMode() string
}

func InitDB(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.GetHost(), cfg.GetPort(), cfg.GetUserName(), cfg.GetPassword(), cfg.GetDatabase(), cfg.GetSSLMode(),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.GetSchema() + ".",
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.GetMaxConns())
	sqlDB.SetConnMaxIdleTime(cfg.GetMaxWaitForConnection())
	return db, nil
}
