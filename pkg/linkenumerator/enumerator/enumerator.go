package enumerator

import (
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

type Enumerator interface {
	SetWriter(writer writer.Writer)
	SetToken(token string)
	Enumerate() error
}

type enumeratorBase struct {
	client *req.Client
	token  string
	writer writer.Writer
}

func newEnumeratorBase() enumeratorBase {
	return enumeratorBase{
		client: req.C().ImpersonateChrome().SetTimeout(10 * time.Second),
	}
}

func (c *enumeratorBase) SetWriter(writer writer.Writer) {
	c.writer = writer
}

func (c *enumeratorBase) SetToken(token string) {
	c.token = token
	c.client.SetCommonBearerAuthToken(token)
}

func (c *enumeratorBase) fetch(url string) (*req.Response, error) {
	res, err := c.client.R().Get(url)

	if err != nil || res.GetStatusCode() != 200 {
		logrus.Errorf(
			"[Enumerator] fetch failed: code=%d, msg=%s, err=%v",
			res.GetStatusCode(),
			res.String(),
			err,
		)
		return nil, err
	}

	return res, nil
}
