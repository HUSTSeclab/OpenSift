package api

import (
	"encoding/json"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/bitbucket"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/cargo"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/gitlab"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/haskell"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/npm"
	packagist "github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/packagist"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/pypi"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/ruby"
	"github.com/imroc/req/v3"
)

const (
	PER_PAGE      = 100
	TIME_INTERVAL = 2
	TIMEOUT       = 1000

	BITBUCKET_ENUMERATE_API_URL = "https://api.bitbucket.org/2.0/repositories?pagelen=200"
	GITLAB_ENUMERATE_API_URL    = "https://gitlab.com/api/v4/projects"
	GITEE_ENUMERATE_API_URL     = "https://api.indexea.com/v1/search/widget/wjawvtmm7r5t25ms1u3d"
	CRATES_IO_ENUMERATE_API_URL = "https://crates.io/api/v1/crates"
	PACKAGIST_LIST_API_URL      = "https://packagist.org/packages/list.json"
	PACKAGIST_ENUMERATE_API_URL = "https://packagist.org/packages/"
	GOLANG_INDEX_API_URL        = "https://index.golang.org/index"
	HASKELL_INDEX_API_URL       = "https://hackage.haskell.org/packages/"
	HASKELL_ENUMERATE_API_URL   = "https://hackage.haskell.org"
	NPM_INDEX_API_URL           = "https://github.com/nice-registry/all-the-package-repos/raw/refs/heads/master/data/packages.json"
	NPM_ENUMERATE_API_URL       = "https://registry.npmjs.org/"
	NUGET_INDEX_URL             = "https://azuresearch-ea.nuget.org/query"
	PYPI_INDEX_API_URL          = "https://pypi.org/simple/"
	PYPI_ENUMERAE_API_URL       = "https://pypi.org/pypi"
	RUBY_INDEX_API_URL          = "https://rubygems.org/names"
	RUBY_ENUMERATE_API_URL      = "https://rubygems.org/api/v1/gems/"

	GITLAB_TOTAL_PAGES = 100000

	BITBUCKET_ENUMERATE_PAGE = 40 //* repo_num = page * 10
	GITLAB_ENUMERATE_PAGE    = 20 //* repo_num = page * 100
	GITEE_ENUMERATE_PAGE     = 20 //* repo_num = page * 100
	CRATES_IO_ENUMERATE_PAGE = 20
)

func FromGitlab(res *req.Response) (*gitlab.Response, error) {
	resp := &gitlab.Response{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromBitbucket(res *req.Response) (*bitbucket.Response, error) {
	resp := &bitbucket.Response{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromCargo(res *req.Response) (*cargo.Response, error) {
	resp := &cargo.Response{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromHaskell(res *req.Response) (*haskell.Response, error) {
	resp := &haskell.Response{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromPackagist(res *req.Response) (*packagist.ListResponse, error) {
	resp := &packagist.ListResponse{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromPackagistDetail(res *req.Response) (*packagist.PackageResponse, error) {
	resp := &packagist.PackageResponse{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func FromRubyNames(res *req.Response) ([]string, error) {
	names := strings.Split(string(res.Bytes()), "\n")
	return names, nil
}

func FromRubyDetail(res *req.Response) (*ruby.Response, error) {
	detail := &ruby.Response{}
	if err := json.Unmarshal(res.Bytes(), detail); err != nil {
		return nil, err
	}
	return detail, nil
}

func FromPypiIndex(data []byte) (*pypi.IndexResp, error) {
	var r pypi.IndexResp
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func FromPypiPackage(data []byte) (*pypi.PackageResp, error) {
	var r pypi.PackageResp
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func FromNpm(res *req.Response) (*npm.NpmResponse, error) {
	resp := &npm.NpmResponse{}
	if err := json.Unmarshal(res.Bytes(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}
