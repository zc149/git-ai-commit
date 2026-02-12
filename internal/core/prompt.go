package core

import (
	"fmt"
	"git-ai-commit/internal/git"
	"strings"
)

// GeneratePrompt는 LLM에 전달할 프롬프트를 생성합니다.
func GeneratePrompt(diff *git.DiffResult, detail string, lang string) string {
	var builder strings.Builder

	// 헤더
	if lang == "ko" {
		builder.WriteString("다음 정보를 기반으로 Conventional Commit 메시지를 생성하세요.\n\n")
	} else {
		builder.WriteString("Generate Conventional Commit messages based on the following information.\n\n")
	}

	// 추천 타입과 설명
	if lang == "ko" {
		builder.WriteString(fmt.Sprintf("추천 타입: %s (%s)\n", diff.CommitType, getCommitTypeDescription(diff.CommitType, lang)))
	} else {
		builder.WriteString(fmt.Sprintf("Recommended type: %s (%s)\n", diff.CommitType, getCommitTypeDescription(diff.CommitType, lang)))
	}

	// 추천 scope
	if len(diff.Scopes) > 0 {
		if lang == "ko" {
			builder.WriteString(fmt.Sprintf("추천 scope: %s\n", strings.Join(diff.Scopes, ", ")))
		} else {
			builder.WriteString(fmt.Sprintf("Recommended scope: %s\n", strings.Join(diff.Scopes, ", ")))
		}
	}

	builder.WriteString("\n")

	// 변경 패턴 분석
	changePattern := analyzeChangePattern(diff.Files, lang)
	if changePattern != "" {
		if lang == "ko" {
			builder.WriteString("변경 패턴: " + changePattern + "\n\n")
		} else {
			builder.WriteString("Change pattern: " + changePattern + "\n\n")
		}
	}

	// 디렉토리 구조 분석
	dirAnalysis := analyzeDirectoryStructure(diff.Files, lang)
	if dirAnalysis != "" {
		if lang == "ko" {
			builder.WriteString("디렉토리 구조:\n" + dirAnalysis + "\n")
		} else {
			builder.WriteString("Directory structure:\n" + dirAnalysis + "\n")
		}
	}

	// 변경 내용 요약
	if lang == "ko" {
		builder.WriteString("변경 내용 요약:\n")
	} else {
		builder.WriteString("Changes summary:\n")
	}
	if len(diff.Files) == 0 {
		builder.WriteString("변경된 파일이 없습니다.\n")
	} else {
		for _, file := range diff.Files {
			builder.WriteString(fmt.Sprintf("- %s (%s)", file.Path, file.FileType.String()))

			if file.IsNew {
				builder.WriteString(" [새 파일]")
			}
			if file.IsDeleted {
				builder.WriteString(" [삭제됨]")
			}

			builder.WriteString("\n")

			// 변경 내용의 일부를 추가
			if file.Changes != "" {
				summary := summarizeChanges(file.Changes)
				if summary != "" {
					builder.WriteString(fmt.Sprintf("  %s\n", summary))
				}
			}
		}
	}

	builder.WriteString("\n")

	// 요구사항
	if lang == "ko" {
		builder.WriteString("요구사항:\n")
	} else {
		builder.WriteString("Requirements:\n")
	}
	builder.WriteString("- 간결할 것\n")
	if lang == "ko" {
		builder.WriteString("- Conventional Commit 형식 (type(scope): message)\n")
	} else {
		builder.WriteString("- Conventional Commit format (type(scope): message)\n")
	}
	if lang == "ko" {
		builder.WriteString("- 3개의 후보 생성\n")
	} else {
		builder.WriteString("- Generate 3 candidates\n")
	}
	if lang == "ko" {
		builder.WriteString("- 번호로 구분 (예: 1) feat(auth): ...)\n")
	} else {
		builder.WriteString("- Numbered format (e.g., 1) feat(auth): ...)\n")
	}

	// 디테일 레벨에 따른 추가 요구사항
	switch detail {
	case "high":
		if lang == "ko" {
			builder.WriteString("- **반드시** 다중 줄 형식을 사용할 것\n")
			builder.WriteString("- 제목 줄 뒤에 빈 줄을 두고 상세 내용 작성\n")
			builder.WriteString("- 실제 예시:\n")
			builder.WriteString("  1) feat(auth): 사용자 인증 기능 추가\n")
			builder.WriteString("  \n")
			builder.WriteString("     - JWT 토큰 기반 인증 구현\n")
			builder.WriteString("     - 로그인/로그아웃 API 추가\n")
			builder.WriteString("     - 사용자 세션 관리 개선\n")
			builder.WriteString("  \n")
			builder.WriteString("  2) fix(database): 연결 풀 누수 수정\n")
			builder.WriteString("  \n")
			builder.WriteString("     - 연결 해제 로직 수정\n")
			builder.WriteString("     - 타임아웃 설정 추가\n")
			builder.WriteString("     - 메모리 사용량 30% 감소\n")
		} else {
			builder.WriteString("- **MUST** use multi-line format\n")
			builder.WriteString("- Add blank line after title, then provide details\n")
			builder.WriteString("- Real examples:\n")
			builder.WriteString("  1) feat(auth): Add user authentication\n")
			builder.WriteString("  \n")
			builder.WriteString("     - Implement JWT token authentication\n")
			builder.WriteString("     - Add login/logout APIs\n")
			builder.WriteString("     - Improve user session management\n")
			builder.WriteString("  \n")
			builder.WriteString("  2) fix(database): Fix connection pool leak\n")
			builder.WriteString("  \n")
			builder.WriteString("     - Fix connection disposal logic\n")
			builder.WriteString("     - Add timeout settings\n")
			builder.WriteString("     - Reduce memory usage by 30%\n")
		}
	case "medium":
		if lang == "ko" {
			builder.WriteString("- 적절한 디테일 수준 유지\n")
			builder.WriteString("- 한 줄 또는 간단한 다중 줄 형식\n")
		} else {
			builder.WriteString("- Maintain appropriate detail level\n")
			builder.WriteString("- Single line or simple multi-line format\n")
		}
	case "low":
		if lang == "ko" {
			builder.WriteString("- 최소한의 설명\n")
			builder.WriteString("- 한 줄 형식 권장\n")
		} else {
			builder.WriteString("- Minimal description\n")
			builder.WriteString("- Single line format recommended\n")
		}
	}

	return builder.String()
}

