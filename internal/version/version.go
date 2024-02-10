package version

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/client"
	"github.com/AlinScreciu/gocd-go-api-client/internal/constants"
)

type Version struct {
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Doc struct {
			Href string `json:"href"`
		} `json:"doc"`
	} `json:"_links"`
	Version     string `json:"version"`
	BuildNumber string `json:"build_number"`
	GitSha      string `json:"git_sha"`
	FullVersion string `json:"full_version"`
	CommitURL   string `json:"commit_url"`
}

const (
	endpoint = "/api/version"
)

func GetVersion(c *client.Client) (*Version, error) {
	return client.Get[Version](c, endpoint, constants.AcceptV1, "version")
}
