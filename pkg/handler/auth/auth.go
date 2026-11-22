package auth_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"karat-server/common"
	manager "karat-server/pkg/manager/auth"
)

type HttpHandler struct {
	handler AuthHandler
}

type handler struct {
	service manager.AuthManager
}

type AuthHandler interface {
	Session(ctx context.Context, idToken string) (*manager.Session, error)
}

func NewHandler(service manager.AuthManager) AuthHandler {
	return &handler{service: service}
}

func NewHttpHandler(h AuthHandler) *HttpHandler {
	return &HttpHandler{handler: h}
}

func (c *HttpHandler) Init(_ context.Context, router *gin.Engine) {
	routes := map[string]map[string]common.HandlerFunc{
		http.MethodPost: {
			"/session": c.Session,
		},
	}
	for method, route := range routes {
		for path, h := range route {
			router.Handle(method, "/auth"+path, common.GenericHandler(h))
		}
	}
}
