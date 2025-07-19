package main

import (
	"sync"

	"github.com/HUSTSecLab/OpenSift/pkg/collector/alpine"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/archlinux"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/aur"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/centos"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/debian"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/deepin"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/fedora"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/gentoo"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/homebrew"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/nix"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/openeuler"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/openkylin"
	"github.com/HUSTSecLab/OpenSift/pkg/collector/ubuntu"
	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/spf13/pflag"
)

var (
	flagType    = pflag.String("type", "", "type of the distribution")
	flagGenDot  = pflag.String("gendot", "", "output dot file")
	workerCount = pflag.Int("worker", 1, "number of workers")
	batchSize   = pflag.Int("batch", 1000, "batch size")
	downloadDir = pflag.String("downloadDir", "./download", "download directory")
)

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	if *flagType == "" {
		var wg sync.WaitGroup
		wg.Add(11)

		go func() {
			defer wg.Done()
			archlinux.NewArchLinuxCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			debian.NewDebianCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			deepin.NewDeepinCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			ubuntu.NewUbuntuCollector().Collect(*flagGenDot)
		}()
		// go func() {
		// 	defer wg.Done()
		// 	nix.NewNixCollector().Collect(*workerCount, *batchSize, *flagGenDot)
		// }()
		go func() {
			defer wg.Done()
			homebrew.NewHomebrewCollector().Collect(*flagGenDot, *downloadDir)
		}()
		// go func() {
		// 	defer wg.Done()
		// 	gentoo.NewGentooCollector().Collect(*flagGenDot)
		// }()
		go func() {
			defer wg.Done()
			fedora.NewFedoraCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			centos.NewCentosCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			alpine.NewAlpineCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			aur.NewAurCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			openeuler.NewOpenEulerCollector().Collect(*flagGenDot)
		}()
		go func() {
			defer wg.Done()
			openkylin.NewOpenKylinCollector().Collect(*flagGenDot)
		}()

		wg.Wait()
	} else {
		switch *flagType {
		case "archlinux":
			archlinux.NewArchLinuxCollector().Collect(*flagGenDot)
		case "debian":
			debian.NewDebianCollector().Collect(*flagGenDot)
		case "deepin":
			deepin.NewDeepinCollector().Collect(*flagGenDot)
		case "ubuntu":
			ubuntu.NewUbuntuCollector().Collect(*flagGenDot)
		case "nix":
			nix.NewNixCollector().Collect(*workerCount, *batchSize, *flagGenDot)
		case "homebrew":
			homebrew.NewHomebrewCollector().Collect(*flagGenDot, *downloadDir)
		case "gentoo":
			gentoo.NewGentooCollector().Collect(*flagGenDot, *downloadDir)
		case "fedora":
			fedora.NewFedoraCollector().Collect(*flagGenDot)
		case "centos":
			centos.NewCentosCollector().Collect(*flagGenDot)
		case "alpine":
			alpine.NewAlpineCollector().Collect(*flagGenDot)
		case "aur":
			aur.NewAurCollector().Collect(*flagGenDot)
		case "openeuler":
			openeuler.NewOpenEulerCollector().Collect(*flagGenDot)
		case "openkylin":
			openkylin.NewOpenKylinCollector().Collect(*flagGenDot)
		}
	}
}
