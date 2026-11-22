package user_handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"karat-server/common"
	"karat-server/pkg/manager/port"
	manager "karat-server/pkg/manager/user"
)

func (h *HttpHandler) Me(ctx *gin.Context, r *http.Request) (interface{}, error) {
	return h.handler.Me(r.Context(), r)
}

func (h *handler) Me(ctx context.Context, _ *http.Request) (*manager.User, error) {
	current, err := common.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return h.service.Get(ctx, port.User{
		ID:       current.ID,
		GoogleID: current.GoogleID,
		Email:    current.Email,
		Name:     current.Name,
		PhotoURL: current.PhotoURL,
	})
}
