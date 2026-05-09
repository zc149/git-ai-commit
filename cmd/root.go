package cmd

import (
	"errors"
	"flag"
	"fmt"
	"git-ai-commit/internal/cache"
	"git-ai-commit/internal/config"
	"git-ai-commit/internal/core"
	"git-ai-commit/internal/git"
	"git-ai-commit/internal/llm"
	"git-ai-commit/internal/ui"
	"git-ai-commit/internal/version"
	"os"
)

// RootCommand는 메인 명령어입니다.
type RootCommand struct {
	config *config.Config
	detail string
	lang   string
}

// NewRootCommand는 새로운 RootCommand 인스턴스를 생성합니다.
func NewRootCommand(cfg *config.Config, detail string, lang string) *RootCommand {
	return &RootCommand{
		config: cfg,
		detail: detail,
		lang:   lang,
	}
}

// Run은 메인 명령어를 실행합니다.
func (r *RootCommand) Run() error {
	// 언어 설정 확인
	lang := r.getLanguage()

	fmt.Println("🤖 Git AI Commit")
	fmt.Println("================")

	// 1. staged된 파일 확인
	files, err := git.GetStagedFiles()
	if err != nil {
		return fmt.Errorf("%s: %w", r.getMessage("error_staged_failed", lang), err)
	}

	if len(files) == 0 {
		fmt.Println("\n❌ " + r.getMessage("error_no_staged_files", lang))
		fmt.Println(r.getMessage("hint_use_git_add", lang))
		return nil
	}

	fmt.Printf("\n✅ %s\n", r.formatFileCount(len(files), lang))
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}

	// 2. diff 분석 및 파싱
	diffResult, err := git.GetCachedDiff()
	if err != nil {
		return fmt.Errorf("%s: %w", r.getMessage("error_diff_failed", lang), err)
	}

	// diff hash 계산
	diffHash := git.CalculateDiffHash(diffResult.RawDiff)

	fmt.Printf("\n📊 %s: %s\n", r.getMessage("label_recommended_type", lang), diffResult.CommitType)
	if len(diffResult.Scopes) > 0 {
		fmt.Printf("   %s: %s\n", r.getMessage("label_recommended_scope", lang), diffResult.Scopes)
	}

	// 3. 캐시 매니저 초기화 및 이전 메시지 로드
	cacheManager, err := cache.NewCacheManager()
	if err != nil {
		return fmt.Errorf("캐시 매니저 생성 실패: %w", err)
	}

	cachedData, err := cacheManager.Load(diffHash)
	var prevMessage string
	if err == nil && cachedData != nil {
		prevMessage = cachedData.Message
	}

	// 4. 사용할 모델 결정
	model := r.config.Model
	if model == "" {
		model = r.config.GetFirstAvailableModel()
	}

	if model == "" {
		return errors.New(r.getMessage("error_no_api_key", lang))
	}

	fmt.Printf("🤖 %s: %s\n", r.getMessage("label_using_model", lang), model)

	// 4. API 키 가져오기
	apiKey, err := r.config.GetAPIKey(model)
	if err != nil {
		return fmt.Errorf("%s: %w", r.getMessage("error_get_api_key", lang), err)
	}

	// 5. LLM 제공자 생성
	provider, err := llm.NewProvider(model, apiKey)
	if err != nil {
		return fmt.Errorf("%s: %w", r.getMessage("error_create_provider", lang), err)
	}

	// 5. 커밋 메시지 생성
	detail := r.getDetailLevel()
	fmt.Printf("📝 %s: %s\n", r.getMessage("label_detail_level", lang), detail)
	fmt.Println("\n🔄 " + r.getMessage("generating_messages", lang))
	generator := core.NewGenerator(provider)
	messages, err := generator.Generate(diffResult, detail, lang)
	if err != nil {
		return fmt.Errorf("%s: %w", r.getMessage("error_generate_failed", lang), err)
	}

	fmt.Println("✅ " + r.getMessage("candidates_generated", lang))

	// 6. 사용자 선택 (재추천 루프)
	selector := ui.NewSelector(lang)
	var selectedMessage string

	for {
		selectedMessage, err = selector.Select(messages, prevMessage)

		// 에러 타입 확인
		if err != nil {
			// 재추천 요청
			if _, ok := err.(*ui.RegenerateError); ok {
				fmt.Println("\n🔄 " + r.getMessage("regenerating_messages", lang))
				messages, err = generator.Generate(diffResult, detail, lang)
				if err != nil {
					return fmt.Errorf("%s: %w", r.getMessage("error_generate_failed", lang), err)
				}
				fmt.Println("✅ " + r.getMessage("candidates_generated", lang))
				continue
			}

			// 이전 메시지 사용
			if prevMsgErr, ok := err.(*ui.UsePrevMessageError); ok {
				selectedMessage = prevMsgErr.Message
				break
			}

			// 그 외 에러 (종료 등)
			return err
		}

		// 정상 선택
		break
	}

	// 7. 선택한 메시지 캐시에 저장
	if err := cacheManager.Save(diffHash, selectedMessage); err != nil {
		// 캐시 저장 실패는 치명적이지 않으므로 계속 진행
		fmt.Println("⚠️ " + r.getMessage("warning_cache_save_failed", lang))
	}

	// 8. 커밋 실행
	fmt.Printf("\n🎯 %s: %s\n", r.getMessage("label_commit_message", lang), selectedMessage)
	fmt.Println("\n🚀 " + r.getMessage("executing_commit", lang))

	if err := git.Commit(selectedMessage); err != nil {
		return err
	}

	fmt.Println("\n✨ " + r.getMessage("commit_complete", lang))
	return nil
}

