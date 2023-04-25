package services

import (
	"context"
	"errors"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slices"
	"time"
)

type OrchestratorService struct {
	logger                   ports.Logger
	externalMapsRepo         ports.ExternalMapProviderRepo
	rawInternalMapRepo       ports.InternalMapRepo
	processedInternalMapRepo ports.InternalMapRepo
	frontendMapDataRepo      ports.FrontendMapDataRepo
	mapTileServerRepo        ports.MapTileServerRepo
}

func NewOrchestratorService(logger ports.Logger, externalMapsRepo ports.ExternalMapProviderRepo, rawInternalMapRepo ports.InternalMapRepo, processedInternalMapRepo ports.InternalMapRepo, frontendMapDataRepo ports.FrontendMapDataRepo, mapTileServerRepo ports.MapTileServerRepo) *OrchestratorService {
	return &OrchestratorService{
		logger:                   logger,
		externalMapsRepo:         externalMapsRepo,
		processedInternalMapRepo: processedInternalMapRepo,
		rawInternalMapRepo:       rawInternalMapRepo,
		frontendMapDataRepo:      frontendMapDataRepo,
		mapTileServerRepo:        mapTileServerRepo,
	}
}

// SyncInternalWithExternalMaps syncs the maps from the external map repo with the maps we have in the internal map repo
func (srv *OrchestratorService) SyncInternalWithExternalMaps(ctx context.Context, request domain.SyncMapRequest) (*int, error) {
	newMaps, err := srv.findNewMaps(ctx, request.Bounds, request.MapType, request.SelectedDates)
	if err != nil {
		srv.logger.Error(ctx, "Error when finding new maps", "error", err)
		return nil, err
	}
	count, err := srv.addNewMaps(ctx, newMaps)
	if err != nil {
		srv.logger.Error(ctx, "Error when adding new maps.")
		return nil, err
	}
	return &count, nil
}

func (srv *OrchestratorService) ListRawInternalMaps(ctx context.Context) ([]domain.Map, error) {
	return srv.rawInternalMapRepo.List(ctx, srv.externalMapsRepo.GetProvider(), domain.BoundsUnspecified, domain.MapTypeUnspecified)
}

func (srv *OrchestratorService) ListProcessedInternalMaps(ctx context.Context) ([]domain.Map, error) {
	return srv.processedInternalMapRepo.List(ctx, domain.MapProviderUnspecified, domain.BoundsUnspecified, domain.MapTypeUnspecified)
}

func (srv *OrchestratorService) ListPublishedMaps(ctx context.Context) ([]domain.PublishedMap, error) {
	return srv.frontendMapDataRepo.List(ctx)
}

func (srv *OrchestratorService) PublishMap(ctx context.Context, m domain.Map) error {
	theMap, err := srv.processedInternalMapRepo.Download(ctx, m)
	if err != nil {
		return err
	}
	if theMap == nil {
		msg := "A map was somehow nil after it was downloaded"
		srv.logger.Error(ctx, msg, "map", m)
		return errors.New(msg)
	}
	publishedMap, err := srv.mapTileServerRepo.Publish(ctx, *theMap)
	if err != nil {
		srv.logger.Error(ctx, "While publishing the map", "map", m, "error", err)
		return err
	}
	if publishedMap == nil {
		msg := "the map returned from the map tile server repo was somehow nil"
		srv.logger.Error(ctx, msg)
		return errors.New(msg)
	}
	if err != srv.frontendMapDataRepo.Upsert(ctx, *publishedMap) {
		return err
	}
	return nil
}

// DeleteMap deletes a map from all the repos
func (srv *OrchestratorService) DeleteMap(ctx context.Context, m domain.Map) {
	if err := srv.mapTileServerRepo.Delete(ctx, m); err != nil {
		srv.logger.Error(ctx, "Map was not deleted from the tile server repo", "error", err)
	} else {
		srv.logger.Debug(ctx, "Map was deleted from the tile server repo", "map", m)
	}
	if err := srv.frontendMapDataRepo.Delete(ctx, m); err != nil {
		srv.logger.Error(ctx, "Map was not deleted from the frontend repo", "error", err)
	} else {
		srv.logger.Debug(ctx, "Map was deleted from the frontend repo", "map", m)
	}
	if err := srv.processedInternalMapRepo.Delete(ctx, m); err != nil {
		srv.logger.Error(ctx, "Map was not deleted from the processed internal repo", "error", err)
	} else {
		srv.logger.Debug(ctx, "Map was deleted from the processed internal repo", "map", m)
	}
	if err := srv.rawInternalMapRepo.Delete(ctx, m); err != nil {
		srv.logger.Error(ctx, "Map was not deleted from the raw internal repo", "error", err)
	} else {
		srv.logger.Debug(ctx, "Map was deleted from the raw internal repo", "map", m)
	}
}

func (srv *OrchestratorService) findNewMaps(ctx context.Context, cropper domain.Bounds, mapType domain.MapType, selectedDates domain.SelectedDates) ([]domain.Map, error) {
	validate := validator.New()
	if err := validate.Struct(selectedDates); err != nil {
		return nil, errors.New("the selected dates were not valid")
	}
	provider := srv.externalMapsRepo.GetProvider()
	internalMaps, err := srv.rawInternalMapRepo.List(ctx, provider, cropper, mapType)
	if err != nil {
		return nil, errors.New("get internal maps from internalMapRepository has failed, error: " + err.Error())
	}
	srv.logger.Debug(ctx, "Successfully extracted internal maps", "internalMaps", internalMaps)

	sourceMaps, err := srv.externalMapsRepo.List(ctx, cropper, mapType)
	if err != nil {
		return nil, errors.New("get source maps from sourceMapRepository has failed")
	}
	srv.logger.Debug(ctx, "Successfully obtained maps from the source.")

	var newMaps []domain.Map
	for _, sourceMap := range sourceMaps {
		if slices.Contains(selectedDates.Months, time.Month(sourceMap.Date.Month)) && slices.Contains(selectedDates.Years, sourceMap.Date.Year) && !slices.Contains(internalMaps, sourceMap) {
			srv.logger.Debug(ctx, "Found a map we do not have in our internal repo", "sourceMap", sourceMap)
			newMaps = append(newMaps, sourceMap)
		}
	}
	return newMaps, nil
}

func (srv *OrchestratorService) addNewMaps(ctx context.Context, newMaps []domain.Map) (int, error) {
	var count int
	for _, newMap := range newMaps {
		err := srv.rawInternalMapRepo.Create(ctx, newMap)
		if err != nil {
			srv.logger.Error(ctx, "Error when adding a new map", "error", err)
		} else {
			count += 1
		}
	}
	return count, nil
}
