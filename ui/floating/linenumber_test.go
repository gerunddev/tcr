package floating

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestFixture represents a single test case from diff-line-dataset.json
type TestFixture struct {
	Iteration          int    `json:"iteration"`
	Diff               string `json:"diff"`
	ChosenDiffLine     string `json:"chosenDiffLine"`
	ChosenLineContent  string `json:"chosenLineContent"`
	ExpectedLineNumber int    `json:"expectedLineNumber"`
}

func loadTestFixtures(t *testing.T) []TestFixture {
	t.Helper()

	data, err := os.ReadFile("../../diff-line-dataset.json")
	if err != nil {
		t.Fatalf("Failed to load test fixtures: %v", err)
	}

	var fixtures []TestFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("Failed to parse test fixtures: %v", err)
	}

	return fixtures
}

func TestExtractLineNumberFromDiffLine(t *testing.T) {
	fixtures := loadTestFixtures(t)

	for _, tc := range fixtures {
		t.Run(
			fmt.Sprintf("iteration_%d_line_%d", tc.Iteration, tc.ExpectedLineNumber),
			func(t *testing.T) {
				got := ExtractLineNumberFromDiffLine(tc.ChosenDiffLine)
				if got != tc.ExpectedLineNumber {
					t.Errorf("ExtractLineNumberFromDiffLine() = %d, want %d\nInput: %q",
						got, tc.ExpectedLineNumber, tc.ChosenDiffLine)
				}
			},
		)
	}
}
