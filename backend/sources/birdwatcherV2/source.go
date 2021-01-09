package birdwatcherV2

import (
	"github.com/alice-lg/alice-lg/backend/api"
	"github.com/alice-lg/alice-lg/backend/caches"
	"github.com/alice-lg/alice-lg/backend/sources"
	"github.com/alice-lg/alice-lg/backend/sources/birdwatcher"
	"sort"
)

type Birdwatcher interface {
	sources.Source
}

type GenericBirdwatcher struct {
	config birdwatcher.Config
	client *Client

	// Caches: Neighbors
	neighborsCache *caches.NeighborsCache

	// Caches: Routes
	routesRequiredCache    *caches.RoutesCache
	routesNotExportedCache *caches.RoutesCache

	// Mutices:
	routesFetchMutex *LockMap
}

func NewBirdwatcher(config birdwatcher.Config) birdwatcher.Birdwatcher {
	client := NewClient(config.Api)

	// Cache settings:
	// TODO: Maybe read from config file
	neighborsCacheDisable := false

	routesCacheDisabled := false
	routesCacheMaxSize := 128

	// Initialize caches
	neighborsCache := caches.NewNeighborsCache(neighborsCacheDisable)
	routesRequiredCache := caches.NewRoutesCache(routesCacheDisabled, routesCacheMaxSize)
	routesNotExportedCache := caches.NewRoutesCache(routesCacheDisabled, routesCacheMaxSize)

	var birdwatcher birdwatcher.Birdwatcher

	if config.Type == "single_table" {
		singleTableBirdwatcher := new(SingleTableBirdwatcher)

		singleTableBirdwatcher.config = config
		singleTableBirdwatcher.client = client

		singleTableBirdwatcher.neighborsCache = neighborsCache

		singleTableBirdwatcher.routesRequiredCache = routesRequiredCache
		singleTableBirdwatcher.routesNotExportedCache = routesNotExportedCache

		singleTableBirdwatcher.routesFetchMutex = NewLockMap()

		birdwatcher = singleTableBirdwatcher
	} else if config.Type == "multi_table" {
		multiTableBirdwatcher := new(MultiTableBirdwatcher)

		multiTableBirdwatcher.config = config
		multiTableBirdwatcher.client = client

		multiTableBirdwatcher.neighborsCache = neighborsCache

		multiTableBirdwatcher.routesRequiredCache = routesRequiredCache
		multiTableBirdwatcher.routesNotExportedCache = routesNotExportedCache

		multiTableBirdwatcher.routesFetchMutex = NewLockMap()

		birdwatcher = multiTableBirdwatcher
	}

	return birdwatcher
}

func (b *GenericBirdwatcher) filterRoutesByPeerOrLearntFrom(routes api.Routes, peer string, learntFrom string) api.Routes {
	resultRoutes := make(api.Routes, 0, len(routes))

	// Choose routes with next_hop == gateway of this neighbour
	for _, route := range routes {
		if (route.Gateway == peer) ||
			(route.Gateway == learntFrom) ||
			(route.LearntFrom == peer) {
			resultRoutes = append(resultRoutes, route)
		}
	}

	// Sort routes for deterministic ordering
	sort.Sort(resultRoutes)
	routes = resultRoutes

	return routes
}


