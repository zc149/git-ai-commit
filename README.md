# Git AI Commit

AI를 활용하여 Git 커밋 메시지를 자동으로 생성하는 CLI 도구입니다. 다양한 LLM(Claude, OpenAI, Codex, GLM, Gemini)을 지원합니다.

## 기능

- ✅ Git diff 자동 분석
- 🤖 AI 기반 커밋 메시지 생성 (Conventional Commit 형식)
- 🎯 다중 후보 메시지 제공 및 사용자 선택
- 🔄 다양한 LLM 제공자 지원 (Claude, OpenAI, Codex, GLM, Gemini)
- 📊 스마트한 커밋 타입 및 scope 추천
- 🎨 사용자 친화적인 TUI 인터페이스

## 지원하는 LLM

- **Claude** (Anthropic)
- **OpenAI** (GPT-4)
- **Codex** (OpenAI)
- **GLM** (Zhipu AI)
- **Gemini** (Google)

## 설치

### 빌드

```bash
go build -o git-ai-commit
```

### 사용 가능한 바이너리 (선택 사항)

```bash
# 바이너리를 PATH에 추가
sudo mv git-ai-commit /usr/local/bin/
```

## 사용법

### 1. 환경변수 설정

```bash
export AI_COMMIT_API_KEY="your-api-key-here"
export AI_COMMIT_MODEL="claude"  # 기본값: claude
export AI_COMMIT_DETAIL="medium"  # low, medium, high (기본값: medium)
```

### 2. Git 파일 Stage

```bash
git add .
```

### 3. 실행

```bash
./git-ai-commit
```

### 4. 메시지 선택

AI가 생성한 3개의 커밋 메시지 후보 중 하나를 선택하거나, 직접 입력할 수 있습니다.

## 환경변수

| 변수 | 설명 | 기본값 | 필수 |
|------|------|--------|------|
| `AI_COMMIT_API_KEY` | LLM API 키 | - | ✅ |
| `AI_COMMIT_MODEL` | 사용할 LLM 모델 | `claude` | ❌ |
| `AI_COMMIT_DETAIL` | 디테일 레벨 (`low`, `medium`, `high`) | `medium` | ❌ |

## 지원하는 모델

- `claude` - Claude 3.5 Sonnet
- `openai` - GPT-4
- `codex` - Code Davinci 003
- `glm` - GLM-4
- `gemini` - Gemini Pro

## Conventional Commit 형식

이 도구는 [Conventional Commits](https://www.conventionalcommits.org/) 형식을 따릅니다:

```
type(scope): description
```

### 타입 (Type)

- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서 변경
- `style`: 코드 스타일 변경 (포맷팅 등)
- `refactor`: 코드 리팩토링
- `test`: 테스트 관련
- `build`: 빌드 시스템 또는 의존성 변경
- `chore`: 그 외 작업

## 예시

### 기본 사용

```bash
# 1. 파일 변경 후 stage
git add main.go

# 2. git-ai-commit 실행
./git-ai-commit

# 3. 메시지 후보 중 선택
🤖 Git AI Commit
================

✅ 1개의 파일이 staged되었습니다:
  - main.go

📊 추천 커밋 타입: refactor

🔄 AI가 커밋 메시지를 생성 중...
✅ 커밋 메시지 후보가 생성되었습니다.

=== 커밋 메시지 후보 ===
1) refactor(core): 메시지 생성 로직 개선
2) refactor(generator): diff 분석 최적화
3) refactor: 커밋 메시지 생성 프로세스 리팩토링
c) 사용자 직접 입력
q) 종료

선택 (1-{} 또는 c/q): 1

🎯 커밋 메시지: refactor(core): 메시지 생성 로직 개선

🚀 커밋을 실행합니다...

✨ 커밋 완료!
```

### 다른 모델 사용

```bash
export AI_COMMIT_MODEL="openai"
export AI_COMMIT_API_KEY="sk-..."
./git-ai-commit
```

### 높은 디테일 레벨

```bash
export AI_COMMIT_DETAIL="high"
./git-ai-commit
```

## 프로젝트 구조

```
git-ai-commit/
├── cmd/
│   └── root.go          # CLI 메인 명령어
├── internal/
│   ├── core/
│   │   ├── generator.go  # 커밋 메시지 생성기
│   │   └── prompt.go     # 프롬프트 생성
│   ├── git/
│   │   ├── commit.go     # git commit 실행
│   │   └── diff.go       # git diff 파싱
│   ├── llm/
│   │   ├── provider.go   # LLM 제공자 인터페이스
│   │   ├── claude.go     # Claude 구현
│   │   ├── openai.go     # OpenAI 구현
│   │   ├── codex.go      # Codex 구현
│   │   ├── glm.go        # GLM 구현
│   │   ├── gemini.go     # Gemini 구현
│   │   └── utils.go      # 유틸리티 함수
│   ├── model/
│   │   └── types.go      # 공통 타입 정의
│   └── ui/
│       └── selector.go   # 사용자 선택 인터페이스
├── docs/
│   └── claude/           # 프로젝트 문서
├── main.go               # 진입점
└── README.md
```

## 기여

기여를 환영합니다! Pull Request를 제출하거나 Issue를 생성해주세요.

## 라이선스

MIT License