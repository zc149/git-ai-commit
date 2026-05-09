package core

import (
	"strings"
	"testing"

	"git-ai-commit/internal/git"
)

func TestGeneratePromptUsesRequestedLanguageForRequirements(t *testing.T) {
	diff := &git.DiffResult{
		Files: []git.FileChange{
			{
				Path:         "internal/git/diff.go",
				FileType:     git.FileTypeSource,
				AddedLines:   1,
				DeletedLines: 0,
				Changes:      "+func Parse() {}",
			},
		},
		CommitType: "feat",
		Scopes:     []string{"git"},
	}

	prompt := GeneratePrompt(diff, "low", "en")
	if !strings.Contains(prompt, "- Be concise") {
		t.Fatal("expected English concise requirement")
	}
	if strings.Contains(prompt, "- 간결할 것") {
		t.Fatal("English prompt must not include Korean concise requirement")
	}
}
