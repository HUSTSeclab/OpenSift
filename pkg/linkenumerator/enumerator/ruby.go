package enumerator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/ruby"
	"github.com/sirupsen/logrus"
)

type RubyEnumerator struct {
	enumeratorBase
	take int
}

func NewRubyEnumerator(take int) *RubyEnumerator {
	return &RubyEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func GetBestRubyUrl(resp *ruby.Response) string {
	if resp.HomepageURI != "" {
		return resp.HomepageURI
	}
	if resp.SourceCodeURI != "" {
		return resp.SourceCodeURI
	}
	if resp.GemURI != "" {
		return resp.GemURI
	}
	return ""
}

func (c *RubyEnumerator) Enumerate() error {
	if err := c.writer.Open(); err != nil {
		logrus.Error("Open writer: ", err)
		return err
	}
	defer c.writer.Close()

	req := c.client.R()
	req.SetURL(api.RUBY_INDEX_API_URL)
	logrus.Info("Downloading RubyGems package name list...")

	resp := req.Do()
	if resp.IsErrorState() {
		logrus.Error("RubyGems fetch error: ", resp.Err)
		return resp.Err
	}

	names, err := api.FromRubyNames(resp)
	if err != nil {
		logrus.Error("Failed to parse RubyGems names: ", err)
		return err
	}
	if c.take > 0 && c.take < len(names) {
		names = names[:c.take]
	}

	maxConcurrency := 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	rateLimiter := time.Tick(time.Second / 10)

	collected := 0

	for _, name := range names {
		<-rateLimiter
		sem <- struct{}{}
		wg.Add(1)
		go func(gemName string) {
			defer wg.Done()
			defer func() { <-sem }()
			url := fmt.Sprintf("%s/%s.json", api.RUBY_ENUMERATE_API_URL, gemName)

			res, err := c.fetch(url)
			if err != nil || res == nil || res.IsErrorState() {
				logrus.Warnf("Fetch RubyGem failed: %s, error: %v", gemName, err)
				return
			}

			resp, err := api.FromRubyDetail(res)
			if err != nil {
				logrus.Warnf("Parse RubyGem failed: %s, error: %v", gemName, err)
				return
			}

			u := GetBestRubyUrl(resp)

			c.writer.Write(resp.Name)
			c.writer.Write(resp.Version)
			c.writer.Write(fmt.Sprintf("%d", resp.VersionDownloads))
			c.writer.Write(u)
			c.writer.Write(fmt.Sprintf("%d", resp.Downloads))

			if len(resp.Dependencies.Runtime) > 0 {
				deps := make([]string, 0, len(resp.Dependencies.Runtime))
				for _, dep := range resp.Dependencies.Runtime {
					deps = append(deps, dep.Name)
				}
				c.writer.Write("Dependencies: [" + strings.Join(deps, ", ") + "]")
			} else {
				c.writer.Write("Dependencies: []")
			}

			c.writer.Write("\n")

			collected++
		}(name)
	}
	wg.Wait()
	logrus.Infof("Enumerator has collected and written %d packages", collected)
	return nil
}
