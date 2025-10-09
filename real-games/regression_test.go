package tests

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

// TestRegressionSuite runs all regression tests from results files
func TestRegressionSuite(t *testing.T) {
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
			suffixResults := ".expected"
			fileName := strings.TrimSuffix(filepath.Base(testFile), ".input")
			parts := strings.Split(fileName, "_")
			if len(parts) == 1 {
				suffixResults = ".actual"
				parts = strings.Split(fileName, "-")
			}
			strategyName := parts[0]

			// Construct paths for .input and results files
			inputPath := testFile
			resultsPath := strings.TrimSuffix(testFile, ".input") + suffixResults

			if *debug {
				fmt.Printf("\n=== DEBUG: Processing test file %s ===\n", testFile)
			}

			// Read the input events
			inputEvents, err := readInputFile(inputPath)
			if err != nil {
				t.Fatalf("Failed to read input file %s: %v", inputPath, err)
			}

			if *debug {
				fmt.Printf("Input events (%d):\n", len(inputEvents))
				for i, event := range inputEvents {
					fmt.Printf("  %d: %s,%.2f,%.2f,%.2f\n", i+1, event.Result, event.OddF, event.OddX, event.OddL)
				}
			}

			// Read the results output
			resultsOutput, err := readResultsFile(resultsPath)
			if err != nil {
				t.Fatalf("Failed to read results file %s: %v", resultsPath, err)
			}

			if *debug {
				fmt.Printf("Results output (%d records):\n", len(resultsOutput))
				for i, record := range resultsOutput {
					fmt.Printf("  %d: %s,%s,%.2f,%.2f,%.2f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f\n",
						i+1, record.Result, record.Pattern, record.OddF, record.OddX, record.OddL,
						record.BetF, record.BetX, record.BetL, record.LossF, record.LossX, record.LossL,
						record.Total, record.UF, record.UX, record.UL)
				}
			}

			// Process the input events and generate actual output
			flags := trainer.Flags{
				Input:    "",
				Output:   "",
				Verbose:  false,
				Debug:    *debug,
				Report:   "",
				Hockey:   false,
				Strategy: strategyName,
				Real:     false,
				Force:    false,
				Testing:  true,
			}
			actualOutput, err := processInputEvents(inputEvents, flags)
			if err != nil {
				t.Fatalf("Failed to process input events: %v", err)
			}

			if *debug {
				fmt.Printf("Actual output (%d records):\n", len(actualOutput))
				for i, record := range actualOutput {
					fmt.Printf("  %d: %s,%s,%.2f,%.2f,%.2f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f\n",
						i+1, record.Result, record.Pattern, record.OddF, record.OddX, record.OddL,
						record.BetF, record.BetX, record.BetL, record.LossF, record.LossX, record.LossL,
						record.Total, record.UF, record.UX, record.UL)
				}
			}

			// Compare the actual output with results output
			assert.Equal(t, resultsOutput, actualOutput, "Output mismatch for %s", testFile)

			if *debug {
				fmt.Printf("=== DEBUG: Test %s completed ===\n\n", testFile)
			}
		})
	}
}

// readInputFile reads and parses an .input file
func readInputFile(filename string) ([]common.Event, error) {
	return common.ReadInputFile(filename)
}

// readResultsFile reads and parses an results file
func readResultsFile(filename string) ([]trainer.TrainerRecord, error) {
	return trainer.ReadCSV(filename)
}

// processInputEvents processes input events and generates output rows using the trainer package
func processInputEvents(events []common.Event, flags trainer.Flags) ([]trainer.TrainerRecord, error) {
	strategyName := flags.Strategy

	if flags.Debug {
		fmt.Printf("\n=== DEBUG: Starting event processing ===\n")
		fmt.Printf("Processing %d events with %s strategy\n", len(events), strategyName)
	}

	// Convert events to the format results by the trainer package
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
	strategy, err := trainer.GetStrategy(strategyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy: %v", err)
	}

	if flags.Debug {
		fmt.Printf("Strategy loaded: %s - %s\n", strategy.Name(), strategy.Description())
		fmt.Printf("Event sequence (oldest to newest): %s\n", strings.Join(eventStrings, "/"))
		fmt.Printf("\n=== DEBUG: Step-by-step processing ===\n")
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

	// Reverse records to match results format (newest first)
	records = trainer.ReverseRecords(records)

	if flags.Debug {
		fmt.Printf("\n=== DEBUG: Final records (newest to oldest) ===\n")
		for i, record := range records {
			fmt.Printf("Record %d: %s,%s,%.2f,%.2f,%.2f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f\n",
				i+1, record.Result, record.Pattern, record.OddF, record.OddX, record.OddL,
				record.BetF, record.BetX, record.BetL, record.LossF, record.LossX, record.LossL,
				record.Total, record.UF, record.UX, record.UL)
		}
		fmt.Printf("=== DEBUG: Event processing completed ===\n\n")
	}

	return records, nil
}
