package enumerator

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	packagist "github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/packagist"
	"github.com/sirupsen/logrus"
)

type PackagistEnumerator struct {
	enumeratorBase
	take int
}

func NewPackagistEnumerator(take int) *PackagistEnumerator {
	return &PackagistEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func GetPackagistBestUrl(ver packagist.Version, pkg packagist.Package) string {
	if ver.Homepage != "" {
		return ver.Homepage
	}
	if pkg.Repository != "" {
		return pkg.Repository
	}
	if ver.Source.URL != "" {
		return ver.Source.URL
	}
	return ""
}

func (c *PackagistEnumerator) Enumerate() error {

	if err := c.writer.Open(); err != nil {
		logrus.Panic("Open writer: ", err)
		return err
	}
	defer c.writer.Close()

	u := api.PACKAGIST_LIST_API_URL

	res, err := c.fetch(u)
	if err != nil {
		logrus.Panic("Fetch packagist list: ", err)
		return err
	}

	listResp, err := api.FromPackagist(res)
	if err != nil {
		logrus.Panic("Parse packagist list: ", err)
		return err
	}

	maxConcurrency := 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	collected := 0

	for _, fullname := range listResp.PackageNames {
		if c.take > 0 && collected >= c.take {
			break
		}
		parts := strings.SplitN(fullname, "/", 2)
		if len(parts) != 2 {
			continue
		}
		vendor, name := parts[0], parts[1]
		sem <- struct{}{}
		wg.Add(1)
		go func(vendor, name string) {
			defer wg.Done()
			defer func() { <-sem }()

			url := api.PACKAGIST_ENUMERATE_API_URL + vendor + "/" + name + ".json"

			res, err := c.fetch(url)
			if err != nil {
				logrus.Warnf("Fetch packagist package %s/%s failed: %v", vendor, name, err)
				return
			}

			detail, err := api.FromPackagistDetail(res)
			if err != nil {
				logrus.Warnf("Parse packagist package %s/%s failed: %v", vendor, name, err)
				return
			}

			pkg := detail.Package

			var firstVer packagist.Version

			if len(pkg.Versions) > 0 {
				keys := make([]string, 0, len(pkg.Versions))
				for k := range pkg.Versions {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				firstVer = pkg.Versions[keys[0]]
			}

			u := GetPackagistBestUrl(firstVer, pkg)

			c.writer.Write(pkg.Name)
			c.writer.Write(u)
			c.writer.Write(strconv.Itoa(pkg.Downloads.Total))
			c.writer.Write(firstVer.Version)

			if len(firstVer.Require) > 0 {
				deps := make([]string, 0, len(firstVer.Require))
				for dep := range firstVer.Require {
					deps = append(deps, dep)
				}
				c.writer.Write("[" + strings.Join(deps, ", ") + "]")
			} else {
				c.writer.Write("[]")
			}

			c.writer.Write("\n")

		}(vendor, name)
		collected++
	}
	wg.Wait()
	logrus.Infof("Enumerator has collected and written %d packages", collected)
	return nil
}
