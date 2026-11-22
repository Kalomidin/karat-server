package item_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	manager "karat-server/pkg/manager/item"
)

func (h *HttpHandler) ReturnToStock(ctx *gin.Context, req *http.Request) (interface{}, error) {
	id, err := itemID(ctx)
	if err != nil {
		return nil, err
	}
	return h.handler.ReturnToStock(req.Context(), req, id)
}

func (h *handler) ReturnToStock(ctx context.Context, req *http.Request, id uuid.UUID) (*manager.Item, error) {
	userID, err := userID(req)
	if err != nil {
		return nil, err
	}
	return h.service.ReturnToStock(ctx, userID, id)
}
