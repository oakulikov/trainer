package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const Events = "X/F/L/X/F/F/X/F/F/X/X/F/F/X/X/F/F/X/F/F/X/F/F/X/X/F/F/F/X/F/L/F/X/X/F/F/X/L/L/X/F/L/F/F/F/X/L/F/F/X/X/L/X/F/F/X/F/F/L/F/F/F/L/F/L/X/F/L/F/L/X/L/F/L/F/F/F/L/L/X/X/F/F/F/L/X/L/F/F/X/L/L/F/F/X/X/F/X/L/F/F/F/X/L/X/L/F/L/F/F/L/F/F/X/F/X/X/F/F/F/F/F/X/F/X/L/L/F/F/F/F/L/L/F/L/F/X/F/F/X/L/L/L/X/X/L/L/F/X/F/F/F/F/F/F/F/F/F/L/F/F/X/L/F/F/X/L/X/X/F/X/F/X/L/F/X/F/F/F/X/F/X/F/X/X/X/F/L/L/X/F/F/F/L/F/F/L/F/L/F/X/F/X/F/F/X/F/F/X/F/F/X/F/F/L/F/F/L/F/F/F/F/F/F/F/F/F/F/L/F/L/F/F/F/F/F/F/X/F/F/F/F/F/F/L/F/F/F/F/F/X/F/F/X/X/L/L/L/F/X/X/X/F/L/F/L/X/X/F/X/F/F/F/F/X/F/L/X/L/L/L/F/F/X/F/F/F/F/X/L/L/F/X/F/F/F/F/F/X/F/F/X/F/F/F/F/F/X/L/F/F/L/F/X/X/F/X/L/X/F/F/F/L/L/F/F/F/X/F/L/L/F/L/F/L/F/L"

// Pattern определяет структуру паттерна
type Pattern struct {
	ID          string
	Description string
	Sequence    string
}

// Зарегистрированные паттерны
var patterns = []Pattern{}

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
func (pd *PatternDetector) AddEvent(event string, eventNumber int, ux, ul float64) []string {
	pd.recentEvents = append(pd.recentEvents, event)
	if len(pd.recentEvents) > pd.windowSize {
		pd.recentEvents = pd.recentEvents[1:]
	}

	detectedPatterns := []string{}

	// Проверяем паттерны
	for _, pattern := range patterns {
		if pd.checkPattern(pattern, ux, ul) {
			detectedPatterns = append(detectedPatterns, pattern.ID)
			fmt.Printf("⚠️ Событие номер %d: обнаружен паттерн %s - %s\n", eventNumber, pattern.ID, pattern.Description)
		}
	}

	return detectedPatterns
}

// checkPattern проверяет конкретный паттерн
func (pd *PatternDetector) checkPattern(pattern Pattern, ux, ul float64) bool {
	switch pattern.Sequence {
	case "CRITICAL_ACCUMULATION":
		return ux > 3 && ul > 3
	case "NON_X_5":
		return pd.countConsecutive("X") >= 5
	case "NON_L_5":
		return pd.countConsecutive("L") >= 5
	case "F_SERIES_5":
		return pd.countConsecutiveF() >= 5
	}
	return false
}

// countConsecutive считает последовательные события без указанного типа
func (pd *PatternDetector) countConsecutive(eventType string) int {
	count := 0
	for i := len(pd.recentEvents) - 1; i >= 0; i-- {
		if pd.recentEvents[i] != eventType {
			count++
		} else {
			break
		}
	}
	return count
}

