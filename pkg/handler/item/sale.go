package item_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karat-server/common"
	manager "karat-server/pkg/manager/item"
)

func (h *HttpHandler) Sell(ctx *gin.Context, req *http.Request) (interface{}, error) {
	id, err := itemID(ctx)
	if err != nil {
		return nil, err
	}
	var body saleRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return nil, common.ErrInvalidJSON
	}
	return h.handler.Sell(req.Context(), req, id, body)
}

func (h *handler) Sell(ctx context.Context, req *http.Request, id uuid.UUID, body saleRequest) (*manager.Item, error) {
	if body.SoldOn == nil {
		return nil, common.ErrInvalidDate
	}
	if body.SoldAmount == nil {
		return nil, common.ErrInvalidPrice
	}
	userID, err := userID(req)
	if err != nil {
		return nil, err
	}
	return h.service.Sell(ctx, userID, id, manager.Sale{SoldOn: *body.SoldOn, SoldAmount: *body.SoldAmount})
}
