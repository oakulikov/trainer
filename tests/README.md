# Testing Framework for Trainer Application

This directory contains test files for the trainer application using the custom testing framework.

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
go run main.go test_runner.go -test
```

### Run tests with verbose output:
```bash
go run main.go test_runner.go -test -verbose
```

### Run tests from custom directory:
```bash
go run main.go test_runner.go -test -tests-dir /path/to/tests
```

## Creating New Tests

1. Create an `.input` file with your test data
2. Run the application manually to generate expected output:
   ```bash
   go run main.go test_runner.go -input "your/events/here" -output temp.csv
   ```
3. Copy the content from `temp.csv` to your `.expected` file
4. Run the test framework to verify

## Test Examples

### 001_simple_test
- Events: X → F → L
- Tests basic functionality with mixed results
- Verifies correct calculation of bets and losses

### 002_f_streak_test
- Events: F → F → F → F
- Tests behavior with consecutive F results
- Verifies streak handling and loss accumulation

## Test Output

The test framework provides:
- ✅ PASSED / ❌ FAILED status for each test
- Detailed difference reports for failed tests
- Summary with pass/fail counts and success rate
- Verbose mode showing step-by-step processing

## Comparison Logic

Tests compare actual vs expected output with tolerance:
- Odds: ±0.01 tolerance
- Bets/Losses/Totals: ±1.0 tolerance
- Exact match for strings (results, patterns)
- All fields must match for test to pass