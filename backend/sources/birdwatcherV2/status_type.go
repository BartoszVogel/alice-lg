package birdwatcherV2

import (
	"fmt"
	"net/http"
	"time"
)

type StatusResponse struct {
	API      API       `json:"api"`
	CachedAt time.Time `json:"cached_at"`
	Status   Status    `json:"status"`
	TTL      time.Time `json:"ttl"`
}

type Status struct {
	CurrentServer string `json:"current_server"`
	LastReboot    string `json:"last_reboot"`
	LastReconfig  string `json:"last_reconfig"`
	Message       string `json:"message"`
	RouterID      string `json:"router_id"`
	Version       string `json:"version"`
}

func (c *Client) GetStatus()(*StatusResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/status", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res := StatusResponse{}
	if err := c.Get(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
