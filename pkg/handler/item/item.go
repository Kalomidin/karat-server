package item_handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karat-server/common"
	manager "karat-server/pkg/manager/item"
	"karat-server/storage"
)

type HttpHandler struct {
	handler    ItemHandler
	middleware common.Middleware
	images     storage.Store
}

type handler struct {
	service manager.ItemManager
}

type ItemHandler interface {
	List(ctx context.Context, req *http.Request) ([]manager.Item, error)
	Create(ctx context.Context, req *http.Request, body itemRequest) ([]manager.Item, error)
	Get(ctx context.Context, req *http.Request, id uuid.UUID) (*manager.Item, error)
	Update(ctx context.Context, req *http.Request, id uuid.UUID, body itemRequest) (*manager.Item, error)
	Delete(ctx context.Context, req *http.Request, id uuid.UUID) error
	Sell(ctx context.Context, req *http.Request, id uuid.UUID, body saleRequest) (*manager.Item, error)
	ReturnToStock(ctx context.Context, req *http.Request, id uuid.UUID) (*manager.Item, error)
}

func NewHandler(service manager.ItemManager) ItemHandler {
	return &handler{service: service}
}

func NewHttpHandler(h ItemHandler, middleware common.Middleware, images storage.Store) *HttpHandler {
	return &HttpHandler{handler: h, middleware: middleware, images: images}
}

func (c *HttpHandler) Init(_ context.Context, router *gin.Engine) {
	auth := c.middleware.HandlerWithAuth
	router.POST("/item/upload-url", common.GenericHandler(auth(c.UploadURL)))
	routes := map[string]map[string]common.HandlerFunc{
		http.MethodGet: {
			"":     auth(c.List),
			"/:id": auth(c.Get),
		},
		http.MethodPost: {
			"":            auth(c.Create),
			"/:id/sale":   auth(c.Sell),
			"/:id/return": auth(c.ReturnToStock),
		},
		http.MethodPut: {
			"/:id": auth(c.Update),
		},
		http.MethodDelete: {
			"/:id": auth(c.Delete),
		},
	}
	for method, route := range routes {
		for path, h := range route {
			router.Handle(method, "/item"+path, common.GenericHandler(h))
		}
	}
}

func (c *HttpHandler) ownedImage(req *http.Request, key *string) error {
	if key == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*key)
	*key = trimmed
	if trimmed == "" {
		return nil
	}
	user, err := common.CurrentUser(req.Context())
	if err != nil {
		return err
	}
	if c.images == nil || !c.images.BelongsTo(user.GoogleID, trimmed) {
		return common.ErrInvalidImage
	}
	return nil
}

func (c *HttpHandler) UploadURL(ctx *gin.Context, req *http.Request) (interface{}, error) {
	var body uploadRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return nil, common.ErrInvalidJSON
	}
	user, err := common.CurrentUser(req.Context())
	if err != nil {
		return nil, err
	}
	signed, err := c.images.PresignUpload(req.Context(), user.GoogleID, body.ContentType)
	switch {
	case errors.Is(err, storage.ErrNotConfigured):
		return nil, common.ErrImageUnavailable
	case errors.Is(err, storage.ErrInvalidContent):
		return nil, common.ErrInvalidImage
	case err != nil:
		return nil, err
	}
	return uploadResponse{
		UploadURL:   signed.URL,
		ImageKey:    signed.Key,
		ReadURL:     signed.ReadURL,
		ContentType: signed.ContentType,
	}, nil
}

func userID(req *http.Request) (uuid.UUID, error) {
	user, err := common.CurrentUser(req.Context())
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func itemID(ctx *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return uuid.Nil, common.ErrMissing
	}
	return id, nil
}
