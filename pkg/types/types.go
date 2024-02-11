package types

type Links struct {
	Self struct {
		Href string `json:"href"`
	} `json:"self"`
	Doc struct {
		Href string `json:"href"`
	} `json:"doc"`
}

type Properties struct {
	Key            string `json:"key"`
	Value          string `json:"value,omitempty"`
	EncryptedValue string `json:"encrypted_value,omitempty"`
}

type PackageRepo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Package struct {
	Links         Links        `json:"_links,omitempty"`
	Id            string       `json:"id"`
	Name          string       `json:"name"`
	AutoUpdate    bool         `json:"auto_update"`
	PackageRepo   PackageRepo  `json:"package_repo"`
	Configuration []Properties `json:"configuration"`
}

type AllPackages struct {
	Links    Links `json:"_links,omitempty"`
	Embedded struct {
		Packages []Package `json:"packages,omitempty"`
	} `json:"_embedded"`
}

type CurrentUser struct {
	Links          Links  `json:"_links,omitempty"`
	LoginName      string `json:"login_name,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	Enabled        bool   `json:"enabled"`
	Email          string `json:"email,omitempty"`
	EmailMe        bool   `json:"email_me"`
	CheckinAliases []any  `json:"checkin_aliases,omitempty"`
}

type Version struct {
	Links       Links  `json:"_links,omitempty"`
	Version     string `json:"version,omitempty"`
	BuildNumber string `json:"build_number,omitempty"`
	GitSha      string `json:"git_sha,omitempty"`
	FullVersion string `json:"full_version,omitempty"`
	CommitURL   string `json:"commit_url,omitempty"`
}
