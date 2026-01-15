package floating

import "testing"

func TestCalculateLineNumber(t *testing.T) {
	tests := []struct {
		name       string
		diff       string
		cursorLine int
		want       int
	}{
		{
			name: "simple add at line 5",
			diff: `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -3,6 +3,7 @@ package main
 import "fmt"

 func main() {
+    fmt.Println("new line")
     fmt.Println("hello")
 }`,
			cursorLine: 8, // The +fmt.Println("new line") line
			want:       6, // Line 6 in new file
		},
		{
			name: "context line after hunk header",
			diff: `@@ -10,5 +10,6 @@ func foo() {
     x := 1
     y := 2
+    z := 3
     return x + y
 }`,
			cursorLine: 1, // "    x := 1" context line
			want:       10, // First line after @@ header at +10 is line 10
		},
		{
			name: "deleted line should not increment",
			diff: `@@ -1,4 +1,3 @@
 line1
-deleted
 line2
 line3`,
			cursorLine: 3, // "line2" - after deleted line
			want:       2,
		},
		{
			name: "cursor on hunk header returns start line",
			diff: `@@ -1,3 +1,3 @@
 context
 more`,
			cursorLine: 0, // On @@ line itself
			want:       1, // Falls back to cursorLine+1 when currentLine==0
		},
		{
			name: "multiple hunks - cursor on added line",
			diff: `@@ -1,3 +1,4 @@
 line1
+added1
 line2
@@ -10,3 +11,4 @@
 line10
+added10
 line11`,
			cursorLine: 6, // "+added10" line
			want:       12, // Line 12 in new file (11=line10, 12=added10)
		},
		{
			name: "brand new file",
			diff: `diff --git a/new.go b/new.go
new file mode 100644
@@ -0,0 +1,3 @@
+package main
+
+func main() {}`,
			cursorLine: 3, // "package main"
			want:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateLineNumber(tt.diff, tt.cursorLine)
			if got != tt.want {
				t.Errorf("CalculateLineNumber() = %d, want %d", got, tt.want)
			}
		})
	}
}
