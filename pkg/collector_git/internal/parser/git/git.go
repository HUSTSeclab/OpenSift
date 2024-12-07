package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/logger"
	parser "github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/parser"
	url "github.com/HUSTSecLab/criticality_score/pkg/collector_git/internal/parser/url"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/licensecheck"
)

type Repo struct {
	Name    string
	Owner   string
	Source  string
	URL     string
	License string
	// is_maintained bool
	Languages        string
	Ecosystems       string
	CreatedSince     time.Time
	UpdatedSince     time.Time
	ContributorCount int
	OrgCount         int
	CommitFrequency  float64
}

func NewRepo() Repo {
	return Repo{
		Name:             parser.UNKNOWN_NAME,
		Owner:            parser.UNKNOWN_OWNER,
		Source:           parser.UNKNOWN_SOURCE,
		URL:              parser.UNKNOWN_URL,
		License:          parser.UNKNOWN_LICENSE,
		Languages:        parser.UNKNOWN_LANGUAGES,
		Ecosystems:       parser.UNKNOWN_ECOSYSTEMS,
		CreatedSince:     parser.UNKNOWN_TIME,
		UpdatedSince:     parser.UNKNOWN_TIME,
		ContributorCount: parser.UNKNOWN_COUNT,
		OrgCount:         parser.UNKNOWN_COUNT,
		CommitFrequency:  parser.UNKNOWN_FREQUENCY,
	}
}

