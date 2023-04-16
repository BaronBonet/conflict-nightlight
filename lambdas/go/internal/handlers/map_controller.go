package handlers

import (
	"context"
	"fmt"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/externalmapsrepository"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters/internalmapsrepository"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure"
)

type SyncMapsRequest struct {
	Cropper        string `json:"cropper"`
	MapType        string `json:"mapType"`
	SelectedMonths []int  `json:"selectedMonths"`
	SelectedYears  []int  `json:"selectedYears"`
}

type MapControllerLambdaEventHandler struct {
	logger ports.Logger
}

func NewMapControllerLambdaEventHandler(logger ports.Logger) *MapControllerLambdaEventHandler {
	return &MapControllerLambdaEventHandler{logger: logger}
}

func (handler *MapControllerLambdaEventHandler) HandleEvent(_ context.Context, request SyncMapsRequest) {
	ctx := infrastructure.NewContext()

	writeDir := infrastructure.GetEnvOrDefault("WRITE_DIR", "/tmp")

	bounds := domain.StringToBounds(request.Cropper)
	if bounds == domain.BoundsUnspecified {
		handler.logger.Fatal(ctx, "Unknown bounds", "supplied bounds", bounds)
	}

	internalMapsRepo := internalmapsrepository.NewAWSInternalMapsRepository(ctx,
		handler.logger,
		infrastructure.GetEnvOrDefault("RAW_TIF_BUCKET", "conflict-nightlight-raw-tif"),
		infrastructure.GetEnvOrDefault("SOURCE_URL_KEY", "source-url"),
		infrastructure.GetEnvOrDefault("DOWNLOAD_RAW_TIF_QUEUE", "conflict-nightlight-download-and-crop-raw-tif-request"),
		writeDir,
	)

	externalMapsRepo := externalmapsrepository.NewEogdataExternalMapsRepository(handler.logger, fmt.Sprintf("%s/eog_cache", writeDir))

	service := services.NewMapService(handler.logger, internalMapsRepo, externalMapsRepo)
	handler.logger.Info(ctx, "Received Sync Maps request", "sync maps request", request)

	count := service.SyncInternalWithSource(ctx, bounds, domain.StringToMapType(request.MapType), domain.SelectedDates{
		Months: request.SelectedMonths,
		Years:  request.SelectedYears,
	})

	handler.logger.Info(ctx, "Success", "new-maps-added", count)
}
