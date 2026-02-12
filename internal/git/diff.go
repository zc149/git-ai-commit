package git

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"git-ai-commit/internal/worker"
)

// 의존성 파일 Map (패키지 초기화 시 한 번만 생성)
var (
	dependencyFiles map[string]bool
	depFileOnce     sync.Once
)

// FileType은 파일의 유형을 나타냅니다.
type FileType int

const (
	FileTypeSource FileType = iota
	FileTypeTest
	FileTypeDoc
	FileTypeConfig
)

func (ft FileType) String() string {
	switch ft {
	case FileTypeSource:
		return "source"
	case FileTypeTest:
		return "test"
	case FileTypeDoc:
		return "doc"
	case FileTypeConfig:
		return "config"
	default:
		return "unknown"
	}
}

// FileChange는 단일 파일의 변경 정보를 담습니다.
type FileChange struct {
	Path      string   // 파일 경로
	FileType  FileType // 파일 타입
	IsNew     bool     // 새 파일 여부
	IsDeleted bool     // 삭제된 파일 여부
	Changes   string   // 변경된 내용 (diff 내용)
}

// DiffResult는 파싱된 diff 결과를 담습니다.
type DiffResult struct {
	Files      []FileChange // 변경된 파일 목록
	CommitType string       // 추론된 커밋 타입
	Scopes     []string     // 추론된 scope 목록
	RawDiff    string       // 원본 diff 문자열
}

// GetCachedDiff는 git diff --cached 명령을 실행하여 결과를 반환합니다.
func GetCachedDiff() (*DiffResult, error) {
	// git diff --cached 실행
	cmd := exec.Command("git", "diff", "--cached")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff --cached failed: %w, stderr: %s", err, stderr.String())
	}

	rawDiff := stdout.String()

	// 빈 diff 처리
	if strings.TrimSpace(rawDiff) == "" {
		return &DiffResult{
			Files:      []FileChange{},
			CommitType: "",
			Scopes:     []string{},
			RawDiff:    rawDiff,
		}, nil
	}

	// 병렬 diff 파싱 시도
	fileCount := estimateFileCount(rawDiff)
	workers := worker.GetOptimalWorkerCount(fileCount)

	parsedFiles, err := worker.ParseDiffParallel(rawDiff, workers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	var result *DiffResult

	// 병렬 파싱 결과가 있으면 사용, 없으면 순차 파싱
	if parsedFiles != nil {
		result = &DiffResult{
			Files:   convertParsedFiles(parsedFiles),
			RawDiff: rawDiff,
		}
	} else {
		// 순차 파싱
		result, err = ParseDiff(rawDiff)
		if err != nil {
			return nil, fmt.Errorf("failed to parse diff: %w", err)
		}
		result.RawDiff = rawDiff
	}

	result.CommitType = InferCommitType(result.Files)
	result.Scopes = InferScopes(result.Files)

	return result, nil
}

// convertParsedFiles는 worker.ParsedFile을 git.FileChange로 변환합니다.
func convertParsedFiles(parsedFiles []worker.ParsedFile) []FileChange {
	files := make([]FileChange, len(parsedFiles))
	for i, pf := range parsedFiles {
		files[i] = FileChange{
			Path:      pf.Path,
			FileType:  FileType(pf.FileType),
			IsNew:     pf.IsNew,
			IsDeleted: pf.IsDeleted,
			Changes:   pf.Changes,
		}
	}
	return files
}

// estimateFileCount는 diff에서 대략적인 파일 수를 추정합니다.
func estimateFileCount(diff string) int {
	count := 0
	scanner := bufio.NewScanner(strings.NewReader(diff))

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "diff --git") {
			count++
		}
	}

	return count
}

