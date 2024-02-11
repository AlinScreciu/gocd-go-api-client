package client

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/version"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
)

func (c *Client) GetVersion() (*types.Version, error) {
	return version.GetVersion(c.client)
}
