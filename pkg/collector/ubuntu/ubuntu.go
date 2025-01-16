package ubuntu

import (
	"log"
	"regexp"
	"strings"

	collector "github.com/HUSTSecLab/criticality_score/pkg/collector/internal"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
)

type UbuntuCollector struct {
	collector.CollecterInterface
}

func (dc *UbuntuCollector) Collect(outputPath string) {
	adc := storage.GetDefaultAppDatabaseContext()
	data := dc.GetPackageInfo(collector.UbuntuURL)
	dc.ParseInfo(data)
	dc.GetDep()
	dc.PageRank(0.85, 20)
	dc.GetDepCount()
	dc.UpdateDistRepoCount(adc)
	dc.CalculateDistImpact()
	dc.UpdateOrInsertDatabase(adc)
	dc.UpdateOrInsertDistDependencyDatabase(adc)
	err := dc.GenerateDependencyGraph(outputPath)
	if err != nil {
		log.Printf("Error generating dependency graph: %v\n", err)
		return
	}
}

func (dc *UbuntuCollector) ParseInfo(data string) {
	var currentPkg *collector.PackageInfo
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		switch {
		case strings.Contains(line, "Package"):
			if currentPkg != nil {
				dc.SetPkgInfo(currentPkg.Name, currentPkg)
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				currentPkg = &collector.PackageInfo{Name: strings.TrimSpace(parts[1])}
			}
		case strings.Contains(line, "Version"):
			if currentPkg != nil {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					currentPkg.Version = strings.TrimSpace(parts[1])
				}
			}
		case strings.Contains(line, "Description"):
			if currentPkg != nil {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					description := strings.TrimSpace(parts[1])
					if len(description) > 255 {
						description = description[:255]
					}
					currentPkg.Description = description
				}
			}
		case strings.Contains(line, "Homepage"):
			if currentPkg != nil {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					currentPkg.Homepage = strings.TrimSpace(parts[1])
				}
			}
		case strings.Contains(line, "Depends"):
			if currentPkg != nil {
				depLine := strings.TrimPrefix(line, "Depends: ")
				re := regexp.MustCompile(`[\w\-\.|]+(?:\s*\([^)]+\))?`)
				matches := re.FindAllString(depLine, -1)

				var cleanedDeps []string
				for _, match := range matches {
					for _, subPkg := range strings.Split(match, "|") {
						subPkg = strings.TrimSpace(subPkg)
						if idx := strings.Index(subPkg, " ("); idx != -1 {
							subPkg = subPkg[:idx]
						}
						cleanedDeps = append(cleanedDeps, subPkg)
					}
				}
				currentPkg.DirectDepends = cleanedDeps
			}
		}
	}
	if currentPkg != nil {
		dc.SetPkgInfo(currentPkg.Name, currentPkg)
	}
}

func NewUbuntuCollector() *UbuntuCollector {
	return &UbuntuCollector{
		CollecterInterface: collector.NewCollector(repository.Ubuntu, repository.DistPackageTablePrefix("ubuntu")),
	}
}