// ParseDiff는 diff 문자열을 구조체로 변환합니다.
func ParseDiff(diff string) (*DiffResult, error) {
	result := &DiffResult{
		Files: []FileChange{},
	}

	scanner := bufio.NewScanner(strings.NewReader(diff))
	var currentFile *FileChange
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		// 새 파일 헤더 감지
		if strings.HasPrefix(line, "diff --git") {
			// 이전 파일이 있다면 저장
			if currentFile != nil {
				currentFile.Changes = strings.Join(lines, "\n")
				result.Files = append(result.Files, *currentFile)
				lines = []string{}
			}

			// 새 파일 정보 추출
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				path := strings.TrimPrefix(parts[3], "b/")
				currentFile = &FileChange{
					Path:      path,
					FileType:  ClassifyFileType(path),
					IsNew:     false,
					IsDeleted: false,
				}
			}
			continue
		}

		// 새 파일 표시
		if strings.HasPrefix(line, "new file mode") {
			if currentFile != nil {
				currentFile.IsNew = true
			}
			continue
		}

		// 삭제된 파일 표시
		if strings.HasPrefix(line, "deleted file mode") {
			if currentFile != nil {
				currentFile.IsDeleted = true
			}
			continue
		}

		// 변경 라인 수집
		if currentFile != nil {
			lines = append(lines, line)
		}
	}

	// 마지막 파일 저장
	if currentFile != nil {
		currentFile.Changes = strings.Join(lines, "\n")
		result.Files = append(result.Files, *currentFile)
	}

	return result, nil
}

// ClassifyFileType은 파일 경로에서 파일 타입을 결정합니다.
func ClassifyFileType(path string) FileType {
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	// 테스트 파일
	if strings.HasSuffix(base, "_test.go") ||
		strings.HasSuffix(base, ".spec.js") ||
		strings.HasSuffix(base, ".test.ts") ||
		strings.HasSuffix(base, ".test.jsx") ||
		strings.HasSuffix(base, ".spec.tsx") {
		return FileTypeTest
	}

	// 문서 파일
	if base == "README.md" ||
		base == "CHANGELOG.md" ||
		base == "CONTRIBUTING.md" ||
		ext == ".md" ||
		ext == ".txt" ||
		ext == ".rst" {
		return FileTypeDoc
	}

	// 설정 파일
	if base == "package.json" ||
		base == "package-lock.json" ||
		base == "go.mod" ||
		base == "go.sum" ||
		base == "Cargo.toml" ||
		base == "pom.xml" ||
		base == "build.gradle" ||
		base == "requirements.txt" ||
		base == "Makefile" ||
		base == "Dockerfile" ||
		ext == ".yml" ||
		ext == ".yaml" ||
		ext == ".toml" ||
		ext == ".json" ||
		ext == ".xml" ||
		ext == ".ini" ||
		ext == ".conf" ||
		ext == ".cfg" {
		return FileTypeConfig
	}

	// 기본적으로 소스 파일
	return FileTypeSource
}