// RunWithArgs는 명령줄 인자를 받아 실행합니다.
func RunWithArgs(args []string) error {
	// 플래그 정의
	versionFlag := flag.Bool("v", false, "버전 정보 출력")
	detailFlag := flag.String("detail", "", "디테일 레벨: low, medium, high")
	langFlag := flag.String("lang", "", "언어: en, ko")

	// 플래그 파싱
	flag.CommandLine.Parse(args)

	// 버전 출력
	if *versionFlag {
		fmt.Printf("git-ai-commit %s\n", version.Version)
		return nil
	}

	// 디테일 레벨 유효성 검사
	if *detailFlag != "" {
		valid := false
		for _, level := range []string{"low", "medium", "high"} {
			if *detailFlag == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("잘못된 디테일 레벨: %s (low, medium, high 중 하나를 입력하세요)", *detailFlag)
		}
	}

	// 언어 유효성 검사
	if *langFlag != "" {
		valid := false
		for _, l := range []string{"en", "ko"} {
			if *langFlag == l {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("잘못된 언어 설정: %s (en, ko 중 하나를 입력하세요)", *langFlag)
		}
	}

	// 설정 로드
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("설정 로드 실패: %w", err)
	}

	cmd := NewRootCommand(cfg, *detailFlag, *langFlag)
	return cmd.Run()
}

// getLanguage는 언어를 반환합니다.
// 우선순위: 명령줄 옵션 > 환경 변수 > 기본값
func (r *RootCommand) getLanguage() string {
	if r.lang != "" {
		return r.lang
	}
	return getEnvWithDefault("AI_COMMIT_LANG", "en")
}

// getDetailLevel은 디테일 레벨을 반환합니다.
// 우선순위: 명령줄 옵션 > 환경 변수 > 기본값
func (r *RootCommand) getDetailLevel() string {
	if r.detail != "" {
		return r.detail
	}
	return getEnvWithDefault("AI_COMMIT_DETAIL", "low")
}

// getMessage는 언어에 따른 메시지를 반환합니다.
func (r *RootCommand) getMessage(key, lang string) string {
	messages := map[string]map[string]string{
		"error_staged_failed": {
			"en": "Failed to check staged files",
			"ko": "staged 파일 확인 실패",
		},
		"error_no_staged_files": {
			"en": "No staged files",
			"ko": "staged된 파일이 없습니다",
		},
		"hint_use_git_add": {
			"en": "Stage files using git add and try again",
			"ko": "git add를 사용하여 파일을 stage한 후 다시 시도해주세요",
		},
		"error_diff_failed": {
			"en": "Failed to analyze diff",
			"ko": "diff 분석 실패",
		},
		"error_no_api_key": {
			"en": "No API key available. Please set API key in .env file or environment variables",
			"ko": "사용 가능한 API 키가 없습니다. .env 파일 또는 환경변수에 API 키를 설정해주세요",
		},
		"error_get_api_key": {
			"en": "Failed to get API key",
			"ko": "API 키 가져오기 실패",
		},
		"error_create_provider": {
			"en": "Failed to create LLM provider",
			"ko": "LLM 제공자 생성 실패",
		},
		"error_generate_failed": {
			"en": "Failed to generate commit messages",
			"ko": "커밋 메시지 생성 실패",
		},
		"label_recommended_type": {
			"en": "Recommended commit type",
			"ko": "추천 커밋 타입",
		},
		"label_recommended_scope": {
			"en": "Recommended scope",
			"ko": "추천 scope",
		},
		"label_using_model": {
			"en": "Using model",
			"ko": "사용 모델",
		},
		"label_detail_level": {
			"en": "Detail level",
			"ko": "디테일 레벨",
		},
		"label_commit_message": {
			"en": "Commit message",
			"ko": "커밋 메시지",
		},
		"generating_messages": {
			"en": "AI is generating commit messages...",
			"ko": "AI가 커밋 메시지를 생성 중...",
		},
		"candidates_generated": {
			"en": "Commit message candidates generated",
			"ko": "커밋 메시지 후보가 생성되었습니다",
		},
		"regenerating_messages": {
			"en": "Regenerating candidates...",
			"ko": "새로운 후보를 생성 중...",
		},
		"warning_cache_save_failed": {
			"en": "Failed to save commit message to cache",
			"ko": "커밋 메시지 캐시 저장 실패",
		},
		"executing_commit": {
			"en": "Executing commit...",
			"ko": "커밋을 실행합니다...",
		},
		"commit_complete": {
			"en": "Commit complete!",
			"ko": "커밋 완료!",
		},
	}

	if msgMap, ok := messages[key]; ok {
		if msg, ok := msgMap[lang]; ok {
			return msg
		}
		return msgMap["en"] // 기본값은 영어
	}
	return key
}

// formatFileCount는 파일 수를 언어에 맞게 포맷팅합니다.
func (r *RootCommand) formatFileCount(count int, lang string) string {
	if lang == "ko" {
		return fmt.Sprintf("%d개의 파일이 staged되었습니다", count)
	}
	return fmt.Sprintf("%d file%s staged", count, map[bool]string{true: "s", false: ""}[count > 1])
}

// getEnvWithDefault는 환경변수를 가져오거나 기본값을 반환합니다.
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
