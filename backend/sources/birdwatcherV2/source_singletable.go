package birdwatcherV2

import (
	"github.com/alice-lg/alice-lg/backend/api"
)

type SingleTableBirdwatcher struct {
	GenericBirdwatcher
}

func (s SingleTableBirdwatcher) ExpireCaches() int {
	panic("implement me")
}

func (s SingleTableBirdwatcher) Status() (*api.StatusResponse, error) {
	panic("implement me")
}

func (s SingleTableBirdwatcher) Neighbours() (*api.NeighboursResponse, error) {
	panic("implement me")
}

func (s SingleTableBirdwatcher) NeighboursStatus() (*api.NeighboursStatusResponse, error) {
	panic("implement me")
}

func (s SingleTableBirdwatcher) Routes(neighbourId string) (*api.RoutesResponse, error) {
	panic("implement me")
}

func (s SingleTableBirdwatcher) RoutesReceived(neighbourId string) (*api.RoutesResponse, error) {
	panic("implement me")
}

func (s SingleTableBirdwatcher) RoutesFiltered(neighbourId string) (*api.RoutesResponse, error) {
	panic("implement me")
}

func (s SingleTableBirdwatcher) RoutesNotExported(neighbourId string) (*api.RoutesResponse, error) {
	panic("implement me")
}

func (s SingleTableBirdwatcher) AllRoutes() (*api.RoutesResponse, error) {
	panic("implement me")
}

