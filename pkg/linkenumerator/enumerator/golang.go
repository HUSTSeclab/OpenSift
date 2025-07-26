package enumerator

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	goapi "github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/golang"
	"github.com/sirupsen/logrus"
)

type GoEnumerator struct {
	enumeratorBase
	take int
}

func NewGoEnumerator(take int) *GoEnumerator {
	return &GoEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func fetchGoModules(limit int) ([]goapi.GoModule, error) {
	var (
		resp *http.Response
		err  error
	)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err = client.Get(api.GOLANG_INDEX_API_URL)
	if err != nil {
		logrus.Warnf("GET https://index.golang.org/index failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var modules []goapi.GoModule
	scanner := bufio.NewScanner(resp.Body)
	count := 0
	for scanner.Scan() {
		var entry struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
			Time    string `json:"Timestamp"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			modules = append(modules, goapi.GoModule{
				Path:    entry.Path,
				Version: entry.Version,
				Time:    entry.Time,
			})
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return modules, scanner.Err()
}

func fetchGoModuleDetail(mod *goapi.GoModule) {
	url := "https://pkg.go.dev/" + mod.Path + "?tab=overview"
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buf := make([]byte, 64*1024)
	n, _ := resp.Body.Read(buf)
	html := string(buf[:n])

	if idx := strings.Index(html, `property="og:url" content="`); idx != -1 {
		start := idx + len(`property="og:url" content="`)
		end := strings.Index(html[start:], `"`)
		if end != -1 {
			mod.Homepage = html[start : start+end]
		}
	}

	if idx := strings.Index(html, `Repository</span>`); idx != -1 {
		sub := html[idx:]
		if hrefIdx := strings.Index(sub, `<a href="`); hrefIdx != -1 {
			hrefStart := hrefIdx + len(`<a href="`)
			hrefEnd := strings.Index(sub[hrefStart:], `"`)
			if hrefEnd != -1 {
				mod.Repository = sub[hrefStart : hrefStart+hrefEnd]
			}
		}
	}
}

func (c *GoEnumerator) Enumerate() error {
	if err := c.writer.Open(); err != nil {
		logrus.Panic("Open writer: ", err)
		return err
	}
	defer c.writer.Close()

	modules, err := fetchGoModules(c.take)
	if err != nil {
		logrus.Panic("Fetch go modules: ", err)
		return err
	}

	var wg sync.WaitGroup
	maxConcurrency := 10
	sem := make(chan struct{}, maxConcurrency)
	rateLimiter := time.Tick(time.Second / 10)

	collected := 0

	for i := range modules {
		<-rateLimiter
		sem <- struct{}{}
		wg.Add(1)
		go func(mod *goapi.GoModule) {
			defer wg.Done()
			defer func() { <-sem }()
			fetchGoModuleDetail(mod)
			c.writer.Write(mod.Path)
			c.writer.Write(mod.Version)
			c.writer.Write(mod.Homepage)
			c.writer.Write(mod.Repository)
			c.writer.Write("\n")
			collected++
		}(&modules[i])
	}
	wg.Wait()
	logrus.Infof("GoEnumerator has collected and written %d modules", collected)
	return nil
}
