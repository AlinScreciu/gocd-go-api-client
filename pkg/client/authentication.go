package client

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/authentication"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
)

func (c *Client) GetCurrentUser() (*types.CurrentUser, error) {
	return authentication.GetCurrentUser(c.client)
}
