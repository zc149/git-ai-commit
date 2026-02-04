package model

import "git-ai-commit/internal/git"

// Config는 애플리케이션 설정을 나타냅니다.
type Config struct {
	APIKey      string // AI API 키
	Model       string // 사용할 모델 (claude, openai 등)
	DetailLevel string // 디테일 레벨 (low, medium, high)
}

// CommitMessage는 AI가 생성한 커밋 메시지 후보입니다.
type CommitMessage struct {
	Message string // Conventional Commit 형식 메시지
	Index   int    // 후보 번호
}

// GeneratorInput은 메시지 생성기의 입력입니다.
type GeneratorInput struct {
	DiffResult *git.DiffResult // 파싱된 diff 결과
	Detail     string          // 디테일 레벨
}

// CommitRequest는 git commit 요청입니다.
type CommitRequest struct {
	Message    string // 커밋 메시지
	DryRun     bool   // dry-run 모드 여부
	AllowEmpty bool   // 빈 커밋 허용 여부
}
