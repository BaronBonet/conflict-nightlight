package handlers

import (
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
)

type MapControllerLambdaEventHandler struct {
	logger ports.Logger
	srv    ports.OrchestratorService
}

func NewMapControllerLambdaHandler(logger ports.Logger, srv ports.OrchestratorService) *MapControllerLambdaEventHandler {
	return &MapControllerLambdaEventHandler{logger: logger, srv: srv}
}

func (handler *MapControllerLambdaEventHandler) HandleEvent(_ context.Context, request *conflict_nightlightv1.SyncMapRequest) {
	ctx := infrastructure.NewContext()

	if request.Bounds == conflict_nightlightv1.Bounds_BOUNDS_UNSPECIFIED {
		handler.logger.Fatal(ctx, "Bounds cannot be unspecified")
	}
	if request.MapType == conflict_nightlightv1.MapType_MAP_TYPE_UNSPECIFIED {
		handler.logger.Fatal(ctx, "Map type cannot be unspecified")
	}
	count, err := handler.srv.SyncInternalWithExternalMaps(ctx, prototransformers.ProtoToSyncMapsRequest(request))
	if err != nil {
		handler.logger.Fatal(ctx, "Error while publishing map.", "error", err)
	}
	if *count > 0 {
		handler.logger.Info(ctx, "Successfully scraped and requested for new maps", "new-maps-added", count)
	} else {
		handler.logger.Info(ctx, "No new maps found")
	}
}