// summarizeChanges는 diff 변경 내용을 요약합니다.
func summarizeChanges(changes string) string {
	lines := strings.Split(changes, "\n")
	var summaryLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 빈 줄 건너뛰기
		if line == "" {
			continue
		}

		// diff 헤더 및 메타데이터 건너뛰기
		if strings.HasPrefix(line, "diff --git") ||
			strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "+++") ||
			strings.HasPrefix(line, "@@") {
			continue
		}

		// 실제 코드 변경 라인만 추출
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			// 한 줄짜리 요약만 수집
			if len(summaryLines) < 3 {
				summaryLines = append(summaryLines, line)
			} else {
				summaryLines = append(summaryLines, "...")
				break
			}
		}
	}

	summary := strings.Join(summaryLines, " ")

	// 너무 긴 경우 자르기
	if len(summary) > 100 {
		summary = summary[:100] + "..."
	}

	return summary
}

// getCommitTypeDescription는 커밋 타입에 대한 설명을 반환합니다.
func getCommitTypeDescription(commitType string, lang string) string {
	descriptions := map[string]map[string]string{
		"feat": {
			"ko": "새로운 기능 추가",
			"en": "New feature addition",
		},
		"fix": {
			"ko": "버그 수정",
			"en": "Bug fix",
		},
		"build": {
			"ko": "빌드 시스템 또는 의존성 변경",
			"en": "Build system or dependency changes",
		},
		"docs": {
			"ko": "문서 변경",
			"en": "Documentation changes",
		},
		"test": {
			"ko": "테스트 코드 추가 또는 수정",
			"en": "Test code additions or modifications",
		},
		"refactor": {
			"ko": "코드 리팩토링 (기능 변경 없음)",
			"en": "Code refactoring (no functional changes)",
		},
		"chore": {
			"ko": "기타 작업 (설정, 빌드 등)",
			"en": "Other tasks (config, build, etc.)",
		},
	}

	if desc, ok := descriptions[commitType]; ok {
		if d, ok2 := desc[lang]; ok2 {
			return d
		}
		return desc["en"]
	}
	return ""
}

