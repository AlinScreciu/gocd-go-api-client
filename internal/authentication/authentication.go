package authentication

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/client"
	"github.com/AlinScreciu/gocd-go-api-client/internal/constants"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
)

const (
	endpoint = "/api/current_user"
)

func GetCurrentUser(c *client.Client) (*types.CurrentUser, error) {
	return client.Get[types.CurrentUser](c, endpoint, constants.AcceptV1, "authentication")
}
