package tests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/holygun/go-trainer/trainer"

	"github.com/stretchr/testify/assert"
)

// TestRegressionSuite runs all regression tests from .input/.expected files
func TestRegressionSuite(t *testing.T) {
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
			actualOutput, err := processInputEvents(inputEvents)
			if err != nil {
				t.Fatalf("Failed to process input events: %v", err)
			}

			// Compare the actual output with expected output
			assert.Equal(t, expectedOutput, actualOutput, "Output mismatch for %s", testFile)
		})
	}
}

// Event represents a single event from the input file
type Event struct {
	Result string
	OddF   float64
	OddX   float64
	OddL   float64
}

// readInputFile reads and parses an .input file
func readInputFile(filename string) ([]Event, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)

	// Skip header line
	if !scanner.Scan() {
		return events, nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid line format in %s: %s", filename, line)
		}

		oddF, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid oddF value in %s: %s", filename, parts[1])
		}

		oddX, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid oddX value in %s: %s", filename, parts[2])
		}

		oddL, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid oddL value in %s: %s", filename, parts[3])
		}

		events = append(events, Event{
			Result: parts[0],
			OddF:   oddF,
			OddX:   oddX,
			OddL:   oddL,
		})
	}

	return events, scanner.Err()
}

// readExpectedFile reads and parses an .expected file
func readExpectedFile(filename string) ([]trainer.TrainerRecord, error) {
	return trainer.ReadCSV(filename)
}

// processInputEvents processes input events and generates output rows using the trainer package
func processInputEvents(events []Event) ([]trainer.TrainerRecord, error) {
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
	strategy, err := trainer.GetStrategy("xlWithSupport")
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy: %v", err)
	}

	// Generate records using the trainer package with specified odds
	records := trainer.GenerateRecordsWithOdds(eventStrings, odds, false, false, strategy)

	// Reverse records to match expected format (newest first)
	records = trainer.ReverseRecords(records)

	return records, nil
}