// countConsecutiveF считает последовательные события F
func (pd *PatternDetector) countConsecutiveF() int {
	count := 0
	for i := len(pd.recentEvents) - 1; i >= 0; i-- {
		if pd.recentEvents[i] == "F" {
			count++
		} else {
			break
		}
	}
	return count
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
	DefaultBetF: 5000,
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

func main() {
	// Парсинг аргументов командной строки
	var (
		inputString = flag.String("input", "", "Строка событий F/X/L")
		outputFile  = flag.String("output", "trainer_output.csv", "Имя выходного CSV файла")
		verbose     = flag.Bool("verbose", false, "Подробный вывод")
		debug       = flag.Bool("debug", false, "Режим отладки с детальным логированием")
		printReport = flag.String("report", "", "Имя входного CSV файла")
		// lossFXL     = flag.String("loss_fxl", "", "Убытки F,X,L для анализа (например: 5000,10000,20000)")
		// rateFXL     = flag.String("rate_fxl", "", "Коэффициенты F,X,L для анализа (например: 2,3.5,4)")
	)
	flag.Parse()

	if *printReport != "" {
		readCSVAndPrint(*printReport)
		return
	}

	if *inputString == "" {
		*inputString = Events
	}

	// Парсинг событий
	events := parseEvents(*inputString)
	if len(events) == 0 {
		log.Fatal("Не найдено корректных событий F/X/L во входной строке")
	}

	fmt.Printf("📊 Обработка %d событий: %v\n", len(events), strings.Join(events, "/"))

	// Реверсируем для обработки от старых к новым
	eventsFromOldest := reverseSlice(events)

	// Генерация записей
	records := generateRecords(eventsFromOldest, *verbose, *debug)

	// Реверсируем обратно для отображения новых сверху
	records = reverseRecords(records)

	// Сохранение в CSV
	if err := saveToCSV(records, *outputFile); err != nil {
		log.Fatalf("Ошибка сохранения CSV: %v", err)
	}

	fmt.Printf("✅ Данные сохранены в %s\n", *outputFile)

	// Генерация и вывод статистики
	generateStatsAndPrint(records, eventsFromOldest)
}

func readCSVAndPrint(filename string) {
	records, err := readCSV(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	eventsFromNewest := []string{}
	for i := 0; i < len(records); i++ {
		eventsFromNewest = append(eventsFromNewest, records[i].Result)
	}
	// Генерация и вывод статистики
	generateStatsAndPrint(records, reverseSlice(eventsFromNewest))
}

func generateStatsAndPrint(records []TrainerRecord, eventsFromOldest []string) {
	stats := calculateStats(records, eventsFromOldest)
	printReport(stats, records)
}

// parseEvents парсит строку событий F/X/L
func parseEvents(input string) []string {
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
func reverseSlice(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// reverseRecords реверсирует слайс записей
func reverseRecords(records []TrainerRecord) []TrainerRecord {
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
func generateOdds(verbose bool, debug bool) (float64, float64, float64) {
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

			// DEBUG: Логируем опасные комбинации коэффициентов
			if debug && (oddF < 1.85 || oddX < 3.4 || oddL < 4.2) {
				fmt.Printf("DEBUG ODDS: Опасные коэффициенты - F=%.2f, X=%.2f, L=%.2f, margin=%.3f\n",
					oddF, oddX, oddL, margin)
			}

			return oddF, oddX, oddL
		}
	}

	if verbose {
		fmt.Printf("DEBUG: [generateOdds] fallback to default odds")
	}

	// Значения по умолчанию
	return 2, 3.5, 4
}

// xlWithSupport реализует стратегию "Ставка с поддержкой"
func xlWithSupport(current *TrainerRecord, previous *TrainerRecord) {
	lossF := previous.LossF
	lossX := previous.LossX
	lossL := previous.LossL
	total := previous.Total
	uf := previous.UF
	ux := previous.UX
	ul := previous.UL

	baseAmount := config.DefaultBetF

	// Инициализация потерь
	if uf == 0 {
		lossF = baseAmount
	}
	if ux == 0 {
		lossX = baseAmount
	}
	if ul == 0 {
		lossL = baseAmount
	}

	fullCoverage := ""
	partialCoverage := ""

	if uf > 0 || ux > 0 || ul > 0 {
		ratio := 0.3
		realLoss := lossF + lossX + lossL - baseAmount*3
		lossF = baseAmount
		lossX = baseAmount
		lossL = baseAmount

		if ux > 3 && ul > 3 {
			fullCoverage = "XL"
		} else {
			smallPart := roundUp(ratio * realLoss)
			lossX += smallPart
			lossL += roundUp(realLoss - smallPart)
			fullCoverage = "X"
			if lossL > baseAmount*2 {
				partialCoverage = "L"
			}
		}
	}

	betX := calcBet(lossX, current.OddX)
	betL := calcBet(lossL, current.OddL)

	// Корректировка lossF в зависимости от покрытия
	if fullCoverage == "XL" {
		lossF += betX + betL
	} else if fullCoverage == "X" {
		lossF += betX
		if partialCoverage == "L" {
			lossF += betL - baseAmount*2
		}
	} else if fullCoverage == "L" {
		lossF += betL
		if partialCoverage == "X" {
			lossF += betX - baseAmount*2
		}
	}

	betF := calcBet(lossF, current.OddF)

	// Обработка результата
	if current.Result == "F" {
		// Серии
		uf = 0
		ux++
		ul++
		// Потери
		lossF = 0
		if fullCoverage == "XL" {
			// При полном покрытии XL убытки не растут (остаются как есть)
			// lossX остается прежним
			// lossL остается прежним
		} else if fullCoverage == "X" {
			// X был покрыт полностью, убытки не растут
			// lossX остается прежним
			if partialCoverage == "L" {
				lossL += betL - baseAmount*2
			} else {
				lossL += betL
			}
		} else if fullCoverage == "L" {
			// L был покрыт полностью, убытки не растут
			// lossL остается прежним
			if partialCoverage == "X" {
				lossX += betX - baseAmount*2
			} else {
				lossX += betX
			}
		} else {
			lossX += betX
			lossL += betL
		}
	} else if current.Result == "X" {
		// Серии
		uf++
		ux = 0
		ul++
		// Потери
		lossF += betF
		lossX = 0
		lossL += betL
	} else if current.Result == "L" {
		// Серии
		uf++
		ux++
		ul = 0
		// Потери
		lossF += betF
		lossX += betX
		lossL = 0
	}
	total += baseAmount

	// Обновляем текущую запись
	current.BetF = betF
	current.BetX = betX
	current.BetL = betL
	current.LossF = lossF
	current.LossX = lossX
	current.LossL = lossL
	current.Total = total
	current.UF = uf
	current.UX = ux
	current.UL = ul
}

// generateRecords генерирует записи для событий
func generateRecords(eventsFromOldest []string, verbose bool, debug bool) []TrainerRecord {
	records := make([]TrainerRecord, len(eventsFromOldest))
	detector := NewPatternDetector()

	// Начальная запись (предыдущая для первого события)
	previous := TrainerRecord{
		Result: "N",
		Total:  0,
	}

	for i, event := range eventsFromOldest {
		oddF, oddX, oddL := generateOdds(verbose, debug)

		current := TrainerRecord{
			EventNumber: i + 1,
			Result:      event,
			OddF:        oddF,
			OddX:        oddX,
			OddL:        oddL,
		}

		// Применяем стратегию
		xlWithSupport(&current, &previous)

		// Детектируем паттерны
		detectedPatterns := detector.AddEvent(event, i+1, previous.UX, previous.UL)
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

	return records
}

func readCSV(filename string) ([]TrainerRecord, error) {
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

// saveToCSV сохраняет записи в CSV файл
func saveToCSV(records []TrainerRecord, filename string) error {
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

// calculateStats вычисляет статистику
func calculateStats(records []TrainerRecord, eventsFromOldest []string) Stats {
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

// printReport выводит отчет
func printReport(stats Stats, records []TrainerRecord) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    📊 ОТЧЕТ ТРЕНАЖЕРА")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n📈 ОБЩАЯ СТАТИСТИКА:\n")
	fmt.Printf("   Всего записей: %d\n", stats.TotalRecords)

	if len(records) > 0 {
		fmt.Printf("   Итоговый результат: %.0f\n", records[0].Total)
	}

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

	fmt.Println("\n" + strings.Repeat("=", 60))
}