func GetBlobs(r *git.Repository) (*[]*object.Blob, error) {
	bIter, err := r.BlobObjects()

	if err != nil {
		return nil, err
	}

	blobs := make([]*object.Blob, 0)

	err = bIter.ForEach(func(b *object.Blob) error {
		// fmt.Println(b)
		blobs = append(blobs, b)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &blobs, nil
}

func GetBranches(r *git.Repository) (*[]*plumbing.Reference, error) {
	rIter, err := r.Branches()

	if err != nil {
		return nil, err
	}

	refs := make([]*plumbing.Reference, 0)
	err = rIter.ForEach(func(r *plumbing.Reference) error {
		// fmt.Println(r)
		refs = append(refs, r)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &refs, nil
}

func GetCommits(r *git.Repository) (*[]*object.Commit, error) {
	cIter, err := r.CommitObjects()

	if err != nil {
		return nil, err
	}

	commits := make([]*object.Commit, 0)
	err = cIter.ForEach(func(c *object.Commit) error {
		// fmt.Println(c)
		commits = append(commits, c)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &commits, nil
}

func GetConfig(r *git.Repository) (*gitconfig.Config, error) {
	c, err := r.Config()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func GetObjects(r *git.Repository) (*[]*object.Object, error) {
	oIter, err := r.Objects()

	if err != nil {
		return nil, err
	}

	objs := make([]*object.Object, 0)
	err = oIter.ForEach(func(o object.Object) error {
		// fmt.Println(o)
		objs = append(objs, &o)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &objs, nil
}

func GetReferences(r *git.Repository) (*[]*plumbing.Reference, error) {
	rIter, err := r.References()

	if err != nil {
		return nil, err
	}

	refs := make([]*plumbing.Reference, 0)
	err = rIter.ForEach(func(r *plumbing.Reference) error {
		// fmt.Println(r)
		refs = append(refs, r)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &refs, nil
}

func GetRemotes(r *git.Repository) (*[]*git.Remote, error) {
	remotes, err := r.Remotes()

	if err != nil {
		return nil, err
	}

	return &remotes, nil
}

func GetTags(r *git.Repository) (*[]*object.Tag, error) {
	tIter, err := r.TagObjects()

	if err != nil {
		return nil, err
	}

	tags := make([]*object.Tag, 0)
	err = tIter.ForEach(func(t *object.Tag) error {
		// fmt.Println(t)
		tags = append(tags, t)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &tags, nil
}

func GetTagRefs(r *git.Repository) (*[]*plumbing.Reference, error) {
	rIter, err := r.Tags()

	if err != nil {
		return nil, err
	}

	refs := make([]*plumbing.Reference, 0)
	err = rIter.ForEach(func(r *plumbing.Reference) error {
		// fmt.Println(r)
		refs = append(refs, r)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &refs, nil
}

func GetTrees(r *git.Repository) (*[]*object.Tree, error) {
	tIter, err := r.TreeObjects()

	if err != nil {
		return nil, err
	}

	trees := make([]*object.Tree, 0)
	err = tIter.ForEach(func(t *object.Tree) error {
		// fmt.Println(t)
		trees = append(trees, t)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &trees, nil
}

func GetWorkTree(r *git.Repository) (*git.Worktree, error) {
	wt, err := r.Worktree()

	if err != nil {
		return nil, err
	}

	return wt, nil
}

func GetURL(r *git.Repository) (string, error) {
	//? In most cases, the Remote URLs of Git Fetch and Git Push are the same, but we take the former one
	remotes, err := GetRemotes(r)

	if err != nil {
		logger.Error(err)
		return "", err
	}

	if len(*remotes) == 0 {
		return "", nil
	}

	var u string

	if len((*remotes)[0].Config().URLs) > 0 {
		u = (*remotes)[0].Config().URLs[0]
	}

	for _, remote := range *remotes {
		if remote.Config().Name == parser.DEFAULT_REMOTE_NAME {
			if len(remote.Config().URLs) > 0 {
				u = remote.Config().URLs[0]
				break
			}
		}
	}

	return u, nil
}

func GetLanguages(filename string, l *map[string]int) {
	v, ok := parser.LANGUAGE_FILENAMES[filename]
	if ok {
		(*l)[v] += 1
	} else {
		ex := filepath.Ext(filename)
		v, ok = parser.LANGUAGE_EXTENSIONS[ex]
		if ok {
			(*l)[v] += 1
		}
	}
}

func GetEcosystem(filename string, e *map[string]int) {
	v, ok := parser.ECOSYSTEM_MAP[filename]
	if ok {
		(*e)[v] += 1
	}
}

func GetLicense(f *object.File) (string, error) {
	text, err := f.Contents()
	if err != nil {
		return "", err
	}
	cov := licensecheck.Scan([]byte(text))
	license := cov.Match[0].ID

	return license, nil
}

func (repo *Repo) WalkLog(r *git.Repository) error {
	cIter, err := r.Log(&git.LogOptions{
		//* From:  ref.Hash(),
		All:   true,
		Since: &parser.BEGIN_TIME,
		Until: &parser.END_TIME,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return err
	}

	contributors := make(map[string]int, 0)
	orgs := make(map[string]int, 0)
	var commit_count float64 = 0

	latest_commit, err := cIter.Next()
	if err != nil {
		return err
	}

	author := fmt.Sprintf("%s(%s)", latest_commit.Author.Name, latest_commit.Author.Email)
	e := strings.Split(latest_commit.Author.Email, "@")
	org := e[len(e)-1]

	repo.UpdatedSince = latest_commit.Committer.When
	contributors[author]++
	orgs[org]++

	if latest_commit.Author.When.After(parser.LAST_YEAR) {
		commit_count++
	}

	flag := true
	created_since := latest_commit.Committer.When

	err = cIter.ForEach(func(c *object.Commit) error {
		author := fmt.Sprintf("%s(%s)", c.Author.Name, c.Author.Email)
		e = strings.Split(c.Author.Email, "@")
		org := e[len(e)-1]

		//! It made sense that this `if`` statement is not necessary but sometimes there are errors
		if created_since.After(c.Committer.When) {
			created_since = c.Committer.When
		}
		contributors[author]++
		orgs[org]++

		if flag {
			if created_since.After(parser.LAST_YEAR) {
				commit_count++
			} else {
				flag = false
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	repo.CreatedSince = created_since
	repo.ContributorCount = len(contributors)
	repo.OrgCount = len(orgs)
	repo.CommitFrequency = commit_count / 52

	return nil
}

func (repo *Repo) WalkRepo(r *git.Repository) error {

	ref, err := r.Head()
	if err != nil {
		return err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	languages := make(map[string]int, 0)
	ecosystems := make(map[string]int, 0)

	fIter := tree.Files()
	err = fIter.ForEach(func(f *object.File) error {
		filename := filepath.Base(f.Name)
		GetLanguages(filename, &languages)
		GetEcosystem(filename, &ecosystems)
		if repo.License != parser.UNKNOWN_LICENSE {
			if _, ok := parser.LICENSE_FILENAMES[filename]; ok {
				license, err := GetLicense(f)
				if err != nil {
					logger.Error(err)
				} else {
					repo.License = license
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	l := ""
	for k, v := range languages {
		if v >= parser.LANGUAGE_THRESHOLD {
			l += fmt.Sprintf("%s ", k)
		}
	}

	e := ""
	for k, v := range ecosystems {
		if v >= parser.ECOSYSTEM_THRESHOLD {
			e += fmt.Sprintf("%s ", k)
		}
	}

	if len(l) != 0 {
		repo.Languages = l[:len(l)-1]
	}
	if len(e) != 0 {
		repo.Ecosystems = e[:len(e)-1]
	}

	return nil
}

func ParseRepo(r *git.Repository) (*Repo, error) {

	repo := NewRepo()

	u, err := GetURL(r)
	if err != nil {
		logger.Errorf("Failed to Get RepoURL for %v", err)
	}

	var name, owner, source string
	if u == "" {
		source = parser.UNKNOWN_SOURCE
		u = parser.UNKNOWN_URL
		name = parser.UNKNOWN_NAME
		owner = parser.UNKNOWN_OWNER
	} else {
		uu := url.ParseURL(u)

		if uu.Pathname == "" {
			name = parser.UNKNOWN_NAME
			owner = parser.UNKNOWN_OWNER
		} else {
			path := strings.Split(uu.Pathname, "/")
			name = strings.Split(path[len(path)-1], ".")[0]
			owner = path[len(path)-2]
		}

		if uu.Resource == "" {
			source = parser.UNKNOWN_SOURCE
		} else {
			source = uu.Resource
		}
	}

	err = repo.WalkRepo(r)
	if err != nil {
		logger.Errorf("Failed to Walk Repo for %v", err)
	}

	err = repo.WalkLog(r)
	if err != nil {
		logger.Errorf("Failed to Walk Log for %v", err)
	}

	repo.Name = name
	repo.Owner = owner
	repo.Source = source
	repo.URL = u

	return &repo, nil
}
