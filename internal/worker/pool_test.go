package worker

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseDiffParallelPreservesFileStateAndOrder(t *testing.T) {
	diff := strings.Join([]string{
		newFileDiff("internal/a.go", "+package internal"),
		modifiedFileDiff("internal/file with space.go", "-old", "+new"),
		deletedFileDiff("internal/c.go", "-package internal"),
		newFileDiff("docs/readme.md", "+# Docs"),
	}, "\n")

	files, err := ParseDiffParallel(diff, 2)
	if err != nil {
		t.Fatalf("ParseDiffParallel returned error: %v", err)
	}
	if files == nil {
		t.Fatal("ParseDiffParallel returned nil for a 4-file diff")
	}
	if len(files) != 4 {
		t.Fatalf("expected 4 files, got %d", len(files))
	}

	paths := []string{files[0].Path, files[1].Path, files[2].Path, files[3].Path}
	expectedPaths := []string{"internal/a.go", "internal/file with space.go", "internal/c.go", "docs/readme.md"}
	if !reflect.DeepEqual(paths, expectedPaths) {
		t.Fatalf("paths mismatch\nexpected: %#v\nactual:   %#v", expectedPaths, paths)
	}

	if !files[0].IsNew {
		t.Fatalf("expected %s to be marked new", files[0].Path)
	}
	if files[1].IsNew || files[1].IsDeleted {
		t.Fatalf("expected %s to be marked modified only", files[1].Path)
	}
	if !files[2].IsDeleted {
		t.Fatalf("expected %s to be marked deleted", files[2].Path)
	}
	if !files[3].IsNew {
		t.Fatalf("expected %s to be marked new", files[3].Path)
	}
	if strings.Contains(files[0].Changes, "new file mode") {
		t.Fatalf("status metadata leaked into Changes: %q", files[0].Changes)
	}
}

func TestSplitFileDiffsRetainsLongLines(t *testing.T) {
	longLine := "+" + strings.Repeat("x", 70*1024)
	diff := newFileDiff("internal/long.go", longLine)

	fileDiffs := splitFileDiffs(diff)
	if len(fileDiffs) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(fileDiffs))
	}
	if !strings.Contains(fileDiffs[0].Body, longLine) {
		t.Fatal("expected long diff line to be retained")
	}
}

func newFileDiff(path string, addedLine string) string {
	return fmt.Sprintf(`diff --git a/%[1]s b/%[1]s
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/%[1]s
@@ -0,0 +1 @@
%[2]s
`, path, addedLine)
}

func modifiedFileDiff(path string, removedLine string, addedLine string) string {
	return fmt.Sprintf(`diff --git a/%[1]s b/%[1]s
index 1111111..2222222 100644
--- a/%[1]s
+++ b/%[1]s
@@ -1 +1 @@
%[2]s
%[3]s
`, path, removedLine, addedLine)
}

func deletedFileDiff(path string, removedLine string) string {
	return fmt.Sprintf(`diff --git a/%[1]s b/%[1]s
deleted file mode 100644
index 1111111..0000000
--- a/%[1]s
+++ /dev/null
@@ -1 +0,0 @@
%[2]s
`, path, removedLine)
}
