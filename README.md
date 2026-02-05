# Git AI Commit

AI를 활용하여 Git 커밋 메시지를 자동으로 생성하는 CLI 도구입니다. Groq의 고성능 LLM(Llama 3.3-70B)을 사용합니다.

## 기능

- ✅ Git diff 자동 분석
- 🤖 AI 기반 커밋 메시지 생성 (Conventional Commit 형식)
- 🎯 다중 후보 메시지 제공 및 사용자 선택
- 🚀 Groq LLM 제공자 지원 (무료, 빠름)
- 📊 스마트한 커밋 타입 및 scope 추천
- 🎨 사용자 친화적인 TUI 인터페이스
- 🌍 다국어 지원 (한국어, 영어)

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
export AI_COMMIT_LANG="en"  # en, ko (기본값: en)
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

### 언어 설정

사용 언어를 설정할 수 있습니다 (기본값: 영어):

#### 명령줄 옵션으로 설정 (우선순위 1)

```bash
# 영어 (기본)
./git-ai-commit --lang en

# 한국어
./git-ai-commit --lang ko
```

#### 환경변수로 설정 (우선순위 2)

```bash
export AI_COMMIT_LANG="ko"
./git-ai-commit
```

**우선순위:** 명령줄 옵션 > 환경 변수 > 기본값(`en`)

### 디테일 레벨 설정

커밋 메시지의 상세도를 조절할 수 있습니다:

#### 명령줄 옵션으로 설정 (우선순위 1)

```bash
# 간단한 메시지
./git-ai-commit --detail low

# 중간 길이 (기본)
./git-ai-commit --detail medium
./git-ai-commit

# 상세한 메시지
./git-ai-commit --detail high
```

#### 환경변수로 설정 (우선순위 2)

```bash
export AI_COMMIT_DETAIL="high"
./git-ai-commit
```

**우선순위:** 명령줄 옵션 > 환경 변수 > 기본값(`medium`)

#### 디테일 레벨 설명

- **low**: 간단하고 짧은 커밋 메시지 (한 줄 위주)
- **medium**: 적절한 길이의 커밋 메시지 (기본값)
- **high**: 상세하고 긴 커밋 메시지 (변경 내용을 상세히 설명)

### 4. 메시지 선택

AI가 생성한 3개의 커밋 메시지 후보 중 하나를 선택하거나, 직접 입력할 수 있습니다.

## 환경변수

| 변수 | 설명 | 기본값 | 필수 |
|------|------|--------|------|
| `AI_COMMIT_GROQ_API_KEY` | Groq API 키 | - | ✅ |
| `AI_COMMIT_MODEL` | 사용할 LLM 모델 (현재는 groq만 지원) | `groq` | ❌ |
| `AI_COMMIT_DETAIL` | 디테일 레벨 (`low`, `medium`, `high`) | `medium` | ❌ |
| `AI_COMMIT_LANG` | 언어 설정 (`en`, `ko`) | `en` | ❌ |

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
📝 디테일 레벨: medium

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

### 다양한 언어 사용

```bash
# 영어 메시지 생성 (기본)
./git-ai-commit --lang en

# 한국어 메시지 생성
./git-ai-commit --lang ko

# 환경변수로 설정
export AI_COMMIT_LANG="ko"
./git-ai-commit
```

### 다양한 디테일 레벨 사용

```bash
# 간단한 메시지
./git-ai-commit --detail low

# 상세한 메시지
./git-ai-commit --detail high

# 환경변수로 설정
export AI_COMMIT_DETAIL="high"
./git-ai-commit
```

### 영어 모드 예시

```bash
./git-ai-commit --lang en

🤖 Git AI Commit

✅ 1 file staged
  - main.go

📊 Recommended commit type: refactor
🤖 Using model: groq
📝 Detail level: medium

🔄 AI is generating commit messages...
✅ Commit message candidates generated.

=== Commit Message Candidates ===
1) refactor(core): improve message generation logic
2) refactor(generator): optimize diff analysis
3) refactor: refactor commit message generation process
c) Custom input
q) Quit

Select (1-3 or c/q): 1

🎯 Commit message: refactor(core): improve message generation logic

🚀 Executing commit...

✨ Commit complete!
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