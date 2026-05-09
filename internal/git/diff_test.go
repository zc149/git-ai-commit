package git

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseDiffRetainsLongLines(t *testing.T) {
	longLine := "+" + strings.Repeat("x", 70*1024)
	diff := fmt.Sprintf(`diff --git a/internal/long.go b/internal/long.go
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/internal/long.go
@@ -0,0 +1 @@
%s
`, longLine)

	result, err := ParseDiff(diff)
	if err != nil {
		t.Fatalf("ParseDiff returned error: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	if !result.Files[0].IsNew {
		t.Fatal("expected file to be marked new")
	}
	if !strings.Contains(result.Files[0].Changes, longLine) {
		t.Fatal("expected long diff line to be retained")
	}
}

func TestParseDiffHandlesPathWithSpaces(t *testing.T) {
	diff := `diff --git a/internal/file with space.go b/internal/file with space.go
index 1111111..2222222 100644
--- a/internal/file with space.go	
+++ b/internal/file with space.go	
@@ -1 +1 @@
-old
+new
`

	result, err := ParseDiff(diff)
	if err != nil {
		t.Fatalf("ParseDiff returned error: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	if result.Files[0].Path != "internal/file with space.go" {
		t.Fatalf("expected path with spaces to be preserved, got %q", result.Files[0].Path)
	}
}
