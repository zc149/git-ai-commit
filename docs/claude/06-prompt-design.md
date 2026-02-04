# 프롬프트 설계

## 목적

LLM에게 최소한의 정보로 정확한 커밋 메시지를 생성하도록 한다.

---

## 입력 데이터 구조

- 변경 파일 목록  
- 추천 타입  
- 추천 scope  
- 변경 요약  

---

## 기본 프롬프트 템플릿

다음 정보를 기반으로 Conventional Commit 메시지를 생성하세요.

추천 타입: {type}  
추천 scope: {scope}

변경 내용 요약:
{summary}

요구사항:

- 간결할 것
- Conventional Commit 형식
- 3개의 후보 생성

---

## 응답 형식

1) type(scope): message  
2) type(scope): message  
3) type(scope): message  
