# AI API 연동

## 제공자 인터페이스

모든 LLM 제공자는 다음 인터페이스를 구현한다.

interface Provider {
  Generate(prompt string) (string, error)
}

---

## 지원 제공자

- Claude
- OpenAI
- Gemini

---

## 설정 방식

설정 파일 경로:

~/.git-ai/config.json

형식:

{
  "default_model": "claude",
  "api_keys": {
    "claude": "..."
  }
}

---

## 호출 흐름

1. prompt 생성  
2. 선택된 provider 호출  
3. 결과 파싱  
4. 후보 목록 반환  
