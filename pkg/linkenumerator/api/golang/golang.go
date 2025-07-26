package GO

type GoModule struct {
	Path         string   `json:"path"`
	Version      string   `json:"version"`
	Time         string   `json:"time"`
	Description  string   `json:"description"`
	Homepage     string   `json:"homepage"`
	Repository   string   `json:"repository"`
	License      string   `json:"license"`
	Dependencies []string `json:"dependencies"`
}
