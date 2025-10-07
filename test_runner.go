package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TestInput represents the parsed input test file
type TestInput struct {
	Events []TestEvent
}

// TestEvent represents a single event with result and odds
type TestEvent struct {
	Result string
	OddF   float64
	OddX   float64
	OddL   float64
}

// TestResult represents the result of running a test
type TestResult struct {
	TestName    string
	Passed      bool
	InputFile   string
	OutputFile  string
	Error       string
	Differences []string
}

// parseTestInputFile parses a test input file
func parseTestInputFile(filename string) (*TestInput, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening input file %s: %v", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	events := []TestEvent{}
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip header line
		if strings.HasPrefix(line, "result,") {
			continue
		}

		// Parse event line
		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid format at line %d in %s: expected 4 fields, got %d", lineNum, filename, len(parts))
		}

		result := strings.TrimSpace(parts[0])
		if result != "F" && result != "X" && result != "L" {
			return nil, fmt.Errorf("invalid result '%s' at line %d in %s: must be F, X, or L", result, lineNum, filename)
		}

		oddF, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid oddF '%s' at line %d in %s: %v", parts[1], lineNum, filename, err)
		}

		oddX, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid oddX '%s' at line %d in %s: %v", parts[2], lineNum, filename, err)
		}

		oddL, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid oddL '%s' at line %d in %s: %v", parts[3], lineNum, filename, err)
		}

		events = append(events, TestEvent{
			Result: result,
			OddF:   oddF,
			OddX:   oddX,
			OddL:   oddL,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input file %s: %v", filename, err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found in input file %s", filename)
	}

	return &TestInput{Events: events}, nil
}

// processTestInput processes test input and generates records using existing logic
func processTestInput(testInput *TestInput, verbose bool, hockey bool) ([]TrainerRecord, error) {
	// Extract events from test input
	events := make([]string, len(testInput.Events))
	for i, event := range testInput.Events {
		events[i] = event.Result
	}

	// Process events from oldest to newest (as they appear in input file)
	records := make([]TrainerRecord, len(events))
	detector := NewPatternDetector()

	// Initial record (previous for first event)
	previous := TrainerRecord{
		Result: "N",
		Total:  0,
	}

	for i, event := range events {
		testEvent := testInput.Events[i]

		current := TrainerRecord{
			EventNumber: i + 1,
			Result:      event,
			OddF:        testEvent.OddF,
			OddX:        testEvent.OddX,
			OddL:        testEvent.OddL,
		}

		// Apply the strategy with fixed odds
		xlWithSupport(&current, &previous, hockey)

		// Detect patterns
		detectedPatterns := detector.AddEvent(event, i+1, current)
		if len(detectedPatterns) > 0 {
			current.Pattern = strings.Join(detectedPatterns, "_")
		}

		records[i] = current
		previous = current

		if verbose {
			fmt.Printf("Событие %d: %s, Ставки: F=%.0f X=%.0f L=%.0f, Total=%.0f\n",
				i+1, event, current.BetF, current.BetX, current.BetL, current.Total)
		}
	}

	// Reverse records for display (newest first)
	records = reverseRecords(records)
	return records, nil
}

// readExpectedFile reads the expected output file
func readExpectedFile(filename string) ([]TrainerRecord, error) {
	return readCSV(filename)
}

