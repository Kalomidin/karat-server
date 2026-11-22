package common

import (
	"errors"
	"karat-server/common/response"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HandlerFunc func(ctx *gin.Context, req *http.Request) (interface{}, error)

func GenericHandler(h HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := h(ctx, ctx.Request)
		if err != nil {
			var coded *Error
			if errors.As(err, &coded) {
				if sendErr := response.NewErrorMessage(coded.ErrorCode(), coded.StatusCode()).Send(ctx.Writer); sendErr != nil {
					log.Printf("failed to send response: %v", sendErr)
				}
				return
			}
			log.Printf("request failed: %v", err)
			if sendErr := response.NewErrorMessage("internal", http.StatusInternalServerError).Send(ctx.Writer); sendErr != nil {
				log.Printf("failed to send response: %v", sendErr)
			}
			return
		}
		if sendErr := response.NewSuccess(resp, http.StatusOK).Send(ctx.Writer); sendErr != nil {
			log.Printf("failed to send response: %v", sendErr)
		}
	}
}
