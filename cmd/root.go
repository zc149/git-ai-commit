package cmd

import (
	"fmt"
	"git-ai-commit/internal/core"
	"git-ai-commit/internal/git"
	"git-ai-commit/internal/llm"
	"git-ai-commit/internal/ui"
	"os"
)

// Config는 애플리케이션 설정입니다.
type Config struct {
	Model  string
	APIKey string
	Detail string
}

// RootCommand는 메인 명령어입니다.
type RootCommand struct {
	config *Config
}

// NewRootCommand는 새로운 RootCommand 인스턴스를 생성합니다.
func NewRootCommand(config *Config) *RootCommand {
	return &RootCommand{
		config: config,
	}
}

// Run은 메인 명령어를 실행합니다.
func (r *RootCommand) Run() error {
	fmt.Println("🤖 Git AI Commit")
	fmt.Println("================")

	// 1. staged된 파일 확인
	files, err := git.GetStagedFiles()
	if err != nil {
		return fmt.Errorf("staged 파일 확인 실패: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("\n❌ staged된 파일이 없습니다.")
		fmt.Println("git add를 사용하여 파일을 stage한 후 다시 시도해주세요.")
		return nil
	}

	fmt.Printf("\n✅ %d개의 파일이 staged되었습니다:\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}

	// 2. diff 분석 및 파싱
	diffResult, err := git.GetCachedDiff()
	if err != nil {
		return fmt.Errorf("diff 분석 실패: %w", err)
	}

	fmt.Printf("\n📊 추천 커밋 타입: %s\n", diffResult.CommitType)
	if len(diffResult.Scopes) > 0 {
		fmt.Printf("   추천 scope: %s\n", diffResult.Scopes)
	}

	// 3. LLM 제공자 생성
	provider, err := llm.NewProvider(r.config.Model, r.config.APIKey)
	if err != nil {
		return fmt.Errorf("LLM 제공자 생성 실패: %w", err)
	}

	// 4. 커밋 메시지 생성
	fmt.Println("\n🔄 AI가 커밋 메시지를 생성 중...")
	generator := core.NewGenerator(provider)
	messages, err := generator.Generate(diffResult, r.config.Detail)
	if err != nil {
		return fmt.Errorf("커밋 메시지 생성 실패: %w", err)
	}

	fmt.Println("✅ 커밋 메시지 후보가 생성되었습니다.")

	// 5. 사용자 선택
	selector := ui.NewSelector()
	selectedMessage, err := selector.Select(messages)
	if err != nil {
		return err
	}

	// 6. 커밋 실행
	fmt.Printf("\n🎯 커밋 메시지: %s\n", selectedMessage)
	fmt.Println("\n🚀 커밋을 실행합니다...")

	if err := git.Commit(selectedMessage); err != nil {
		return err
	}

	fmt.Println("\n✨ 커밋 완료!")
	return nil
}

// RunWithArgs는 명령줄 인자를 받아 실행합니다.
func RunWithArgs(args []string) error {
	config := &Config{
		Model:  getEnvWithDefault("AI_COMMIT_MODEL", "claude"),
		APIKey: getEnvWithDefault("AI_COMMIT_API_KEY", ""),
		Detail: getEnvWithDefault("AI_COMMIT_DETAIL", "medium"),
	}

	// API 키 확인
	if config.APIKey == "" {
		return fmt.Errorf("AI_COMMIT_API_KEY 환경변수가 설정되지 않았습니다")
	}

	cmd := NewRootCommand(config)
	return cmd.Run()
}

// getEnvWithDefault는 환경변수를 가져오거나 기본값을 반환합니다.
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
