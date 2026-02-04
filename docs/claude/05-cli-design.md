# CLI 디자인

## 기본 명령

git ai-commit

---

## 옵션

--model  
사용할 AI 모델 지정

--detail  
low / medium / high

--dry-run  
실제 커밋 없이 메시지만 출력

---

## 사용자 흐름

1. 명령 실행  
2. 메시지 후보 출력  
3. 번호 선택  
4. git commit 실행  

---

## 출력 포맷

AI 커밋 메시지 제안:

1) feat(auth): add login  
2) fix(auth): handle null  
3) refactor(auth): cleanup  

선택:
