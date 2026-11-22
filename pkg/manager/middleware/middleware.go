package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"karat-server/common"
	"karat-server/pkg/manager/port"
	"karat-server/pkg/provider"
)

type middleware struct {
	provider provider.Client
	users    port.UserPort
}

func NewMiddleware(provider provider.Client, users port.UserPort) common.Middleware {
	return &middleware{provider: provider, users: users}
}

func (m *middleware) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.Next()
			return
		}
		token, ok := bearer(header)
		if !ok {
			abort(ctx, common.ErrUnauthorized)
			return
		}
		details, err := m.provider.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			abort(ctx, common.ErrUnauthorized)
			return
		}
		user, err := m.users.UpsertGoogleUser(ctx.Request.Context(), port.GoogleUser{
			GoogleID: details.GoogleID,
			Email:    details.Email,
			Name:     details.Name,
			PhotoURL: details.PhotoURL,
		})
		if err != nil {
			abort(ctx, common.NewError(http.StatusInternalServerError, "internal"))
			return
		}
		ctx.Request = ctx.Request.WithContext(common.WithUser(ctx.Request.Context(), common.User{
			ID:       user.ID,
			GoogleID: user.GoogleID,
			Email:    user.Email,
			Name:     user.Name,
			PhotoURL: user.PhotoURL,
		}))
		ctx.Next()
	}
}

func (m *middleware) HandlerWithAuth(handler common.HandlerFunc) common.HandlerFunc {
	return func(ctx *gin.Context, req *http.Request) (interface{}, error) {
		if _, err := common.CurrentUser(req.Context()); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func bearer(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}

func abort(ctx *gin.Context, err *common.Error) {
	ctx.AbortWithStatusJSON(err.StatusCode(), gin.H{"error": err.ErrorCode()})
}
