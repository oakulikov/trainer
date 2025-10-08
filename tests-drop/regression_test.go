package testsdrop

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/holygun/go-trainer/common"
	"github.com/holygun/go-trainer/trainer"

	"github.com/stretchr/testify/assert"
)

// Debug flag for detailed output
var debug = flag.Bool("debug", false, "enable debug output for tests")

// TestRegressionSuite runs all regression tests from .input/.expected files
func TestRegressionSuite(t *testing.T) {
	// Явно вызываем flag.Parse() чтобы обработать флаги
	flag.Parse()

	// Find all .input files in the current directory
	testFiles, err := filepath.Glob("*.input")
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	if len(testFiles) == 0 {
		t.Log("No regression tests found.")
		return
	}

	for _, testFile := range testFiles {
		t.Run(testFile, func(t *testing.T) {
			// Construct paths for .input and .expected files
			inputPath := testFile
			expectedPath := strings.TrimSuffix(testFile, ".input") + ".expected"

			// Read the input events
			inputEvents, err := readInputFile(inputPath)
			if err != nil {
				t.Fatalf("Failed to read input file %s: %v", inputPath, err)
			}

			// Read the expected output
			expectedOutput, err := readExpectedFile(expectedPath)
			if err != nil {
				t.Fatalf("Failed to read expected file %s: %v", expectedPath, err)
			}

			// Process the input events and generate actual output
			flags := trainer.Flags{
				Input:    "",
				Output:   "",
				Verbose:  false,
				Debug:    *debug,
				Report:   "",
				Hockey:   false,
				Strategy: "xlDrop",
			}
			actualOutput, err := processInputEvents(inputEvents, flags)
			if err != nil {
				t.Fatalf("Failed to process input events: %v", err)
			}

			// Сохранение в CSV
			// if err := trainer.SaveToCSV(actualOutput, strings.TrimSuffix(testFile, ".input")+".actual"); err != nil {
			// 	t.Fatalf("Ошибка сохранения CSV: %v", err)
			// }

			// Compare the actual output with expected output
			assert.Equal(t, expectedOutput, actualOutput, "Output mismatch for %s", testFile)
		})
	}
}

// readInputFile reads and parses an .input file
func readInputFile(filename string) ([]common.Event, error) {
	return common.ReadInputFile(filename)
}

// readExpectedFile reads and parses an .expected file
func readExpectedFile(filename string) ([]trainer.TrainerRecord, error) {
	return trainer.ReadCSV(filename)
}

// processInputEvents processes input events and generates output rows using the trainer package
func processInputEvents(events []common.Event, flags trainer.Flags) ([]trainer.TrainerRecord, error) {
	// Convert events to the format expected by the trainer package
	eventStrings := make([]string, len(events))
	odds := make([]struct{ OddF, OddX, OddL float64 }, len(events))

	for i, event := range events {
		eventStrings[i] = event.Result
		odds[i] = struct{ OddF, OddX, OddL float64 }{
			OddF: event.OddF,
			OddX: event.OddX,
			OddL: event.OddL,
		}
	}

	// Get the default strategy
	strategy, err := trainer.GetStrategy("xlDrop")
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy: %v", err)
	}

	// Generate records using the trainer package with specified odds
	records := trainer.GenerateRecordsWithOdds(eventStrings, odds, flags, strategy)

	if flags.Debug {
		fmt.Printf("\n=== DEBUG: Generated records (oldest to newest) ===\n")
		for i, record := range records {
			fmt.Printf("Step %d: %s,%.2f,%.2f,%.2f -> bets: %.0f,%.0f,%.0f, losses: %.0f,%.0f,%.0f, total: %.0f, streaks: %.0f,%.0f,%.0f, pattern: %s\n",
				i+1, record.Result, record.OddF, record.OddX, record.OddL,
				record.BetF, record.BetX, record.BetL, record.LossF, record.LossX, record.LossL,
				record.Total, record.UF, record.UX, record.UL, record.Pattern)
		}
	}

	// Reverse records to match expected format (newest first)
	records = trainer.ReverseRecords(records)

	return records, nil
}
