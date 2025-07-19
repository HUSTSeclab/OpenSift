package main

import (
	"log"
	"os"

	"github.com/HUSTSecLab/OpenSift/scripts/git-metrics-sync/internal/gmsync"
	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/spf13/pflag"
)

var batchSize = pflag.Int("batch", 1000, "batch size")

func main() {
	os.Args = []string{"cmd/git-metrics-sync", "-h"}
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	log.Println("Starting synchronization...")
	gmsync.Run()
	log.Println("Synchronization complete.")
	gmsync.Union_repo(*batchSize)
}
