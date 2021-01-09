package birdwatcherV2

import (
	"github.com/alice-lg/alice-lg/backend/api"
	"github.com/alice-lg/alice-lg/backend/sources/birdwatcher"
	"log"
	"strconv"
	"time"
)

// Parse partial routes response
func parseRoutesData(birdRoutes []Route, config birdwatcher.Config) api.Routes {
	routes := api.Routes{}

	for _, data := range birdRoutes {
		age := parseRelativeServerTime(data.Age, config)
		rtype := data.Type
		bgpInfo := parseRouteBgpInfo(data.Bgp)

		route := &api.Route{
			Id:          mustString(data.Network, "unknown"),
			NeighbourId: mustString(data.FromProtocol, "unknown neighbour"),

			Network:    mustString(data.Network, "unknown net"),
			Interface:  mustString(data.Interface, "unknown interface"),
			Gateway:    mustString(data.Gateway, "unknown gateway"),
			Metric:     data.Metric,
			Primary:    data.Primary,
			LearntFrom: data.LearntFrom,
			Age:        age,
			Type:       rtype,
			Bgp:        bgpInfo,
		}

		routes = append(routes, route)
	}
	return routes
}

// Parse neighbour uptime
func parseRelativeServerTime(uptime interface{}, config birdwatcher.Config) time.Duration {
	serverTime, _ := parseServerTime(uptime, config.ServerTimeShort, config.Timezone)
	return time.Since(serverTime)
}

// Convert server time string to time
func parseServerTime(value interface{}, layout, timezone string) (time.Time, error) {
	svalue, ok := value.(string)
	if !ok {
		return time.Time{}, nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.ParseInLocation(layout, svalue, loc)
	if err != nil {
		return time.Time{}, err
	}

	return t.UTC(), nil
}

// Parse route bgp info
func parseRouteBgpInfo(data Bgp) api.BgpInfo {

	asPath := mustIntList(data.AsPath)
	communities := data.Communities
	largeCommunities := data.LargeCommunities
	extCommunities := parseExtBgpCommunities(data.ExtCommunities)

	localPref, _ := strconv.Atoi(mustString(data.LocalPref, "0"))
	med, _ := strconv.Atoi(mustString(data.Med, "0"))

	bgp := api.BgpInfo{
		Origin:           mustString(data.Origin, "unknown"),
		AsPath:           asPath,
		NextHop:          mustString(data.NextHop, "unknown"),
		LocalPref:        localPref,
		Med:              med,
		Communities:      communities,
		ExtCommunities:   extCommunities,
		LargeCommunities: largeCommunities,
	}
	return bgp
}

// Extract extended communtieis
func parseExtBgpCommunities(data []api.ExtCommunity) []api.ExtCommunity {
	var communities []api.ExtCommunity
	for _, cdata := range data {
		if len(cdata) != 3 {
			log.Println("Ignoring malformed ext community:", cdata)
			continue
		}
		communities = append(communities, api.ExtCommunity{
			cdata[0],
			cdata[1],
			cdata[2],
		})
	}

	return communities
}

// Make api status from response:
// The api status is always included in a birdwatcher response
func parseApiStatus(birdApi API, responseTtl time.Time, config birdwatcher.Config) (api.ApiStatus, error) {
	// Parse TTL
	ttl, err := parseServerTime(
		responseTtl,
		config.ServerTime,
		config.Timezone,
	)
	if err != nil {
		return api.ApiStatus{}, err
	}

	// Parse Cache Status
	cacheStatus, _ := parseCacheStatus(birdApi.CacheStatus, config)

	status := api.ApiStatus{
		Version:         birdApi.Version,
		ResultFromCache: birdApi.ResultFromCache,
		Ttl:             ttl,
		CacheStatus:     cacheStatus,
	}

	return status, nil
}

// Parse cache status from api response
func parseCacheStatus(cacheStatus CacheStatus, config birdwatcher.Config) (api.CacheStatus, error) {
	cachedAtTime, err := parseServerTime(cacheStatus.CachedAt.Date, config.ServerTime, config.Timezone)
	if err != nil {
		return api.CacheStatus{}, err
	}

	status := api.CacheStatus{
		CachedAt: cachedAtTime,
	}

	return status, nil
}

// Parse birdwatcher status
func parseBirdwatcherStatus(bird *StatusResponse, config birdwatcher.Config) (api.Status, error) {

	// Get special fields
	serverTime, _ := parseServerTime(
		bird.Status.CurrentServer,
		config.ServerTimeShort,
		config.Timezone,
	)

	lastReboot, _ := parseServerTime(
		bird.Status.LastReboot,
		config.ServerTimeShort,
		config.Timezone,
	)

	if config.ShowLastReboot == false {
		lastReboot = time.Time{}
	}

	lastReconfig, _ := parseServerTime(
		bird.Status.LastReconfig,
		config.ServerTimeExt,
		config.Timezone,
	)

	// Make status response
	status := api.Status{
		ServerTime:   serverTime,
		LastReboot:   lastReboot,
		LastReconfig: lastReconfig,
		Backend:      "bird",
		Version:      mustString(bird.Status.Version, "unknown"),
		Message:      mustString(bird.Status.Message, "unknown"),
		RouterId:     mustString(bird.Status.RouterID, "unknown"),
	}

	return status, nil
}
