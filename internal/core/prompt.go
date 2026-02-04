package core

import (
	"fmt"
	"git-ai-commit/internal/git"
	"strings"
)

// GeneratePrompt는 LLM에 전달할 프롬프트를 생성합니다.
func GeneratePrompt(diff *git.DiffResult, detail string) string {
	var builder strings.Builder

	// 헤더
	builder.WriteString("다음 정보를 기반으로 Conventional Commit 메시지를 생성하세요.\n\n")

	// 추천 타입
	builder.WriteString(fmt.Sprintf("추천 타입: %s\n", diff.CommitType))

	// 추천 scope
	if len(diff.Scopes) > 0 {
		builder.WriteString(fmt.Sprintf("추천 scope: %s\n", strings.Join(diff.Scopes, ", ")))
	}

	builder.WriteString("\n")

	// 변경 내용 요약
	builder.WriteString("변경 내용 요약:\n")
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
	builder.WriteString("요구사항:\n")
	builder.WriteString("- 간결할 것\n")
	builder.WriteString("- Conventional Commit 형식 (type(scope): message)\n")
	builder.WriteString("- 3개의 후보 생성\n")
	builder.WriteString("- 번호로 구분 (예: 1) feat(auth): ...)\n")

	// 디테일 레벨에 따른 추가 요구사항
	switch detail {
	case "high":
		builder.WriteString("- 변경 내용을 자세히 설명\n")
		builder.WriteString("- 영향받는 기능 명시\n")
	case "medium":
		builder.WriteString("- 적절한 디테일 수준 유지\n")
	case "low":
		builder.WriteString("- 최소한의 설명\n")
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
