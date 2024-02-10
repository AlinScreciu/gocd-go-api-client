package client

import "github.com/AlinScreciu/gocd-go-api-client/internal/version"

func (c *Client) GetVersion() (*version.Version, error) {
	return version.GetVersion(c.client)
}
