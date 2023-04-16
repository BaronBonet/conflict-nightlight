package main

import (
	"context"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/adapters"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/handlers"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	request := conflict_nightlightv1.RequestWrapper{Message: &conflict_nightlightv1.RequestWrapper_PublishMapProductRequest{PublishMapProductRequest: &conflict_nightlightv1.PublishMapProductRequest{Map: &conflict_nightlightv1.Map{
		Date: &conflict_nightlightv1.Date{
			Day:   0,
			Month: 1,
			Year:  2021,
		},
		MapType: conflict_nightlightv1.MapType(domain.StringToMapType("Monthly")),
		Bounds:  1,
		MapSource: &conflict_nightlightv1.MapSource{
			MapProvider: 1,
			Url:         "example.com",
		},
	}}}}
	v, err := protojson.Marshal(&request)
	if err != nil {
		panic("cannot marshal")
	}

	s := "1234"
	event := events.SQSEvent{Records: []events.SQSMessage{{
		Body: string(v),
		MessageAttributes: map[string]events.SQSMessageAttribute{"correlation-id": {
			StringValue: &s,
			DataType:    "string",
		},
		}}}}
	logger := adapters.NewZapLogger(zap.NewDevelopmentConfig(), true)
	handler := handlers.NewMapPublisherLambdaHandler(logger)
	handler.HandleEvent(context.TODO(), event)
}
