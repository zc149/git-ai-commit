# MVP 스펙

## MVP 목표

최소한의 기능으로 실제 사용 가능한 버전을 만든다.

---

## 필수 기능

1. git diff 읽기  
2. 변경 내용 분석  
3. AI로 메시지 생성  
4. 후보 메시지 출력  
5. 선택 후 커밋 실행  

---

## 지원 명령어

기본 명령:

git ai-commit

옵션:

git ai-commit --model=claude  
git ai-commit --detail=medium  
git ai-commit --dry-run  

---

## 제외 범위

MVP에서는 다음 기능은 포함하지 않는다.

- AST 기반 코드 분석  
- 자동 커밋 분리  
- PR 자동 생성  
- IDE 플러그인  

---

## 성공 기준

다음 조건을 만족하면 MVP 성공으로 본다.

- 한 줄 명령으로 동작  
- 실제 프로젝트에서 사용 가능  
- 설치가 간단  
- 메시지가 충분히 정확  
