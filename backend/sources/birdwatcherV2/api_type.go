package birdwatcherV2

import "time"

type API struct {
	Version         string      `json:"Version"`
	ResultFromCache bool        `json:"result_from_cache"`
	CacheStatus     CacheStatus `json:"cache_status"`
}

type CacheStatus struct {
	CachedAt CachedAt `json:"cached_at"`
}

type CachedAt struct {
	Date         time.Time `json:"date"`
	TimezoneType string    `json:"timezone_type"`
	Timezone     string    `json:"timezone"`
}
