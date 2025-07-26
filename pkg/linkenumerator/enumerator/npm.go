package enumerator

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/npm"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
)

type NpmEnumerator struct {
	enumeratorBase
	take int
}

func NewNpmEnumerator(take int) *NpmEnumerator {
	return &NpmEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func (c *NpmEnumerator) Enumerate() error {
	if err := c.writer.Open(); err != nil {
		logger.Error("Open writer: ", err)
		return err
	}
	defer c.writer.Close()

	req := c.client.R()
	req.SetURL(api.NPM_INDEX_API_URL)
	logger.Info("Downloading npm package name list...")
	resp := req.Do()
	if resp.IsErrorState() {
		logger.Error("NPM fetch error: ", resp.Err)
		return resp.Err
	}

	var data map[string]*string
	if err := json.Unmarshal(resp.Bytes(), &data); err != nil {
		logger.Error("Failed to parse npm data: ", err)
		return err
	}

	names := make([]string, 0, len(data))
	for name := range data {
		names = append(names, name)
	}
	if c.take > 0 && c.take < len(names) {
		names = names[:c.take]
	}

	maxConcurrency := 10
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, name := range names {
		sem <- struct{}{}
		wg.Add(1)
		go func(pkgName string) {
			defer wg.Done()
			defer func() { <-sem }()
			res, err := c.fetch(api.NPM_ENUMERATE_API_URL + pkgName)
			if err != nil {
				logger.Error("Fetch npm package failed: ", pkgName, err)
				return
			}
			npmResp, err := api.FromNpm(res)
			if err != nil {
				logger.Error("Parse npm package failed: ", pkgName, err)
				return
			}
			latest := npmResp.DistTags["latest"]
			versionInfo, ok := npmResp.Versions[latest]
			if !ok {
				versionInfo = npm.NpmVersion{}
			}
			c.writer.Write(npmResp.Name)
			c.writer.Write(versionInfo.Homepage)
			c.writer.Write(versionInfo.Repository.URL)
			deps := make([]string, 0, len(versionInfo.Dependencies))
			for dep := range versionInfo.Dependencies {
				deps = append(deps, dep)
			}
			c.writer.Write(strings.Join(deps, ", "))
			c.writer.Write("\n")
		}(name)
	}
	wg.Wait()
	return nil
}
