package user_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"karat-server/common"
	manager "karat-server/pkg/manager/user"
)

type HttpHandler struct {
	handler    UserHandler
	middleware common.Middleware
}

type handler struct {
	service manager.UserManager
}

type UserHandler interface {
	Me(ctx context.Context, req *http.Request) (*manager.User, error)
}

func NewHandler(service manager.UserManager) UserHandler {
	return &handler{service: service}
}

func NewHttpHandler(h UserHandler, middleware common.Middleware) *HttpHandler {
	return &HttpHandler{handler: h, middleware: middleware}
}

func (c *HttpHandler) Init(_ context.Context, router *gin.Engine) {
	routes := map[string]map[string]common.HandlerFunc{
		http.MethodGet: {
			"/me": c.middleware.HandlerWithAuth(c.Me),
		},
	}
	for method, route := range routes {
		for path, h := range route {
			router.Handle(method, "/user"+path, common.GenericHandler(h))
		}
	}
}
