# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] - 2026-02-06

### Added
- Worker Pool 패턴 도입으로 병렬 diff 파싱 기능 추가
- 대규모 커밋 처리 성능 향상 (50+ 파일: ~50% 속도 향상)
- 동적 worker 수 조절 (파일 수에 따라 2-8 workers)
- 조건부 병렬화 (소규모 커밋은 순차 처리로 오버헤드 방지)
- `internal/worker/pool.go`: Worker Pool 구현

### Improved
- diff 파싱 성능 최적화 (순차 → 병렬 처리)
- CPU 리소스 효율적 활용 (I/O 병렬 처리)
- 대규모 프로젝트에서의 응답 속도 개선

### Technical Details
- `internal/worker/pool.go`: Worker Pool, 병렬 diff 파싱, 파일 타입 분류
- `internal/git/diff.go`: ParseDiffParallel 통합, convertParsedFiles 함수
- 순환 참조 문제 해결 (worker 패키지 독립화)

상세 내용은 [docs/refactoring/v0.3.0-parallel-diff-parsing.md](docs/refactoring/v0.3.0-parallel-diff-parsing.md) 참고

## [0.2.0] - 2025-02-06

### Added
- 커밋 타입 자동 추론 기능 (feat, fix, build, docs, test, refactor, chore)
- Scope 자동 추론 기능 (디렉토리 기반 분석)
- 변경 패턴 분석 (새 기능/버그 수정/설정 변경 등)
- 디렉토리 구조 분석 및 LLM 전달

### Improved
- 커밋 타입 정확도 향상 (build → feat 오류 수정)
- Scope 추천 개선 (과도한 파일 나열 → 적절한 단일 scope)
- LLM 프롬프트 강화 (더 풍부한 컨텍스트 제공)

### Technical Details
- `internal/git/diff.go`: InferCommitType, InferScopes 함수 구현
- `internal/core/prompt.go`: 프롬프트 강화 (3개 분석 함수 추가)

상세 내용은 [docs/refactoring/v0.2.0-ai-recommendation-improvement.md](docs/refactoring/v0.2.0-ai-recommendation-improvement.md) 참고

## [0.1.0] - 2025-02-05

### Initial Release
- LLM 기반 커밋 메시지 생성
- Groq API 연동
- Conventional Commit 형식 지원
- 다국어 지원 (한국어, 영어)
- 디테일 레벨 조절 (low, medium, high)

상세 내용은 [docs/refactoring/v0.1.0-initial-release.md](docs/refactoring/v0.1.0-initial-release.md) 참고