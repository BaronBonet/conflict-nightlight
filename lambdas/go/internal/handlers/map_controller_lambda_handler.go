package handlers

import (
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
)

type SyncMapsRequest struct {
	Bounds         string `json:"bounds"`
	MapType        string `json:"mapType"`
	SelectedMonths []int  `json:"selectedMonths"`
	SelectedYears  []int  `json:"selectedYears"`
}

type MapControllerLambdaEventHandler struct {
	logger ports.Logger
	srv    *services.OrchestratorService
}

func NewMapControllerLambdaHandler(logger ports.Logger, srv *services.OrchestratorService) *MapControllerLambdaEventHandler {
	return &MapControllerLambdaEventHandler{logger: logger, srv: srv}
}

func (handler *MapControllerLambdaEventHandler) HandleEvent(_ context.Context, request *conflict_nightlightv1.SyncMapRequest) {
	ctx := infrastructure.NewContext()
	bounds := domain.Bounds(request.Bounds)
	if bounds == domain.BoundsUnspecified {
		handler.logger.Fatal(ctx, "Unknown bounds", "supplied bounds", bounds)
	}
	count, err := handler.srv.SyncInternalWithExternalMaps(ctx, prototransformers.ProtoToSyncMapsRequest(request))
	if err != nil {
		handler.logger.Fatal(ctx, "Error while publishing map.", "error", err)
	}
	handler.logger.Info(ctx, "Successfully scraped and requested for new maps", "new-maps-added", count)
}
