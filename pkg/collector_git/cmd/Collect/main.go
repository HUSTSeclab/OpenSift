/*
 * @Date: 2023-11-11 22:44:26
 * @LastEditTime: 2024-12-09 19:32:13
 * @Description: Collect Local Repo
 */
package main

import (
	// collector "github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/collector"
	"os"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/config"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/io/database"
	psql "github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/io/database/psql"
	csv "github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/io/file/csv"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/logger"
	git "github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/parser/git"
	url "github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/parser/url"
	"github.com/bytedance/gopkg/util/gopool"

	//"fmt"
	"sync"
)

func getPath() string {
	var path string
	if len(os.Args) == 2 {
		path = os.Args[1]
	} else {
		path = ""
	}
	return path
}

func main() {
	path := getPath()
	urls, err := csv.GetCSVInput(path)
	if err != nil {
		logger.Fatalf("Failed to read %s", path)
	}

	var wg sync.WaitGroup
	logger.Infof("%d urls in total", len(urls))

	ch := make(chan database.GitMetrics, config.BATCH_SIZE)

	db, err := psql.InitDB()
	if err != nil {
		logger.Fatal("Failed to connect database")
	}
	psql.CreateTable(db)

	wg.Add(1)
	gopool.Go(func() {
		defer wg.Done()
		var data database.GitMetrics
		var ok bool
		for {
			data, ok = <-ch
			if !ok {
				break
			}
			psql.InsertTable(db, &data)
		}
	})

	wg.Add(len(urls))
	for index, input := range urls {
		if index%10 == 0 {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}

		gopool.Go(func() {
			defer wg.Done()
			u := url.ParseURL(input[0])
			r, err := collector.Collect(&u)
			if err != nil {
				logger.Panicf("Collecting %s Failed", input)
			}

			repo, err := git.ParseRepo(r)
			if err != nil {
				logger.Panicf("[!] Paring %s Failed", input)
			}

			output := database.Repo2Metrics(repo)
			ch <- output
		})
	}
	wg.Wait()
	close(ch)
}
