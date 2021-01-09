package birdwatcherV2

import (
	"fmt"
	"github.com/alice-lg/alice-lg/backend/api"
	"github.com/alice-lg/alice-lg/backend/sources/birdwatcher"
	"log"
	"sort"
	"strings"
)

type MultiTableBirdwatcher struct {
	GenericBirdwatcher
}

func (mt MultiTableBirdwatcher) Status() (*api.StatusResponse, error) {
	// Query birdwatcher
	bird, err := mt.client.GetStatus()
	if err != nil {
		return nil, err
	}

	// Use api status from first request
	apiStatus, err := parseApiStatus(bird.API, bird.TTL, mt.config)
	if err != nil {
		return nil, err
	}

	// Parse the status
	birdStatus, err := parseBirdwatcherStatus(bird, mt.config)
	if err != nil {
		return nil, err
	}

	response := &api.StatusResponse{
		Api:    apiStatus,
		Status: birdStatus,
	}

	return response, nil
}

func (mt MultiTableBirdwatcher) ExpireCaches() int {
	count := mt.routesRequiredCache.Expire()
	count += mt.routesNotExportedCache.Expire()
	return count
}

func (mt MultiTableBirdwatcher) Neighbours() (*api.NeighboursResponse, error) {
	// Check if we hit the cache
	response := mt.neighborsCache.Get()
	if response != nil {
		return response, nil
	}

	// Query birdwatcher
	birdProtocols, err := mt.client.GetProtocols()
	if err != nil {
		return nil, err
	}

	apiStatus, err := parseApiStatus(birdProtocols.API, birdProtocols.TTL, mt.config)
	if err != nil {
		return nil, err
	}
	// Parse the neighbors
	neighbours, err := parseNeighbours(filterProtocolsBgp(birdProtocols), mt.config)
	if err != nil {
		return nil, err
	}

	pipes := filterProtocolsPipe(birdProtocols)
	tree := parseProtocolToTableTree(birdProtocols.Protocols)

	// Now determine the session count for each neighbor and check if the pipe
	// did filter anything
	filtered := make(map[string]int)
	for table, _ := range tree {
		allRoutesImported := 0
		pipeRoutesImported := 0

		// Sum up all routes from all peers for a table
		for _, protocol := range tree[table] {
			// Skip peers that are not up (start/down)
			if !isProtocolUp(protocol.State) {
				continue
			}
			allRoutesImported += protocol.Routes.Imported

			pipeName := mt.getMasterPipeName(table)

			if pipe, ok := pipes[pipeName]; ok {
				pipeRoutesImported = pipe.Routes.Imported
			} else {
				continue
			}
		}

		// If no routes were imported, there is nothing left to filter
		if allRoutesImported == 0 {
			continue
		}

		// If the pipe did not filter anything, there is nothing left to do
		if pipeRoutesImported == allRoutesImported {
			continue
		}

		if len(tree[table]) == 1 {
			// Single router
			for _, protocol := range tree[table] {
				filtered[protocol.Protocol] = allRoutesImported - pipeRoutesImported
			}
		} else {
			// Multiple routers
			if pipeRoutesImported == 0 {
				// 0 is a special condition, which means that the pipe did filter ALL routes of
				// all peers. Therefore we already know the amount of filtered routes and don't have
				// to query birdwatcher again.
				for _, protocol := range tree[table] {
					// Skip peers that are not up (start/down)
					if !isProtocolUp(protocol.State) {
						continue
					}
					filtered[protocol.Protocol] = protocol.Routes.Imported
				}
			} else {
				// Otherwise the pipe did import at least some routes which means that
				// we have to query birdwatcher to get the count for each peer.
				for neighborAddress, protocol := range tree[table] {
					table := protocol.Table
					pipe := mt.getMasterPipeName(table)

					count, err := mt.client.GetCount("/routes/pipe/filtered/count?table=" + table + "&pipe=" + pipe + "&address=" + neighborAddress)
					if err != nil {
						log.Println("WARNING Could not retrieve filtered routes count:", err)
						log.Println("Is the 'pipe_filtered_count' module active in birdwatcher?")
						return nil, err
					}

					if _, ok := count["routes"]; ok {
						filtered[protocol.Protocol] = int(count["routes"].(float64))
					}
				}
			}
		}
	}

	// Update the results with the information about filtered routes from the pipe
	for _, neighbor := range neighbours {
		if pipeRoutesFiltered, ok := filtered[neighbor.Id]; ok {
			neighbor.RoutesAccepted -= pipeRoutesFiltered
			neighbor.RoutesFiltered += pipeRoutesFiltered
		}
	}

	response = &api.NeighboursResponse{
		Api:        apiStatus,
		Neighbours: neighbours,
	}

	// Cache result
	mt.neighborsCache.Set(response)

	return response, nil // dereference for now
}

