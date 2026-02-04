# Diff 분석 규칙

## 입력

git diff --cached 결과를 사용한다.

---

## 분석 단계

1. 파일 목록 추출  
2. 파일 타입 분류  
3. 변경 패턴 분석  
4. 커밋 타입 추론  
5. scope 추론  

---

## 파일 타입 분류 규칙

테스트 파일:
- *_test.go
- *.spec.js
- *.test.ts

문서 파일:
- README.md
- *.md

설정 파일:
- package.json
- go.mod
- *.yml

---

## 커밋 타입 추론 규칙

테스트 파일만 변경 → test  
문서만 변경 → docs  
의존성 변경 → build  
새 파일 추가 → feat  
버그 관련 키워드 → fix  
그 외 → refactor  

---

## Scope 추론

경로 기반으로 결정한다.

예시:

src/auth/... → scope: auth  
src/payment/... → scope: payment  
