# Git AI Commit

AI를 활용하여 Git 커밋 메시지를 자동으로 생성하는 CLI 도구입니다. Groq의 고성능 LLM(Llama 3.3-70B)을 사용합니다.

## 기능

- ✅ Git diff 자동 분석
- 🤖 AI 기반 커밋 메시지 생성 (Conventional Commit 형식)
- 🎯 다중 후보 메시지 제공 및 사용자 선택
- 🚀 Groq LLM 제공자 지원 (무료, 빠름)
- 📊 스마트한 커밋 타입 및 scope 추천
- 🎨 사용자 친화적인 TUI 인터페이스

## 지원하는 LLM

- **Groq** - Llama 3.3-70B-Versatile (완전 무료, 매우 빠름)

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
export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"
```

**Groq API 키 받는 방법:**
1. [console.groq.com](https://console.groq.com)에서 계정 생성
2. API Keys 메뉴에서 새 키 생성
3. 키를 환경변수에 설정

### 선택 사항

```bash
export AI_COMMIT_MODEL="groq"  # 기본값 (현재 유일한 옵션)
export AI_COMMIT_DETAIL="medium"  # low, medium, high (기본값: medium)
```

### 영구 설정 (선택 사항)

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가
echo 'export AI_COMMIT_GROQ_API_KEY="your-api-key"' >> ~/.zshrc
source ~/.zshrc
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
| `AI_COMMIT_GROQ_API_KEY` | Groq API 키 | - | ✅ |
| `AI_COMMIT_MODEL` | 사용할 LLM 모델 (현재는 groq만 지원) | `groq` | ❌ |
| `AI_COMMIT_DETAIL` | 디테일 레벨 (`low`, `medium`, `high`) | `medium` | ❌ |

### API 키 우선순위

Groq는 다음 환경 변수 중 첫 번째로 설정된 값을 사용합니다:
- `AI_COMMIT_GROQ_API_KEY` > `GROQ_API_KEY`

## 지원하는 모델

- `groq` - Llama 3.3-70B-Versatile
  - 완전 무료
  - 매우 빠른 추론 속도
  - 높은 성능

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
🤖 사용 모델: groq

🔄 AI가 커밋 메시지를 생성 중...
✅ 커밋 메시지 후보가 생성되었습니다.

=== 커밋 메시지 후보 ===
1) refactor(core): 메시지 생성 로직 개선
2) refactor(generator): diff 분석 최적화
3) refactor: 커밋 메시지 생성 프로세스 리팩토링
c) 사용자 직접 입력
q) 종료

선택 (1-3 또는 c/q): 1

🎯 커밋 메시지: refactor(core): 메시지 생성 로직 개선

🚀 커밋을 실행합니다...

✨ 커밋 완료!
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
│   │   ├── groq.go       # Groq 구현
│   │   └── utils.go      # 유틸리티 함수
│   ├── model/
│   │   └── types.go      # 공통 타입 정의
│   ├── config/
│   │   └── config.go     # 설정 관리
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