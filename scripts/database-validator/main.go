package main

import (
	"log"

	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/scripts/database-validator/internal/checkvalid"
	"github.com/spf13/pflag"
)

var flagOutputFile = pflag.String("output", "output.csv", "path to the output file")
var flagCheckCloneValid = pflag.Bool("checkCloneValid", false, "check clone valid")
var flagMaxThreads = pflag.Int("maxThreads", 10, "max threads")

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)
	db, err := storage.GetDefaultAppDatabaseContext().GetDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()
	invalidLinks := checkvalid.CheckVaild(db, *flagCheckCloneValid, *flagMaxThreads)
	checkvalid.WriteCsv(invalidLinks, *flagOutputFile)
	log.Println("checkvalid finished")
}
