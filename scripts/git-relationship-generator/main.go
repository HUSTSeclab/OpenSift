package main

import (
	"log"

	"github.com/HUSTSecLab/OpenSift/scripts/git-relationship-generator/internal/pkgdep2git"
	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/spf13/pflag"
)

var batchSize = pflag.Int("batch", 1000, "batch size for updating scores")

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)
	db, err := storage.GetDefaultAppDatabaseContext().GetDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()

	depMap := pkgdep2git.FetchAlldep(db)
	gitdepMap := pkgdep2git.GenGitDep(db, depMap)
	err = pkgdep2git.BatchUpdate(db, *batchSize, gitdepMap)
	if err != nil {
		log.Fatalf("Error updating database: %v", err)
	}
}
