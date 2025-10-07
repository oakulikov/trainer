package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/holygun/go-trainer/trainer"
)

const EVENTS = "X/F/L/X/F/F/X/F/F/X/X/F/F/X/X/F/F/X/F/F/X/F/F/X/X/F/F/F/X/F/L/F/X/X/F/F/X/L/L/X/F/L/F/F/F/X/L/F/F/X/X/L/X/F/F/X/F/F/L/F/F/F/L/F/L/X/F/L/F/L/X/L/F/L/F/F/F/L/L/X/X/F/F/F/L/X/L/F/F/X/L/L/F/F/X/X/F/X/L/F/F/F/X/L/X/L/F/L/F/F/L/F/F/X/F/X/X/F/F/F/F/F/X/F/X/L/L/F/F/F/F/L/L/F/L/F/X/F/F/X/L/L/L/X/X/L/L/F/X/F/F/F/F/F/F/F/F/F/L/F/F/X/L/F/F/X/L/X/X/F/X/F/X/L/F/X/F/F/F/X/F/X/F/X/X/X/F/L/L/X/F/F/F/L/F/F/L/F/L/F/X/F/X/F/F/X/F/F/X/F/F/X/F/F/L/F/F/L/F/F/F/F/F/F/F/F/F/F/L/F/L/F/F/F/F/F/F/X/F/F/F/F/F/F/L/F/F/F/F/F/X/F/F/X/X/L/L/L/F/X/X/X/F/L/F/L/X/X/F/X/F/F/F/F/X/F/L/X/L/L/L/F/F/X/F/F/F/F/X/L/L/F/X/F/F/F/F/F/X/F/F/X/F/F/F/F/F/X/L/F/F/L/F/X/X/F/X/L/X/F/F/F/L/L/F/F/F/X/F/L/L/F/L/F/L/F/L"
const EVENTS_HOCKEY = "F/X/X/X/L/L/L/L/L/L/F/F/X/X/X/F/F/F/X/L/X/X/X/F/X/L/L/F/X/L/X/F/X/F/L/X/F/F/F/X/L/X/X/X/F/F/F/L/F/F/L/F/L/L/L/F/X/F/L/F/L/L/F/L/X/F/F/F/L/F/F/F/F/F/L/F/F/X/F/F/L/X/F/F/F/F/F/F/L/F/X/F/X/F/X/X/F/F/F/F/F/F/X/L"

func main() {
	// Парсинг аргументов командной строки
	var (
		inputString  = flag.String("input", "", "Строка событий F/X/L")
		outputFile   = flag.String("output", "trainer_output.csv", "Имя выходного CSV файла")
		verbose      = flag.Bool("verbose", false, "Подробный вывод")
		debug        = flag.Bool("debug", false, "Подробный вывод в тестах")
		printReport  = flag.String("report", "", "Имя входного CSV файла")
		hockey       = flag.Bool("hockey", false, "События хоккея")
		strategyName = flag.String("strategy", "xlWithSupport", "Имя стратегии для использования")
	)
	flag.Parse()

	// Создание структуры флагов
	flags := trainer.Flags{
		Input:    *inputString,
		Output:   *outputFile,
		Verbose:  *verbose,
		Debug:    *debug,
		Report:   *printReport,
		Hockey:   *hockey,
		Strategy: *strategyName,
	}

	if flags.Report != "" {
		readCSVAndPrint(flags.Report)
		return
	}

	if flags.Input == "" && flags.Hockey {
		flags.Input = EVENTS_HOCKEY
	} else if flags.Input == "" {
		flags.Input = EVENTS
	}

	// Парсинг событий
	events := trainer.ParseEvents(flags.Input)
	if len(events) == 0 {
		log.Fatal("Не найдено корректных событий F/X/L во входной строке")
	}

	fmt.Printf("📊 Обработка %d событий: %v\n", len(events), strings.Join(events, "/"))

	// Реверсируем для обработки от старых к новым
	eventsFromOldest := trainer.ReverseSlice(events)

	// Получение стратегии
	strategy, err := trainer.GetStrategy(flags.Strategy)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("📈 Используется стратегия: %s - %s\n", strategy.Name(), strategy.Description())

	// Генерация записей
	records := trainer.GenerateRecords(eventsFromOldest, flags, strategy)

	// Реверсируем обратно для отображения новых сверху
	records = trainer.ReverseRecords(records)

	// Сохранение в CSV
	if err := trainer.SaveToCSV(records, flags.Output); err != nil {
		log.Fatalf("Ошибка сохранения CSV: %v", err)
	}

	fmt.Printf("✅ Данные сохранены в %s\n", flags.Output)

	// Генерация и вывод статистики
	generateStatsAndPrint(records, eventsFromOldest)
}

func readCSVAndPrint(filename string) {
	records, err := trainer.ReadCSV(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	eventsFromNewest := []string{}
	for i := 0; i < len(records); i++ {
		eventsFromNewest = append(eventsFromNewest, records[i].Result)
	}
	// Генерация и вывод статистики
	generateStatsAndPrint(records, trainer.ReverseSlice(eventsFromNewest))
}

func generateStatsAndPrint(records []trainer.TrainerRecord, eventsFromOldest []string) {
	stats := trainer.CalculateStats(records, eventsFromOldest)
	trainer.PrintReport(stats, records)
}
