package provider

import (
	"context"
	"karat-server/pkg/provider/google"
	"karat-server/pkg/provider/model"
)

type Client interface {
	VerifyIDToken(ctx context.Context, raw string) (*model.UserDetails, error)
}

type providerClient struct {
	google *google.Client
}

func NewProviderClient(googleClient *google.Client) Client {
	return &providerClient{google: googleClient}
}

func (p *providerClient) VerifyIDToken(ctx context.Context, raw string) (*model.UserDetails, error) {
	return p.google.VerifyIDToken(ctx, raw)
}
