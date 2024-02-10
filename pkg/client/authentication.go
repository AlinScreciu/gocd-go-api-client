package client

import "github.com/AlinScreciu/gocd-go-api-client/internal/authentication"

func (c *Client) GetCurrentUser() (*authentication.CurrentUser, error) {
	return authentication.GetCurrentUser(c.client)
}
