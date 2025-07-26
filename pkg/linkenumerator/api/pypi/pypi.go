package pypi

type IndexResp struct {
	Projects []IndexItem `json:"projects"`
}

type IndexItem struct {
	LastSerial int    `json:"_last-serial"`
	Name       string `json:"name"`
}

type PackageResp struct {
	Info            Info             `json:"info"`
	LastSerial      int64            `json:"last_serial"`
	Releases        map[string][]URL `json:"releases"`
	Urls            []URL            `json:"urls"`
	Vulnerabilities []interface{}    `json:"vulnerabilities"`
}

type Info struct {
	Name         string      `json:"name"`
	HomePage     string      `json:"home_page"`
	ProjectUrls  ProjectUrls `json:"project_urls"`
	Version      string      `json:"version"`
	Summary      string      `json:"summary"`
	License      string      `json:"license"`
	RequiresDist []string    `json:"requires_dist"`
}

type ProjectUrls struct {
	Documentation string `json:"Documentation"`
	Homepage      string `json:"Homepage"`
	Source        string `json:"Source"`
}

type URL struct {
	URL string `json:"url"`
}
