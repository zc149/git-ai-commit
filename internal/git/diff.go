package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"git-ai-commit/internal/worker"
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
				// 새 소스 파일 추가는 feat에 강력한 점수
				typeScore["feat"] += 10
			} else {
				// 기존 소스 파일 수정
				typeScore["refactor"] += 3
				typeScore["fix"] += 1
			}

		case FileTypeTest:
			if file.IsNew {
				typeScore["test"] += 8
			} else {
				typeScore["test"] += 3
			}

		case FileTypeDoc:
			typeScore["docs"] += 8

		case FileTypeConfig:
			if isDependencyFile(path) {
				hasDependencyFile = true
				// 의존성 변경 = build
				typeScore["build"] += 5
			} else {
				hasRegularConfig = true
				// 일반 설정 = chore
				typeScore["chore"] += 3
			}
		}
	}

	// 새 디렉토리가 2개 이상 = 새 기능 추가
	if len(newDirectories) >= 2 {
		typeScore["feat"] += 20
	}

	// 새 소스 파일이 3개 이상 = 새 기능
	if newSourceFileCount >= 3 {
		typeScore["feat"] += 15
	}

	// 의존성 파일만 변경됨 = build
	if hasDependencyFile && sourceFileCount == 0 && !hasRegularConfig {
		typeScore["build"] += 10
		typeScore["chore"] -= 5
	}

	// 일반 설정 파일만 변경됨 = chore
	if hasRegularConfig && sourceFileCount == 0 && !hasDependencyFile {
		typeScore["chore"] += 10
		typeScore["build"] -= 5
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

// isDependencyFile은 파일이 의존성 관련 파일인지 확인합니다.
func isDependencyFile(path string) bool {
	base := filepath.Base(path)

	// Node.js/JavaScript
	if base == "package.json" || base == "package-lock.json" || base == "yarn.lock" || base == "pnpm-lock.yaml" {
		return true
	}

	// Go
	if base == "go.mod" || base == "go.sum" {
		return true
	}

	// Python
	if base == "requirements.txt" || base == "Pipfile" || base == "poetry.lock" || base == "pyproject.toml" {
		return true
	}

	// Java/Maven/Gradle
	if base == "pom.xml" || base == "build.gradle" || base == "build.gradle.kts" || base == "gradle.properties" {
		return true
	}

	// Ruby
	if base == "Gemfile" || base == "Gemfile.lock" {
		return true
	}

	// PHP
	if base == "composer.json" || base == "composer.lock" {
		return true
	}

	// Rust
	if base == "Cargo.toml" || base == "Cargo.lock" {
		return true
	}

	// .NET
	if strings.HasSuffix(base, ".csproj") || base == "packages.config" {
		return true
	}

	// Swift/CocoaPods
	if base == "Podfile" || base == "Podfile.lock" || base == "Package.swift" {
		return true
	}

	// Dart/Flutter
	if base == "pubspec.yaml" || base == "pubspec.lock" {
		return true
	}

	// Composer (PHP)
	if base == "composer.json" || base == "composer.lock" {
		return true
	}

	// CMake
	if base == "CMakeLists.txt" || base == "CMakeCache.txt" {
		return true
	}

	// Conan
	if base == "conanfile.txt" || base == "conanfile.py" {
		return true
	}

	// Vcpkg
	if base == "vcpkg.json" {
		return true
	}

	// Bazel
	if base == "WORKSPACE" || base == "BUILD" || base == "BUILD.bazel" {
		return true
	}

	// Buck
	if base == "BUCK" {
		return true
	}

	// Leiningen (Clojure)
	if base == "project.clj" {
		return true
	}

	// Mix (Elixir)
	if base == "mix.exs" {
		return true
	}

	// Rebar3 (Erlang)
	if base == "rebar.config" {
		return true
	}

	// NuGet.config
	if base == "NuGet.config" || base == "nuget.config" {
		return true
	}

	return false
}

// InferScopes는 파일 경로에서 scope를 추론합니다.
func InferScopes(files []FileChange) []string {
	if len(files) == 0 {
		return []string{}
	}

	// 1. 디렉토리별 파일 수 집계
	dirCounts := make(map[string]int)
	sourceDirs := make(map[string]bool) // 소스 파일이 있는 디렉토리
	configDirs := make(map[string]bool) // 설정 파일만 있는 디렉토리

	for _, file := range files {
		path := file.Path
		parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

		// 각 레벨의 디렉토리 집계
		for i := 1; i < len(parts); i++ {
			dir := strings.Join(parts[:i], string(filepath.Separator))
			dirCounts[dir]++
		}

		// 파일 유형별 분류
		if file.FileType == FileTypeSource {
			// 소스 파일의 모든 상위 디렉토리 표시
			for i := 1; i < len(parts); i++ {
				dir := strings.Join(parts[:i], string(filepath.Separator))
				sourceDirs[dir] = true
			}
		} else if file.FileType == FileTypeConfig {
			// 설정 파일만 있는 디렉토리 확인
			for i := 1; i < len(parts); i++ {
				dir := strings.Join(parts[:i], string(filepath.Separator))
				if !sourceDirs[dir] {
					configDirs[dir] = true
				}
			}
		}
	}

	// 2. 최대 공통 디렉토리 찾기
	commonPrefix := findCommonPrefix(files)

	// 3. 주요 scope 결정
	var scopes []string

	if commonPrefix != "" {
		// 공통 접두사가 있으면 그걸 메인 scope로
		scopes = append(scopes, simplifyScopeName(commonPrefix))
	} else {
		// 공통 접두사가 없으면 소스 파일이 있는 디렉토리 중에서
		// 가장 파일이 많은 것 선택
		maxCount := 0
		var primaryDir string
		for dir, count := range dirCounts {
			if sourceDirs[dir] && count > maxCount {
				maxCount = count
				primaryDir = dir
			}
		}
		if primaryDir != "" {
			scopes = append(scopes, simplifyScopeName(primaryDir))
		}
	}

	// 4. 2번째 scope (필요한 경우)
	// 메인 scope의 바로 하위 디렉토리 중 파일이 많은 것
	if len(scopes) > 0 && len(dirCounts) > 5 {
		primaryScope := scopes[0]
		for dir, count := range dirCounts {
			if strings.HasPrefix(dir, primaryScope+string(filepath.Separator)) &&
				sourceDirs[dir] &&
				count >= 3 {
				// 하위 디렉토리 이름만 추출
				subDir := strings.TrimPrefix(dir, primaryScope+string(filepath.Separator))
				if !strings.Contains(subDir, string(filepath.Separator)) {
					scopes = append(scopes, simplifyScopeName(subDir))
					break
				}
			}
		}
	}

	// 5. scope 너무 많으면 제한 (최대 2개)
	if len(scopes) > 2 {
		scopes = scopes[:2]
	}

	return scopes
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
