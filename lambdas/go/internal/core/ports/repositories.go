package ports

import (
	"context"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
)

// ExternalMapProviderRepository is an interface for interacting with where we get our maps
type ExternalMapProviderRepository interface {
	List(ctx context.Context, bounds domain.Bounds, mapType domain.MapType) ([]domain.Map, error)
	GetProvider() domain.MapProvider
}

// InternalMapRepository is an interface for interacting with the internal maps we have
type InternalMapRepository interface {
	List(ctx context.Context, provider domain.MapProvider, bounds domain.Bounds, mapType domain.MapType) ([]domain.Map, error)
	// Create does not work with the processedInternalRepository implementation of the InternalMapRepository
	Create(ctx context.Context, m domain.Map) error
	Download(ctx context.Context, m domain.Map) *domain.LocalMap
}

// ProductMapRepository is an interface for interacting with the product we show the end user.
type ProductMapRepository interface {
	Publish(ctx context.Context, m domain.LocalMap)
}