// InferCommitType은 파일 변화를 기반으로 커밋 타입을 추론합니다.
func InferCommitType(files []FileChange) string {
	if len(files) == 0 {
		return "chore"
	}

	// 파일 유형별 가중치 점수
	typeScore := make(map[string]int)
	typeScore["feat"] = 0
	typeScore["fix"] = 0
	typeScore["build"] = 0
	typeScore["docs"] = 0
	typeScore["test"] = 0
	typeScore["refactor"] = 0
	typeScore["chore"] = 0

	sourceFileCount := 0
	newSourceFileCount := 0
	modifiedSourceFileCount := 0
	newDirectories := make(map[string]bool)

	hasDependencyFile := false
	hasRegularConfig := false

	for _, file := range files {
		path := file.Path
		parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

		// 새 디렉토리 감지
		if len(parts) >= 2 && file.IsNew {
			dir := strings.Join(parts[:len(parts)-1], string(filepath.Separator))
			newDirectories[dir] = true
		}

		switch file.FileType {
		case FileTypeSource:
			sourceFileCount++
			if file.IsNew {
				newSourceFileCount++
				// 새 소스 파일 추가는 feat에 매우 강력한 점수
				typeScore["feat"] += 15
			} else if !file.IsDeleted {
				modifiedSourceFileCount++
				typeScore["refactor"] += 10
				typeScore["fix"] += 5
			}

		case FileTypeTest:
			if file.IsNew {
				typeScore["test"] += 8
			} else {
				typeScore["test"] += 3
			}

		case FileTypeDoc:
			typeScore["docs"] += 3

		case FileTypeConfig:
			if isDependencyFile(path) {
				hasDependencyFile = true
				// 의존성 변경 점수는 낮게 설정 (소스 파일이 우선)
				typeScore["build"] += 2
			} else {
				hasRegularConfig = true
				// 일반 설정 점수도 낮게 설정
				typeScore["chore"] += 2
			}
		}
	}

	// 새 디렉토리가 2개 이상 = 새 기능 추가 (가중치 증가)
	if len(newDirectories) >= 2 {
		typeScore["feat"] += 30
	}

	// 새 소스 파일이 2개 이상 = 새 기능 (임계값 낮춤, 가중치 증가)
	if newSourceFileCount >= 2 {
		typeScore["feat"] += 30
	}

	// 의존성 파일만 변경됨 (소스 파일이 없는 경우) = build
	if hasDependencyFile && sourceFileCount == 0 && !hasRegularConfig {
		typeScore["build"] += 15
		typeScore["chore"] -= 5
	}

	// 일반 설정 파일만 변경됨 (소스 파일이 없는 경우) = chore
	if hasRegularConfig && sourceFileCount == 0 && !hasDependencyFile {
		typeScore["chore"] += 15
		typeScore["build"] -= 5
	}

	// 소스 파일이 있는 경우: 의존성/설정 변경은 무시하고 소스 파일 유형 우선
	// 새 소스 파일이 있으면 무조건 feat
	if newSourceFileCount > 0 {
		typeScore["feat"] += 50
		typeScore["build"] -= 20
		typeScore["chore"] -= 20
	}

	// 최대 점수 타입 선택
	maxScore := 0
	var bestType string
	for commitType, score := range typeScore {
		if score > maxScore {
			maxScore = score
			bestType = commitType
		}
	}

	if bestType == "" {
		return "chore"
	}
	return bestType
}

// init는 패키지 초기화 시 의존성 파일 Map을 생성합니다.
func init() {
	depFileOnce.Do(func() {
		dependencyFiles = make(map[string]bool)

		// Node.js/JavaScript
		dependencyFiles["package.json"] = true
		dependencyFiles["package-lock.json"] = true
		dependencyFiles["yarn.lock"] = true
		dependencyFiles["pnpm-lock.yaml"] = true

		// Go
		dependencyFiles["go.mod"] = true
		dependencyFiles["go.sum"] = true

		// Python
		dependencyFiles["requirements.txt"] = true
		dependencyFiles["Pipfile"] = true
		dependencyFiles["poetry.lock"] = true
		dependencyFiles["pyproject.toml"] = true

		// Java/Maven/Gradle
		dependencyFiles["pom.xml"] = true
		dependencyFiles["build.gradle"] = true
		dependencyFiles["build.gradle.kts"] = true
		dependencyFiles["gradle.properties"] = true

		// Ruby
		dependencyFiles["Gemfile"] = true
		dependencyFiles["Gemfile.lock"] = true

		// PHP
		dependencyFiles["composer.json"] = true
		dependencyFiles["composer.lock"] = true

		// Rust
		dependencyFiles["Cargo.toml"] = true
		dependencyFiles["Cargo.lock"] = true

		// .NET
		dependencyFiles["packages.config"] = true
		dependencyFiles["NuGet.config"] = true
		dependencyFiles["nuget.config"] = true

		// Swift/CocoaPods
		dependencyFiles["Podfile"] = true
		dependencyFiles["Podfile.lock"] = true
		dependencyFiles["Package.swift"] = true

		// Dart/Flutter
		dependencyFiles["pubspec.yaml"] = true
		dependencyFiles["pubspec.lock"] = true

		// CMake
		dependencyFiles["CMakeLists.txt"] = true
		dependencyFiles["CMakeCache.txt"] = true

		// Conan
		dependencyFiles["conanfile.txt"] = true
		dependencyFiles["conanfile.py"] = true

		// Vcpkg
		dependencyFiles["vcpkg.json"] = true

		// Bazel
		dependencyFiles["WORKSPACE"] = true
		dependencyFiles["BUILD"] = true
		dependencyFiles["BUILD.bazel"] = true

		// Buck
		dependencyFiles["BUCK"] = true

		// Leiningen (Clojure)
		dependencyFiles["project.clj"] = true

		// Mix (Elixir)
		dependencyFiles["mix.exs"] = true

		// Rebar3 (Erlang)
		dependencyFiles["rebar.config"] = true
	})
}

