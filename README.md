# Git AI Commit

`git-ai-commit`은 staged diff를 분석해서 Conventional Commit 메시지를 추천하고, 선택한 메시지로 바로 커밋까지 실행하는 CLI 도구입니다.

핵심은 단순한 AI 호출이 아니라 **Git diff를 빠르게 구조화하고, 변경 의도에 맞는 type/scope를 추천한 뒤 LLM에 더 정확한 컨텍스트를 전달하는 것**입니다.

## Why

커밋 메시지는 짧지만, 좋은 메시지를 쓰려면 staged 변경을 다시 읽고 의도를 정리해야 합니다. `git-ai-commit`은 이 과정을 줄여줍니다.

- staged 파일과 diff를 자동 분석합니다.
- 변경 파일의 종류, 추가/삭제 라인 수, 테스트/문서/설정/의존성 변경 여부를 파악합니다.
- `feat`, `fix`, `refactor`, `docs`, `test`, `build`, `chore` 중 적절한 타입을 추천합니다.
- 경로 기반으로 scope를 추천합니다.
- AI가 생성한 3개의 후보 중 선택하거나 직접 입력할 수 있습니다.
- 같은 diff에 대해 이전에 선택한 메시지를 캐시해 재사용할 수 있습니다.

## Features

- Git staged diff 자동 분석
- 점수 기반 commit type/scope 추천
- Groq LLM 기반 커밋 메시지 후보 생성
- Conventional Commit 형식 지원
- 후보 재생성, 직접 입력, 이전 메시지 재사용
- 한국어/영어 출력 지원
- detail level 조절: `low`, `medium`, `high`

## Requirements

- Git
- Go 1.25 이상, 소스에서 빌드할 경우
- Groq API key

