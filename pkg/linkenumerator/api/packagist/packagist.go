package packagist

type ListResponse struct {
	PackageNames []string `json:"packageNames"`
}

type Maintainer struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Homepage string `json:"homepage"`
	Role     string `json:"role"`
}

type Source struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	Reference string `json:"reference"`
}

type Dist struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	Shasum    string `json:"shasum"`
	Reference string `json:"reference"`
}

type Support struct {
	Issues string `json:"issues"`
	Source string `json:"source"`
}

type Funding struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type Version struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Homepage          string            `json:"homepage"`
	Version           string            `json:"version"`
	VersionNormalized string            `json:"version_normalized"`
	License           []string          `json:"license"`
	Authors           []Author          `json:"authors"`
	Source            Source            `json:"source"`
	Dist              Dist              `json:"dist"`
	Type              string            `json:"type"`
	Support           Support           `json:"support"`
	Funding           []Funding         `json:"funding"`
	Time              string            `json:"time"`
	Require           map[string]string `json:"require"`
}

type Package struct {
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Time             string             `json:"time"`
	Maintainers      []Maintainer       `json:"maintainers"`
	Versions         map[string]Version `json:"versions"`
	Type             string             `json:"type"`
	Repository       string             `json:"repository"`
	GithubStars      int                `json:"github_stars"`
	GithubWatchers   int                `json:"github_watchers"`
	GithubForks      int                `json:"github_forks"`
	GithubOpenIssues int                `json:"github_open_issues"`
	Language         string             `json:"language"`
	Dependents       int                `json:"dependents"`
	Suggesters       int                `json:"suggesters"`
	Downloads        struct {
		Total   int `json:"total"`
		Monthly int `json:"monthly"`
		Daily   int `json:"daily"`
	} `json:"downloads"`
	Favers int `json:"favers"`
}

type PackageResponse struct {
	Package Package `json:"package"`
}