// compareRecords compares actual and expected records
func compareRecords(actual, expected []TrainerRecord) []string {
	differences := []string{}

	if len(actual) != len(expected) {
		differences = append(differences, fmt.Sprintf("Record count mismatch: actual=%d, expected=%d", len(actual), len(expected)))
		return differences
	}

	for i := 0; i < len(actual); i++ {
		actualRec := actual[i]
		expectedRec := expected[i]

		// Compare all fields with tolerance for floating point
		if actualRec.EventNumber != expectedRec.EventNumber {
			differences = append(differences, fmt.Sprintf("Event %d: EventNumber mismatch - actual=%d, expected=%d", i+1, actualRec.EventNumber, expectedRec.EventNumber))
		}

		if actualRec.Result != expectedRec.Result {
			differences = append(differences, fmt.Sprintf("Event %d: Result mismatch - actual=%s, expected=%s", i+1, actualRec.Result, expectedRec.Result))
		}

		if !floatEqual(actualRec.OddF, expectedRec.OddF, 0.01) {
			differences = append(differences, fmt.Sprintf("Event %d: OddF mismatch - actual=%.2f, expected=%.2f", i+1, actualRec.OddF, expectedRec.OddF))
		}

		if !floatEqual(actualRec.OddX, expectedRec.OddX, 0.01) {
			differences = append(differences, fmt.Sprintf("Event %d: OddX mismatch - actual=%.2f, expected=%.2f", i+1, actualRec.OddX, expectedRec.OddX))
		}

		if !floatEqual(actualRec.OddL, expectedRec.OddL, 0.01) {
			differences = append(differences, fmt.Sprintf("Event %d: OddL mismatch - actual=%.2f, expected=%.2f", i+1, actualRec.OddL, expectedRec.OddL))
		}

		if !floatEqual(actualRec.BetF, expectedRec.BetF, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: BetF mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.BetF, expectedRec.BetF))
		}

		if !floatEqual(actualRec.BetX, expectedRec.BetX, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: BetX mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.BetX, expectedRec.BetX))
		}

		if !floatEqual(actualRec.BetL, expectedRec.BetL, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: BetL mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.BetL, expectedRec.BetL))
		}

		if !floatEqual(actualRec.LossF, expectedRec.LossF, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: LossF mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.LossF, expectedRec.LossF))
		}

		if !floatEqual(actualRec.LossX, expectedRec.LossX, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: LossX mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.LossX, expectedRec.LossX))
		}

		if !floatEqual(actualRec.LossL, expectedRec.LossL, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: LossL mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.LossL, expectedRec.LossL))
		}

		if !floatEqual(actualRec.Total, expectedRec.Total, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: Total mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.Total, expectedRec.Total))
		}

		if !floatEqual(actualRec.UF, expectedRec.UF, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: UF mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.UF, expectedRec.UF))
		}

		if !floatEqual(actualRec.UX, expectedRec.UX, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: UX mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.UX, expectedRec.UX))
		}

		if !floatEqual(actualRec.UL, expectedRec.UL, 1.0) {
			differences = append(differences, fmt.Sprintf("Event %d: UL mismatch - actual=%.0f, expected=%.0f", i+1, actualRec.UL, expectedRec.UL))
		}

		if actualRec.Pattern != expectedRec.Pattern {
			differences = append(differences, fmt.Sprintf("Event %d: Pattern mismatch - actual=%s, expected=%s", i+1, actualRec.Pattern, expectedRec.Pattern))
		}
	}

	return differences
}

// floatEqual compares two floats with tolerance
func floatEqual(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// runSingleTest runs a single test case
func runSingleTest(inputFile, expectedFile string, verbose bool, hockey bool) TestResult {
	testName := strings.TrimSuffix(filepath.Base(inputFile), ".input")

	result := TestResult{
		TestName:   testName,
		InputFile:  inputFile,
		OutputFile: expectedFile,
		Passed:     false,
	}

	// Parse input file
	testInput, err := parseTestInputFile(inputFile)
	if err != nil {
		result.Error = fmt.Sprintf("Error parsing input file: %v", err)
		return result
	}

	// Process test input
	actualRecords, err := processTestInput(testInput, verbose, hockey)
	if err != nil {
		result.Error = fmt.Sprintf("Error processing test input: %v", err)
		return result
	}

	// Read expected file
	expectedRecords, err := readExpectedFile(expectedFile)
	if err != nil {
		result.Error = fmt.Sprintf("Error reading expected file: %v", err)
		return result
	}

	// Compare records
	differences := compareRecords(actualRecords, expectedRecords)
	result.Differences = differences
	result.Passed = len(differences) == 0

	return result
}

// runTests runs all tests in the tests directory
func runTests(testsDir string, verbose bool, hockey bool) []TestResult {
	results := []TestResult{}

	// Find all .input files
	inputFiles, err := filepath.Glob(filepath.Join(testsDir, "*.input"))
	if err != nil {
		log.Printf("Error finding test files: %v", err)
		return results
	}

	for _, inputFile := range inputFiles {
		expectedFile := strings.TrimSuffix(inputFile, ".input") + ".expected"

		// Check if expected file exists
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			result := TestResult{
				TestName:   strings.TrimSuffix(filepath.Base(inputFile), ".input"),
				InputFile:  inputFile,
				OutputFile: expectedFile,
				Passed:     false,
				Error:      fmt.Sprintf("Expected file %s not found", expectedFile),
			}
			results = append(results, result)
			continue
		}

		result := runSingleTest(inputFile, expectedFile, verbose, hockey)
		results = append(results, result)
	}

	return results
}

// printTestResults prints test results
func printTestResults(results []TestResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    🧪 TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	passed := 0
	failed := 0

	for _, result := range results {
		if result.Passed {
			passed++
			fmt.Printf("✅ %s: PASSED\n", result.TestName)
		} else {
			failed++
			fmt.Printf("❌ %s: FAILED\n", result.TestName)
			if result.Error != "" {
				fmt.Printf("   Error: %s\n", result.Error)
			}
			for _, diff := range result.Differences {
				fmt.Printf("   %s\n", diff)
			}
		}
	}

	fmt.Printf("\n📊 SUMMARY:\n")
	fmt.Printf("   Total: %d\n", len(results))
	fmt.Printf("   Passed: %d\n", passed)
	fmt.Printf("   Failed: %d\n", failed)
	fmt.Printf("   Success Rate: %.1f%%\n", float64(passed)/float64(len(results))*100)

	if failed > 0 {
		fmt.Printf("\n❌ Some tests failed. Check the differences above.\n")
	} else {
		fmt.Printf("\n✅ All tests passed!\n")
	}
}