Groq API key는 [console.groq.com](https://console.groq.com)에서 생성할 수 있습니다.

## Quick Start

```bash
git clone https://github.com/zc149/git-ai-commit.git
cd git-ai-commit

go build -o git-ai-commit main.go
git config --global alias.ai-commit "!$(pwd)/git-ai-commit"

export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"

git add .
git ai-commit
```

Windows에서는 Git Bash 사용을 권장합니다. CMD/PowerShell에서는 Git alias와 환경변수 전달 방식 때문에 정상 동작하지 않을 수 있습니다.

## Installation

### Homebrew

```bash
brew install zc149/git-ai-commit/git-ai-commit
```

### From Source

macOS/Linux:

```bash
git clone https://github.com/zc149/git-ai-commit.git
cd git-ai-commit
go build -o git-ai-commit main.go
git config --global alias.ai-commit "!$(pwd)/git-ai-commit"
```

Windows Git Bash:

```bash
git clone https://github.com/zc149/git-ai-commit.git
cd git-ai-commit
go build -o git-ai-commit.exe main.go
git config --global alias.ai-commit "!$(pwd)/git-ai-commit.exe"
```

### GitHub Releases

GitHub Releases에서 운영체제에 맞는 바이너리를 내려받아 PATH에 있는 디렉토리로 이동한 뒤 Git alias를 설정합니다.

macOS/Linux 예시:

```bash
chmod +x git-ai-commit-darwin-arm64
sudo mv git-ai-commit-darwin-arm64 /usr/local/bin/git-ai-commit
git config --global alias.ai-commit "!/usr/local/bin/git-ai-commit"
```

Linux 예시:

```bash
wget https://github.com/zc149/git-ai-commit/releases/latest/download/git-ai-commit-linux-amd64
chmod +x git-ai-commit-linux-amd64
sudo mv git-ai-commit-linux-amd64 /usr/local/bin/git-ai-commit
git config --global alias.ai-commit "!/usr/local/bin/git-ai-commit"
```

Windows Git Bash 예시:

```bash
mkdir -p ~/bin
mv ~/Downloads/git-ai-commit-windows-amd64.exe ~/bin/git-ai-commit.exe
git config --global alias.ai-commit "!~/bin/git-ai-commit.exe"
```

## API Key

macOS/Linux:

```bash
export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"
```

영구 적용:

```bash
echo 'export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"' >> ~/.zshrc
source ~/.zshrc
```

Windows Git Bash:

```bash
echo 'export AI_COMMIT_GROQ_API_KEY="your-groq-api-key"' >> ~/.bashrc
source ~/.bashrc
```

`GROQ_API_KEY`도 fallback으로 사용할 수 있습니다.

## Usage

1. 커밋할 파일을 stage합니다.

```bash
git add .
```

2. 커밋 메시지를 생성합니다.

```bash
git ai-commit
```

3. 후보 중 하나를 선택합니다.

```text
🤖 Git AI Commit
================

✅ 2 files staged
  - internal/git/analysis.go
  - internal/git/analysis_test.go

📊 Recommended commit type: feat
   Recommended scope: git
🤖 Using model: groq
📝 Detail level: low

🔄 AI is generating commit messages...
✅ Commit message candidates generated

=== Commit Message Candidates ===
1) feat(git): add structured diff analysis
2) feat(analysis): improve commit type scoring
3) refactor(git): refine diff-based commit recommendations
c) Custom input
r) Regenerate candidates
q) Quit

Select (1-3 or c/r/q): 1
```

선택하면 해당 메시지로 `git commit`이 실행됩니다.

### Selection Options

| 입력 | 동작 |
|------|------|
| `1`, `2`, `3` | 후보 메시지 선택 |
| `c` | 직접 커밋 메시지 입력 |
| `r` | 후보 다시 생성 |
| `p` | 같은 diff에 대해 캐시된 이전 메시지 사용 |
| `q` | 종료 |

`p` 옵션은 같은 diff hash에 대해 이전에 선택한 메시지가 있을 때만 표시됩니다.

## Options

```bash
git ai-commit --lang ko
git ai-commit --lang en

git ai-commit --detail low
git ai-commit --detail medium
git ai-commit --detail high

git ai-commit -v
```

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--lang` | 출력 언어: `en`, `ko` | `en` |
| `--detail` | 메시지 상세도: `low`, `medium`, `high` | `low` |
| `-v` | 버전 출력 | - |

### Detail Levels

- `low`: 짧은 한 줄 메시지 중심
- `medium`: 한 줄 또는 간단한 다중 줄 메시지
- `high`: 제목과 bullet body를 포함한 상세 메시지

`high` 예시:

```text
feat(auth): implement OAuth2 authentication

- Add Google OAuth provider
- Add GitHub OAuth provider
- Update token refresh flow
```

## Environment Variables

| 변수 | 설명 | 기본값 | 필수 |
|------|------|--------|------|
| `AI_COMMIT_GROQ_API_KEY` | Groq API key | - | 예 |
| `GROQ_API_KEY` | Groq API key fallback | - | 아니오 |
| `AI_COMMIT_MODEL` | 사용할 모델 provider. 현재 `groq`만 지원 | `groq` | 아니오 |
| `AI_COMMIT_DETAIL` | detail level: `low`, `medium`, `high` | `low` | 아니오 |
| `AI_COMMIT_LANG` | 출력 언어: `en`, `ko` | `en` | 아니오 |

우선순위:

1. CLI option: `--detail`, `--lang`
2. Environment variable: `AI_COMMIT_DETAIL`, `AI_COMMIT_LANG`
3. Default: `low`, `en`

## How It Works

```text
staged files
  -> git diff --cached
  -> parallel diff parser
  -> structured analysis
  -> commit type/scope recommendation
  -> prompt generation
  -> Groq LLM candidates
  -> user selection
  -> git commit
```

분석기는 다음 정보를 사용합니다.

- 파일 타입: source, test, docs, config, dependencies
- 변경 상태: new, modified, deleted
- 추가/삭제 라인 수
- public symbol 추가 여부
- bug/error-handling 신호
- 테스트와 소스가 함께 바뀌었는지
- 의존성 파일만 바뀌었는지
- 공통 경로 기반 scope

## Conventional Commit

생성되는 메시지는 Conventional Commit 형식을 따릅니다.

```text
type(scope): description
```

지원하는 주요 type:

- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서 변경
- `refactor`: 기능 변화 없는 코드 개선
- `test`: 테스트 추가 또는 수정
- `build`: 빌드 시스템 또는 의존성 변경
- `chore`: 기타 유지보수 작업

## Troubleshooting

### No staged files

먼저 파일을 stage해야 합니다.

```bash
git add .
git ai-commit
```

### No API key available

Groq API key가 설정되어 있는지 확인합니다.

```bash
echo "$AI_COMMIT_GROQ_API_KEY"
```

### Windows에서 동작하지 않음

Git Bash에서 실행하세요. CMD/PowerShell은 Git alias와 환경변수 전달 방식 때문에 지원이 제한됩니다.

## Development

### Build

```bash
go build -o git-ai-commit main.go
```

### Test

```bash
go test ./...
```

### Release Build

```bash
./build.sh
```

## Project Structure

```text
git-ai-commit/
├── cmd/
│   └── root.go              # CLI flow
├── internal/
│   ├── cache/               # diff hash based message cache
│   ├── config/              # environment variable config
│   ├── core/                # prompt generation and message generation
│   ├── git/
│   │   ├── analysis.go      # structured diff analysis and type scoring
│   │   ├── commit.go        # git commit execution
│   │   └── diff.go          # staged diff parsing
│   ├── llm/                 # Groq provider and response parsing
│   ├── model/               # shared model types
│   ├── ui/                  # CLI selection prompt
│   ├── version/             # build-time version
│   └── worker/              # parallel diff parser
├── docs/
├── main.go
├── build.sh
└── README.md
```

## License

MIT License
