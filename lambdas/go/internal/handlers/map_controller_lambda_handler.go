package handlers

import (
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/golang/protobuf/jsonpb"
	"strings"
)

type MapControllerLambdaEventHandler struct {
	logger ports.Logger
	srv    *services.OrchestratorService
}

func NewMapControllerLambdaHandler(logger ports.Logger, srv *services.OrchestratorService) *MapControllerLambdaEventHandler {
	return &MapControllerLambdaEventHandler{logger: logger, srv: srv}
}

func (handler *MapControllerLambdaEventHandler) HandleEvent(_ context.Context, payload string) {
	ctx := infrastructure.NewContext()

	request := &conflict_nightlightv1.SyncMapRequest{}

	if err := jsonpb.Unmarshal(strings.NewReader(payload), request); err != nil {
		handler.logger.Fatal(ctx, "Error while unmarshaling JSON payload.", "error", err)
	}
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
	handler.logger.Info(ctx, "Successfully scraped and requested for new maps", "new-maps-added", count)
}
