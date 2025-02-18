package score

import (
	"math"
	"time"

	log "github.com/HUSTSecLab/criticality_score/pkg/logger"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/sqlutil"
)

type LinkScore struct {
	// GitMetrics       []*repository.GitMetric
	GitMetadataScore GitMetadataScore
	// LangEcosystems   []*repository.LangEcosystem
	LangEcoScore LangEcoScore
	// DistDependencies []*repository.DistDependency
	DistScore DistScore
	Score     float64
}

type GitMetadata struct {
	Id               int64
	CreatedSince     time.Time
	UpdatedSince     time.Time
	ContributorCount int
	CommitFrequency  float64
	Org_Count        int
}

type GitMetadataScore struct {
	GitMetrics       []*repository.GitMetric
	GitMetadataScore float64
}

type DistMetadata struct {
	Id        int64
	DepImpact float64
	DepCount  int
	PageRank  float64
	Type      repository.DistType
}

type LangEcoMetadata struct {
	Id            int64
	Type          repository.LangEcosystemType
	LangEcoImpact float64
	DepCount      int
}

type DistScore struct {
	DistDependencies []*repository.DistDependency
	DistImpact       float64
	DistPageRank     float64
	DistScore        float64
}

type LangEcoScore struct {
	LangEcosystems  []*repository.LangEcosystem
	LangEcoImpact   float64
	LangEcoPageRank float64
	LangEcoScore    float64
}

var SigmoidWeight = 1.2

// Define weights (αi) and max thresholds (Ti)
var weights = map[string]map[string]float64{
	"gitMetadataScore": {
		"created_since":     1,
		"updated_since":     -1,
		"contributor_count": 2,
		"commit_frequency":  1,
		"org_count":         1,
		"gitMetadataScore":  0.2,
	},
	"distScore": {
		"dist_impact":   1,
		"dist_pagerank": 1,
		"distScore":     0.5,
	},
	"langEcoScore": {
		"lang_eco_impact": 1,
		"langEcoScore":    0.3,
	},
}

var thresholds = map[string]map[string]float64{
	"gitMetadataScore": {
		"created_since":     120,
		"updated_since":     120,
		"contributor_count": 40000,
		"commit_frequency":  1000,
		"org_count":         8400,
		"gitMetadataScore":  4,
	},
	"distScore": {
		"dist_impact":   22,
		"dist_pagerank": 3,
		"distScore":     2,
	},
	"langEcoScore": {
		"lang_eco_impact": 0.5,
		"langEcoScore":    1,
	},
}

var PackageList = map[repository.DistType]int{
	repository.Debian:   0,
	repository.Arch:     0,
	repository.Nix:      0,
	repository.Homebrew: 0,
	repository.Gentoo:   0,
	repository.Alpine:   0,
	repository.Fedora:   0,
	repository.Ubuntu:   0,
	repository.Deepin:   0,
	repository.Aur:      0,
	repository.Centos:   0,
}

var PackageWeight = map[repository.LangEcosystemType]float64{
	repository.Npm:   1.5,
	repository.Go:    1.4,
	repository.Maven: 1.3,
	repository.Pypi:  1.2,
	repository.NuGet: 1.1,
	repository.Cargo: 1,
}

func (langEcoMetadata *LangEcoMetadata) ParseLangEcoMetadata(langEcosystem *repository.LangEcosystem) {
	langEcoMetadata.Id = *langEcosystem.ID
	langEcoMetadata.Type = *langEcosystem.Type
	langEcoMetadata.DepCount = *langEcosystem.DepCount
	langEcoMetadata.LangEcoImpact = *langEcosystem.LangEcoImpact
}

func (distMetadata *DistMetadata) PraseDistMetadata(distLink *repository.DistDependency) {
	distMetadata.Id = *distLink.ID
	distMetadata.DepCount = *distLink.DepCount
	distMetadata.DepImpact = *distLink.DepImpact
	distMetadata.PageRank = *distLink.PageRank
	distMetadata.Type = *distLink.Type
}

