package handlers

import (
	"context"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
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

func (handler *MapControllerLambdaEventHandler) HandleEvent(_ context.Context, request SyncMapsRequest) {
	ctx := infrastructure.NewContext()
	bounds := domain.StringToBounds(request.Bounds)
	if bounds == domain.BoundsUnspecified {
		handler.logger.Fatal(ctx, "Unknown bounds", "supplied bounds", bounds)
	}
	count, err := handler.srv.SyncInternalWithExternal(ctx, bounds, domain.StringToMapType(request.MapType), domain.SelectedDates{
		Months: request.SelectedMonths,
		Years:  request.SelectedYears,
	})
	if err != nil {
		handler.logger.Fatal(ctx, "Error while publishing map.", "error", err)
	}
	handler.logger.Info(ctx, "Successfully scraped and requested for new maps", "new-maps-added", count)
}
