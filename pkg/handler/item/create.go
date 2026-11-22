package item_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	manager "karat-server/pkg/manager/item"
)

func (h *HttpHandler) Create(ctx *gin.Context, req *http.Request) (interface{}, error) {
	body, err := readItem(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.ownedImage(req, body.ImageKey); err != nil {
		return nil, err
	}
	return h.handler.Create(req.Context(), req, body)
}

func (h *handler) Create(ctx context.Context, req *http.Request, body itemRequest) (*manager.Item, error) {
	userID, err := userID(req)
	if err != nil {
		return nil, err
	}
	in, err := body.input()
	if err != nil {
		return nil, err
	}
	return h.service.Create(ctx, userID, in)
}
