package birdwatcherV2

import (
	"fmt"
	"net/http"
	"time"
)

type ProtocolResponse struct {
	API       API                 `json:"api"`
	CachedAt  time.Time           `json:"cached_at"`
	Protocols map[string]Protocol `json:"protocols"`
	TTL       time.Time           `json:"ttl"`
}

type ProtocolShortResponse struct {
	API       API                      `json:"api"`
	CachedAt  time.Time                `json:"cached_at"`
	Protocols map[string]ProtocolShort `json:"protocols"`
	TTL       time.Time                `json:"ttl"`
}

type ExportUpdates struct {
	Accepted int `json:"accepted"`
	Filtered int `json:"filtered"`
	Received int `json:"received"`
	Rejected int `json:"rejected"`
}
type ExportWithdraws struct {
	Accepted int `json:"accepted"`
	Received int `json:"received"`
}
type ImportUpdates struct {
	Accepted int `json:"accepted"`
	Filtered int `json:"filtered"`
	Ignored  int `json:"ignored"`
	Received int `json:"received"`
	Rejected int `json:"rejected"`
}
type ImportWithdraws struct {
	Accepted int `json:"accepted"`
	Ignored  int `json:"ignored"`
	Received int `json:"received"`
	Rejected int `json:"rejected"`
}
type RouteChanges struct {
	ExportUpdates   ExportUpdates   `json:"export_updates"`
	ExportWithdraws ExportWithdraws `json:"export_withdraws"`
	ImportUpdates   ImportUpdates   `json:"import_updates"`
	ImportWithdraws ImportWithdraws `json:"import_withdraws"`
}
type Routes struct {
	Exported  int `json:"exported"`
	Filtered  int `json:"filtered"`
	Imported  int `json:"imported"`
	Preferred int `json:"preferred"`
}
type Protocol struct {
	//Action           string       `json:"action"`
	//AfAnnounced      string       `json:"af_announced"`
	//BgpNextHop       string       `json:"bgp_next_hop"`
	//BgpState         string       `json:"bgp_state"`
	BirdProtocol string `json:"bird_protocol"`
	//Connection       string       `json:"connection"`
	Description string `json:"description"`
	//HoldTimer        string       `json:"hold_timer"`
	//ImportLimit      int        `json:"import_limit"`
	//InputFilter      string       `json:"input_filter"`
	//KeepaliveTimer   string       `json:"keepalive_timer"`
	//LocalAs          int          `json:"local_as"`
	NeighborAddress string `json:"neighbor_address"`
	NeighborAs      int    `json:"neighbor_as"`
	//NeighborID       string       `json:"neighbor_id"`
	//OutputFilter     string       `json:"output_filter"`
	//Preference       int          `json:"preference"`
	Protocol string `json:"protocol"`
	//ReceiveLimit     int          `json:"receive_limit"`
	//RouteChangeStats string       `json:"route_change_stats"`
	//RouteChanges     RouteChanges `json:"route_changes"`
	Routes Routes `json:"routes"`
	//Rx               string       `json:"rx"`
	//Session          string       `json:"session"`
	//SourceAddress    string       `json:"source_address"`
	State        string `json:"state"`
	StateChanged string `json:"state_changed"`
	Table        string `json:"table"`
	//Tx               string       `json:"tx"`
	LastError string `json:"last_error"`
}

type ProtocolShort struct {
	Info  string `json:"info"`
	Proto string `json:"proto"`
	Since string `json:"since"`
	State string `json:"state"`
	Table string `json:"table"`
}

func (c *Client) GetProtocols() (*ProtocolResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/protocols", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res := ProtocolResponse{}
	if err := c.Get(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) GetCount(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.baseURL, url), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res := make(map[string]interface{})
	if err := c.Get(req, &res); err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *Client) GetProtocolsShort() (*ProtocolShortResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/protocols/short?uncached=true", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res := ProtocolShortResponse{}
	if err := c.Get(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
