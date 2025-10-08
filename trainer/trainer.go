package trainer

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/holygun/go-trainer/common"
)

// Flags содержит все флаги командной строки
type Flags struct {
	Input    string
	Output   string
	Verbose  bool
	Debug    bool
	Report   string
	Hockey   bool
	Strategy string
	Real     bool
}

const DEFAULT_BET = 10000
const PARTIAL_COVERAGE_MULT = 1

// Pattern определяет структуру паттерна
type Pattern struct {
	ID          string
	Description string
}

// Зарегистрированные паттерны
var patterns = []Pattern{
	{"RED", "three or more metrics > 20 * DEFAULT_BET"},
	{"YELLOW", "two or more metrics > 10 * DEFAULT_BET or one metric > 20 * DEFAULT_BET"},
	{"GREEN", "one metric > 10 * DEFAULT_BET"},
}

// Зарегистрированные флаги для фильтрации файлов
var registeredFlags = map[string]bool{
	"hockey": false,
}

// RegisterFlag регистрирует новый флаг
func RegisterFlag(flag string) {
	registeredFlags[flag] = true
}

// IsFlagRegistered проверяет, зарегистрирован ли флаг
func IsFlagRegistered(flag string) bool {
	_, ok := registeredFlags[flag]
	return ok
}

// GetRegisteredFlags возвращает список зарегистрированных флагов
func GetRegisteredFlags() map[string]bool {
	return registeredFlags
}

// PatternDetector детектор паттернов
type PatternDetector struct {
	recentEvents []string
	windowSize   int
}

// NewPatternDetector создает новый детектор
func NewPatternDetector() *PatternDetector {
	return &PatternDetector{
		recentEvents: make([]string, 0),
		windowSize:   10,
	}
}

// AddEvent добавляет событие и проверяет паттерны
func (pd *PatternDetector) AddEvent(event string, eventNumber int, record TrainerRecord) []string {
	pd.recentEvents = append(pd.recentEvents, event)
	if len(pd.recentEvents) > pd.windowSize {
		pd.recentEvents = pd.recentEvents[1:]
	}

	detectedPatterns := []string{}

	// Проверяем паттерны
	for _, pattern := range patterns {
		if pd.checkPattern(pattern, record) {
			detectedPatterns = append(detectedPatterns, pattern.ID)
			fmt.Printf("⚠️ Событие номер %d: обнаружен паттерн %s - %s\n", eventNumber, pattern.ID, pattern.Description)
			break
		}
	}

	return detectedPatterns
}

// checkPattern проверяет конкретный паттерн
func (pd *PatternDetector) checkPattern(pattern Pattern, record TrainerRecord) bool {
	metrics := []float64{
		record.BetF,
		record.BetX,
		record.BetL,
		record.LossF,
		record.LossX,
		record.LossL,
	}
	switch pattern.ID {
	case "RED":
		threshold := 20.0 * DEFAULT_BET
		count := 0
		for _, value := range metrics {
			if value > threshold {
				count += 1
			}
		}
		return count >= 3
	case "YELLOW":
		small_threshold, big_threshold := 10.0*DEFAULT_BET, 20.0*DEFAULT_BET
		small_count, big_count := 0, 0
		for _, value := range metrics {
			if value > big_threshold {
				big_count += 1
			}
			if value > small_threshold {
				small_count += 1
			}
		}
		return small_count >= 2 || big_count >= 1
	case "GREEN":
		threshold := 10.0 * DEFAULT_BET
		count := 0
		for _, value := range metrics {
			if value > threshold {
				count += 1
			}
		}
		return count >= 1
	}
	return false
}

// TrainerRecord представляет одну запись в CSV
type TrainerRecord struct {
	EventNumber int     // Номер события
	Result      string  // F, X или L
	OddF        float64 // Коэффициент F
	OddX        float64 // Коэффициент X
	OddL        float64 // Коэффициент L
	BetF        float64 // Ставка F
	BetX        float64 // Ставка X
	BetL        float64 // Ставка L
	LossF       float64 // Убыток F
	LossX       float64 // Убыток X
	LossL       float64 // Убыток L
	Total       float64 // Итого
	UF          float64 // Серия без F
	UX          float64 // Серия без X
	UL          float64 // Серия без L
	Pattern     string  // Обнаруженные паттерны
}

