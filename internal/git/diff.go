package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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

	// diff 파싱
	result, err := ParseDiff(rawDiff)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	result.RawDiff = rawDiff
	result.CommitType = InferCommitType(result.Files)
	result.Scopes = InferScopes(result.Files)

	return result, nil
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

	hasTest := false
	hasDoc := false
	hasConfig := false
	hasNewFile := false
	hasDeletedFile := false

	for _, file := range files {
		switch file.FileType {
		case FileTypeTest:
			hasTest = true
		case FileTypeDoc:
			hasDoc = true
		case FileTypeConfig:
			hasConfig = true
		}

		if file.IsNew {
			hasNewFile = true
		}
		if file.IsDeleted {
			hasDeletedFile = true
		}
	}

	// 우선순위 규칙에 따른 타입 결정
	if hasTest && !hasDoc && !hasConfig {
		return "test"
	}

	if hasDoc && !hasTest && !hasConfig {
		return "docs"
	}

	if hasConfig && !hasDoc && !hasTest {
		return "build"
	}

	// 의존성 변경 확인
	for _, file := range files {
		if file.FileType == FileTypeConfig &&
			(strings.Contains(file.Path, "package.json") ||
				strings.Contains(file.Path, "go.mod") ||
				strings.Contains(file.Path, "Cargo.toml") ||
				strings.Contains(file.Path, "pom.xml")) {
			// 변경 내용에서 버전 변경 감지
			if strings.Contains(file.Changes, "+") &&
				(strings.Contains(file.Changes, "version") ||
					strings.Contains(file.Changes, "\"")) {
				return "build"
			}
		}
	}

	// 새 파일만 추가된 경우
	if hasNewFile && !hasDeletedFile {
		return "feat"
	}

	// 변경 내용에서 버그 관련 키워드 검사
	bugKeywords := []string{"fix", "bug", "error", "issue", "resolve"}
	for _, file := range files {
		for _, keyword := range bugKeywords {
			if strings.Contains(strings.ToLower(file.Changes), keyword) {
				return "fix"
			}
		}
	}

	// 기본값
	return "refactor"
}

// InferScopes는 파일 경로에서 scope를 추론합니다.
func InferScopes(files []FileChange) []string {
	if len(files) == 0 {
		return []string{}
	}

	scopes := make(map[string]bool)
	primaryScope := ""

	for _, file := range files {
		// 경로 파싱
		parts := strings.Split(filepath.Clean(file.Path), string(filepath.Separator))

		// 첫 번째 디렉토리를 scope로 사용
		if len(parts) > 0 && parts[0] != "" {
			// 일반적인 패키지 구조의 경우
			if len(parts) > 1 && (parts[0] == "src" || parts[0] == "lib" || parts[0] == "app") {
				if len(parts) > 2 {
					scope := parts[1]
					scopes[scope] = true
					if primaryScope == "" {
						primaryScope = scope
					}
				}
			} else {
				// 그 외 경우 첫 번째 디렉토리 사용
				scope := parts[0]
				// 내부 패키지 폴더가 아니라면 추가
				if !strings.HasPrefix(scope, ".") &&
					scope != "internal" &&
					scope != "vendor" &&
					scope != "node_modules" {
					scopes[scope] = true
					if primaryScope == "" {
						primaryScope = scope
					}
				}
			}
		}
	}

	// scope가 없으면 빈 슬라이스 반환
	if len(scopes) == 0 {
		return []string{}
	}

	// 결과 리스트 생성 (정렬)
	result := make([]string, 0, len(scopes))
	for scope := range scopes {
		result = append(result, scope)
	}

	// 주요 scope를 맨 앞으로
	if len(result) > 1 && primaryScope != "" {
		for i, scope := range result {
			if scope == primaryScope {
				// swap
				result[0], result[i] = result[i], result[0]
				break
			}
		}
	}

	return result
}
