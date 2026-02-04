package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Commit은 git commit을 실행합니다.
func Commit(message string) error {
	// commit 실행
	cmd := exec.Command("git", "commit", "-m", message)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit 실패: %w\n%s", err, string(output))
	}

	fmt.Println(string(output))
	return nil
}

// IsCleanWorkingTree는 working tree가 clean한지 확인합니다.
func IsCleanWorkingTree() (bool, error) {
	// status 확인
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status 실패: %w", err)
	}

	return len(strings.TrimSpace(string(output))) == 0, nil
}

// GetStagedFiles는 staged된 파일들을 반환합니다.
func GetStagedFiles() ([]string, error) {
	// staged 파일 확인
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached 실패: %w", err)
	}

	if len(output) == 0 {
		return []string{}, nil
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	return files, nil
}
