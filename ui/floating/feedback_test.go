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
			name: "jj diff - added line (green)",
			diff: "[1m[93mfile.go[39m[0m[2m --- Go[0m\n" +
				"[2m1 [0m[2m1 [0mpackage main\n" +
				"[2m2 [0m[2m2 [0m\n" +
				"[92;1m3 [0m[92mfunc newFunc() {}[0m",
			cursorLine: 3, // The added line
			want:       3,
		},
		{
			name: "jj diff - context line (dim)",
			diff: "[1m[93mfile.go[39m[0m[2m --- Go[0m\n" +
				"[2m1 [0m[2m1 [0mpackage main\n" +
				"[2m2 [0m[2m2 [0mfunc main() {}\n" +
				"[2m3 [0m[2m3 [0mfunc other() {}",
			cursorLine: 2, // Context line at new file line 2
			want:       2,
		},
		{
			name: "jj diff - header line falls back to cursorLine+1",
			diff: "[1m[93mfile.go[39m[0m[2m --- Go[0m\n" +
				"[2m1 [0m[2m1 [0mpackage main",
			cursorLine: 0, // Header line, no line number
			want:       1, // Fallback
		},
		{
			name: "jj diff - cursor beyond diff length",
			diff: "[2m1 [0mline1\n[2m2 [0mline2",
			cursorLine: 10,
			want:       11, // Fallback
		},
		{
			name: "jj diff - added line with space before number",
			diff: "[92;1m 5 [0m  newLine();",
			cursorLine: 0,
			want:       5,
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