func (gitMetadata *GitMetadata) ParseMetadata(gitMetic *repository.GitMetric) {
	gitMetadata.Id = *gitMetic.ID
	if !sqlutil.IsNull(gitMetic.CreatedSince) {
		gitMetadata.CreatedSince = **gitMetic.CreatedSince
	}
	if !sqlutil.IsNull(gitMetic.UpdatedSince) {
		gitMetadata.UpdatedSince = **gitMetic.UpdatedSince
	}
	if !sqlutil.IsNull(gitMetic.ContributorCount) {
		gitMetadata.ContributorCount = **gitMetic.ContributorCount
	}
	if !sqlutil.IsNull(gitMetic.CommitFrequency) {
		gitMetadata.CommitFrequency = **gitMetic.CommitFrequency
	}
	if !sqlutil.IsNull(gitMetic.OrgCount) {
		gitMetadata.Org_Count = **gitMetic.OrgCount
	}
}

func (langEcoScore *LangEcoScore) CalculateLangEcoScore() {
	langEcoScore.LangEcoScore = weights["langEcoScore"]["lang_eco_impact"] * LogNormalize(langEcoScore.LangEcoImpact, thresholds["langEcoScore"]["lang_eco_impact"])
}

func NewLangEcoScore() *LangEcoScore {
	return &LangEcoScore{}
}

func (gitMetadataScore *GitMetadataScore) CalculateGitMetadataScore(gitMetadata *GitMetadata) {
	var score float64
	var createdSinceScore, updatedSinceScore, contributorCountScore, commitFrequencyScore, orgCountScore float64

	monthsSinceCreation := time.Since(gitMetadata.CreatedSince).Hours() / (24 * 30)
	createdSinceScore = weights["gitMetadataScore"]["created_since"] * LogNormalize(monthsSinceCreation, thresholds["gitMetadataScore"]["created_since"])
	score += createdSinceScore

	monthsSinceUpdate := time.Since(gitMetadata.UpdatedSince).Hours() / (24 * 30)
	updatedSinceScore = weights["gitMetadataScore"]["updated_since"] * LogNormalize(monthsSinceUpdate, thresholds["gitMetadataScore"]["updated_since"])
	score += updatedSinceScore

	contributorCountScore = weights["gitMetadataScore"]["contributor_count"] * LogNormalize(float64(gitMetadata.ContributorCount), thresholds["gitMetadataScore"]["contributor_count"])
	score += contributorCountScore

	commitFrequencyScore = weights["gitMetadataScore"]["commit_frequency"] * LogNormalize(gitMetadata.CommitFrequency, thresholds["gitMetadataScore"]["commit_frequency"])
	score += commitFrequencyScore

	orgCountScore = weights["gitMetadataScore"]["org_count"] * LogNormalize(float64(gitMetadata.Org_Count), thresholds["gitMetadataScore"]["org_count"])
	score += orgCountScore

	gitMetadataScore.GitMetadataScore = score
	gitMetadataScore.GitMetrics = []*repository.GitMetric{
		{
			ID: sqlutil.ToData(gitMetadata.Id),
		},
	}
}

func NewGitMetadata() *GitMetadata {
	return &GitMetadata{}
}

func (distScore *DistScore) CalculateDistScore() {
	distScore.DistScore = weights["distScore"]["dist_impact"]*LogNormalize(distScore.DistImpact, thresholds["distScore"]["dist_impact"]) + weights["distScore"]["dist_pagerank"]*LogNormalize(distScore.DistPageRank, thresholds["distScore"]["dist_pagerank"])
}

func (linkScore *LinkScore) CalculateScore() {
	score := 0.0

	score += weights["gitMetadataScore"]["gitMetadataScore"] * LogNormalize(linkScore.GitMetadataScore.GitMetadataScore, thresholds["gitMetadataScore"]["gitMetadataScore"]) * 100

	score += weights["langEcoScore"]["langEcoScore"] * LogNormalize(linkScore.LangEcoScore.LangEcoScore, thresholds["langEcoScore"]["langEcoScore"]) * 100

	score += weights["distScore"]["distScore"] * LogNormalize(linkScore.DistScore.DistScore, thresholds["distScore"]["distScore"]) * 100

	linkScore.Score = score
}

func NewGitMetadataScore() *GitMetadataScore {
	return &GitMetadataScore{}
}

func NewLangEcoMetadata() *LangEcoMetadata {
	return &LangEcoMetadata{}
}