// isDependencyFile은 파일이 의존성 관련 파일인지 확인합니다.
// O(1) 성능: Map 기반 룩업 사용
func isDependencyFile(path string) bool {
	base := filepath.Base(path)

	// Map 룩업 (O(1))
	if dependencyFiles[base] {
		return true
	}

	// .NET 프로젝트 파일 확장자 체크
	if strings.HasSuffix(base, ".csproj") || strings.HasSuffix(base, ".vbproj") || strings.HasSuffix(base, ".fsproj") {
		return true
	}

	return false
}

// InferScopes는 파일 경로에서 scope를 추론합니다.
func InferScopes(files []FileChange) []string {
	if len(files) == 0 {
		return []string{}
	}

	// 1. 파일 유형별 분류
	var sourceFiles []FileChange
	var configFiles []FileChange
	var dependencyFiles []FileChange
	var docFiles []FileChange
	var testFiles []FileChange

	for _, file := range files {
		switch file.FileType {
		case FileTypeSource:
			sourceFiles = append(sourceFiles, file)
		case FileTypeConfig:
			if isDependencyFile(file.Path) {
				dependencyFiles = append(dependencyFiles, file)
			} else {
				configFiles = append(configFiles, file)
			}
		case FileTypeDoc:
			docFiles = append(docFiles, file)
		case FileTypeTest:
			testFiles = append(testFiles, file)
		}
	}

	// 2. 소스 파일이 있는 경우: 소스 파일 기반으로 scope 결정
	if len(sourceFiles) > 0 {
		return inferScopeFromSourceFiles(sourceFiles)
	}

	// 3. 소스 파일이 없는 경우: 다른 파일 유형으로 scope 결정
	if len(dependencyFiles) > 0 {
		// 의존성 파일만 있는 경우
		return inferScopeFromConfigFiles(append(dependencyFiles, configFiles...))
	}

	if len(configFiles) > 0 {
		// 일반 설정 파일만 있는 경우
		return inferScopeFromConfigFiles(configFiles)
	}

	if len(docFiles) > 0 {
		return []string{"docs"}
	}

	if len(testFiles) > 0 {
		return []string{"test"}
	}

	return []string{}
}

// inferScopeFromSourceFiles는 소스 파일에서 scope를 추론합니다.
func inferScopeFromSourceFiles(sourceFiles []FileChange) []string {
	// 최상위 디렉토리별 소스 파일 수 집계
	topLevelDirs := make(map[string]int)
	for _, file := range sourceFiles {
		path := file.Path
		parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
		if len(parts) > 0 {
			topDir := parts[0]
			topLevelDirs[topDir]++
		}
	}

	// 최대 공통 디렉토리 찾기
	commonPrefix := findCommonPrefix(sourceFiles)

	// 주요 scope 결정
	var scopes []string

	if commonPrefix != "" {
		// 공통 접두사가 있으면 그걸 메인 scope로
		scopes = append(scopes, simplifyScopeName(commonPrefix))
	} else if len(topLevelDirs) >= 3 {
		// 3개 이상의 다른 최상위 디렉토리에 소스 파일이 있으면 "multiple" 사용
		scopes = append(scopes, "multiple")
	} else {
		// 소스 파일이 가장 많은 최상위 디렉토리 선택
		maxCount := 0
		var primaryDir string
		for dir, count := range topLevelDirs {
			if count > maxCount {
				maxCount = count
				primaryDir = dir
			}
		}
		if primaryDir != "" {
			scopes = append(scopes, simplifyScopeName(primaryDir))
		}
	}

	return scopes
}

