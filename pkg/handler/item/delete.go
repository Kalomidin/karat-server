package item_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *HttpHandler) Delete(ctx *gin.Context, req *http.Request) (interface{}, error) {
	id, err := itemID(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.handler.Delete(req.Context(), req, id); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}

func (h *handler) Delete(ctx context.Context, req *http.Request, id uuid.UUID) error {
	userID, err := userID(req)
	if err != nil {
		return err
	}
	return h.service.Delete(ctx, userID, id)
}
