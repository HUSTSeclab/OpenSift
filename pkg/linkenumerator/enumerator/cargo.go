package enumerator

import (
	"fmt"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/linkenumerator/api"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/sirupsen/logrus"
)

// Todo Use channel to receive and write data
func (c *enumeratorBase) EnumerateCargo() {
	api_url := api.CRATES_IO_ENUMERATE_API_URL
	var wg sync.WaitGroup
	wg.Add(1)
	for page := 1; page <= 1; page++ {
		time.Sleep(api.TIME_INTERVAL * time.Second)
		gopool.Go(func() {
			defer wg.Done()
			u := fmt.Sprintf(
				"%s?%s=%s&%s=%d&%s=%d",
				api_url,
				"sort", "downloads",
				"per_page", api.PER_PAGE,
				"page", page,
			)
			_, err := c.fetch(u)
			if err != nil {
				logrus.Panic("Cargo", err)
			}
			// _ := res.Bytes()
			// err = os.WriteFile(config.OUTPUT_DIR+config.CRATES_IO_OUTPUT_FILEPATH, data, 0644)
			if err != nil {
				logger.Panic("Cargo", err)
			}
		})
	}
	wg.Wait()
}
