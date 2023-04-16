package main

import (
	"github.com/BaronBonet/conflict-nightlight/internal/adapters"
	"github.com/BaronBonet/conflict-nightlight/internal/handlers"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

func main() {
	logger := adapters.NewZapLogger(zap.NewProductionConfig(), false)
	lambdaHandler := handlers.NewMapPublisherLambdaHandler(logger)
	lambda.Start(lambdaHandler.HandleEvent)
}
