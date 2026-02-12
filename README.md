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

## 설치

### 🚀 빠른 시작 (개발자용)

**소스 코드에서 바로 빌드해서 사용하고 싶다면:**

**⚠️ Windows 사용자 주의:**
- Windows에서는 **Git Bash**에서만 사용 가능합니다
- CMD/PowerShell에서는 환경변수 전달 문제로 작동하지 않습니다
- Git Bash는 Git for Windows 설치 시 자동으로 제공됩니다

```bash
# 1. 저장소 복제
git clone https://github.com/zc149/git-ai-commit.git
cd git-ai-commit

# 2. 빌드
go build -o git-ai-commit.exe main.go  # Windows
go build -o git-ai-commit main.go       # macOS/Linux

# 3. Git alias 설정 (필수!)
git config --global alias.ai-commit "!$(pwd)/git-ai-commit.exe"  # Windows
git config --global alias.ai-commit "!$(pwd)/git-ai-commit"       # macOS/Linux

# 4. API 키 설정
# Windows (Git Bash에서 실행)
echo 'export AI_COMMIT_GROQ_API_KEY="your-api-key"' >> ~/.bashrc
source ~/.bashrc

# macOS/Linux
echo 'export AI_COMMIT_GROQ_API_KEY="your-api-key"' >> ~/.zshrc  # 또는 ~/.bashrc
source ~/.zshrc

# 5. 사용 (Git Bash에서)
git add .
git ai-commit
```

---

### 📦 바이너리 설치 (일반 사용자용)

### macOS

#### Homebrew로 설치 (추천)

```bash
brew install zc149/git-ai-commit/git-ai-commit
```

#### GitHub Releases에서 설치

