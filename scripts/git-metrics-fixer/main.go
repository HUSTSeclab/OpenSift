// This file can manual fix the git metrics of a repository
package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/config"
	collector "github.com/HUSTSecLab/OpenSift/pkg/gitfile/collector"
	git "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/git"
	url "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/OpenSift/pkg/logger"
	scores "github.com/HUSTSecLab/OpenSift/pkg/score"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/sqlutil"
	gogit "github.com/go-git/go-git/v5"
	"github.com/spf13/pflag"
)

var (
	flagUpdateDB   = pflag.Bool("update-db", false, "Whether to update the database")
	flagUpdateLink = pflag.String("update-link", "", "Which link to update")
	flagUpdateList = pflag.String("list", "", "Which file to update")
	workpoolSize   = pflag.Int("workpool", 50, "workpool size")
	filePath       = pflag.String("file", "", "file path")
)

func main() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "This program collects metrics for git repositories.\n")
		fmt.Fprintf(os.Stderr, "This tool can be used to fix the git metrics of a repository manually.\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
	}

	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)
	ac := storage.GetDefaultAppDatabaseContext()

	// updateDB := *flagUpdateDB
	link := *flagUpdateLink
	file := *flagUpdateList

	var links []string
	var err error

	if file != "" {
		links, err = ReadFileLines(file)
		if err != nil {
			logger.Panicf("Reading file %s Failed", file)
		}
	} else if link != "" {
		links = append(links, link)
	} else {
		logger.Panicf("No link or file provided")
	}
	scores.UpdatePackageList(ac)

	var wg sync.WaitGroup
	var mu sync.Mutex
	workpool := make(chan struct{}, *workpoolSize)

	for _, link := range links {
		wg.Add(1)
		workpool <- struct{}{}
		go func(link string) {
			defer wg.Done()
			defer func() { <-workpool }()
			logger.Infof("Collecting %s", link)
			r := &gogit.Repository{}
			u, _ := url.ParseURL(link)
			r, err := collector.Collect(&u, *filePath, nil)
			if err != nil {
				logger.Println("Collecting Failed:", link)
				return
			}

			repo, err := git.ParseRepo(r)
			if err != nil {
				logger.Infof("Parsing %s Failed", link)
			}
			logger.Infof("%s Collected", repo.Name)

			// repo.Show()
			gitMetric := &repository.GitMetric{
				GitLink:          &link,
				CommitFrequency:  sqlutil.ToNullable(repo.CommitFrequency),
				ContributorCount: sqlutil.ToNullable(repo.ContributorCount),
				CreatedSince:     sqlutil.ToNullable(repo.CreatedSince),
				UpdatedSince:     sqlutil.ToNullable(repo.UpdatedSince),
				OrgCount:         sqlutil.ToNullable(repo.OrgCount),
			}

			mu.Lock()
			InsertGitMeticAndFetch(ac, gitMetric)
			mu.Unlock()
		}(link)
	}
	wg.Wait()
}

func InsertGitMeticAndFetch(ac storage.AppDatabaseContext, gitMetadata *repository.GitMetric) map[string]*scores.GitMetadata {
	repo := repository.NewGitMetricsRepository(ac)
	repo.InsertOrUpdate(gitMetadata)
	gitMetric := scores.FetchGitMetricsSingle(ac, *gitMetadata.GitLink)
	return gitMetric
}

func ReadFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