func (mt MultiTableBirdwatcher) NeighboursStatus() (*api.NeighboursStatusResponse, error) {
	birdProtocols, err := mt.client.GetProtocolsShort()
	if err != nil {
		return nil, err
	}
	apiStatus, err := parseApiStatus(birdProtocols.API, birdProtocols.TTL, mt.config)
	if err != nil {
		return nil, err
	}

	// Parse the neighbors short
	neighbours, err := parseNeighboursShort(birdProtocols, mt.config)
	if err != nil {
		return nil, err
	}

	response := &api.NeighboursStatusResponse{
		Api:        apiStatus,
		Neighbours: neighbours,
	}

	return response, nil // dereference for now
}

func (mt MultiTableBirdwatcher) Routes(neighbourId string) (*api.RoutesResponse, error) {
	response := &api.RoutesResponse{}
	// Fetch required routes first (received and filtered)
	// However: Store in separate cache for faster access
	required, err := mt.fetchRequiredRoutes(neighbourId)
	if err != nil {
		return nil, err
	}

	// Optional: NoExport
	_, notExported, err := mt.fetchNotExportedRoutes(neighbourId)
	if err != nil {
		return nil, err
	}

	response.Api = required.Api
	response.Imported = required.Imported
	response.Filtered = required.Filtered
	response.NotExported = notExported

	return response, nil
}

func (mt MultiTableBirdwatcher) RoutesReceived(neighbourId string) (*api.RoutesResponse, error) {
	response := &api.RoutesResponse{}

	// Check if we have a cache hit
	cachedRoutes := mt.routesRequiredCache.Get(neighbourId)
	if cachedRoutes != nil {
		response.Api = cachedRoutes.Api
		response.Imported = cachedRoutes.Imported
		return response, nil
	}

	// Fetch required routes first (received and filtered)
	routes, err := mt.fetchRequiredRoutes(neighbourId)
	if err != nil {
		return nil, err
	}

	response.Api = routes.Api
	response.Imported = routes.Imported

	return response, nil
}

func (mt MultiTableBirdwatcher) RoutesFiltered(neighbourId string) (*api.RoutesResponse, error) {
	response := &api.RoutesResponse{}

	// Check if we have a cache hit
	cachedRoutes := mt.routesRequiredCache.Get(neighbourId)
	if cachedRoutes != nil {
		response.Api = cachedRoutes.Api
		response.Filtered = cachedRoutes.Filtered
		return response, nil
	}

	// Fetch required routes first (received and filtered)
	routes, err := mt.fetchRequiredRoutes(neighbourId)
	if err != nil {
		return nil, err
	}

	response.Api = routes.Api
	response.Filtered = routes.Filtered

	return response, nil
}