// inferScopeFromConfigFiles는 설정 파일에서 scope를 추론합니다.
func inferScopeFromConfigFiles(configFiles []FileChange) []string {
	if len(configFiles) == 0 {
		return []string{}
	}

	// 모든 파일이 같은 최상위 디렉토리에 있는지 확인
	topLevelDirs := make(map[string]bool)
	for _, file := range configFiles {
		path := file.Path
		parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
		if len(parts) > 0 {
			topLevelDirs[parts[0]] = true
		}
	}

	// 하나의 최상위 디렉토리만 있으면 그걸 사용
	if len(topLevelDirs) == 1 {
		for dir := range topLevelDirs {
			return []string{simplifyScopeName(dir)}
		}
	}

	// 여러 디렉토리에 파일이 있으면:
	// - 의존성 파일이 포함되어 있는지 확인
	hasDependency := false
	for _, file := range configFiles {
		if isDependencyFile(file.Path) {
			hasDependency = true
			break
		}
	}

	if hasDependency {
		return []string{"config"}
	}

	// 일반 설정 파일만 있고 여러 디렉토리에 있으면 "config" 사용
	return []string{"config"}
}

// findCommonPrefix는 모든 파일의 공통 경로 접두사를 찾습니다.
func findCommonPrefix(files []FileChange) string {
	if len(files) == 0 {
		return ""
	}

	paths := make([][]string, len(files))
	for i, file := range files {
		paths[i] = strings.Split(filepath.Clean(file.Path), string(filepath.Separator))
	}

	common := paths[0]
	for i := 1; i < len(paths); i++ {
		common = commonPrefix(common, paths[i])
		if len(common) == 0 {
			return ""
		}
	}

	// 마지막이 파일 이름인 경우 제거
	if len(common) > 1 {
		return strings.Join(common[:len(common)-1], "/")
	} else if len(common) == 1 && isDirectoryName(common[0]) {
		return common[0]
	}
	return ""
}

// commonPrefix는 두 경로 배열의 공통 접두사를 반환합니다.
func commonPrefix(a, b []string) []string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	i := 0
	for i < minLen && a[i] == b[i] {
		i++
	}

	return a[:i]
}

// isDirectoryName은 이름이 디렉토리 이름일 가능성이 있는지 확인합니다.
func isDirectoryName(name string) bool {
	// 확장자가 없거나 일반적인 확장자가 아니면 디렉토리로 간주
	ext := filepath.Ext(name)
	if ext == "" {
		return true
	}

	// 일반적인 파일 확장자
	commonExts := []string{".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".kt", ".rb", ".php",
		".cs", ".cpp", ".c", ".h", ".hpp", ".swift", ".rs", ".dart", ".lua", ".r",
		".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".conf", ".cfg", ".md", ".txt"}

	for _, commonExt := range commonExts {
		if ext == commonExt {
			return false
		}
	}

	return true
}

// simplifyScopeName은 scope 이름을 단순화합니다.
func simplifyScopeName(scope string) string {
	// 너무 긴 scope는 줄이기
	if len(scope) > 20 {
		parts := strings.Split(scope, "/")
		if len(parts) > 1 {
			return parts[len(parts)-1] // 마지막 부분만
		}
		return scope[:20]
	}
	return scope
}

// CalculateDiffHash는 diff 내용의 SHA256 해시를 계산합니다.
func CalculateDiffHash(rawDiff string) string {
	hash := sha256.Sum256([]byte(rawDiff))
	return hex.EncodeToString(hash[:])
}
