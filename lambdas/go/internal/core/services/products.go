package services

import (
	"context"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
)

type ProductService struct {
	logger                         ports.Logger
	processedInternalMapRepository ports.InternalMapRepository
	productMapRepository           ports.ProductMapRepository
}

func NewProductService(logger ports.Logger, processedInternalMapRepository ports.InternalMapRepository, productMapRepository ports.ProductMapRepository) *ProductService {
	return &ProductService{
		logger:                         logger,
		processedInternalMapRepository: processedInternalMapRepository,
		productMapRepository:           productMapRepository,
	}
}

func (srv *ProductService) PublishMap(ctx context.Context, m domain.Map) {
	theMap := srv.processedInternalMapRepository.Download(ctx, m)

	if theMap == nil {
		srv.logger.Fatal(ctx, "a map was somehow nil")
	}
	// A lot of logic is currently handled by the Publish method, perhaps we break up the productMapRepository into two
	srv.productMapRepository.Publish(ctx, *theMap)
	srv.logger.Info(ctx, "Successfully Published map to mapbox and updated our internal repo")
}
