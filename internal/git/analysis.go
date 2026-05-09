package git

import (
	"regexp"
	"sort"
	"strings"
)

var commitTypeOrder = []string{"feat", "fix", "build", "docs", "test", "refactor", "chore"}

var publicSymbolPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\+\s*func\s+[A-Z][A-Za-z0-9_]*\s*\(`),
	regexp.MustCompile(`^\+\s*type\s+[A-Z][A-Za-z0-9_]*\b`),
	regexp.MustCompile(`^\+\s*(const|var)\s+[A-Z][A-Za-z0-9_]*\b`),
	regexp.MustCompile(`^\+\s*export\s+(default\s+)?(async\s+)?(function|class|const|let|var|interface|type)\b`),
}

var fixSignalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bfix(e[ds])?\b`),
	regexp.MustCompile(`(?i)\bbug(s)?\b`),
	regexp.MustCompile(`(?i)\berr\s*!=\s*nil\b`),
	regexp.MustCompile(`(?i)(==|!=)\s*(nil|null)\b`),
	regexp.MustCompile(`(?i)\b(error|errors|failure|failed|panic|exception|timeout)\b`),
	regexp.MustCompile(`(?i)\b(validate|validation|guard|fallback|regression|missing|invalid)\b`),
	regexp.MustCompile(`(?i)\bnot\s+found\b`),
}

var featureSignalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(add|adds|added|create|creates|created)\b`),
	regexp.MustCompile(`(?i)\b(implement|implements|implemented|support|supports|enable|enables|enabled)\b`),
	regexp.MustCompile(`(?i)\b(endpoint|route|handler|command|feature)\b`),
}

// DiffAnalysis는 파일 diff에서 추출한 구조화된 분석 결과입니다.
type DiffAnalysis struct {
	TotalFiles          int
	SourceFiles         int
	TestFiles           int
	DocFiles            int
	ConfigFiles         int
	DependencyFiles     int
	NewFiles            int
	DeletedFiles        int
	ModifiedFiles       int
	NewSourceFiles      int
	ModifiedSourceFiles int
	DeletedSourceFiles  int
	AddedLines          int
	DeletedLines        int
	Signals             []string
	TypeScores          []CommitTypeScore
}

// CommitTypeScore는 특정 Conventional Commit 타입의 점수와 근거입니다.
type CommitTypeScore struct {
	Type    string
	Score   int
	Reasons []string
}

type scoreAccumulator struct {
	score   int
	reasons []string
}

// AnalyzeFiles는 파일 변경 목록을 요약하고 점수 기반 타입 추천 결과를 생성합니다.
func AnalyzeFiles(files []FileChange) DiffAnalysis {
	analysis := summarizeFiles(files)
	scores := newTypeScores()

	if analysis.TotalFiles == 0 {
		addScore(scores, "chore", 1, "no changed files")
		analysis.Signals = append(analysis.Signals, "no changed files")
		analysis.TypeScores = flattenScores(scores)
		return analysis
	}

	scoreByFileMix(&analysis, scores)
	scoreByContentSignals(files, &analysis, scores)
	if !hasPositiveScore(scores) {
		addScore(scores, "chore", 1, "general maintenance change")
		analysis.Signals = append(analysis.Signals, "general maintenance change")
	}

	analysis.Signals = uniqueSorted(analysis.Signals)
	analysis.TypeScores = flattenScores(scores)
	return analysis
}

// RecommendedType는 가장 높은 점수를 받은 커밋 타입을 반환합니다.
func (a DiffAnalysis) RecommendedType() string {
	if len(a.TypeScores) == 0 {
		return "chore"
	}
	return a.TypeScores[0].Type
}

func summarizeFiles(files []FileChange) DiffAnalysis {
	analysis := DiffAnalysis{TotalFiles: len(files)}

	for _, file := range files {
		addedLines, deletedLines := file.AddedLines, file.DeletedLines
		if addedLines == 0 && deletedLines == 0 && file.Changes != "" {
			addedLines, deletedLines = countChangedLines(file.Changes)
		}

		analysis.AddedLines += addedLines
		analysis.DeletedLines += deletedLines

		if file.IsNew {
			analysis.NewFiles++
		} else if file.IsDeleted {
			analysis.DeletedFiles++
		} else {
			analysis.ModifiedFiles++
		}

		switch file.FileType {
		case FileTypeSource:
			analysis.SourceFiles++
			if file.IsNew {
				analysis.NewSourceFiles++
			} else if file.IsDeleted {
				analysis.DeletedSourceFiles++
			} else {
				analysis.ModifiedSourceFiles++
			}
		case FileTypeTest:
			analysis.TestFiles++
		case FileTypeDoc:
			analysis.DocFiles++
		case FileTypeConfig:
			if isDependencyFile(file.Path) {
				analysis.DependencyFiles++
			} else {
				analysis.ConfigFiles++
			}
		}
	}

	return analysis
}

func scoreByFileMix(analysis *DiffAnalysis, scores map[string]*scoreAccumulator) {
	onlyDocs := analysis.DocFiles > 0 && analysis.SourceFiles == 0 && analysis.TestFiles == 0 && analysis.ConfigFiles == 0 && analysis.DependencyFiles == 0
	onlyTests := analysis.TestFiles > 0 && analysis.SourceFiles == 0 && analysis.DocFiles == 0 && analysis.ConfigFiles == 0 && analysis.DependencyFiles == 0
	onlyDeps := analysis.DependencyFiles > 0 && analysis.SourceFiles == 0 && analysis.TestFiles == 0 && analysis.DocFiles == 0 && analysis.ConfigFiles == 0
	onlyConfig := analysis.ConfigFiles > 0 && analysis.SourceFiles == 0 && analysis.TestFiles == 0 && analysis.DocFiles == 0 && analysis.DependencyFiles == 0
	depsAndConfig := analysis.DependencyFiles > 0 && analysis.ConfigFiles > 0 && analysis.SourceFiles == 0 && analysis.TestFiles == 0 && analysis.DocFiles == 0

	switch {
	case onlyDocs:
		addScore(scores, "docs", 100, "only documentation files changed")
		analysis.Signals = append(analysis.Signals, "documentation-only change")
	case onlyTests:
		addScore(scores, "test", 100, "only test files changed")
		analysis.Signals = append(analysis.Signals, "test-only change")
	case onlyDeps:
		addScore(scores, "build", 100, "only dependency files changed")
		analysis.Signals = append(analysis.Signals, "dependency-only change")
	case onlyConfig:
		addScore(scores, "chore", 80, "only configuration files changed")
		analysis.Signals = append(analysis.Signals, "configuration-only change")
	case depsAndConfig:
		addScore(scores, "build", 70, "dependency and build configuration changed")
		addScore(scores, "chore", 25, "configuration changed with dependencies")
		analysis.Signals = append(analysis.Signals, "dependency/configuration change")
	}

	if analysis.SourceFiles > 0 {
		if analysis.NewSourceFiles > 0 {
			addScore(scores, "feat", 60, "new source files added")
			analysis.Signals = append(analysis.Signals, "new source files")
		}
		if analysis.ModifiedSourceFiles > 0 {
			addScore(scores, "refactor", 25, "existing source files modified")
			analysis.Signals = append(analysis.Signals, "existing source modified")
		}
		if analysis.DeletedSourceFiles > 0 && analysis.NewSourceFiles == 0 {
			addScore(scores, "refactor", 20, "source files removed")
			analysis.Signals = append(analysis.Signals, "source files removed")
		}
	}

	if analysis.TestFiles > 0 && analysis.SourceFiles > 0 {
		addScore(scores, "test", 10, "tests changed with source")
		analysis.Signals = append(analysis.Signals, "tests changed with source")
	}
	if analysis.DocFiles > 0 && analysis.SourceFiles > 0 {
		addScore(scores, "docs", 5, "documentation updated with source")
	}
	if analysis.DependencyFiles > 0 && analysis.SourceFiles > 0 {
		addScore(scores, "build", 5, "dependency files changed with source")
	}
}

func scoreByContentSignals(files []FileChange, analysis *DiffAnalysis, scores map[string]*scoreAccumulator) {
	for _, file := range files {
		if file.FileType != FileTypeSource {
			continue
		}

		added := addedDiffLines(file.Changes)
		changedText := strings.Join(changedDiffLines(file.Changes), "\n")
		addedText := strings.Join(added, "\n")

		if len(added) > 0 && hasPublicSymbol(added) {
			addScore(scores, "feat", 35, "public source symbol added")
			analysis.Signals = append(analysis.Signals, "public API/symbol added")
		}
		if matchesAnyPattern(changedText, fixSignalPatterns) {
			addScore(scores, "fix", 55, "bug or error-handling signal in source diff")
			analysis.Signals = append(analysis.Signals, "bug/error-handling signal")
		}
		if file.IsNew && matchesAnyPattern(addedText, featureSignalPatterns) {
			addScore(scores, "feat", 20, "feature-oriented source additions")
			analysis.Signals = append(analysis.Signals, "feature-oriented additions")
		}
	}
}

func newTypeScores() map[string]*scoreAccumulator {
	scores := make(map[string]*scoreAccumulator, len(commitTypeOrder))
	for _, commitType := range commitTypeOrder {
		scores[commitType] = &scoreAccumulator{}
	}
	return scores
}

func addScore(scores map[string]*scoreAccumulator, commitType string, score int, reason string) {
	acc, ok := scores[commitType]
	if !ok {
		acc = &scoreAccumulator{}
		scores[commitType] = acc
	}
	acc.score += score
	if reason != "" {
		acc.reasons = append(acc.reasons, reason)
	}
}

func hasPositiveScore(scores map[string]*scoreAccumulator) bool {
	for _, acc := range scores {
		if acc != nil && acc.score > 0 {
			return true
		}
	}
	return false
}

func flattenScores(scores map[string]*scoreAccumulator) []CommitTypeScore {
	type priorityScore struct {
		CommitTypeScore
		priority int
	}

	var flattened []priorityScore
	for priority, commitType := range commitTypeOrder {
		acc := scores[commitType]
		if acc == nil {
			continue
		}
		flattened = append(flattened, priorityScore{
			CommitTypeScore: CommitTypeScore{
				Type:    commitType,
				Score:   acc.score,
				Reasons: uniqueSorted(acc.reasons),
			},
			priority: priority,
		})
	}

	sort.SliceStable(flattened, func(i, j int) bool {
		if flattened[i].Score == flattened[j].Score {
			return flattened[i].priority < flattened[j].priority
		}
		return flattened[i].Score > flattened[j].Score
	})

	result := make([]CommitTypeScore, len(flattened))
	for i, score := range flattened {
		result[i] = score.CommitTypeScore
	}
	return result
}

func countChangedLines(changes string) (added int, deleted int) {
	for _, line := range strings.Split(changes, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.HasPrefix(line, "+") {
			added++
			continue
		}
		if strings.HasPrefix(line, "-") {
			deleted++
		}
	}
	return added, deleted
}

func changedDiffLines(changes string) []string {
	var lines []string
	for _, line := range strings.Split(changes, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			lines = append(lines, line)
		}
	}
	return lines
}

func addedDiffLines(changes string) []string {
	var lines []string
	for _, line := range strings.Split(changes, "\n") {
		if strings.HasPrefix(line, "+++") {
			continue
		}
		if strings.HasPrefix(line, "+") {
			lines = append(lines, line)
		}
	}
	return lines
}

func hasPublicSymbol(lines []string) bool {
	for _, line := range lines {
		for _, pattern := range publicSymbolPatterns {
			if pattern.MatchString(line) {
				return true
			}
		}
	}
	return false
}

func matchesAnyPattern(text string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(values))
	var result []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
