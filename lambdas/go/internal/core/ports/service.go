package ports

import (
	"context"

	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
)

type OrchestratorService interface {
	SyncInternalWithExternalMaps(ctx context.Context, request domain.SyncMapRequest) (*int, error)
	ListRawInternalMaps(ctx context.Context) ([]domain.Map, error)
	ListProcessedInternalMaps(ctx context.Context) ([]domain.Map, error)
	ListPublishedMaps(ctx context.Context) ([]domain.PublishedMap, error)
	PublishMap(ctx context.Context, m domain.Map) error
	DeleteMap(ctx context.Context, m domain.Map)
}