func (mt MultiTableBirdwatcher) RoutesNotExported(neighbourId string) (*api.RoutesResponse, error) {
	// Check if we have a cache hit
	response := mt.routesNotExportedCache.Get(neighbourId)
	if response != nil {
		return response, nil
	}

	// Fetch not exported routes
	apiStatus, routes, err := mt.fetchNotExportedRoutes(neighbourId)
	if err != nil {
		return nil, err
	}

	response = &api.RoutesResponse{
		Api:         *apiStatus,
		NotExported: routes,
	}

	// Cache result
	mt.routesNotExportedCache.Set(neighbourId, response)

	return response, nil
}

func (mt *MultiTableBirdwatcher) AllRoutes() (*api.RoutesResponse, error) {
	response := &api.RoutesResponse{}

	// Query birdwatcher
	birdProtocols, err := mt.client.GetProtocols()
	if err != nil {
		return nil, err
	}
	// Iterate over all the protocols and fetch the filtered routes for everyone
	protocolsBgp := filterProtocolsBgp(birdProtocols)
	for protocolId, protocolsData := range protocolsBgp {
		peer := protocolsData.NeighborAddress
		learntFrom := peer

		// Fetch filtered routes
		filtered, err := mt.fetchFilteredRoutes(protocolId, protocolsData.Table)
		if err != nil {
			continue
		}

		// Perform route deduplication
		filtered = mt.filterRoutesByPeerOrLearntFrom(filtered, peer, learntFrom)
		response.Filtered = append(response.Filtered, filtered...)
	}

	// Fetch received routes first
	birdImported, err := mt.client.GetRoutes("/routes/table/master")
	if err != nil {
		return nil, err
	}

	// Use api status from first request
	apiStatus, err := parseApiStatus(birdImported.API, birdImported.Ttl, mt.config)
	if err != nil {
		return nil, err
	}

	response.Api = apiStatus

	// Parse the routes
	imported := parseRoutesData(birdImported.Routes, mt.config)
	// Sort routes for deterministic ordering
	sort.Sort(imported)
	response.Imported = imported

	return response, nil
}

func filterProtocolsBgp(bird *ProtocolResponse) map[string]Protocol {
	return filterProtocols(bird.Protocols, "BGP")
}

func filterProtocolsPipe(bird *ProtocolResponse) map[string]Protocol {
	return filterProtocols(bird.Protocols, "Pipe")
}

func filterProtocols(protocols map[string]Protocol, protocol string) map[string]Protocol {
	response := make(map[string]Protocol, len(protocols))
	for protocolId, protocolData := range protocols {
		if protocolData.BirdProtocol == protocol {
			response[protocolId] = protocolData
		}
	}
	return response
}

func (mt *MultiTableBirdwatcher) fetchFilteredRoutes(neighborId string, table string) (api.Routes, error) {
	// Stage 1 filters
	birdFiltered, err := mt.client.GetRoutes("/routes/filtered/" + neighborId)
	if err != nil {
		log.Println("WARNING Could not retrieve filtered routes:", err)
		log.Println("Is the 'routes_filtered' module active in birdwatcher?")
		return nil, err
	}

	// Parse the routes
	filtered := parseRoutesData(birdFiltered.Routes, mt.config)

	// Stage 2 filters
	pipeName := mt.getMasterPipeName(table)

	// If there is no pipe to master, there is nothing left to do
	if pipeName == "" {
		return filtered, nil
	}

	// Query birdwatcher
	birdPipeFiltered, err := mt.client.GetRoutes("/routes/pipe/filtered/?table=" + table + "&pipe=" + pipeName)
	if err != nil {
		log.Println("WARNING Could not retrieve filtered routes:", err)
		log.Println("Is the 'pipe_filtered' module active in birdwatcher?")
		return nil, err
	}

	// Parse the routes
	pipeFiltered := parseRoutesData(birdPipeFiltered.Routes, mt.config)

	// Sort routes for deterministic ordering
	filtered = append(filtered, pipeFiltered...)
	sort.Sort(filtered)

	return filtered, nil
}

