package handler

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"karat-server/common"
	auth_handler "karat-server/pkg/handler/auth"
	currency_handler "karat-server/pkg/handler/currency"
	item_handler "karat-server/pkg/handler/item"
	user_handler "karat-server/pkg/handler/user"
	auth_manager "karat-server/pkg/manager/auth"
	currency_repo "karat-server/pkg/manager/currency/repository"
	item_manager "karat-server/pkg/manager/item"
	item_repo "karat-server/pkg/manager/item/repository"
	user_manager "karat-server/pkg/manager/user"
	user_repo "karat-server/pkg/manager/user/repository"
	"karat-server/pkg/provider"
	"karat-server/storage"
)

var schemaName = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

func InitHandlers(router *gin.Engine, db *gorm.DB, middleware common.Middleware, schema string, tokens provider.Client, locationName string, images storage.Store) error {
	if !schemaName.MatchString(schema) {
		return fmt.Errorf("schema name %q is not allowed", schema)
	}
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS " + schema).Error; err != nil {
		return err
	}

	users := user_repo.NewRepository(db)
	currencies := currency_repo.NewRepository(db)
	items := item_repo.NewRepository(db)
	if err := users.Migrate(); err != nil {
		return err
	}
	if err := currencies.Migrate(); err != nil {
		return err
	}
	if err := items.Migrate(); err != nil {
		return err
	}

	userPort := user_manager.NewUserPort(users)
	authHandler := auth_handler.NewHttpHandler(auth_handler.NewHandler(auth_manager.NewAuthManager(userPort, tokens)))
	authHandler.Init(nil, router)
	currency_handler.NewHttpHandler(currencies).Init(nil, router)

	userHandler := user_handler.NewHttpHandler(user_handler.NewHandler(user_manager.NewUserManager(users)), middleware)
	userHandler.Init(nil, router)

	location, err := loadLocation(locationName)
	if err != nil {
		return err
	}
	itemHandler := item_handler.NewHttpHandler(item_handler.NewHandler(item_manager.NewItemManager(items, currencies, location, images.PresignRead)), middleware, images)
	itemHandler.Init(nil, router)
	return nil
}
