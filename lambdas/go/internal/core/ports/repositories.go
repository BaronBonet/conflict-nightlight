package ports

import (
	"context"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
)

// ExternalMapProviderRepo is an interface for interacting with where we get our maps
type ExternalMapProviderRepo interface {
	List(ctx context.Context, bounds domain.Bounds, mapType domain.MapType) ([]domain.Map, error)
	GetProvider() domain.MapProvider
}

// InternalMapRepo is an interface for interacting with the internal maps we have
type InternalMapRepo interface {
	List(ctx context.Context, provider domain.MapProvider, bounds domain.Bounds, mapType domain.MapType) ([]domain.Map, error)
	// Create does not work with the processedInternalRepository implementation of the InternalMapRepo
	//  TODO make it work
	Create(ctx context.Context, m domain.Map) error
	Download(ctx context.Context, m domain.Map) (*domain.LocalMap, error)
	Delete(ctx context.Context, m domain.Map) error
}

// MapTileServerRepo is the interface for interacting with the (currently) external service where our maps are hosted, that the frontend can display
type MapTileServerRepo interface {
	Publish(ctx context.Context, m domain.LocalMap) (*domain.PublishedMap, error)
	Delete(ctx context.Context, m domain.Map) error
}

// FrontendMapDataRepo is an interface for interacting with the pointers to the published maps the frontend can display
type FrontendMapDataRepo interface {
	Upsert(ctx context.Context, m domain.PublishedMap) error
	List(ctx context.Context) ([]domain.PublishedMap, error)
	Delete(ctx context.Context, m domain.Map) error
}