func (mt *MultiTableBirdwatcher) getMasterPipeName(table string) string {
	ptPrefix := mt.config.PeerTablePrefix
	if strings.HasPrefix(table, ptPrefix) {
		return mt.config.PipeProtocolPrefix + table[len(ptPrefix):]
	} else {
		return ""
	}
}

// Parse neighbours response
func parseNeighbours(protocols map[string]Protocol, config birdwatcher.Config) (api.Neighbours, error) {
	rsId := config.Id
	neighbours := api.Neighbours{}

	// Iterate over protocols map:
	for protocolId, proto := range protocols {
		routes := proto.Routes

		uptime := parseRelativeServerTime(proto.StateChanged, config)
		lastError := mustString(proto.LastError, "")

		routesReceived := routes.Imported + routes.Filtered

		neighbour := &api.Neighbour{
			Id: protocolId,

			Address:     mustString(proto.NeighborAddress, "error"),
			Asn:         proto.NeighborAs,
			State:       strings.ToLower(mustString(proto.State, "unknown")),
			Description: mustString(proto.Description, "no description"),

			RoutesReceived:  routesReceived,
			RoutesAccepted:  routes.Imported,
			RoutesFiltered:  routes.Filtered,
			RoutesExported:  routes.Exported, //TODO protocol_exported?
			RoutesPreferred: routes.Preferred,

			Uptime:    uptime,
			LastError: lastError,

			RouteServerId: rsId,
		}
		neighbours = append(neighbours, neighbour)
	}
	sort.Sort(neighbours)
	return neighbours, nil
}

func parseProtocolToTableTree(protocols map[string]Protocol) map[string]map[string]Protocol {
	response := make(map[string]map[string]Protocol)
	for _, protocol := range protocols {
		if protocol.BirdProtocol == "BGP" {
			table := protocol.Table
			neighborAddress := protocol.NeighborAddress
			if _, ok := response[table]; !ok {
				response[table] = make(map[string]Protocol)
			}
			if _, ok := response[table][neighborAddress]; !ok {
				response[table][neighborAddress] = Protocol{}
			}
			response[table][neighborAddress] = protocol
		}
	}
	return response
}

/*
RoutesRequired is a specialized request to fetch:

- RoutesExported and
- RoutesFiltered

from Birdwatcher. As the not exported routes can be very many
these are optional and can be loaded on demand using the
RoutesNotExported() API.

A route deduplication is applied.
*/
func (mt *MultiTableBirdwatcher) fetchRequiredRoutes(neighborId string) (*api.RoutesResponse, error) {
	// Allow only one concurrent request for this neighbor
	// to our backend server.
	mt.routesFetchMutex.Lock(neighborId)
	defer mt.routesFetchMutex.Unlock(neighborId)

	// Check if we have a cache hit
	response := mt.routesRequiredCache.Get(neighborId)
	if response != nil {
		return response, nil
	}

	// Query birdwatcher
	birdProtocols, err := mt.client.GetProtocols()
	if err != nil {
		return nil, err
	}

	protocols := birdProtocols.Protocols

	if _, ok := protocols[neighborId]; !ok {
		return nil, fmt.Errorf("Invalid Neighbor")
	}

	protocol := protocols[neighborId]
	peer := protocol.NeighborAddress

	// First: get routes received
	apiStatus, receivedRoutes, err := mt.fetchReceivedRoutes(peer)
	if err != nil {
		return nil, err
	}

	// Second: get routes filtered
	filteredRoutes, err := mt.fetchFilteredRoutes(neighborId, protocol.Table)
	if err != nil {
		return nil, err
	}

	// Perform route deduplication
	importedRoutes := api.Routes{}
	if len(receivedRoutes) > 0 {
		peer := receivedRoutes[0].Gateway
		learntFrom := mustString(receivedRoutes[0].LearntFrom, peer)

		filteredRoutes = mt.filterRoutesByPeerOrLearntFrom(filteredRoutes, peer, learntFrom)
		importedRoutes = mt.filterRoutesByDuplicates(receivedRoutes, filteredRoutes)
	}

	response = &api.RoutesResponse{
		Api:      *apiStatus,
		Imported: importedRoutes,
		Filtered: filteredRoutes,
	}

	// Cache result
	mt.routesRequiredCache.Set(neighborId, response)

	return response, nil
}

