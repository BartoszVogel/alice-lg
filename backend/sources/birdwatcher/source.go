package birdwatcher

import (
	"github.com/alice-lg/alice-lg/backend/api"
	"github.com/alice-lg/alice-lg/backend/caches"

	"log"
	"sort"
)

const (
	NEIGHBOR_SUMMARY_ENDPOINT = "/neighbors/summary"
)

type Birdwatcher struct {
	config         Config
	client         *Client
	neighborsCache *caches.NeighborsCache
	routesCache    *caches.RoutesCache

	hasNeighborSummary bool
}

func NewBirdwatcher(config Config) *Birdwatcher {
	client := NewClient(config.Api)

	neighborsCache := caches.NewNeighborsCache(false)
	routesCache := caches.NewRoutesCache(false, 128)
	// TODO: Make LRU routes cache max size configurable

	// Check if we have a neighbor summary endpoint:
	hasNeighborSummary := true
	if config.DisableNeighborSummary {
		hasNeighborSummary = false
		log.Println("Config override: Disable neighbor summary; using `show protocols all`")
	}

	_, err := client.GetJson(NEIGHBOR_SUMMARY_ENDPOINT)
	if err != nil {
		hasNeighborSummary = false
	} else {
		if !config.DisableNeighborSummary {
			log.Println("Using neighbor-summary capabilities on:", config.Name)
		}
	}

	birdwatcher := &Birdwatcher{
		config:         config,
		client:         client,
		neighborsCache: neighborsCache,
		routesCache:    routesCache,

		hasNeighborSummary: hasNeighborSummary,
	}
	return birdwatcher
}

func (self *Birdwatcher) Status() (api.StatusResponse, error) {
	bird, err := self.client.GetJson("/status")
	if err != nil {
		return api.StatusResponse{}, err
	}

	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.StatusResponse{}, err
	}

	birdStatus, err := parseBirdwatcherStatus(bird, self.config)
	if err != nil {
		return api.StatusResponse{}, err
	}

	response := api.StatusResponse{
		Api:    apiStatus,
		Status: birdStatus,
	}

	return response, nil
}

// Get neighbors
func (self *Birdwatcher) Neighbours() (api.NeighboursResponse, error) {
	// Check if we hit the cache
	cachedResponse := self.neighborsCache.Get()
	if cachedResponse != nil {
		return *cachedResponse, nil // dereference for now...
	}

	var (
		response api.NeighboursResponse
		err      error
	)

	if self.hasNeighborSummary {
		response, err = self.summaryNeighbors()
	} else {
		response, err = self.bgpProtocolsNeighbors()
	}

	if err != nil {
		return api.NeighboursResponse{}, err
	}

	self.neighborsCache.Set(&response)

	return response, nil
}

// Get neighbors from neighbors summary
func (self *Birdwatcher) summaryNeighbors() (api.NeighboursResponse, error) {
	// Query birdwatcher
	bird, err := self.client.GetJson(NEIGHBOR_SUMMARY_ENDPOINT)
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	neighbors, err := parseNeighborSummary(bird, self.config)
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	response := api.NeighboursResponse{
		Api:        apiStatus,
		Neighbours: neighbors,
	}

	return response, nil
}

// Get neighbors from protocols
func (self *Birdwatcher) bgpProtocolsNeighbors() (api.NeighboursResponse, error) {

	// Query birdwatcher
	bird, err := self.client.GetJson("/protocols/bgp")
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	neighbours, err := parseNeighbours(bird, self.config)
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	response := api.NeighboursResponse{
		Api:        apiStatus,
		Neighbours: neighbours,
	}

	return response, nil // dereference for now
}

