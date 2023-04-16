package services

import (
	"context"
	"errors"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slices"
)

type RawMapService struct {
	logger                ports.Logger
	internalMapRepository ports.InternalMapRepository
	sourceMapRepository   ports.ExternalMapProviderRepository
}

func NewMapService(logger ports.Logger, internalMapRepository ports.InternalMapRepository, sourceMapRepository ports.ExternalMapProviderRepository) *RawMapService {
	return &RawMapService{logger: logger, internalMapRepository: internalMapRepository, sourceMapRepository: sourceMapRepository}
}

func (srv *RawMapService) SyncInternalWithSource(ctx context.Context, cropper domain.Bounds, mapType domain.MapType, selectedDates domain.SelectedDates) int {

	newMaps, err := srv.findNewMaps(ctx, cropper, mapType, selectedDates)
	if err != nil {
		srv.logger.Fatal(ctx, "Error when finding new maps", "error", err)
	}
	count, err := srv.addNewMaps(ctx, newMaps)
	if err != nil {
		srv.logger.Fatal(ctx, "Error when adding new maps.")
	}
	srv.logger.Info(ctx, "Successfully requested to download and crop new maps", "number of new maps", count)
	return count
}

func (srv *RawMapService) findNewMaps(ctx context.Context, cropper domain.Bounds, mapType domain.MapType, selectedDates domain.SelectedDates) ([]domain.Map, error) {

	validate := validator.New()
	if err := validate.Struct(selectedDates); err != nil {
		return nil, errors.New("the selected dates were not valid")
	}

	internalMaps, err := srv.internalMapRepository.List(ctx, srv.sourceMapRepository.GetProvider(), cropper, mapType)
	if err != nil {
		return nil, errors.New("get internal maps from internalMapRepository has failed, error: " + err.Error())
	}
	srv.logger.Debug(ctx, "Successfully extracted internal maps", "internalMaps", internalMaps)

	sourceMaps, err := srv.sourceMapRepository.List(ctx, cropper, mapType)
	if err != nil {
		return nil, errors.New("get source maps from sourceMapRepository has failed")
	}
	srv.logger.Debug(ctx, "Successfully obtained maps from the source.")

	var newMaps []domain.Map
	for _, sourceMap := range sourceMaps {
		if slices.Contains(selectedDates.Months, sourceMap.Date.Month) && slices.Contains(selectedDates.Years, sourceMap.Date.Year) && !slices.Contains(internalMaps, sourceMap) {
			srv.logger.Debug(ctx, "Found a map we do not have in our internal repo", "sourceMap", sourceMap)
			newMaps = append(newMaps, sourceMap)
		}
	}

	return newMaps, nil
}

func (srv *RawMapService) addNewMaps(ctx context.Context, newMaps []domain.Map) (int, error) {
	var count int
	for _, newMap := range newMaps {
		err := srv.internalMapRepository.Create(ctx, newMap)
		if err != nil {
			srv.logger.Error(ctx, "Error when adding a new map", "error", err)
		} else {
			count += 1
		}
	}
	return count, nil
}
