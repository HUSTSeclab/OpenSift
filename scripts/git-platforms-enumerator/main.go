package main

import (
	"strings"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/enumerator"
	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// dateFlag implements the flag.Value interface to simplify the input and validation of
// dates from the command line.
type dateFlag time.Time

const dateFormat = "2006-01-02"

func (d *dateFlag) Set(value string) error {
	t, err := time.Parse(dateFormat, value)
	if err != nil {
		return err
	}
	*d = dateFlag(t)
	return nil
}

func (d *dateFlag) String() string {
	return (*time.Time)(d).Format(dateFormat)
}

func (d *dateFlag) Time() time.Time {
	return time.Time(*d)
}

func (d *dateFlag) Type() string {
	return "date"
}

func main() {
	// flags
	var (
		flagPlatforms   = pflag.String("platforms", "", "comma separated list of platforms to enumerate")
		flagOutputType  = pflag.String("output", "stdout", "output type: allow stdout, file, db")
		flagOutputFilev = pflag.String("output-file", "", "output file")
		flagJobs        = pflag.IntP("jobs", "j", 10, "number of concurrent jobs")
		flagTake        = pflag.Int("take", 1000, "number of repositories to enumerate, only for gitlab and bitbucket")
	)

	// github flags
	var (
		flagMinStars        = pflag.Int("min-stars", 100, "minimum number of stars")
		flagStarOverlap     = pflag.Int("star-overlap", 5, "minimum number of stars overlap")
		flagRequireMinStars = pflag.Bool("require-min-stars", false, "require minimum number of stars")
		flagQuery           = pflag.String("query", "is:public", "sets the base query")
		flagStartDate       = dateFlag(enumerator.GithubEpochDate)
		flagEndDate         = dateFlag(time.Now().UTC().Truncate(time.Hour * 24))
	)

	pflag.Var(&flagStartDate, "start-date", "start date for the search")
	pflag.Var(&flagEndDate, "end-date", "end date for the search")
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	platforms := strings.Split(*flagPlatforms, ",")

	var platformList = map[string]struct {
		Enumerator  func() enumerator.Enumerator
		TablePrefix string
	}{
		"github": {
			Enumerator: func() enumerator.Enumerator {
				return enumerator.NewGithubEnumerator(&enumerator.GithubEnumeratorConfig{
					MinStars:        *flagMinStars,
					StarOverlap:     *flagStarOverlap,
					RequireMinStars: *flagRequireMinStars,
					Query:           *flagQuery,
					StartDate:       flagStartDate.Time(),
					EndDate:         flagEndDate.Time(),
					Workers:         *flagJobs,
				})
			},
			TablePrefix: "github",
		},
		"gitlab": {
			Enumerator: func() enumerator.Enumerator {
				return enumerator.NewGitlabEnumerator(*flagTake, *flagJobs)
			},
			TablePrefix: "gitlab",
		},
		"bitbucket": {
			Enumerator: func() enumerator.Enumerator {
				return enumerator.NewBitBucketEnumerator(*flagTake)
			},
			TablePrefix: "bitbucket",
		},
		"pypi": {
			Enumerator: func() enumerator.Enumerator {
				return enumerator.NewPypiBigQueryEnumerator(&enumerator.PypiBigQueryEnumeratorConfig{
					ProjectID: "magnetic-nimbus-425712-e3",
				})
			},
			TablePrefix: "pypi",
		},
		"pypi_slow": {
			Enumerator: func() enumerator.Enumerator {
				return enumerator.NewPypiEnumerator(&enumerator.PypiEnumeratorConfig{
					Jobs: *flagJobs,
				})
			},
			TablePrefix: "pypi",
		},
		"npm": {
			Enumerator: func() enumerator.Enumerator {
				return enumerator.NewNpmEnumerator()
			},
			TablePrefix: "npm",
		},
	}

	for _, platform := range platforms {
		var w writer.Writer
		var tablePrefix string
		var en enumerator.Enumerator

		if platform, ok := platformList[platform]; ok {
			en = platform.Enumerator()
			tablePrefix = platform.TablePrefix
		} else {
			log.Errorf("unknown platform %s", platform)
			return
		}

		switch *flagOutputType {
		case "stdout":
			w = writer.NewStdOutWriter()
		case "file":
			w = writer.NewTextFileWriter(*flagOutputFilev)
		case "db":
			w = writer.NewDatabaseWriter(storage.GetDefaultAppDatabaseContext(), tablePrefix)
		default:
			panic("unknown output type")
		}

		en.SetWriter(w)

		err := en.Enumerate()
		if err != nil {
			log.WithError(err).Errorf("failed to enumerate %s", platform)
		}
	}
}
