package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// RegenerateError는 재추천 요청을 나타내는 에러입니다.
type RegenerateError struct{}

func (e *RegenerateError) Error() string {
	return "regenerate requested"
}

// UsePrevMessageError는 이전 메시지 사용 요청을 나타내는 에러입니다.
type UsePrevMessageError struct {
	Message string
}

func (e *UsePrevMessageError) Error() string {
	return "use previous message"
}

// Selector는 사용자가 커밋 메시지 후보 중 하나를 선택할 수 있게 하는 인터페이스입니다.
type Selector struct {
	lang string
}

// NewSelector는 새로운 Selector 인스턴스를 생성합니다.
func NewSelector(lang string) *Selector {
	return &Selector{
		lang: lang,
	}
}

// Select는 사용자에게 후보 메시지들을 보여주고 선택을 받습니다.
func (s *Selector) Select(messages []string, prevMessage string) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf(s.getMessage("error_no_candidates"))
	}

	fmt.Println("\n" + s.getMessage("header_candidates"))

	// 이전 메시지가 있으면 표시
	if prevMessage != "" {
		fmt.Printf("p) %s\n", s.formatPrevMessage(prevMessage))
	}

	for i, msg := range messages {
		s.displayFormattedMessage(i+1, msg)
	}
	fmt.Println(s.getMessage("option_custom"))
	fmt.Println(s.getMessage("option_regenerate"))
	fmt.Println(s.getMessage("option_quit"))

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n%s: ", s.formatPrompt(len(messages), prevMessage != ""))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf(s.getMessage("error_read_input"), err)
		}

		choice := strings.TrimSpace(input)

		// 종료
		if choice == "q" || choice == "Q" {
			return "", fmt.Errorf(s.getMessage("error_user_quit"))
		}

		// 이전 메시지 사용
		if choice == "p" || choice == "P" {
			if prevMessage != "" {
				return "", &UsePrevMessageError{Message: prevMessage}
			}
			fmt.Println(s.getMessage("error_no_prev_message"))
			continue
		}

		// 재추천
		if choice == "r" || choice == "R" {
			return "", &RegenerateError{}
		}

		// 직접 입력
		if choice == "c" || choice == "C" {
			return s.getCustomMessage()
		}

		// 숫자 선택
		index, err := strconv.Atoi(choice)
		if err != nil {
			fmt.Println(s.getMessage("error_invalid_choice"))
			continue
		}

		if index < 1 || index > len(messages) {
			fmt.Printf(s.getMessage("error_invalid_range")+"\n", len(messages))
			continue
		}

		return messages[index-1], nil
	}
}

// getCustomMessage는 사용자로부터 직접 커밋 메시지를 입력받습니다.
func (s *Selector) getCustomMessage() (string, error) {
	fmt.Println("\n" + s.getMessage("prompt_custom_message"))

	var lines []string
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf(s.getMessage("error_read_input"), err)
		}

		line = strings.TrimSpace(line)

		// 빈 줄 입력 시 완료
		if line == "" {
			if len(lines) > 0 {
				break
			} else {
				fmt.Println(s.getMessage("error_empty_message"))
				continue
			}
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

// DisplayDiff는 diff 결과를 표시합니다.
func (s *Selector) DisplayDiff(diffStr string) {
	if diffStr == "" {
		fmt.Println(s.getMessage("no_changes"))
		return
	}

	fmt.Println("\n" + s.getMessage("header_diff"))
	fmt.Println(diffStr)
	fmt.Println("================")
}

// getMessage는 언어에 따른 메시지를 반환합니다.
func (s *Selector) getMessage(key string) string {
	messages := map[string]map[string]string{
		"error_no_candidates": {
			"en": "No message candidates to select",
			"ko": "선택할 메시지 후보가 없습니다",
		},
		"header_candidates": {
			"en": "=== Commit Message Candidates ===",
			"ko": "=== 커밋 메시지 후보 ===",
		},
		"option_custom": {
			"en": "c) Custom input",
			"ko": "c) 사용자 직접 입력",
		},
		"option_quit": {
			"en": "q) Quit",
			"ko": "q) 종료",
		},
		"error_read_input": {
			"en": "Failed to read input: %v",
			"ko": "입력 읽기 실패: %v",
		},
		"error_user_quit": {
			"en": "User chose to quit",
			"ko": "사용자가 종료를 선택했습니다",
		},
		"error_invalid_choice": {
			"en": "Invalid choice. Please try again.",
			"ko": "유효하지 않은 선택입니다. 다시 입력해주세요.",
		},
		"error_invalid_range": {
			"en": "Please enter a number between 1 and %d.",
			"ko": "1부터 %d 사이의 숫자를 입력해주세요.",
		},
		"prompt_custom_message": {
			"en": "Enter your custom commit message (empty line to complete):",
			"ko": "커밋 메시지를 직접 입력해주세요 (빈 줄로 완료):",
		},
		"error_empty_message": {
			"en": "Please enter a message.",
			"ko": "메시지를 입력해주세요.",
		},
		"no_changes": {
			"en": "No changes to display.",
			"ko": "변경된 내용이 없습니다.",
		},
		"header_diff": {
			"en": "=== Git Diff ===",
			"ko": "=== Git Diff ===",
		},
		"option_regenerate": {
			"en": "r) Regenerate candidates",
			"ko": "r) 재추천 받기",
		},
		"error_no_prev_message": {
			"en": "No previous message available.",
			"ko": "이전 메시지가 없습니다.",
		},
		"prompt_select": {
			"en": "Select (1-%d or c/r/q)",
			"ko": "선택 (1-%d 또는 c/r/q)",
		},
		"prompt_select_with_prev": {
			"en": "Select (p/1-%d or c/r/q)",
			"ko": "선택 (p/1-%d 또는 c/r/q)",
		},
	}

	if msgMap, ok := messages[key]; ok {
		if msg, ok := msgMap[s.lang]; ok {
			return msg
		}
		return msgMap["en"] // 기본값은 영어
	}
	return key
}

// formatPrompt는 선택 프롬프트를 언어에 맞게 포맷팅합니다.
func (s *Selector) formatPrompt(count int, hasPrev bool) string {
	if hasPrev {
		prompt := s.getMessage("prompt_select_with_prev")
		return fmt.Sprintf(prompt, count)
	}
	prompt := s.getMessage("prompt_select")
	return fmt.Sprintf(prompt, count)
}

// formatPrevMessage는 이전 메시지를 포맷팅합니다.
func (s *Selector) formatPrevMessage(msg string) string {
	if s.lang == "ko" {
		return "이전 메시지 사용 (" + msg + ")"
	}
	return "Use previous message (" + msg + ")"
}

// displayFormattedMessage는 메시지를 포맷팅하여 표시합니다.
func (s *Selector) displayFormattedMessage(index int, msg string) {
	lines := strings.Split(msg, "\n")

	// 첫 번째 줄 (번호와 함께)
	if len(lines) > 0 {
		fmt.Printf("%d) %s\n", index, lines[0])
	}

	// 나머지 줄들 (들여쓰기 적용)
	indent := fmt.Sprintf("    ")
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			fmt.Printf("%s%s\n", indent, line)
		} else {
			fmt.Println()
		}
	}
}

// Add structured format for high detail level
