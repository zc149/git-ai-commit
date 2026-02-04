package llm

// parseCommitMessages는 응답 텍스트에서 커밋 메시지 후보들을 추출합니다.
func parseCommitMessages(text string) []string {
	var messages []string
	lines := splitLines(text)

	for _, line := range lines {
		line = trimWhitespace(line)
		if line == "" {
			continue
		}

		// "1) ", "2) ", "1. ", "2. " 등의 패턴 감지
		if isNumberedFormat(line) {
			msg := removeNumberPrefix(line)
			if msg != "" {
				messages = append(messages, msg)
			}
		}
	}

	return messages
}

// splitLines은 문자열을 줄 단위로 분리합니다.
func splitLines(s string) []string {
	lines := []string{}
	currentLine := ""

	for _, r := range s {
		if r == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(r)
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// trimWhitespace는 문자열의 앞뒤 공백을 제거합니다.
func trimWhitespace(s string) string {
	return trimLeft(trimRight(s))
}

func trimLeft(s string) string {
	for len(s) > 0 {
		r := s[0]
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			s = s[1:]
		} else {
			break
		}
	}
	return s
}

func trimRight(s string) string {
	for len(s) > 0 {
		r := s[len(s)-1]
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			s = s[:len(s)-1]
		} else {
			break
		}
	}
	return s
}

// isNumberedFormat은 문자열이 번호 포맷인지 확인합니다.
func isNumberedFormat(s string) bool {
	if len(s) < 3 {
		return false
	}

	// 첫 글자가 숫자인지 확인
	if s[0] < '0' || s[0] > '9' {
		return false
	}

	// 두 번째 글자가 ) 또는 .인지 확인
	secondChar := s[1]
	return secondChar == ')' || secondChar == '.'
}

// removeNumberPrefix는 번호 접두사를 제거합니다.
func removeNumberPrefix(s string) string {
	if len(s) < 3 {
		return ""
	}

	// 첫 글자가 숫자이고 두 번째가 ) 또는 .이면 제거
	if s[0] >= '0' && s[0] <= '9' {
		secondChar := s[1]
		if secondChar == ')' || secondChar == '.' {
			return trimWhitespace(s[2:])
		}
	}

	return s
}