// Config содержит конфигурацию тренажера
type Config struct {
	DefaultBetF float64
	RoundUp     float64
	OddsRanges  struct {
		OddF        Range
		OddX        Range
		OddL        Range
		MarginRange Range
	}
}

// Range представляет диапазон значений
type Range struct {
	Min float64
	Max float64
}

// Статистика для отчета
type Stats struct {
	TotalRecords     int
	EventCounts      map[string]int
	EventPercentages map[string]float64
	MaxBets          map[string]float64
	MaxLosses        map[string]float64
	MaxStreaks       map[string]int
}

var config = Config{
	DefaultBetF: DEFAULT_BET,
	RoundUp:     50,
	OddsRanges: struct {
		OddF        Range
		OddX        Range
		OddL        Range
		MarginRange Range
	}{
		OddF:        Range{Min: 1.8, Max: 2.1},
		OddX:        Range{Min: 3.3, Max: 3.9},
		OddL:        Range{Min: 4.0, Max: 5.0},
		MarginRange: Range{Min: 1.05, Max: 1.1},
	},
}

// parseEvents парсит строку событий F/X/L
func ParseEvents(input string) []string {
	parts := strings.Split(strings.TrimSpace(input), "/")
	events := []string{}

	for _, part := range parts {
		event := strings.ToUpper(strings.TrimSpace(part))
		if event == "F" || event == "X" || event == "L" {
			events = append(events, event)
		}
	}

	return events
}

// reverseSlice реверсирует слайс строк
func ReverseSlice(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// reverseRecords реверсирует слайс записей
func ReverseRecords(records []TrainerRecord) []TrainerRecord {
	result := make([]TrainerRecord, len(records))
	for i, v := range records {
		result[len(records)-1-i] = v
	}
	return result
}

// roundUp округляет значение вверх до кратного config.RoundUp
func roundUp(value float64) float64 {
	return math.Ceil(value/config.RoundUp) * config.RoundUp
}

// calcBet вычисляет ставку
func calcBet(value, odd float64) float64 {
	return roundUp(value / (odd - 1))
}

// generateOdds генерирует коэффициенты с учетом ограничений
func generateOdds(flags Flags) (float64, float64, float64) {
	rand.Seed(time.Now().UnixNano())
	maxAttempts := 1000

	for i := 0; i < maxAttempts; i++ {
		oddF := config.OddsRanges.OddF.Min + rand.Float64()*(config.OddsRanges.OddF.Max-config.OddsRanges.OddF.Min)
		oddX := config.OddsRanges.OddX.Min + rand.Float64()*(config.OddsRanges.OddX.Max-config.OddsRanges.OddX.Min)
		oddL := config.OddsRanges.OddL.Min + rand.Float64()*(config.OddsRanges.OddL.Max-config.OddsRanges.OddL.Min)

		margin := 1/oddF + 1/oddX + 1/oddL

		if margin >= config.OddsRanges.MarginRange.Min && margin <= config.OddsRanges.MarginRange.Max {
			// Округляем до 2 знаков после запятой
			oddF = math.Round(oddF*100) / 100
			oddX = math.Round(oddX*100) / 100
			oddL = math.Round(oddL*100) / 100

			if flags.Hockey {
				return oddF, oddL, oddX
			}

			return oddF, oddX, oddL
		}
	}

	if flags.Debug {
		fmt.Printf("DEBUG: [generateOdds] fallback to default odds")
	}

	if flags.Hockey {
		return 2, 4, 3.5
	}

	// Значения по умолчанию
	return 2, 3.5, 4
}

// GenerateRecords генерирует записи для событий
func GenerateRecords(eventsFromOldest []string, flags Flags, strategy Strategy) []TrainerRecord {
	records := make([]TrainerRecord, len(eventsFromOldest))
	detector := NewPatternDetector()

	// Начальная запись (предыдущая для первого события)
	previous := TrainerRecord{
		Result: "N",
		Total:  0,
	}

	for i, event := range eventsFromOldest {
		oddF, oddX, oddL := generateOdds(flags)

		current := TrainerRecord{
			EventNumber: i + 1,
			Result:      event,
			OddF:        oddF,
			OddX:        oddX,
			OddL:        oddL,
		}

		// Применяем стратегию
		strategy.Calculate(&current, &previous, flags)

		// Детектируем паттерны
		detectedPatterns := detector.AddEvent(event, i+1, current)
		if len(detectedPatterns) > 0 {
			current.Pattern = strings.Join(detectedPatterns, "_")
		}

		records[i] = current
		previous = current

		if flags.Verbose {
			fmt.Printf("Событие %d: %s, Ставки: F=%.0f X=%.0f L=%.0f, Total=%.0f\n",
				i+1, event, current.BetF, current.BetX, current.BetL, current.Total)
		}
	}

	return records
}

// ReadCSV читает CSV файл и возвращает записи
func ReadCSV(filename string) ([]TrainerRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Skip header if present
	startIdx := 0
	if len(records) > 0 && records[0][0] == "event_number" {
		startIdx = 1
	}

	// Convert string records to TrainerRecord structs
	trainerRecords := make([]TrainerRecord, 0, len(records)-startIdx)

	for i := startIdx; i < len(records); i++ {
		row := records[i]

		// Ensure we have enough columns
		if len(row) < 16 {
			return nil, fmt.Errorf("invalid CSV format at row %d: expected at least 16 columns, got %d", i+1, len(row))
		}

		// Parse each field
		eventNumber, err := strconv.Atoi(row[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing event_number at row %d: %v", i+1, err)
		}

		oddF, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing oddF at row %d: %v", i+1, err)
		}

		oddX, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing oddX at row %d: %v", i+1, err)
		}

		oddL, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing oddL at row %d: %v", i+1, err)
		}

		betF, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing betF at row %d: %v", i+1, err)
		}

		betX, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing betX at row %d: %v", i+1, err)
		}

		betL, err := strconv.ParseFloat(row[7], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing betL at row %d: %v", i+1, err)
		}

		lossF, err := strconv.ParseFloat(row[8], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing lossF at row %d: %v", i+1, err)
		}

		lossX, err := strconv.ParseFloat(row[9], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing lossX at row %d: %v", i+1, err)
		}

		lossL, err := strconv.ParseFloat(row[10], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing lossL at row %d: %v", i+1, err)
		}

		total, err := strconv.ParseFloat(row[11], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing total at row %d: %v", i+1, err)
		}

		uf, err := strconv.ParseFloat(row[12], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing uf at row %d: %v", i+1, err)
		}

		ux, err := strconv.ParseFloat(row[13], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing ux at row %d: %v", i+1, err)
		}

		ul, err := strconv.ParseFloat(row[14], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing ul at row %d: %v", i+1, err)
		}

		// Create TrainerRecord
		record := TrainerRecord{
			EventNumber: eventNumber,
			Result:      row[1],
			OddF:        oddF,
			OddX:        oddX,
			OddL:        oddL,
			BetF:        betF,
			BetX:        betX,
			BetL:        betL,
			LossF:       lossF,
			LossX:       lossX,
			LossL:       lossL,
			Total:       total,
			UF:          uf,
			UX:          ux,
			UL:          ul,
			Pattern:     row[15], // Pattern is the last column
		}

		trainerRecords = append(trainerRecords, record)
	}

	return trainerRecords, nil
}

