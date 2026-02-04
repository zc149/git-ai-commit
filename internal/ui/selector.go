package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Selector는 사용자가 커밋 메시지 후보 중 하나를 선택할 수 있게 하는 인터페이스입니다.
type Selector struct{}

// NewSelector는 새로운 Selector 인스턴스를 생성합니다.
func NewSelector() *Selector {
	return &Selector{}
}

// Select는 사용자에게 후보 메시지들을 보여주고 선택을 받습니다.
func (s *Selector) Select(messages []string) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("선택할 메시지 후보가 없습니다")
	}

	fmt.Println("\n=== 커밋 메시지 후보 ===")
	for i, msg := range messages {
		fmt.Printf("%d) %s\n", i+1, msg)
	}
	fmt.Println("c) 사용자 직접 입력")
	fmt.Println("q) 종료")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n선택 (1-{} 또는 c/q): ", len(messages))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("입력 읽기 실패: %w", err)
		}

		choice := strings.TrimSpace(input)

		// 종료
		if choice == "q" || choice == "Q" {
			return "", fmt.Errorf("사용자가 종료를 선택했습니다")
		}

		// 직접 입력
		if choice == "c" || choice == "C" {
			return s.getCustomMessage()
		}

		// 숫자 선택
		index, err := strconv.Atoi(choice)
		if err != nil {
			fmt.Println("유효하지 않은 선택입니다. 다시 입력해주세요.")
			continue
		}

		if index < 1 || index > len(messages) {
			fmt.Printf("1부터 %d 사이의 숫자를 입력해주세요.\n", len(messages))
			continue
		}

		return messages[index-1], nil
	}
}

// getCustomMessage는 사용자로부터 직접 커밋 메시지를 입력받습니다.
func (s *Selector) getCustomMessage() (string, error) {
	fmt.Println("\n커밋 메시지를 직접 입력해주세요 (빈 줄로 완료):")

	var lines []string
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("입력 읽기 실패: %w", err)
		}

		line = strings.TrimSpace(line)

		// 빈 줄 입력 시 완료
		if line == "" {
			if len(lines) > 0 {
				break
			} else {
				fmt.Println("메시지를 입력해주세요.")
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
		fmt.Println("변경된 내용이 없습니다.")
		return
	}

	fmt.Println("\n=== Git Diff ===")
	fmt.Println(diffStr)
	fmt.Println("================")
}
