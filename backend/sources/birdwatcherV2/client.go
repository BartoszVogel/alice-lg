package birdwatcherV2

import (
	"fmt"
	"github.com/json-iterator/go"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	HTTPClient *http.Client
}

func NewClient(baseUrl string) *Client {
	return &Client{
		baseURL: baseUrl,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

var json = jsoniter.ConfigFastest

func (c *Client) Get(req *http.Request, v interface{}) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Try to unmarshall into errorResponse
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

