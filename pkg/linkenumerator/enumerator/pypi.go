package enumerator

import (
	"fmt"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/pypi"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	"github.com/sirupsen/logrus"
)

type pypiEnumerator struct {
	enumeratorBase
	take int
}

func NewPypiEnumerator(take int) Enumerator {
	return &pypiEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func (c *pypiEnumerator) getPypiIndex() ([]pypi.IndexItem, error) {
	r := c.client.R().
		SetURL(api.PYPI_INDEX_API_URL).
		SetHeader("Accept", "application/vnd.pypi.simple.v1+json")
	resp := r.Do()
	if resp.IsErrorState() {
		logger.Errorf("Pypi index fetch failed: %v", resp)
		return nil, resp.Err
	}
	indexResp, err := api.FromPypiIndex(resp.Bytes())
	if err != nil {
		logger.Errorf("Pypi index unmarshal failed: %v", err)
		return nil, err
	}
	return indexResp.Projects, nil
}

func (c *pypiEnumerator) getPypiPackageInfo(name string) (*pypi.PackageResp, error) {
	r := c.client.R().SetURL(fmt.Sprintf("%s/%s/json", api.PYPI_ENUMERAE_API_URL, name))

	resp := r.Do()
	if resp.IsErrorState() {
		logger.Errorf("Pypi package %s fetch failed: %v", name, resp)
		return nil, resp.Err
	}

	packageResp, err := api.FromPypiPackage(resp.Bytes())
	if err != nil {
		logger.Errorf("Pypi package %s unmarshal failed: %v", name, err)
		return nil, err
	}
	return packageResp, nil
}

func GetBestPypiUrl(info *pypi.Info) string {
	if info.HomePage != "" {
		return info.HomePage
	}
	if info.ProjectUrls.Source != "" {
		return info.ProjectUrls.Source
	}
	return ""
}

func (c *pypiEnumerator) Enumerate() error {
	logger.Info("Fetching pypi index")
	projects, err := c.getPypiIndex()
	if err != nil {
		logger.Errorf("Pypi index fetch failed: %v", err)
		return err
	}

	if err := c.writer.Open(); err != nil {
		return err
	}
	defer c.writer.Close()

	var wg sync.WaitGroup
	maxConcurrency := 10
	sem := make(chan struct{}, maxConcurrency)
	collected := 0

	for _, project := range projects {
		if c.take > 0 && collected >= c.take {
			break
		}
		collected++
		sem <- struct{}{}
		wg.Add(1)
		projName := project.Name
		go func(name string) {
			defer wg.Done()
			defer func() { <-sem }()
			pkg, err := c.getPypiPackageInfo(name)
			if pkg == nil || err != nil {
				logger.Errorf("Pypi package %s fetch failed: %v", name, err)
				return
			}

			u := GetBestPypiUrl(&pkg.Info)

			c.writer.Write(pkg.Info.Name)
			c.writer.Write(u)
			c.writer.Write(pkg.Info.Version)
			c.writer.Write(fmt.Sprintf("%v", pkg.Info.RequiresDist))
			c.writer.Write("\n")
		}(projName)
	}
	wg.Wait()
	logrus.Infof("Enumerator has collected and written %d packages", collected)
	return nil
}