func NewDistMetadata() *DistMetadata {
	return &DistMetadata{}
}

func NewLinkScore(gitMetadataScore *GitMetadataScore, distScore *DistScore, langEcoScore *LangEcoScore) *LinkScore {
	return &LinkScore{
		LangEcoScore:     *langEcoScore,
		DistScore:        *distScore,
		GitMetadataScore: *gitMetadataScore,
	}
}

func NewDistScore() *DistScore {
	return &DistScore{}
}

func LogNormalize(value, threshold float64) float64 {
	return math.Log(value+1) / math.Log(math.Max(value, threshold)+1)
}

func Sigmoid(value, threshold float64) float64 {
	return 1 / (1 + SigmoidWeight*math.Exp(-1*(value-threshold)))
}

func FetchGitMetrics(ac storage.AppDatabaseContext) map[string]*GitMetadata {
	repo := repository.NewGitMetricsRepository(ac)
	linksIter, err := repo.Query()
	linksMap := make(map[string]*GitMetadata)
	if err != nil {
		log.Fatalf("Failed to fetch git links: %v", err)
	}
	for link := range linksIter {
		gitMetadata := NewGitMetadata()
		gitMetadata.ParseMetadata(link)
		linksMap[*link.GitLink] = gitMetadata
	}
	return linksMap
}

func FetchLangEcoMetadata(ac storage.AppDatabaseContext) map[string]*LangEcoScore {
	repo := repository.NewLangEcoLinkRepository(ac)
	LangEcoMap := make(map[string]*LangEcoScore)
	linksIter, err := repo.Query()
	if err != nil {
		log.Fatalf("Failed to fetch lang eco links: %v", err)
	}
	for link := range linksIter {
		langEcoMetadata := NewLangEcoMetadata()
		langEcoMetadata.ParseLangEcoMetadata(link)
		if exists, ok := LangEcoMap[*link.GitLink]; ok && exists != nil {
			LangEcoMap[*link.GitLink].LangEcosystems = append(LangEcoMap[*link.GitLink].LangEcosystems, link)
			LangEcoMap[*link.GitLink].LangEcoImpact += langEcoMetadata.LangEcoImpact * PackageWeight[langEcoMetadata.Type]
		} else {
			LangEcoMap[*link.GitLink] = &LangEcoScore{LangEcosystems: []*repository.LangEcosystem{link}, LangEcoImpact: langEcoMetadata.LangEcoImpact * PackageWeight[langEcoMetadata.Type]}
		}
	}
	return LangEcoMap
}

func FetchDistMetadata(ac storage.AppDatabaseContext) map[string]*DistScore {
	repo := repository.NewDistDependencyRepository(ac)
	distMap := make(map[string]*DistScore)
	linksIter, err := repo.Query()
	if err != nil {
		log.Fatalf("Failed to fetch dist links: %v", err)
	}
	for link := range linksIter {
		distMetadata := NewDistMetadata()
		distMetadata.PraseDistMetadata(link)
		coefficient := PackageList[distMetadata.Type] / PackageList[repository.Homebrew]
		if exists, ok := distMap[*link.GitLink]; ok && exists != nil {
			distMap[*link.GitLink].DistDependencies = append(distMap[*link.GitLink].DistDependencies, link)
			distMap[*link.GitLink].DistImpact += float64(coefficient) * distMetadata.DepImpact
			distMap[*link.GitLink].DistPageRank += float64(coefficient) * distMetadata.PageRank
		} else {
			distMap[*link.GitLink] = &DistScore{DistDependencies: []*repository.DistDependency{link}, DistImpact: float64(coefficient) * distMetadata.DepImpact, DistPageRank: float64(coefficient) * distMetadata.PageRank}
		}
	}
	return distMap
}
func FetchGitLink(ac storage.AppDatabaseContext) []string {
	repo := repository.NewAllGitLinkRepository(ac)
	linksIter, err := repo.Query()
	if err != nil {
		log.Fatalf("Failed to fetch git links: %v", err)
	}
	links := []string{}
	for link := range linksIter {
		links = append(links, link)
	}
	return links
}

