package enumerator

import (
	"testing"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

func Test_enumerateBitbucket(t *testing.T) {
	t.Run("Bitbucket", func(t *testing.T) {
		c := NewBitBucketEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateCargo(t *testing.T) {
	t.Run("Cargo", func(t *testing.T) {
		c := NewCargoEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateGithub(t *testing.T) {
	t.Run("Github", func(t *testing.T) {
		cfg := &GithubEnumeratorConfig{
			MinStars:        1000,
			StarOverlap:     100,
			RequireMinStars: false,
			Query:           "stars:>1000",
			Workers:         2,
			StartDate:       GithubEpochDate,
			EndDate:         time.Now(),
		}
		c := NewGithubEnumerator(cfg)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateGitlab(t *testing.T) {
	t.Run("Gitlab", func(t *testing.T) {
		c := NewGitlabEnumerator(1000, 4)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateGO(t *testing.T) {
	t.Run("Go", func(t *testing.T) {
		c := NewGoEnumerator(100)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateHaskell(t *testing.T) {
	t.Run("Haskell", func(t *testing.T) {
		c := NewHaskellEnumerator(500)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateNpm(t *testing.T) {
	t.Run("Npm", func(t *testing.T) {
		c := NewNpmEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateNuget(t *testing.T) {
	t.Run("Nuget", func(t *testing.T) {
		c := NewNugetEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumeratePackagist(t *testing.T) {
	t.Run("Packagist", func(t *testing.T) {
		c := NewPackagistEnumerator(400)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumeratePypiBigQuery(t *testing.T) {
	t.Run("PypiBigQuery", func(t *testing.T) {
		cfg := &PypiBigQueryEnumeratorConfig{
			ProjectID: "your-gcp-project-id", // 请替换为你的GCP项目ID
		}
		c := NewPypiBigQueryEnumerator(cfg)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumeratePypi(t *testing.T) {
	t.Run("Pypi", func(t *testing.T) {
		c := NewPypiEnumerator(500)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateRuby(t *testing.T) {
	t.Run("Ruby", func(t *testing.T) {
		c := NewRubyEnumerator(200)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}