// Get filtered and exported routes
func (self *Birdwatcher) Routes(neighbourId string) (api.RoutesResponse, error) {
	// Check if we have a cache hit
	response := self.routesCache.Get("all:" + neighbourId)
	if response != nil {
		return *response, nil
	}

	// Exported
	bird, err := self.client.GetJson("/routes/protocol/" + neighbourId)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	// Use api status from first request
	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	imported, err := parseRoutes(bird, self.config)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	gateway := ""
	learnt_from := ""
	if len(imported) > 0 { // infer next_hop ip address from imported[0]
		gateway = imported[0].Gateway                                         //TODO: change mechanism to infer gateway when state becomes available elsewhere.
		learnt_from = mustString(imported[0].Details["learnt_from"], gateway) // also take learnt_from address into account if present.
		// ^ learnt_from is regularly present on routes for remote-triggered blackholing or on filtered routes (e.g. next_hop not in AS-Set)
	}

	// Optional: Filtered
	bird, _ = self.client.GetJson("/routes/filtered/" + neighbourId)
	filtered, err := parseRoutes(bird, self.config)
	if err != nil {
		log.Println("WARNING Could not retrieve filtered routes:", err)
		log.Println("Is the 'routes_filtered' module active in birdwatcher?")
	} else { // we got a filtered routes response => perform routes deduplication
		result_filtered := make(api.Routes, 0, len(filtered))
		result_imported := make(api.Routes, 0, len(imported))

		importedMap := make(map[string]*api.Route) // for O(1) access
		for _, route := range imported {
			importedMap[route.Id] = route
		}
		// choose routes with next_hop == gateway of this neighbour
		for _, route := range filtered {
			if (route.Gateway == gateway) || (route.Gateway == learnt_from) || (route.Details["learnt_from"] == gateway) {
				result_filtered = append(result_filtered, route)
				delete(importedMap, route.Id) // remove routes that are filtered on pipe
			} else if len(imported) == 0 { // in case there are just filtered routes
				result_filtered = append(result_filtered, route)
			}
		}
		sort.Sort(result_filtered)
		filtered = result_filtered
		// map to slice
		for _, route := range importedMap {
			result_imported = append(result_imported, route)
		}
		sort.Sort(result_imported)
		imported = result_imported
	}

	// Optional: NoExport
	bird, _ = self.client.GetJson("/routes/noexport/" + neighbourId)
	noexport, err := parseRoutes(bird, self.config)
	if err != nil {
		log.Println("WARNING Could not retrieve routes not exported:", err)
		log.Println("Is the 'routes_noexport' module active in birdwatcher?")
	} else {
		result_noexport := make(api.Routes, 0, len(noexport))
		// choose routes with next_hop == gateway of this neighbour
		for _, route := range noexport {
			if (route.Gateway == gateway) || (route.Gateway == learnt_from) {
				result_noexport = append(result_noexport, route)
			} else if len(imported) == 0 { // in case there are just filtered routes
				result_noexport = append(result_noexport, route)
			}
		}
	}

	response = &api.RoutesResponse{
		Api:         apiStatus,
		Imported:    imported,
		Filtered:    filtered,
		NotExported: noexport,
	}

	self.routesCache.Set("all:"+neighbourId, response)

	return *response, nil
}

// Make routes lookup
func (self *Birdwatcher) LookupPrefix(prefix string) (api.RoutesLookupResponse, error) {
	// Get RS info
	rs := api.Routeserver{
		Id:   self.config.Id,
		Name: self.config.Name,
	}

	// Query prefix on RS
	bird, err := self.client.GetJson("/routes/prefix?prefix=" + prefix)
	if err != nil {
		return api.RoutesLookupResponse{}, err
	}

	// Parse API status
	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.RoutesLookupResponse{}, err
	}

	// Parse routes
	routes, err := parseRoutes(bird, self.config)

	// Add corresponding neighbour and source rs to result
	results := api.LookupRoutes{}
	for _, src := range routes {
		// Okay. This is actually really hacky.
		// A less bruteforce approach would be highly appreciated
		route := &api.LookupRoute{
			Id: src.Id,

			Routeserver: rs,

			NeighbourId: src.NeighbourId,

			Network:   src.Network,
			Interface: src.Interface,
			Gateway:   src.Gateway,
			Metric:    src.Metric,
			Bgp:       src.Bgp,
			Age:       src.Age,
			Type:      src.Type,

			Details: src.Details,
		}
		results = append(results, route)
	}

	// Make result
	response := api.RoutesLookupResponse{
		Api:    apiStatus,
		Routes: results,
	}
	return response, nil
}

func (self *Birdwatcher) AllRoutes() (api.RoutesResponse, error) {
	bird, err := self.client.GetJson("/routes/dump")
	if err != nil {
		return api.RoutesResponse{}, err
	}
	result, err := parseRoutesDump(bird, self.config)
	return result, err
}