// ReadInputFile читает и парсит .input файл
func ReadInputFile(filename string) ([]common.Event, error) {
	return common.ReadInputFile(filename)
}

// SaveToCSV сохраняет записи в CSV файл
func SaveToCSV(records []TrainerRecord, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Заголовки
	headers := []string{"event_number", "result", "oddF", "oddX", "oddL", "betF", "betX", "betL",
		"lossF", "lossX", "lossL", "total", "uf", "ux", "ul", "pattern"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Данные
	for _, record := range records {
		row := []string{
			strconv.Itoa(record.EventNumber),
			record.Result,
			strconv.FormatFloat(record.OddF, 'f', 2, 64),
			strconv.FormatFloat(record.OddX, 'f', 2, 64),
			strconv.FormatFloat(record.OddL, 'f', 2, 64),
			strconv.FormatFloat(record.BetF, 'f', 0, 64),
			strconv.FormatFloat(record.BetX, 'f', 0, 64),
			strconv.FormatFloat(record.BetL, 'f', 0, 64),
			strconv.FormatFloat(record.LossF, 'f', 0, 64),
			strconv.FormatFloat(record.LossX, 'f', 0, 64),
			strconv.FormatFloat(record.LossL, 'f', 0, 64),
			strconv.FormatFloat(record.Total, 'f', 0, 64),
			strconv.FormatFloat(record.UF, 'f', 0, 64),
			strconv.FormatFloat(record.UX, 'f', 0, 64),
			strconv.FormatFloat(record.UL, 'f', 0, 64),
			record.Pattern,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// CalculateStats вычисляет статистику
func CalculateStats(records []TrainerRecord, eventsFromOldest []string) Stats {
	stats := Stats{
		TotalRecords:     len(records),
		EventCounts:      make(map[string]int),
		EventPercentages: make(map[string]float64),
		MaxBets:          make(map[string]float64),
		MaxLosses:        make(map[string]float64),
		MaxStreaks:       make(map[string]int),
	}

	// Подсчет событий
	for _, event := range eventsFromOldest {
		stats.EventCounts[event]++
	}

	// Проценты
	total := float64(len(eventsFromOldest))
	for event, count := range stats.EventCounts {
		stats.EventPercentages[event] = float64(count) / total * 100
	}

	// Максимальные ставки и убытки
	stats.MaxBets["F"] = 0
	stats.MaxBets["X"] = 0
	stats.MaxBets["L"] = 0
	stats.MaxLosses["F"] = 0
	stats.MaxLosses["X"] = 0
	stats.MaxLosses["L"] = 0

	for _, record := range records {
		if record.BetF > stats.MaxBets["F"] {
			stats.MaxBets["F"] = record.BetF
		}
		if record.BetX > stats.MaxBets["X"] {
			stats.MaxBets["X"] = record.BetX
		}
		if record.BetL > stats.MaxBets["L"] {
			stats.MaxBets["L"] = record.BetL
		}
		if record.LossF > stats.MaxLosses["F"] {
			stats.MaxLosses["F"] = record.LossF
		}
		if record.LossX > stats.MaxLosses["X"] {
			stats.MaxLosses["X"] = record.LossX
		}
		if record.LossL > stats.MaxLosses["L"] {
			stats.MaxLosses["L"] = record.LossL
		}
	}

	// Максимальные серии
	currentStreaks := map[string]int{"F": 0, "X": 0, "L": 0}
	notFStreak := 0
	maxNotFStreak := 0
	lastEvent := ""

	for _, event := range eventsFromOldest {
		// Серии одинаковых событий
		if event == lastEvent {
			currentStreaks[event]++
		} else {
			if lastEvent != "" && currentStreaks[lastEvent] > stats.MaxStreaks[lastEvent] {
				stats.MaxStreaks[lastEvent] = currentStreaks[lastEvent]
			}
			currentStreaks = map[string]int{"F": 0, "X": 0, "L": 0}
			currentStreaks[event] = 1
			lastEvent = event
		}

		// Серия не-F
		if event != "F" {
			notFStreak++
			if notFStreak > maxNotFStreak {
				maxNotFStreak = notFStreak
			}
		} else {
			notFStreak = 0
		}
	}

	// Финальная проверка
	if lastEvent != "" && currentStreaks[lastEvent] > stats.MaxStreaks[lastEvent] {
		stats.MaxStreaks[lastEvent] = currentStreaks[lastEvent]
	}
	stats.MaxStreaks["notF"] = maxNotFStreak

	return stats
}

// PrintReport выводит отчет
func PrintReport(stats Stats, records []TrainerRecord) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    📊 ОТЧЕТ ТРЕНАЖЕРА")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n📊 РАСПРЕДЕЛЕНИЕ СОБЫТИЙ:\n")
	fmt.Printf("   F: %d (%.1f%%)\n", stats.EventCounts["F"], stats.EventPercentages["F"])
	fmt.Printf("   X: %d (%.1f%%)\n", stats.EventCounts["X"], stats.EventPercentages["X"])
	fmt.Printf("   L: %d (%.1f%%)\n", stats.EventCounts["L"], stats.EventPercentages["L"])

	fmt.Printf("\n💰 МАКСИМАЛЬНЫЕ СТАВКИ:\n")
	fmt.Printf("   F: %.0f\n", stats.MaxBets["F"])
	fmt.Printf("   X: %.0f\n", stats.MaxBets["X"])
	fmt.Printf("   L: %.0f\n", stats.MaxBets["L"])

	fmt.Printf("\n📉 МАКСИМАЛЬНЫЕ УБЫТКИ:\n")
	fmt.Printf("   F: %.0f\n", stats.MaxLosses["F"])
	fmt.Printf("   X: %.0f\n", stats.MaxLosses["X"])
	fmt.Printf("   L: %.0f\n", stats.MaxLosses["L"])

	fmt.Printf("\n🔄 МАКСИМАЛЬНЫЕ СЕРИИ:\n")
	fmt.Printf("   F: %d\n", stats.MaxStreaks["F"])
	fmt.Printf("   X: %d\n", stats.MaxStreaks["X"])
	fmt.Printf("   L: %d\n", stats.MaxStreaks["L"])
	fmt.Printf("   Не-F: %d\n", stats.MaxStreaks["notF"])

	fmt.Printf("   Всего записей: %d\n", stats.TotalRecords)
	if len(records) > 0 {
		fmt.Printf("   Итоговый результат: %.0f\n", records[0].Total)
	}
}

// GenerateRecordsWithOdds генерирует записи для событий с заданными коэффициентами
func GenerateRecordsWithOdds(eventsFromOldest []string, odds []struct{ OddF, OddX, OddL float64 }, flags Flags, strategy Strategy) []TrainerRecord {
	records := make([]TrainerRecord, len(eventsFromOldest))
	detector := NewPatternDetector()

	if flags.Debug {
		fmt.Printf("DEBUG: Starting GenerateRecordsWithOdds with %d events\n", len(eventsFromOldest))
		fmt.Printf("DEBUG: Strategy: %s\n", strategy.Name())
	}

	// Начальная запись (предыдущая для первого события)
	previous := TrainerRecord{
		Result: "N",
		Total:  0,
	}

	if flags.Debug {
		fmt.Printf("DEBUG: Initial previous record: Result=%s, Total=%.0f\n", previous.Result, previous.Total)
	}

	for i, event := range eventsFromOldest {
		var oddF, oddX, oddL float64
		if i < len(odds) {
			oddF = odds[i].OddF
			oddX = odds[i].OddX
			oddL = odds[i].OddL
			if flags.Debug {
				fmt.Printf("DEBUG: Event %d: Using provided odds - F=%.2f, X=%.2f, L=%.2f\n", i+1, oddF, oddX, oddL)
			}
		} else {
			oddF, oddX, oddL = generateOdds(flags)
			if flags.Debug {
				fmt.Printf("DEBUG: Event %d: Generated odds - F=%.2f, X=%.2f, L=%.2f\n", i+1, oddF, oddX, oddL)
			}
		}

		current := TrainerRecord{
			EventNumber: i + 1,
			Result:      event,
			OddF:        oddF,
			OddX:        oddX,
			OddL:        oddL,
		}

		if flags.Debug {
			fmt.Printf("DEBUG: Event %d: Before strategy calculation - Result=%s, Previous: UF=%.0f, UX=%.0f, UL=%.0f, LossF=%.0f, LossX=%.0f, LossL=%.0f, Total=%.0f\n",
				i+1, event, previous.UF, previous.UX, previous.UL, previous.LossF, previous.LossX, previous.LossL, previous.Total)
		}

		// Применяем стратегию
		strategy.Calculate(&current, &previous, flags)

		if flags.Debug {
			fmt.Printf("DEBUG: Event %d: After strategy calculation - BetF=%.0f, BetX=%.0f, BetL=%.0f, LossF=%.0f, LossX=%.0f, LossL=%.0f, Total=%.0f, UF=%.0f, UX=%.0f, UL=%.0f\n",
				i+1, current.BetF, current.BetX, current.BetL, current.LossF, current.LossX, current.LossL, current.Total, current.UF, current.UX, current.UL)
		}

		// Детектируем паттерны
		detectedPatterns := detector.AddEvent(event, i+1, current)
		if len(detectedPatterns) > 0 {
			current.Pattern = strings.Join(detectedPatterns, "_")
			if flags.Debug {
				fmt.Printf("DEBUG: Event %d: Pattern detected - %s\n", i+1, current.Pattern)
			}
		}

		records[i] = current
		previous = current

		if flags.Debug {
			fmt.Printf("DEBUG: Event %d: Complete record - %s,%.2f,%.2f,%.2f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f\n",
				i+1, current.Result, current.OddF, current.OddX, current.OddL,
				current.BetF, current.BetX, current.BetL, current.LossF, current.LossX, current.LossL,
				current.Total, current.UF, current.UX, current.UL)
		}
	}

	if flags.Debug {
		fmt.Printf("DEBUG: GenerateRecordsWithOdds completed, generated %d records\n", len(records))
	}

	return records
}
