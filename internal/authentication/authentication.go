package authentication

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/client"
	"github.com/AlinScreciu/gocd-go-api-client/internal/constants"
)

type CurrentUser struct {
	Links struct {
		Doc struct {
			Href string `json:"href,omitempty"`
		} `json:"doc,omitempty"`
		CurrentUser struct {
			Href string `json:"href,omitempty"`
		} `json:"current_user,omitempty"`
		Self struct {
			Href string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Find struct {
			Href string `json:"href,omitempty"`
		} `json:"find,omitempty"`
	} `json:"_links,omitempty"`
	LoginName      string `json:"login_name,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	Enabled        bool   `json:"enabled,omitempty"`
	Email          string `json:"email,omitempty"`
	EmailMe        bool   `json:"email_me,omitempty"`
	CheckinAliases []any  `json:"checkin_aliases,omitempty"`
}

const (
	endpoint = "/api/current_user"
)

func GetCurrentUser(c *client.Client) (*CurrentUser, error) {
	return client.Get[CurrentUser](c, endpoint, constants.AcceptV1, "authentication")
}
