package enumerator

import (
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/bitbucket"
	"github.com/sirupsen/logrus"
)

type BitBucketEnumerator struct {
	enumeratorBase
	take int
}

func NewBitBucketEnumerator(take int) *BitBucketEnumerator {
	return &BitBucketEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func getBestBitBucketGitURL(val *bitbucket.Value) string {
	for _, v := range val.Links.Clone {
		if v.Name == "https" || v.Name == "http" {
			return v.Href
		}
	}
	if len(val.Links.Clone) > 0 {
		return val.Links.Clone[0].Href
	}
	return ""
}

func (c *BitBucketEnumerator) Enumerate() error {
	err := c.writer.Open()
	defer c.writer.Close()
	if err != nil {
		logrus.Panic("Open writer", err)
	}

	u := api.BITBUCKET_ENUMERATE_API_URL
	collected := 0
	for {
		res, err := c.fetch(u)
		if err != nil {
			logrus.Panic("Bitbucket", err)
		}
		resp, err := api.FromBitbucket(res)
		if err != nil {
			logrus.Panic("Bitbucket", err)
		}

		for _, v := range resp.Values {
			url := getBestBitBucketGitURL(&v)
			if strings.HasSuffix(url, ".git") {
				url = url[:len(url)-4]
			}
			c.writer.Write(url)
		}

		collected += len(resp.Values)

		logrus.Infof("Enumerator has collected and written %d repositories", collected)

		if collected >= c.take || resp.Next == "" || len(resp.Values) == 0 {
			break
		}

		u = resp.Next
	}
	return nil
}
