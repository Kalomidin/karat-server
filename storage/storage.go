package storage

import (
	"context"
	"errors"
	"strings"
)

// ErrNotConfigured means Cloudflare R2 credentials are missing.
var ErrNotConfigured = errors.New("image storage is not configured")

// Upload is a one-time permission to PUT one piece photo.
type Upload struct {
	URL         string
	Key         string
	ReadURL     string
	ContentType string
}

// Store issues presigned uploads and read links, and checks key ownership.
type Store interface {
	PresignUpload(ctx context.Context, googleID, contentType string) (Upload, error)
	PresignRead(ctx context.Context, key string) (string, error)
	BelongsTo(googleID, key string) bool
}

func belongs(googleID, key string) bool {
	if googleID == "" || strings.Contains(googleID, "/") || strings.Contains(key, "..") {
		return false
	}
	prefix := "items/" + googleID + "/"
	if !strings.HasPrefix(key, prefix) {
		return false
	}
	name := strings.TrimPrefix(key, prefix)
	return name != "" && !strings.Contains(name, "/")
}
