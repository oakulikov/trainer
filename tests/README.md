# Testing Framework for Trainer Application

This directory contains test files for the trainer application using Go's standard testing framework.

## Test File Format

Each test consists of two files:

### Input File (`.input`)
Contains events from oldest to newest with specific odds:

```
result,oddF,oddX,oddL
X,2.0,3.5,4.0
F,1.9,3.3,4.1
L,1.85,3.6,4.4
```

**Format:**
- First line: Header `result,oddF,oddX,oddL`
- Subsequent lines: Events with result (F/X/L) and odds
- Events are ordered from oldest to newest

### Expected File (`.expected`)
Contains the expected output in CSV format from newest to oldest:

```
event_number,result,oddF,oddX,oddL,betF,betX,betL,lossF,lossX,lossL,total,uf,ux,ul,pattern
3,L,1.85,3.60,4.40,16900,6600,7750,31250,23650,0,30000,2,2,0,
2,F,1.90,3.30,4.10,13800,6150,6250,0,14050,29300,20000,0,1,1,
1,X,2.00,3.50,4.00,10000,4000,3350,20000,0,13350,10000,1,0,1,
```

**Format:**
- Standard CSV output format from the trainer application
- Events are ordered from newest to oldest (reverse of input)
- Contains all calculated fields: bets, losses, totals, streaks, patterns

## Running Tests

### Run all tests:
```bash
go test -v .
```

### Run specific test:
```bash
go test -v -run TestXLWithStrategy/001_simple_test .
```

### Run tests with coverage:
```bash
go test -v -cover .
```

### Run tests with race detection:
```bash
go test -race -v .
```

## Creating New Tests

1. Create an `.input` file with your test data
2. Run the application manually to generate expected output:
   ```bash
   go run main.go -input "your/events/here" -output temp.csv
   ```
3. Copy the content from `temp.csv` to your `.expected` file
4. Run the tests to verify:
   ```bash
   go test -v .
   ```

## Test Examples

### 001_one_event
- Events: X
- Tests single event processing

### 001_simple_test
- Events: X → F → L
- Tests basic functionality with mixed results
- Verifies correct calculation of bets and losses

### 002_f_streak_test
- Events: F → F → F → F
- Tests behavior with consecutive F results
- Verifies streak handling and loss accumulation

## Test Implementation

The test framework uses Go's standard testing package with testify for assertions:

- `trainer_test.go` - Main test file in project root with test logic
- Uses subtests for each input/expected file pair
- Tests all available strategies
- Provides detailed error reporting for failed tests

## Comparison Logic

Tests compare actual vs expected output with tolerance:
- Odds: ±0.01 tolerance
- Bets/Losses/Totals: ±1.0 tolerance
- Exact match for strings (results, patterns)
- All fields must match for test to pass

## Adding New Strategies for Testing

To test a new strategy:

1. Add a new test function or extend the existing one
2. Create a mock implementation of your strategy
3. Add corresponding test files with expected results
4. Run tests to verify implementation