# Changelog

All notable changes to this project will be documented in this file.

## [0.4.0] - Unreleased

### Added
- 커밋 메시지 재추천 기능 (r 키): 후보가 마음에 안 들 때 새로운 후보 생성
- 캐시 시스템: 선택한 메시지를 diff hash 기반으로 저장
- 이전 메시지 사용 기능 (p 키): 같은 diff로 다시 시도할 때 캐시된 메시지 재사용
- 컨트롤 플로우 개선을 위한 에러 타입 (RegenerateError, UsePrevMessageError)

### Fixed
- **치명적 버그**: 대규모 커밋(17개 이상 파일)에서 deadlock 발생 문제 해결
  - 원인: 결과 수집 goroutine이 작업 전송 후 시작되어 output 채널이 가득 참
  - 해결: 결과 수집 goroutine을 먼저 시작하여 Producer-Consumer 패턴 정석 적용
  - 영향 범위: `internal/worker/pool.go`의 `ParseDiffParallel()` 함수
- detail 기본값이 medium으로 나오는 문제: 기본값을 low로 수정
- high 디테일 레벨에서 짧은 메시지가 나오는 문제: 다중 줄 파싱 로직 개선
- feat를 build로 잘못 분류하는 문제: 의존성 파일이 있더라도 새 소스 파일이 많으면 feat로 분류하도록 로직 개선
- scope에 모든 파일 이름이 나오는 문제: 소스 파일이 있는 최상위 디렉토리만 집계하여 더 간결한 scope 추천

### Improved
- 사용자 경험 개선: 마음에 안 드는 후보를 계속 재생성 가능
- 반복 작업 효율성: 좋은 메시지를 찾으면 다음에 바로 재사용 가능
- 커밋 타입 분류 정확도: 새 기능 추가 시 의존성 업데이트가 있더라도 feat로 정확하게 분류
- Scope 추천 품질: 과도한 파일 나열 문제 해결, 더 간결하고 정확한 scope 제공

### Technical Details
- `internal/cache/cache.go`: 캐시 매니저, diff hash 기반 캐싱
- `internal/ui/selector.go`: 재추천 옵션, 이전 메시지 옵션, 커스텀 에러 타입
- `cmd/root.go`: 캐시 통합, 재추천 루프, 메시지 선택 로직
- `internal/git/diff.go`: diff hash 계산 함수

상세 내용은 [docs/refactoring/v0.4.0-message-iteration-and-cache.md](docs/refactoring/v0.4.0-message-iteration-and-cache.md) 참고

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