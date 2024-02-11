package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/AlinScreciu/gocd-go-api-client/internal/client"
	"github.com/AlinScreciu/gocd-go-api-client/internal/logging"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
)

var logger *logging.Logger

func init() {
	logger = logging.NewLoggerWithModule("client")
}

type GoCDClient interface {
	GetVersion() (*types.Version, error)
}

type Client struct {
	client *client.Client
}

func (c *Client) SetDebug() {
	c.client.Debug = true
}

func (c *Client) SetBasicAuth(user, password string) {
	c.client.SetBasicAuth(user, password)
}

func (c *Client) SetAccessToken(token string) {
	c.client.SetAccessToken(token)
}

func NewClient(ctx context.Context, serverUrl string) (*Client, error) {
	url, err := url.Parse(serverUrl)
	if err != nil {
		logger.Errorf("failed to parse '%s' to url: '%s'", url, err)

		return nil, fmt.Errorf("failed to parse '%s' to url: '%w'", serverUrl, err)
	}

	return &Client{
		client: client.NewClient(ctx, url),
	}, nil
}
