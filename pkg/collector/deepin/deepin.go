package deepin

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/lib/pq"
)

var cacheDir = "/tmp/cloc-deepin-cache"

type DepInfo struct {
	Name        string
	Arch        string
	Version     string
	Description string
	Homepage    string
	PageRank    float64
}

type PackageInfo struct {
	DependsCount int
	Description  string
	Homepage     string
	Version      string
	PageRank     float64
}

type DeepinCollector struct {
	packages map[string]map[string]interface{}
}

func NewDeepinCollector() *DeepinCollector {
	return &DeepinCollector{
		packages: make(map[string]map[string]interface{}),
	}
}

func (dc *DeepinCollector) updateOrInsertDatabase(pkgInfoMap map[string]PackageInfo) error {
	db, err := storage.GetDefaultAppDatabaseContext().GetDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for pkgName, pkgInfo := range pkgInfoMap {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM deepin_packages WHERE package = $1)", pkgName).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			_, err := db.Exec("INSERT INTO deepin_packages (package, depends_count, description, homepage, version, page_rank) VALUES ($1, $2, $3, $4, $5, $6)",
				pkgName, pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.Homepage, pkgInfo.Version, pkgInfo.PageRank)
			if err != nil {
				return err
			}
		} else {
			_, err := db.Exec("UPDATE deepin_packages SET depends_count = $1, description = $2, homepage = $3, version = $4, page_rank = $5 WHERE package = $6",
				pkgInfo.DependsCount, pkgInfo.Description, pkgInfo.Homepage, pkgInfo.Version, pkgInfo.PageRank, pkgName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dc *DeepinCollector) storeDependenciesInDatabase(pkgName string, dependencies []DepInfo) error {
	db, err := storage.GetDefaultAppDatabaseContext().GetDatabaseConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, dep := range dependencies {
		_, err := db.Exec("INSERT INTO deepin_relationships (frompackage, topackage) VALUES ($1, $2)", pkgName, dep.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dc *DeepinCollector) getMirrorFile(path string) []byte {
	resp, _ := http.Get("https://mirrors.hust.edu.cn/deepin/" + path)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func (dc *DeepinCollector) getDecompressedFile(path string) string {
	file := dc.getMirrorFile(path)
	reader, _ := gzip.NewReader(strings.NewReader(string(file)))
	defer reader.Close()
	decompressed, _ := ioutil.ReadAll(reader)
	return string(decompressed)
}

func (dc *DeepinCollector) getBeigePackageList() string {
	return dc.getDecompressedFile("beige/dists/beige/main/binary-amd64/Packages.gz")
}

func (dc *DeepinCollector) parseList() {
	content := dc.getBeigePackageList()
	lists := strings.Split(content, "\n\n")
	dc.packages = make(map[string]map[string]interface{})

	for _, packageStr := range lists {
		if strings.TrimSpace(packageStr) == "" {
			continue
		}
		pkg := make(map[string]interface{})
		pkg["__raw"] = packageStr
		var currentKey string

		lines := strings.Split(packageStr, "\n")
		for _, line := range lines {
			if matched, _ := regexp.MatchString(".+:.+", line); matched {
				parts := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				pkg[key] = value
				currentKey = key
			} else if matched, _ := regexp.MatchString(".+:\\s*", line); matched {
				currentKey = strings.Split(line, ":")[0]
				pkg[currentKey] = ""
			} else if matched, _ := regexp.MatchString(" .+", line); matched {
				if currentValue, ok := pkg[currentKey].(string); ok {
					pkg[currentKey] = currentValue + " " + strings.TrimSpace(line)
				}
			}
		}

		if depends, ok := pkg["Depends"].(string); ok {
			depList := strings.Split(depends, ",")
			var depStrings []interface{}
			for _, dep := range depList {
				depInfo := dc.toDep(strings.TrimSpace(dep), packageStr)
				depStrings = append(depStrings, depInfo)
			}
			pkg["Depends"] = depStrings
		}

		if packageName, ok := pkg["Package"].(string); ok {
			dc.packages[packageName] = pkg
		}
	}
}

func (dc *DeepinCollector) toDep(dep string, rawContent string) DepInfo {
	re := regexp.MustCompile(`^(.+?)(:.+?)?(\s\((.+)\))?(\s\|.+)?$`)
	matches := re.FindStringSubmatch(dep)

	depInfo := DepInfo{Name: dep, Arch: "", Version: "", Description: "", Homepage: ""}

	if matches != nil {
		depInfo.Name = matches[1]
		if matches[2] != "" {
			depInfo.Arch = strings.TrimSpace(matches[2])
		}
		if matches[4] != "" {
			depInfo.Version = strings.TrimSpace(matches[4])
		}
	}

	descriptionRegex := regexp.MustCompile(`(?m)^Description:\s*(.*)$`)
	homepageRegex := regexp.MustCompile(`(?m)^Homepage:\s*(.*)$`)

	if descMatches := descriptionRegex.FindStringSubmatch(rawContent); len(descMatches) > 1 {
		depInfo.Description = descMatches[1]
	}

	if homeMatches := homepageRegex.FindStringSubmatch(rawContent); len(homeMatches) > 1 {
		depInfo.Homepage = homeMatches[1]
	}

	return depInfo
}

func (dc *DeepinCollector) generateDependencyGraph(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("digraph {\n")

	packageIndices := make(map[string]int)
	index := 0

	for pkgName, pkgInfo := range dc.packages {
		packageIndices[pkgName] = index
		label := fmt.Sprintf("%s@%s", pkgName, pkgInfo["Version"].(string))
		writer.WriteString(fmt.Sprintf("  %d [label=\"%s\"];\n", index, label))
		index++
	}

	for pkgName, pkgInfo := range dc.packages {
		pkgIndex := packageIndices[pkgName]
		if depends, ok := pkgInfo["Depends"].([]interface{}); ok {
			for _, depInterface := range depends {
				if depInfo, ok := depInterface.(DepInfo); ok {
					if depIndex, ok := packageIndices[depInfo.Name]; ok {
						writer.WriteString(fmt.Sprintf("  %d -> %d [label=\"%s\"];\n", pkgIndex, depIndex, depInfo.Version))
					}
				}
			}
		}
	}

	writer.WriteString("}\n")
	writer.Flush()
	return nil
}

func (dc *DeepinCollector) getAllDep(pkgName string, deps []string) []string {
	deps = append(deps, pkgName)
	if pkg, ok := dc.packages[pkgName]; ok {
		if depends, ok := pkg["Depends"].([]interface{}); ok {
			for _, depInterface := range depends {
				if depMap, ok := depInterface.(DepInfo); ok {
					pkgname := depMap.Name
					if !contains(deps, pkgname) {
						deps = dc.getAllDep(pkgname, deps)
					}
				}
			}
		}
	}
	return deps
}

func (dc *DeepinCollector) rankPage(maxIterations int, dampingFactor float64) map[string]float64 {
	rank := make(map[string]float64)
	N := len(dc.packages)
	for pkgName := range dc.packages {
		rank[pkgName] = 1.0 / float64(N)
	}

	for i := 0; i < maxIterations; i++ {
		newRank := make(map[string]float64)
		for pkgName := range dc.packages {
			newRank[pkgName] = (1 - dampingFactor) / float64(N)
		}

		for pkgName, pkgInfo := range dc.packages {
			if depends, ok := pkgInfo["Depends"].([]interface{}); ok {
				for _, depInterface := range depends {
					if depInfo, ok := depInterface.(DepInfo); ok {
						if _, ok := dc.packages[depInfo.Name]; ok {
							newRank[depInfo.Name] += dampingFactor * rank[pkgName] / float64(len(depends))
						}
					}
				}
			}
		}

		rank = newRank
	}
	return rank
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func (dc *DeepinCollector) Collect(outputPath string) {
	fmt.Println("Getting package list...")
	dc.parseList()
	fmt.Printf("Done, total: %d packages.\n", len(dc.packages))
	fmt.Println("Building dependencies graph...")

	keys := make([]string, 0, len(dc.packages))
	for k := range dc.packages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	depMap := make(map[string][]string)
	for _, pkgName := range keys {
		deps := dc.getAllDep(pkgName, []string{})
		depMap[pkgName] = deps
	}
	fmt.Println("Calculating dependencies count...")
	countMap := make(map[string]int)
	for _, deps := range depMap {
		for _, dep := range deps {
			countMap[dep]++
		}
	}

	pagerank := dc.rankPage(20, 0.85)

	pkgInfoMap := make(map[string]PackageInfo)

	for pkgName, pkgInfo := range dc.packages {
		depCount := countMap[pkgName]

		description, ok := pkgInfo["Description"].(string)
		if !ok {
			description = ""
		}

		homepage, ok := pkgInfo["Homepage"].(string)
		if !ok {
			homepage = ""
		}

		version, ok := pkgInfo["Version"].(string)
		if !ok {
			version = ""
		}

		pagerankValue := pagerank[pkgName]

		pkgInfoMap[pkgName] = PackageInfo{
			DependsCount: depCount,
			Description:  description,
			Homepage:     homepage,
			Version:      version,
			PageRank:     pagerankValue,
		}
	}

	err := dc.updateOrInsertDatabase(pkgInfoMap)
	if err != nil {
		fmt.Printf("Error updating database: %v\n", err)
		return
	}
	for _, pkgInfo := range dc.packages {
		if packageName, ok := pkgInfo["Package"].(string); ok {
			if depends, ok := pkgInfo["Depends"].([]interface{}); ok {
				dependencies := make([]DepInfo, len(depends))
				for i, depInterface := range depends {
					if depInfo, ok := depInterface.(DepInfo); ok {
						dependencies[i] = depInfo
					}
				}
				if err := dc.storeDependenciesInDatabase(packageName, dependencies); err != nil {
					if isUniqueViolation(err) {
						continue
					}
					fmt.Printf("Error storing dependencies for package %s: %v\n", packageName, err)
					return
				}
			}
		}
	}
	fmt.Println("Database updated successfully.")

	if outputPath != "" {
		err := dc.generateDependencyGraph(outputPath)
		if err != nil {
			fmt.Printf("Error generating dependency graph: %v\n", err)
			return
		}
		fmt.Println("Dependency graph generated successfully.")
	}
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
