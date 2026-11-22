package common

import (
	"github.com/gin-gonic/gin"
)

type Middleware interface {
	AuthMiddleware() gin.HandlerFunc
	HandlerWithAuth(handler HandlerFunc) HandlerFunc
}
