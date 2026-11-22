package item_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	manager "karat-server/pkg/manager/item"
)

func (h *HttpHandler) Update(ctx *gin.Context, req *http.Request) (interface{}, error) {
	id, err := itemID(ctx)
	if err != nil {
		return nil, err
	}
	body, err := readItem(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.ownedImage(req, body.ImageKey); err != nil {
		return nil, err
	}
	return h.handler.Update(req.Context(), req, id, body)
}

func (h *handler) Update(ctx context.Context, req *http.Request, id uuid.UUID, body itemRequest) (*manager.Item, error) {
	userID, err := userID(req)
	if err != nil {
		return nil, err
	}
	in, err := body.input()
	if err != nil {
		return nil, err
	}
	return h.service.Update(ctx, userID, id, in)
}
