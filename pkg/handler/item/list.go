package item_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	manager "karat-server/pkg/manager/item"
)

func (h *HttpHandler) List(_ *gin.Context, req *http.Request) (interface{}, error) {
	return h.handler.List(req.Context(), req)
}

func (h *handler) List(ctx context.Context, req *http.Request) ([]manager.Item, error) {
	userID, err := userID(req)
	if err != nil {
		return nil, err
	}
	return h.service.List(ctx, userID)
}
