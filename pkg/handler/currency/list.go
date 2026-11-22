package currency_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"karat-server/common"
	repository "karat-server/pkg/manager/currency/repository"
)

type HttpHandler struct {
	currencies repository.Repository
}

func NewHttpHandler(currencies repository.Repository) *HttpHandler {
	return &HttpHandler{currencies: currencies}
}

func (h *HttpHandler) Init(_ context.Context, router *gin.Engine) {
	router.GET("/currency", common.GenericHandler(h.List))
}

func (h *HttpHandler) List(ctx *gin.Context, _ *http.Request) (interface{}, error) {
	return h.currencies.List(ctx.Request.Context())
}
