package enumerator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/cargo"
	"github.com/sirupsen/logrus"
)

type CargoEnumerator struct {
	enumeratorBase
	take int
}

func NewCargoEnumerator(take int) *CargoEnumerator {
	return &CargoEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func getBestCargoUrl(crate *cargo.Crate) string {
	if crate.Homepage != nil && *crate.Homepage != "" {
		return *crate.Homepage
	}
	if crate.Repository != "" {
		return crate.Repository
	}
	return ""
}

func (c *CargoEnumerator) Enumerate() error {
	if err := c.writer.Open(); err != nil {
		logrus.Panic("Open writer", err)
		return err
	}
	defer c.writer.Close()

	u := api.CRATES_IO_ENUMERATE_API_URL + "?sort=downloads&per_page=100&page=1"
	collected := 0
	maxConcurrency := 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	stop := false

	for !stop {
		sem <- struct{}{}
		wg.Add(1)
		pageURL := u
		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()
			res, err := c.fetch(url)
			if err != nil {
				logrus.Warnf("Cargo fetch error: %v", err)
				return
			}
			resp, err := api.FromCargo(res)
			if err != nil {
				logrus.Warnf("Cargo unmarshal error: %v", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for _, crate := range resp.Crates {
				if collected >= c.take {
					stop = true
					break
				}
				url := getBestCargoUrl(&crate)
				if strings.HasSuffix(url, ".git") {
					url = url[:len(url)-4]
				}
				c.writer.Write(crate.Name)
				c.writer.Write(url)
				c.writer.Write(crate.MaxVersion)
				c.writer.Write(fmt.Sprintf("%d", crate.Downloads))
				c.writer.Write(fmt.Sprintf("%d", crate.RecentDownloads))
				c.writer.Write("\n")
				collected++
			}
			if collected >= c.take || resp.Meta.NextPage == "" || len(resp.Crates) == 0 {
				stop = true
			} else {
				u = api.CRATES_IO_ENUMERATE_API_URL + resp.Meta.NextPage
			}
		}(pageURL)

		if len(sem) == maxConcurrency {
			wg.Wait()
		}
		if stop {
			break
		}
	}
	wg.Wait()
	logrus.Infof("Enumerator has collected and written %d repositories", collected)
	return nil
}
