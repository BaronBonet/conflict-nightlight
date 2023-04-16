package main

import (
	"context"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters"
	"github.com/BaronBonet/conflict-nightlight/internal/handlers"
	"go.uber.org/zap"
)

var syncMapRequest = handlers.SyncMapsRequest{
	Cropper:        "ukraine-and-around",
	MapType:        "monthly",
	SelectedMonths: []int{1, 2},
	SelectedYears:  []int{2021},
}

func main() {
	logger := adapters.NewZapLogger(zap.NewDevelopmentConfig(), true)
	handler := handlers.NewMapControllerLambdaEventHandler(logger)
	handler.HandleEvent(context.TODO(), syncMapRequest)
}
