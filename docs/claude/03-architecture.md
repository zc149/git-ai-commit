# 아키텍처 설계

## 전체 구조

프로젝트는 다음 레이어로 구성된다.

- cmd: CLI 진입점
- core: 비즈니스 로직
- llm: AI 제공자 모듈
- ui: CLI 인터페이스
- config: 설정 관리

---

## 패키지 구조

git-ai-commit
cmd
  root.go

internal
  core
    analyzer.go
    generator.go

  git
    diff.go

  llm
    provider.go
    claude.go
    openai.go

  ui

  model
    types.go

config

main.go

---

## 의존성 방향

cmd → core → llm  
cmd → ui  
core → git  

의존성은 항상 안쪽으로 향한다.
