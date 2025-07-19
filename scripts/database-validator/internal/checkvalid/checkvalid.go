package checkvalid

import (
	"database/sql"
	"encoding/csv"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Metrics struct {
	CreatedSince time.Time
	UpdatedSince time.Time
	Score        float64
}

var repoList = []string{"debian_packages", "arch_packages", "gentoo_packages", "nix_packages", "homebrew_packages"}

func fetchDistroGitlink(gitlink *sql.DB, repo string) []string {
	query := "SELECT git_link FROM " + repo
	rows, err := gitlink.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var gitLinks []string
	for rows.Next() {
		var gitLink sql.NullString
		err := rows.Scan(&gitLink)
		if err != nil {
			panic(err)
		}
		if gitLink.Valid {
			gitLinks = append(gitLinks, gitLink.String)
		}
	}
	return gitLinks
}

func checkDistroValid(gitlink *sql.DB, repo string) [][]string {
	gitLinks := fetchDistroGitlink(gitlink, repo)
	var invalidLinks [][]string
	for _, link := range gitLinks {
		if link == "" || link == "NA" || link == "NaN" {
			continue
		}
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") && !strings.HasPrefix(link, "git://") {
			invalidLinks = append(invalidLinks, []string{link, "invalid protocol"})
		} else if strings.Contains(link, "/tree/") {
			invalidLinks = append(invalidLinks, []string{link, "invalid link"})
		}
	}
	return invalidLinks
}

func fetchMetrics(db *sql.DB) map[string]Metrics {
	query := "SELECT git_link, created_since, updated_since, scores FROM git_metrics"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	MetricsList := make(map[string]Metrics)
	for rows.Next() {
		var gitLink string
		var createdSince sql.NullTime
		var updatedSince sql.NullTime
		var createdSinceres time.Time
		var updatedSinceres time.Time
		var score float64
		err := rows.Scan(&gitLink, &createdSince, &updatedSince, &score)
		if err != nil {
			panic(err)
		}
		if createdSince.Valid {
			createdSinceres = createdSince.Time
		}
		if updatedSince.Valid {
			updatedSinceres = updatedSince.Time
		}
		if !createdSince.Valid {
			createdSinceres = time.Time{}
		}
		if !updatedSince.Valid {
			updatedSinceres = time.Time{}
		}
		MetricsList[gitLink] = Metrics{CreatedSince: createdSinceres, UpdatedSince: updatedSinceres, Score: score}
	}
	return MetricsList
}

func checkMetricsValid(db *sql.DB) [][]string {
	MetricsList := fetchMetrics(db)
	var invalidLinks [][]string
	for link, metrics := range MetricsList {
		duration := metrics.CreatedSince.Sub(metrics.UpdatedSince)
		if duration > 0 {
			invalidLinks = append(invalidLinks, []string{link, "created_since is after updated_since"})
		} else if metrics.Score < 0 {
			invalidLinks = append(invalidLinks, []string{link, "score is less than 0"})
		}
	}
	return invalidLinks
}

func checkCloneValid(db *sql.DB, maxThreads int) [][]string {
	query := "SELECT git_link FROM git_metrics WHERE clone_valid = false"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var gitLinks []string
	for rows.Next() {
		var gitLink string
		err := rows.Scan(&gitLink)
		if err != nil {
			panic(err)
		}
		gitLinks = append(gitLinks, gitLink)
	}
	var invalidLinks [][]string
	sem := make(chan struct{}, maxThreads)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, link := range gitLinks {
		wg.Add(1)
		sem <- struct{}{}
		go func(gitLink string) {
			defer wg.Done()
			defer func() { <-sem }()

			tempDir, err := os.MkdirTemp("", "test_repo_*")
			if err != nil {
				mu.Lock()
				invalidLinks = append(invalidLinks, []string{gitLink, "failed to create temp directory"})
				mu.Unlock()
				return
			}
			defer os.RemoveAll(tempDir)

			cmd := exec.Command("git", "clone", "--depth=1", gitLink, tempDir)
			err = cmd.Start()
			if err != nil {
				mu.Lock()
				invalidLinks = append(invalidLinks, []string{gitLink, "failed to clone"})
				mu.Unlock()
				return
			}
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			select {
			case <-time.After(15 * time.Second):
				cmd.Process.Kill()
				mu.Lock()
				invalidLinks = append(invalidLinks, []string{gitLink, "clone timed out"})
				mu.Unlock()
			case err := <-done:
				if err != nil {
					mu.Lock()
					invalidLinks = append(invalidLinks, []string{gitLink, "failed to clone"})
					mu.Unlock()
				}
			}
		}(link)
	}
	wg.Wait()
	return invalidLinks
}

func checkCloneValidDefault(db *sql.DB, maxThreads int) [][]string {
	query := "SELECT git_link, created_since FROM git_metrics"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var invalidLinks [][]string

	sem := make(chan struct{}, maxThreads)
	var wg sync.WaitGroup

	for rows.Next() {
		var gitLink string
		var createdSince sql.NullTime
		err := rows.Scan(&gitLink, &createdSince)
		if err != nil {
			panic(err)
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(gitLink string, createdSince sql.NullTime) {
			defer wg.Done()
			defer func() { <-sem }()

			if strings.Contains(gitLink, "sourceforge.net") || strings.Contains(gitLink, "sf.net") {
				tempDir, err := os.MkdirTemp("", "test_repo_*")
				if err != nil {
					invalidLinks = append(invalidLinks, []string{gitLink, "failed to create temp directory"})
					return
				}
				defer os.RemoveAll(tempDir)

				cmd := exec.Command("git", "clone", "--depth=1", gitLink, tempDir)
				err = cmd.Start()
				if err != nil {
					invalidLinks = append(invalidLinks, []string{gitLink, "failed to clone"})
					return
				}
				done := make(chan error, 1)
				go func() {
					done <- cmd.Wait()
				}()
				select {
				case <-time.After(15 * time.Second):
					cmd.Process.Kill()
					invalidLinks = append(invalidLinks, []string{gitLink, "clone timed out"})
				case err := <-done:
					if err != nil {
						invalidLinks = append(invalidLinks, []string{gitLink, "failed to clone"})
					}
				}
			} else if createdSince.Valid {
				if createdSince.Time.Sub(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)) == 0 {
					invalidLinks = append(invalidLinks, []string{gitLink, "created_since is 0001-01-01, maybe cannot clone"})
				}
			}
		}(gitLink, createdSince)
	}

	wg.Wait()
	return invalidLinks
}
func CheckVaild(db *sql.DB, checkCloneValidflag bool, maxThreads int) [][]string {
	var invalidLinks [][]string
	for _, repo := range repoList {
		invalidLinks = append(invalidLinks, checkDistroValid(db, repo)...)
	}
	invalidLinks = append(invalidLinks, checkMetricsValid(db)...)
	if checkCloneValidflag {
		invalidLinks = append(invalidLinks, checkCloneValid(db, maxThreads)...)
	} else {
		invalidLinks = append(invalidLinks, checkCloneValidDefault(db, maxThreads)...)
	}
	return invalidLinks
}

func WriteCsv(invalidLinks [][]string, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	for _, value := range invalidLinks {
		link := value[0]
		reason := value[1]
		writer.Write([]string{link, reason})
		writer.Flush()
	}
}
