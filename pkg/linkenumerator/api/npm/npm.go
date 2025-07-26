package npm

type NpmResponse struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Homepage    string                `json:"homepage"`
	Repository  Repository            `json:"repository"`
	Versions    map[string]NpmVersion `json:"versions"`
	DistTags    map[string]string     `json:"dist-tags"`
	Maintainers []Maintainer          `json:"maintainers"`
	Time        map[string]string     `json:"time"`
	License     string                `json:"license"`
}

type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Maintainer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type NpmVersion struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Homepage     string            `json:"homepage"`
	Repository   Repository        `json:"repository"`
	Dependencies map[string]string `json:"dependencies"`
	License      string            `json:"license"`
	Maintainers  []Maintainer      `json:"maintainers"`
}