1. [GitHub Releases](https://github.com/zc149/git-ai-commit/releases) 페이지로 이동
2. 다운로드:
   - **Intel Mac**: `git-ai-commit-darwin-amd64`
   - **Apple Silicon (M1/M2/M3)**: `git-ai-commit-darwin-arm64`
3. 터미널에서 다운로드한 파일 실행 권한 부여:

```bash
chmod +x ~/Downloads/git-ai-commit-darwin-arm64
```

4. `/usr/local/bin`으로 이동 (시스템 PATH에 추가):

```bash
sudo mv ~/Downloads/git-ai-commit-darwin-arm64 /usr/local/bin/git-ai-commit
```

### Windows

**⚠️ 중요: Windows에서는 Git Bash에서만 사용 가능합니다**

Windows의 CMD/PowerShell은 git alias 실행 시 환경변수 전달 문제가 있어 **Git Bash**에서만 사용하세요.

#### Git Bash에서 설치 (권장)

```bash
# 1. GitHub Releases에서 다운로드
# https://github.com/zc149/git-ai-commit/releases/latest
# git-ai-commit-windows-amd64.exe 다운로드

# 2. 다운로드 폴더로 이동
cd ~/Downloads

# 3. 실행 권한 부여
chmod +x git-ai-commit-windows-amd64.exe

# 4. Git alias 설정
git config --global alias.ai-commit "!$(pwd)/git-ai-commit-windows-amd64.exe"

# 5. API 키 설정 (.bashrc에 추가)
echo 'export AI_COMMIT_GROQ_API_KEY="your-api-key"' >> ~/.bashrc
source ~/.bashrc

# 6. 사용
git add .
git ai-commit
```

#### 시스템 전체 설치 (Git Bash)

```bash
# 1. 폴더 생성 및 파일 이동
mkdir -p ~/bin
mv ~/Downloads/git-ai-commit-windows-amd64.exe ~/bin/git-ai-commit.exe

# 2. Git alias 설정
git config --global alias.ai-commit "!~/bin/git-ai-commit.exe"

# 3. API 키 설정
echo 'export AI_COMMIT_GROQ_API_KEY="your-api-key"' >> ~/.bashrc
source ~/.bashrc
```

### Linux

```bash
# 다운로드
wget https://github.com/zc149/git-ai-commit/releases/latest/download/git-ai-commit-linux-amd64

# 실행 권한 부여
chmod +x git-ai-commit-linux-amd64

# /usr/local/bin으로 이동
sudo mv git-ai-commit-linux-amd64 /usr/local/bin/git-ai-commit

# git alias 설정 (중요!)
git config --global alias.ai-commit "!/usr/local/bin/git-ai-commit"
```

## 설치 확인

설치가 완료되면 다음 명령어로 확인:

```bash
git ai-commit --help
```

## API 키 설정

### Groq API 키 발급

1. [console.groq.com](https://console.groq.com)에서 계정 생성
2. API Keys 메뉴에서 새 키 생성

### 환경변수 설정

#### macOS / Linux

```bash
# 현재 터미널 세션에만 적용
export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"

# 영구 적용 (~/.zshrc 또는 ~/.bashrc에 추가)
echo 'export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"' >> ~/.zshrc
source ~/.zshrc
```

**⚠️ 중요:** API 키 설정 후 **새 터미널을 열어야** 환경변수가 적용됩니다.

#### Windows (Git Bash만 지원)

**Windows에서는 Git Bash에서만 사용 가능합니다.**

```bash
# 영구 적용 (~/.bashrc에 추가)
echo 'export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"' >> ~/.bashrc
source ~/.bashrc
```

**⚠️ 주의:** CMD/PowerShell에서는 사용할 수 없습니다. **Git Bash**를 사용하세요.

## 사용법

### 1. Git 파일 Stage

```bash
git add .
```

### 2. AI 커밋 메시지 생성

```bash
git ai-commit
```

### 3. 메시지 후보 중 선택

AI가 생성한 3개의 커밋 메시지 후보 중 하나를 선택하거나, 직접 입력할 수 있습니다.

### 예시

```bash
git ai-commit

🤖 Git AI Commit
================

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

## 옵션

### 언어 설정

기본값은 영어입니다. 한국어로 사용하려면:

```bash
# 명령줄 옵션으로 설정
git ai-commit --lang ko

# 영어 (기본)
git ai-commit --lang en
```

### 디테일 레벨

커밋 메시지의 상세도를 조절할 수 있습니다:

```bash
# 간단한 메시지
git ai-commit --detail low

# 중간 길이 (기본)
git ai-commit --detail medium

# 상세한 메시지
git ai-commit --detail high
```

#### 디테일 레벨 설명

- **low**: 간단하고 짧은 커밋 메시지 (한 줄 위주)
- **medium**: 적절한 길이의 커밋 메시지 (기본값)
- **high**: 정형화된 상세 커밋 메시지
  - 첫 줄: Conventional Commit 형식 (type: message)
  - 두 번째 줄: 빈 줄
  - 세 번째 줄부터: `- `로 시작하는 상세 내용 목록
  
  예시:
  ```
  feat(auth): implement OAuth2 authentication

  - Add Google OAuth provider
  - Add GitHub OAuth provider
  - Update authentication flow
  - Add token refresh mechanism
  ```

### 사용 예시

#### 상세한 메시지 (한국어)

```bash
git ai-commit --lang ko --detail high
```

#### 간단한 메시지 (영어)

```bash
git ai-commit --lang en --detail low
```

## 환경변수

| 변수 | 설명 | 기본값 | 필수 |
|------|------|--------|------|
| `AI_COMMIT_GROQ_API_KEY` | Groq API 키 | - | ✅ |
| `AI_COMMIT_MODEL` | 사용할 LLM 모델 (현재는 groq만 지원) | `groq` | ❌ |
| `AI_COMMIT_DETAIL` | 디테일 레벨 (`low`, `medium`, `high`) | `medium` | ❌ |
| `AI_COMMIT_LANG` | 언어 설정 (`en`, `ko`) | `en` | ❌ |

### 환경변수로 설정

#### macOS / Linux

```bash
# ~/.zshrc 또는 ~/.bashrc에 추가
export AI_COMMIT_DETAIL="high"
export AI_COMMIT_LANG="ko"
```

#### Windows (PowerShell)

```powershell
[System.Environment]::SetEnvironmentVariable('AI_COMMIT_DETAIL', 'high', 'User')
[System.Environment]::SetEnvironmentVariable('AI_COMMIT_LANG', 'ko', 'User')
```

#### Windows (CMD)

```cmd
setx AI_COMMIT_DETAIL "high"
setx AI_COMMIT_LANG "ko"
```

### 우선순위

1. 명령줄 옵션 (`--detail`, `--lang`)
2. 환경변수 (`AI_COMMIT_DETAIL`, `AI_COMMIT_LANG`)
3. 기본값 (`medium`, `en`)

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

## 지원하는 모델

- **Groq** - Llama 3.3-70B-Versatile
  - 완전 무료
  - 매우 빠른 추론 속도
  - 높은 성능

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
├── build.sh              # 빌드 스크립트
└── README.md
```

## 개발

### 빌드

```bash
# 현재 플랫폼용 빌드
go build -o git-ai-commit

# 모든 플랫폼용 빌드
./build.sh
```

## 기여

기여를 환영합니다! Pull Request를 제출하거나 Issue를 생성해주세요.

## 라이선스

MIT License