package version

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/client"
	"github.com/AlinScreciu/gocd-go-api-client/internal/constants"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
)

const (
	endpoint = "/api/version"
)

func GetVersion(c *client.Client) (*types.Version, error) {
	return client.Get[types.Version](c, endpoint, constants.AcceptV1, "version")
}
