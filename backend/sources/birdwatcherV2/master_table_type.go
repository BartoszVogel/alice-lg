package birdwatcherV2

import (
	"fmt"
	"github.com/alice-lg/alice-lg/backend/api"
	"net/http"
	"time"
)

type MasterTable struct {
	API API `json:"api"`
	CachedAt time.Time `json:"cached_at"`
	Routes   []Route `json:"routes"`
	Ttl time.Time `json:"ttl"`
}

type Route struct {
	Age string `json:"age"`
	Bgp Bgp `json:"bgp"`
	FromProtocol string   `json:"from_protocol"`
	Gateway      string   `json:"gateway"`
	Interface    string   `json:"interface"`
	LearntFrom   string   `json:"learnt_from"`
	Metric       int      `json:"metric"`
	Network      string   `json:"network"`
	Primary      bool     `json:"primary"`
	Type         []string `json:"type"`
}

type Bgp struct {
	AsPath           []string `json:"as_path"`
	Communities      []api.Community  `json:"communities"`
	LargeCommunities []api.Community  `json:"large_communities"`
	ExtCommunities   []api.ExtCommunity `json:"ext_communities"`
	LocalPref        string   `json:"local_pref"`
	Med              string   `json:"med"`
	NextHop          string   `json:"next_hop"`
	Origin           string   `json:"origin"`
}

func (c *Client) GetRoutes(path string)(*MasterTable, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.baseURL, path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res := MasterTable{}
	if err := c.Get(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
