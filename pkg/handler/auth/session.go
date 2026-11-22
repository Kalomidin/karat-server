package auth_handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"karat-server/common"
	manager "karat-server/pkg/manager/auth"
)

type sessionRequest struct {
	IDToken string `json:"idToken"`
}

func (h *HttpHandler) Session(ctx *gin.Context, r *http.Request) (interface{}, error) {
	var req sessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !strings.Contains(err.Error(), "EOF") {
		return nil, common.ErrInvalidJSON
	}
	token := strings.TrimSpace(req.IDToken)
	if token == "" {
		token = bearer(r.Header.Get("Authorization"))
	}
	if token == "" {
		return nil, common.ErrUnauthorized
	}
	return h.handler.Session(r.Context(), token)
}

func (h *handler) Session(ctx context.Context, idToken string) (*manager.Session, error) {
	return h.service.Session(ctx, idToken)
}

func bearer(header string) string {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}
