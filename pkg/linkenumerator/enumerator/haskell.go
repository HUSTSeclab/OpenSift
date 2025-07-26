package enumerator

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/api/haskell"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type HaskellEnumerator struct {
	enumeratorBase
	take int
}

func NewHaskellEnumerator(take int) *HaskellEnumerator {
	return &HaskellEnumerator{
		enumeratorBase: newEnumeratorBase(),
		take:           take,
	}
}

func GetBestHaskellUrl(val haskell.Value) string {
	if val.Homepage != "" {
		return val.Homepage
	}
	if val.URL != "" {
		return val.URL
	}
	return ""
}

func (c *HaskellEnumerator) Enumerate() error {
	if err := c.writer.Open(); err != nil {
		logrus.Error("Open writer: ", err)
		return err
	}
	defer c.writer.Close()

	u := api.HASKELL_INDEX_API_URL
	resp, err := http.Get(u)
	if err != nil {
		logrus.Error("Fetch Hackage package list: ", err)
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logrus.Error("Parse Hackage package list: ", err)
		return err
	}

	collected := 0
	maxConcurrency := 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	doc.Find("#content ul li a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if c.take > 0 && collected >= c.take {
			return false
		}
		name := s.Text()
		href, exists := s.Attr("href")
		if !exists || !strings.HasPrefix(href, "/package/") {
			return true
		}
		url := api.HASKELL_ENUMERATE_API_URL + href

		sem <- struct{}{}
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			defer func() { <-sem }()
			val := fetchHaskellValue(name, url)
			val.Name = name
			u := GetBestHaskellUrl(val)

			c.writer.Write(val.Name)
			c.writer.Write(u)
			c.writer.Write(val.Version)
			c.writer.Write(val.Dependencies)
			c.writer.Write(val.Downloads)
			c.writer.Write("\n")
		}(name, url)

		collected++
		return true
	})

	wg.Wait()
	logrus.Infof("Enumerator has collected and written %d Haskell packages", collected)
	return nil
}

func fetchHaskellValue(name, pkgURL string) haskell.Value {
	resp, err := http.Get(pkgURL)
	if err != nil {
		return haskell.Value{}
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return haskell.Value{}
	}

	info := haskell.Value{}

	info.Version = ""
	doc.Find("th:contains('Versions')").Each(func(i int, s *goquery.Selection) {
		td := s.Parent().Find("td")
		if td.Length() > 0 {
			strong := td.Find("strong")
			if strong.Length() > 0 {
				info.Version = strings.TrimSpace(strong.Text())
			} else {
				a := td.Find("a")
				if a.Length() > 0 {
					info.Version = strings.TrimSpace(a.Last().Text())
				}
			}
		}
	})

	info.Dependencies = strings.TrimSpace(doc.Find("th:contains('Dependencies')").Parent().Find("td").Text())
	info.Author = strings.TrimSpace(doc.Find("th:contains('Author')").Parent().Find("td").Text())
	info.Homepage = strings.TrimSpace(doc.Find("th:contains('Home page')").Parent().Find("td").Text())

	info.SourceRepo = ""
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		text := strings.ToLower(s.Text())
		if strings.Contains(text, "source repo") || strings.Contains(text, "source repository") {
			if href, exists := s.Attr("href"); exists && (strings.HasPrefix(href, "http") || strings.HasPrefix(href, "https")) {
				info.SourceRepo = href
			}
		}
	})
	if info.SourceRepo == "" {
		re := regexp.MustCompile(`git clone (https?://[^\s]+)`)
		doc.Find("td").Each(func(i int, s *goquery.Selection) {
			m := re.FindStringSubmatch(s.Text())
			if len(m) > 1 {
				info.SourceRepo = m[1]
			}
		})
	}

	info.ReverseDependencies = strings.TrimSpace(doc.Find("th:contains('Reverse Dependencies')").Parent().Find("td").Text())
	info.Downloads = strings.TrimSpace(doc.Find("th:contains('Downloads')").Parent().Find("td").Text())
	info.Status = strings.TrimSpace(doc.Find("th:contains('Status')").Parent().Find("td").Text())

	return info
}
