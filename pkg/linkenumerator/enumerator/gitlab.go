package enumerator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/sirupsen/logrus"
)

type gitlabEnumerator struct {
	enumeratorBase
	take int
	jobs int
}

func NewGitlabEnumerator(take int, jobs int) Enumerator {
	return &gitlabEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
		jobs:           jobs,
	}
}

func (c *gitlabEnumerator) Enumerate() error {
	if err := c.writer.Open(); err != nil {
		return err
	}
	defer c.writer.Close()

	api_url := api.GITLAB_ENUMERATE_API_URL
	var wg sync.WaitGroup

	collected := 0
	var muCollected sync.Mutex

	pool := gopool.NewPool("gitlab_enumerator", int32(c.jobs), &gopool.Config{})

	for page := 1; page <= c.take/api.PER_PAGE; page++ {
		time.Sleep(api.TIME_INTERVAL * time.Second)
		wg.Add(1)
		pool.Go(func() {
			defer wg.Done()
			u := fmt.Sprintf(
				"%s?%s=%s&%s=%s&%s=%d&%s=%d",
				api_url,
				"order_by", "star_count",
				"sort", "desc",
				"per_page", api.PER_PAGE,
				"page", page,
			)
			res, err := c.fetch(u)
			if err != nil {
				logrus.Errorf("Gitlab fetch failed: %v", err)
				return
			}

			resp, err := api.FromGitlab(res)

			if err != nil {
				logrus.Errorf("Gitlab unmarshal failed: %v", err)
				return
			}

			for _, v := range *resp {
				if strings.HasSuffix(v.HTTPURLToRepo, ".git") {
					v.HTTPURLToRepo = v.HTTPURLToRepo[:len(v.HTTPURLToRepo)-4]
				}
				c.writer.Write(v.Name)
				c.writer.Write(v.HTTPURLToRepo)
				c.writer.Write(fmt.Sprintf("%d", v.StarCount))
				c.writer.Write("\n")
			}

			func() {
				muCollected.Lock()
				defer muCollected.Unlock()
				collected += len(*resp)
			}()

		})
	}
	wg.Wait()
	logrus.Infof("Enumerator has collected and written %d repositories", collected)
	return nil
}
