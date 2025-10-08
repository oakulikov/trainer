package common

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Event represents a single event from the input file
type Event struct {
	Result string
	OddF   float64
	OddX   float64
	OddL   float64
}

// ReadInputFile reads and parses an .input file
func ReadInputFile(filename string) ([]Event, error) {
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
