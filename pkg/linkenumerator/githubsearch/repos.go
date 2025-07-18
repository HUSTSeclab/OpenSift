// Copyright 2022 Criticality Score Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package githubsearch

import (
	"errors"
	"fmt"
	"io"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/githubapi/pagination"
	"github.com/hasura/go-graphql-client"
)

// repo is part of the GitHub GraphQL query and includes the fields
// that will be populated in a response.
type repo struct {
	URL            string
	StargazerCount int
}

// repoQuery is a GraphQL query for iterating over repositories in GitHub.
type repoQuery struct {
	Search struct {
		Nodes []struct {
			Repository repo `graphql:"...on Repository"`
		}
		PageInfo struct {
			EndCursor   string
			HasNextPage bool
		}
		RepositoryCount int
	} `graphql:"search(type: REPOSITORY, query: $query, first: $perPage, after: $endCursor)"`
}

// Total implements the pagination.PagedQuery interface.
func (q *repoQuery) Total() int {
	return q.Search.RepositoryCount
}

// Length implements the pagination.PagedQuery interface.
func (q *repoQuery) Length() int {
	return len(q.Search.Nodes)
}

// Get implements the pagination.PagedQuery interface.
func (q *repoQuery) Get(i int) any {
	return q.Search.Nodes[i].Repository
}

// Reset implements the pagination.PagedQuery interface.
func (q *repoQuery) Reset() {
	q.Search.Nodes = nil
}

// HasNextPage implements the pagination.PagedQuery interface.
func (q *repoQuery) HasNextPage() bool {
	return q.Search.PageInfo.HasNextPage
}

// NextPageVars implements the pagination.PagedQuery interface.
func (q *repoQuery) NextPageVars() map[string]any {
	if q.Search.PageInfo.EndCursor == "" {
		return map[string]any{
			"endCursor": (*graphql.String)(nil),
		}
	} else {
		return map[string]any{
			"endCursor": graphql.String(q.Search.PageInfo.EndCursor),
		}
	}
}

func buildQuery(q string, minStars, maxStars int) string {
	q = q + " sort:stars "
	if maxStars > 0 {
		return q + fmt.Sprintf("stars:%d..%d", minStars, maxStars)
	} else {
		return q + fmt.Sprintf("stars:>=%d", minStars)
	}
}

func (re *Searcher) runRepoQuery(q string) (*pagination.Cursor, error) {
	re.logger.WithFields(map[string]interface{}{
		"query": q,
	}).Debug("Searching GitHub")
	vars := map[string]any{
		"query":   graphql.String(q),
		"perPage": graphql.Int(re.perPage),
	}
	c, err := pagination.Query(re.ctx, re.client, &repoQuery{}, vars)
	if err != nil {
		return nil, fmt.Errorf("repo search query '%s' failed: %w", q, err)
	}
	return c, nil
}

// ReposByStars will call emitter once for each repository returned when searching for baseQuery
// with at least minStars, order from the most stars, to the least.
//
// The emitter function is called with the repository's Url.
//
// The algorithm works to overcome the approx 1000 repository limit returned by a single search
// across 10 pages by:
// - Ordering GitHub's repositories from most stars to least.
// - Iterating through all the repositories returned by each query.
// - Getting the number of stars for the last repository returned.
// - Using this star value, plus an overlap, to be the maximum star limit.
//
// The algorithm fails if the last star value plus overlap has the same or larger value as the
// previous iteration.
func (re *Searcher) ReposByStars(baseQuery string, minStars, overlap int, emitter func(string)) error {
	repos := make(map[string]empty)
	maxStars := -1
	stars := 0

	for {
		q := buildQuery(baseQuery, minStars, maxStars)
		c, err := re.runRepoQuery(q)
		if err != nil {
			return err
		}

		total := c.Total()
		seen := 0
		stars = 0
		for {
			obj, err := c.Next()
			if obj == nil && errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return err
			}
			repo := obj.(repo)
			seen++
			stars = repo.StargazerCount
			if _, ok := repos[repo.URL]; !ok {
				repos[repo.URL] = empty{}
				emitter(repo.URL)
			}
		}
		remaining := total - seen
		re.logger.WithFields(map[string]interface{}{
			"total_available": total,
			"total_returned":  seen,
			"total_remaining": remaining,
			"unique_repos":    len(repos),
			"last_stars":      stars,
			"query":           q,
		}).Debug("Finished iterating through results")
		newMaxStars := stars + overlap
		switch {
		case remaining <= 0:
			// nothing remains, we are done.
			return nil
		case maxStars == -1, newMaxStars < maxStars:
			maxStars = newMaxStars
		default:
			// the gap between "stars" and "maxStars" is less than "overlap", so we can't
			// safely step any lower without skipping stars.
			re.logger.WithFields(map[string]interface{}{
				"min_stars": minStars,
				"stars":     stars,
				"max_stars": maxStars,
				"overlap":   overlap,
			}).Error("Too many repositories for current range")
			return ErrorUnableToListAllResult
		}
	}
}