func UpdatePackageList(ac storage.AppDatabaseContext) {
	repo := repository.NewDistDependencyRepository(ac)
	for distType := range PackageList {
		count, err := repo.QueryDistCountByType(distType)
		if err != nil {
			log.Fatalf("Failed to fetch dist links: %v", err)
		}
		PackageList[distType] = count
	}
}

func UpdateScore(ac storage.AppDatabaseContext, packageScore map[string]*LinkScore) {
	repo := repository.NewScoreRepository(ac)
	scores := []*repository.Score{}
	for link, linkScore := range packageScore {
		score := repository.Score{
			Score:            &linkScore.Score,
			GitLink:          &link,
			DistDependencies: linkScore.DistScore.DistDependencies,
			GitMetrics:       linkScore.GitMetadataScore.GitMetrics,
			LangEcosystems:   linkScore.LangEcoScore.LangEcosystems,
			DistScore:        &linkScore.DistScore.DistScore,
			LangScore:        &linkScore.LangEcoScore.LangEcoScore,
			GitScore:         &linkScore.GitMetadataScore.GitMetadataScore,
		}
		scores = append(scores, &score)
	}
	if err := repo.BatchInsertOrUpdate(scores); err != nil {
		log.Fatalf("Failed to update score: %v", err)
	}
}

func FetchDistMetadataSingle(ac storage.AppDatabaseContext, link string) map[string]*DistScore {
	repo := repository.NewDistDependencyRepository(ac)
	linksMap := []*repository.DistDependency{}
	distMap := make(map[string]*DistScore)
	for PackageType := range PackageList {
		distInfo, err := repo.GetByLink(link, int(PackageType))
		if err != nil {
			log.Fatalf("Failed to fetch dist links: %v", err)
		}
		linksMap = append(linksMap, distInfo)
	}
	for _, link := range linksMap {
		distMetadata := NewDistMetadata()
		distMetadata.PraseDistMetadata(link)
		coefficient := PackageList[distMetadata.Type] / PackageList[repository.Homebrew]
		if exists, ok := distMap[*link.GitLink]; ok && exists != nil {
			distMap[*link.GitLink].DistDependencies = append(distMap[*link.GitLink].DistDependencies, link)
			distMap[*link.GitLink].DistImpact += float64(coefficient) * distMetadata.DepImpact
			distMap[*link.GitLink].DistPageRank += float64(coefficient) * distMetadata.PageRank
		} else {
			distMap[*link.GitLink] = &DistScore{DistDependencies: []*repository.DistDependency{link}, DistImpact: float64(coefficient) * distMetadata.DepImpact, DistPageRank: float64(coefficient) * distMetadata.PageRank}
		}
	}
	return distMap
}

func FetchLangEcoMetadataSingle(ac storage.AppDatabaseContext, link string) map[string]*LangEcoScore {
	repo := repository.NewLangEcoLinkRepository(ac)
	langEcoMap := make(map[string]*LangEcoScore)
	linksIter, err := repo.QueryByLink(link)
	if err != nil {
		log.Fatalf("Failed to fetch lang eco links: %v", err)
	}
	for link := range linksIter {
		langEcoMetadata := NewLangEcoMetadata()
		langEcoMetadata.ParseLangEcoMetadata(link)
		if exists, ok := langEcoMap[*link.GitLink]; ok && exists != nil {
			langEcoMap[*link.GitLink].LangEcosystems = append(langEcoMap[*link.GitLink].LangEcosystems, link)
			langEcoMap[*link.GitLink].LangEcoImpact += langEcoMetadata.LangEcoImpact
		} else {
			langEcoMap[*link.GitLink] = &LangEcoScore{LangEcosystems: []*repository.LangEcosystem{link}, LangEcoImpact: langEcoMetadata.LangEcoImpact}
		}
	}
	return langEcoMap
}
func FetchGitMetricsSingle(ac storage.AppDatabaseContext, link string) map[string]*GitMetadata {
	repo := repository.NewGitMetricsRepository(ac)
	linkInfo, err := repo.QueryByLink(link)
	linksMap := make(map[string]*GitMetadata)
	if err != nil {
		log.Fatalf("Failed to fetch git links: %v", err)
	}
	gitMetadata := NewGitMetadata()
	gitMetadata.ParseMetadata(linkInfo)
	linksMap[*linkInfo.GitLink] = gitMetadata
	return linksMap
}