func (mt *MultiTableBirdwatcher) fetchReceivedRoutes(peer string) (*api.ApiStatus, api.Routes, error) {

	// Query birdwatcher
	bird, err := mt.client.GetRoutes("/routes/peer/" + peer)
	if err != nil {
		return nil, nil, err
	}

	// Use api status from first request
	apiStatus, err := parseApiStatus(bird.API, bird.Ttl, mt.config)
	if err != nil {
		return nil, nil, err
	}

	// Parse the routes
	received, err := parseRoutes(bird.Routes, mt.config)
	if err != nil {
		log.Println("WARNING Could not retrieve received routes:", err)
		log.Println("Is the 'routes_peer' module active in birdwatcher?")
		return &apiStatus, nil, err
	}

	return &apiStatus, received, nil
}

// Parse routes response
func parseRoutes(bird []Route, config birdwatcher.Config) (api.Routes, error) {

	routes := parseRoutesData(bird, config)

	// Sort routes
	sort.Sort(routes)
	return routes, nil
}

func (mt *MultiTableBirdwatcher) filterRoutesByDuplicates(routes api.Routes, filterRoutes api.Routes) api.Routes {
	result_routes := make(api.Routes, 0, len(routes))

	routesMap := make(map[string]*api.Route) // for O(1) access
	for _, route := range routes {
		routesMap[route.Id] = route
	}

	// Remove routes from "routes" that are contained within filterRoutes
	for _, filterRoute := range filterRoutes {
		if _, ok := routesMap[filterRoute.Id]; ok {
			delete(routesMap, filterRoute.Id)
		}
	}

	for _, route := range routesMap {
		result_routes = append(result_routes, route)
	}

	// Sort routes for deterministic ordering
	sort.Sort(result_routes)
	routes = result_routes

	return routes
}

func (mt *MultiTableBirdwatcher) fetchNotExportedRoutes(neighborId string) (*api.ApiStatus, api.Routes, error) {
	// Query birdwatcher
	birdProtocols, err := mt.client.GetProtocols()
	if err != nil {
		return nil, nil, err
	}

	protocols := birdProtocols.Protocols

	if _, ok := protocols[neighborId]; !ok {
		return nil, nil, fmt.Errorf("Invalid Neighbor")
	}

	table := protocols[neighborId].Table
	pipeName := mt.getMasterPipeName(table)

	// Query birdwatcher
	bird, err := mt.client.GetRoutes("/routes/noexport/" + pipeName)

	// Use api status from first request
	apiStatus, err := parseApiStatus(bird.API, bird.Ttl, mt.config)
	if err != nil {
		return nil, nil, err
	}

	notExported, err := parseRoutes(bird.Routes, mt.config)
	if err != nil {
		log.Println("WARNING Could not retrieve routes not exported:", err)
		log.Println("Is the 'routes_noexport' module active in birdwatcher?")
	}

	return &apiStatus, notExported, nil
}

// Parse neighbours response
func parseNeighboursShort(bird *ProtocolShortResponse, config birdwatcher.Config) (api.NeighboursStatus, error) {
	neighbours := api.NeighboursStatus{}
	protocols := bird.Protocols

	// Iterate over protocols map:
	for protocolId, protocol := range protocols {

		uptime := parseRelativeServerTime(protocol.Since, config)

		neighbour := &api.NeighbourStatus{
			Id:    protocolId,
			State: protocol.State,
			Since: uptime,
		}

		neighbours = append(neighbours, neighbour)
	}

	sort.Sort(neighbours)

	return neighbours, nil
}