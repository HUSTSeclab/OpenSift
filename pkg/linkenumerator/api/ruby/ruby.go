package ruby

type Response struct {
	Name             string                 `json:"name"`
	Downloads        int64                  `json:"downloads"`
	Version          string                 `json:"version"`
	VersionCreatedAt string                 `json:"version_created_at"`
	VersionDownloads int64                  `json:"version_downloads"`
	Platform         string                 `json:"platform"`
	Authors          string                 `json:"authors"`
	Info             string                 `json:"info"`
	Licenses         interface{}            `json:"licenses"`
	Metadata         map[string]interface{} `json:"metadata"`
	Yanked           bool                   `json:"yanked"`
	SHA              string                 `json:"sha"`
	SpecSHA          string                 `json:"spec_sha"`
	ProjectURI       string                 `json:"project_uri"`
	GemURI           string                 `json:"gem_uri"`
	HomepageURI      string                 `json:"homepage_uri"`
	WikiURI          string                 `json:"wiki_uri"`
	DocumentationURI string                 `json:"documentation_uri"`
	MailingListURI   string                 `json:"mailing_list_uri"`
	SourceCodeURI    string                 `json:"source_code_uri"`
	BugTrackerURI    string                 `json:"bug_tracker_uri"`
	ChangelogURI     string                 `json:"changelog_uri"`
	FundingURI       string                 `json:"funding_uri"`
	Dependencies     struct {
		Development []struct {
			Name         string `json:"name"`
			Requirements string `json:"requirements"`
		} `json:"development"`
		Runtime []struct {
			Name         string `json:"name"`
			Requirements string `json:"requirements"`
		} `json:"runtime"`
	} `json:"dependencies"`
}
