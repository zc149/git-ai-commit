package git

import (
	"reflect"
	"testing"
)

func TestAnalyzeFilesRecommendedType(t *testing.T) {
	tests := []struct {
		name  string
		files []FileChange
		want  string
	}{
		{
			name: "new public source file is feature",
			files: []FileChange{
				sourceChange("internal/auth/service.go", true, false, "+func Login() error {\n+\treturn nil\n+}"),
			},
			want: "feat",
		},
		{
			name: "error handling source change is fix",
			files: []FileChange{
				sourceChange("internal/auth/service.go", false, false, "-return user.Name\n+if user == nil {\n+\treturn \"\"\n+}\n+return user.Name"),
			},
			want: "fix",
		},
		{
			name: "source cleanup without behavior signal is refactor",
			files: []FileChange{
				sourceChange("internal/auth/service.go", false, false, "-name := strings.TrimSpace(input)\n+normalizedName := strings.TrimSpace(input)"),
			},
			want: "refactor",
		},
		{
			name: "docs only is docs",
			files: []FileChange{
				fileChange("README.md", FileTypeDoc, false, false, "+## Usage"),
			},
			want: "docs",
		},
		{
			name: "tests only is test",
			files: []FileChange{
				fileChange("internal/auth/service_test.go", FileTypeTest, false, false, "+func TestLogin(t *testing.T) {}"),
			},
			want: "test",
		},
		{
			name: "dependency only is build",
			files: []FileChange{
				fileChange("go.mod", FileTypeConfig, false, false, "+require example.com/pkg v1.0.0"),
			},
			want: "build",
		},
		{
			name: "source remains primary when dependency also changes",
			files: []FileChange{
				sourceChange("internal/auth/service.go", true, false, "+func Login() error {\n+\treturn nil\n+}"),
				fileChange("go.mod", FileTypeConfig, false, false, "+require example.com/pkg v1.0.0"),
			},
			want: "feat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := AnalyzeFiles(tt.files)
			if got := analysis.RecommendedType(); got != tt.want {
				t.Fatalf("RecommendedType() = %q, want %q; scores: %#v", got, tt.want, analysis.TypeScores)
			}
		})
	}
}

func TestAnalyzeFilesCollectsLineStatsAndSignals(t *testing.T) {
	analysis := AnalyzeFiles([]FileChange{
		sourceChange("internal/auth/service.go", false, false, "-return user.Name\n+if user == nil {\n+\treturn \"\"\n+}\n+return user.Name"),
	})

	if analysis.AddedLines != 4 || analysis.DeletedLines != 1 {
		t.Fatalf("line stats = +%d/-%d, want +4/-1", analysis.AddedLines, analysis.DeletedLines)
	}
	if !containsString(analysis.Signals, "bug/error-handling signal") {
		t.Fatalf("expected bug/error signal, got %#v", analysis.Signals)
	}
}

func TestInferScopes(t *testing.T) {
	tests := []struct {
		name  string
		files []FileChange
		want  []string
	}{
		{
			name: "common source package uses leaf directory",
			files: []FileChange{
				sourceChange("internal/git/diff.go", false, false, "+diff"),
				sourceChange("internal/git/analysis.go", true, false, "+analysis"),
			},
			want: []string{"git"},
		},
		{
			name: "root source file trims extension",
			files: []FileChange{
				sourceChange("main.go", false, false, "+main"),
			},
			want: []string{"main"},
		},
		{
			name: "equal unrelated source directories use multiple",
			files: []FileChange{
				sourceChange("cmd/root.go", false, false, "+cmd"),
				sourceChange("internal/git/diff.go", false, false, "+diff"),
			},
			want: []string{"multiple"},
		},
		{
			name: "dependency-only change uses deps",
			files: []FileChange{
				fileChange("go.mod", FileTypeConfig, false, false, "+require example.com/pkg v1.0.0"),
			},
			want: []string{"deps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferScopes(tt.files); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("InferScopes() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func sourceChange(path string, isNew bool, isDeleted bool, changes string) FileChange {
	return fileChange(path, FileTypeSource, isNew, isDeleted, changes)
}

func fileChange(path string, fileType FileType, isNew bool, isDeleted bool, changes string) FileChange {
	added, deleted := countChangedLines(changes)
	return FileChange{
		Path:         path,
		FileType:     fileType,
		IsNew:        isNew,
		IsDeleted:    isDeleted,
		AddedLines:   added,
		DeletedLines: deleted,
		Changes:      changes,
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