// analyzeChangePattern은 변경 패턴을 분석하여 설명을 반환합니다.
func analyzeChangePattern(files []git.FileChange, lang string) string {
	if len(files) == 0 {
		return ""
	}

	newFiles := 0
	deletedFiles := 0
	sourceFiles := 0
	configFiles := 0
	testFiles := 0

	newDirectories := make(map[string]bool)

	for _, file := range files {
		if file.IsNew {
			newFiles++
		}
		if file.IsDeleted {
			deletedFiles++
		}

		switch file.FileType {
		case git.FileTypeSource:
			sourceFiles++
		case git.FileTypeConfig:
			configFiles++
		case git.FileTypeTest:
			testFiles++
		}

		// 새 디렉토리 추적
		if file.IsNew {
			parts := strings.Split(file.Path, "/")
			if len(parts) >= 2 {
				dir := strings.Join(parts[:len(parts)-1], "/")
				newDirectories[dir] = true
			}
		}
	}

	var pattern string

	// 변경 패턴 결정
	if newFiles > 0 && deletedFiles == 0 && sourceFiles >= 3 {
		if lang == "ko" {
			pattern = "새로운 기능/모듈 추가"
		} else {
			pattern = "New feature/module addition"
		}
	} else if deletedFiles > 0 {
		if lang == "ko" {
			pattern = "코드/파일 삭제"
		} else {
			pattern = "Code/file deletion"
		}
	} else if testFiles > 0 && sourceFiles == 0 {
		if lang == "ko" {
			pattern = "테스트 코드 변경"
		} else {
			pattern = "Test code changes"
		}
	} else if configFiles > 0 && sourceFiles == 0 {
		if lang == "ko" {
			pattern = "설정 파일 변경"
		} else {
			pattern = "Configuration file changes"
		}
	} else if sourceFiles > 0 && newFiles == 0 {
		if lang == "ko" {
			pattern = "기존 코드 수정"
		} else {
			pattern = "Existing code modifications"
		}
	} else {
		if lang == "ko" {
			pattern = "일반적인 코드 변경"
		} else {
			pattern = "General code changes"
		}
	}

	return pattern
}

// analyzeDirectoryStructure는 디렉토리 구조를 분석하여 요약을 반환합니다.
func analyzeDirectoryStructure(files []git.FileChange, lang string) string {
	if len(files) == 0 {
		return ""
	}

	type dirInfo struct {
		total    int
		source   int
		config   int
		test     int
		doc      int
		newFiles int
	}

	dirStats := make(map[string]*dirInfo)

	for _, file := range files {
		parts := strings.Split(file.Path, "/")
		if len(parts) == 0 {
			continue
		}

		// 첫 번째 디렉토리 기준 집계
		dir := parts[0]
		if dir == "" {
			continue
		}

		if _, ok := dirStats[dir]; !ok {
			dirStats[dir] = &dirInfo{}
		}

		info := dirStats[dir]
		info.total++

		switch file.FileType {
		case git.FileTypeSource:
			info.source++
		case git.FileTypeConfig:
			info.config++
		case git.FileTypeTest:
			info.test++
		case git.FileTypeDoc:
			info.doc++
		}

		if file.IsNew {
			info.newFiles++
		}
	}

	if len(dirStats) == 0 {
		return ""
	}

	var builder strings.Builder

	for dir, info := range dirStats {
		if lang == "ko" {
			builder.WriteString(fmt.Sprintf("- %s: 총 %d개 (소스: %d, 설정: %d, 테스트: %d, 문서: %d, 새 파일: %d)\n",
				dir, info.total, info.source, info.config, info.test, info.doc, info.newFiles))
		} else {
			builder.WriteString(fmt.Sprintf("- %s: total %d (source: %d, config: %d, test: %d, doc: %d, new: %d)\n",
				dir, info.total, info.source, info.config, info.test, info.doc, info.newFiles))
		}
	}

	return builder.String()
}
