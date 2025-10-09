package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
		strategyName = flag.String("strategy", "xlDrop", "Имя стратегии для использования")
		realGames    = flag.Bool("real", false, "Обработка реальных игр из папки real-games")
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
		Real:     *realGames,
	}

	if flags.Report != "" {
		readCSVAndPrint(flags.Report)
		return
	}

	realGamesFlags := map[string]bool{
		"hockey": false,
	}
	if flags.Real {
		if flags.Hockey {
			realGamesFlags["hockey"] = true
		}
		processRealGames(flags, realGamesFlags)
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

// processRealGames обрабатывает реальные игры из папки real-games
func processRealGames(flags trainer.Flags, realGamesFlags map[string]bool) {
	realGamesDir := "real-games"

	// Проверяем существование папки
	if _, err := os.Stat(realGamesDir); os.IsNotExist(err) {
		fmt.Printf("Папка %s не существует. Создаем...\n", realGamesDir)
		if err := os.MkdirAll(realGamesDir, 0755); err != nil {
			log.Fatalf("Ошибка создания папки %s: %v", realGamesDir, err)
		}
		fmt.Printf("Папка %s создана. Добавьте .input файлы для обработки.\n", realGamesDir)
		return
	}

	// Читаем все файлы в папке
	files, err := ioutil.ReadDir(realGamesDir)
	if err != nil {
		log.Fatalf("Ошибка чтения папки %s: %v", realGamesDir, err)
	}

	// Фильтруем .input файлы
	var inputFiles []os.FileInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".input") {
			inputFiles = append(inputFiles, file)
		}
	}

	if len(inputFiles) == 0 {
		fmt.Printf("В папке %s не найдено .input файлов\n", realGamesDir)
		return
	}

	fmt.Printf("Найдено %d .input файлов в папке %s\n", len(inputFiles), realGamesDir)

	// Обрабатываем каждый файл
	for _, file := range inputFiles {
		processInputFile(filepath.Join(realGamesDir, file.Name()), flags, realGamesFlags)
	}
}

// processInputFile обрабатывает один input файл
func processInputFile(filePath string, flags trainer.Flags, realGamesFlags map[string]bool) {
	fileName := filepath.Base(filePath)

	// Проверяем наличие флага в имени файла
	baseName := strings.TrimSuffix(fileName, ".input")
	parts := strings.Split(baseName, "-")

	strategy, err := trainer.GetStrategy(parts[0])
	if err != nil {
		fmt.Printf("Ошибка получения стратегии: %v\n", err)
		return
	}

	var fileFlag string
	if len(parts) > 1 {
		fileFlag = parts[len(parts)-1]
	}

	if flags.Debug {
		fmt.Printf("DEBUG: Файл %s, флаг: %s\n", fileName, fileFlag)
	}

	// Проверяем, нужно ли обрабатывать этот файл
	if fileFlag != "" {
		// Если у файла есть флаг, проверяем, зарегистрирован ли он
		_, ok := realGamesFlags[fileFlag]
		if !ok {
			fmt.Printf("Ошибка: флаг '%s' в файле %s не зарегистрирован\n", fileFlag, fileName)
			fmt.Printf("Зарегистрированные флаги: %v\n", realGamesFlags)
			os.Exit(1)
		}

		// Если есть активные флаги, обрабатываем только файлы с этими флагами
		if len(realGamesFlags) > 0 {
			if !realGamesFlags[fileFlag] {
				if flags.Debug {
					fmt.Printf("DEBUG: Пропускаем файл %s (флаг %s не активен)\n", fileName, fileFlag)
				}
				return
			}
		} else {
			// Если активных флагов нет, пропускаем файлы с флагами
			if flags.Debug {
				fmt.Printf("DEBUG: Пропускаем файл %s (есть флаг %s, но активные флаги не указаны)\n", fileName, fileFlag)
			}
			return
		}
	} else {
		// Если у файла нет флага, но есть активные флаги, пропускаем его
		if hasTrue(realGamesFlags) {
			if flags.Debug {
				fmt.Printf("DEBUG: Пропускаем файл %s (нет флага, но есть активные флаги)\n", fileName)
			}
			return
		}
	}

	actualFilePath := strings.TrimSuffix(filePath, ".input") + ".actual"

	// Проверяем существование actual файла и сравниваем количество строк
	if _, err := os.Stat(actualFilePath); err == nil {
		inputLines, err1 := countLines(filePath)
		actualLines, err2 := countLines(actualFilePath)

		if err1 == nil && err2 == nil && inputLines == actualLines {
			fmt.Printf("Файл %s не требует обновления (количество строк совпадает)\n", fileName)
			return
		}
	}

	fmt.Printf("Обрабатываем файл: %s\n", fileName)

	// Читаем input файл
	events, err := trainer.ReadInputFile(filePath)
	if err != nil {
		fmt.Printf("Ошибка чтения файла %s: %v\n", fileName, err)
		return
	}

	// Извлекаем события и коэффициенты
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

	// Генерируем записи с использованием стратегии
	generatedRecords := trainer.GenerateRecordsWithOdds(eventStrings, odds, flags, strategy)

	// Сохраняем в actual файл
	if err := trainer.SaveToCSV(generatedRecords, actualFilePath); err != nil {
		fmt.Printf("Ошибка сохранения файла %s: %v\n", actualFilePath, err)
		return
	}

	fmt.Printf("Файл %s успешно обработан и сохранен как %s\n", fileName, filepath.Base(actualFilePath))
}

// countLines подсчитывает количество строк в файле
func countLines(filePath string) (int, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	// Убираем пустые строки в конце
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count, nil
}

func hasTrue(m map[string]bool) bool {
	for _, value := range m {
		if value {
			return true
		}
	}
	return false
}
