package haskell

type Response struct {
	Values []Value
}

type Value struct {
	Name                string `json:"name"`
	URL                 string `json:"url"`
	Version             string `json:"version"`
	Dependencies        string `json:"dependencies"`
	License             string `json:"license"`
	Copyright           string `json:"copyright"`
	Author              string `json:"author"`
	Maintainer          string `json:"maintainer"`
	Category            string `json:"category"`
	Homepage            string `json:"homepage"`
	SourceRepo          string `json:"source_repo"`
	UploadedBy          string `json:"uploaded_by"`
	Distributions       string `json:"distributions"`
	ReverseDependencies string `json:"reverse_dependencies"`
	Downloads           string `json:"downloads"`
	Rating              string `json:"rating"`
	Status              string `json:"status"`
}
