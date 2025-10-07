# Руководство по созданию стратегий

Этот документ объясняет, как создавать новые стратегии для приложения тренажера.

## Архитектура стратегий

Все стратегии реализуют интерфейс `Strategy`:

```go
type Strategy interface {
    Name() string
    Description() string
    Calculate(current, previous *TrainerRecord, hockey bool)
}
```

### Методы интерфейса

- `Name() string` - Уникальное имя стратегии (используется в параметре `-strategy`)
- `Description() string` - Краткое описание стратегии
- `Calculate(current, previous *TrainerRecord, hockey bool)` - Основная логика расчета ставок

## Создание новой стратегии

### Шаг 1: Определите структуру стратегии

```go
type MyCustomStrategy struct{}
```

### Шаг 2: Реализуйте методы интерфейса

```go
func (s *MyCustomStrategy) Name() string {
    return "myCustom"
}

func (s *MyCustomStrategy) Description() string {
    return "Моя кастомная стратегия ставок"
}

func (s *MyCustomStrategy) Calculate(current, previous *TrainerRecord, hockey bool) {
    // Ваша логика расчета ставок здесь
    
    // Доступ к полям:
    // current.Result - результат текущего события (F, X, L)
    // current.OddF, current.OddX, current.OddL - коэффициенты
    // previous - предыдущая запись с накопленными значениями
    
    // Установите рассчитанные значения:
    // current.BetF, current.BetX, current.BetL - ставки
    // current.LossF, current.LossX, current.LossL - убытки
    // current.Total - общий итог
    // current.UF, current.UX, current.UL - серии без событий
}
```

### Шаг 3: Зарегистрируйте стратегию

Добавьте вашу стратегию в карту `strategies` в `main.go`:

```go
var strategies = map[string]Strategy{
    "xlWithSupport": &XLWithSupportStrategy{},
    "basic":         &BasicStrategy{},
    "myCustom":      &MyCustomStrategy{},  // Добавьте эту строку
}
```

## Пример: Простая стратегия Мартингейла

```go
type MartingaleStrategy struct{}

func (s *MartingaleStrategy) Name() string {
    return "martingale"
}

func (s *MartingaleStrategy) Description() string {
    return "Стратегия Мартингейла с удвоением ставки после проигрыша"
}

func (s *MartingaleStrategy) Calculate(current, previous *TrainerRecord, hockey bool) {
    baseAmount := config.DefaultBetF
    
    // Определяем, на что ставить (простая логика - ставим на F)
    targetBet := "F"
    
    // Рассчитываем ставку
    var betF, betX, betL float64
    
    if targetBet == "F" {
        // Удваиваем ставку после проигрыша
        if previous.Result != "F" {
            betF = calcBet(previous.LossF*2, current.OddF)
        } else {
            betF = calcBet(baseAmount, current.OddF)
        }
    }
    
    // Обновляем значения
    current.BetF = betF
    current.BetX = betX
    current.BetL = betL
    
    // Расчет убытков и итогов
    if current.Result == "F" {
        current.LossF = 0
        current.Total = previous.Total + baseAmount
    } else {
        current.LossF = previous.LossF + betF
        current.Total = previous.Total + baseAmount
    }
    
    // Обновляем серии
    if current.Result == "F" {
        current.UF = 0
    } else {
        current.UF = previous.UF + 1
    }
    
    current.UX = previous.UX + 1
    current.UL = previous.UL + 1
}
```

## Использование новой стратегии

После регистрации стратегии вы можете использовать ее через командную строку:

```bash
go run main.go test_runner.go -strategy myCustom -input "F/X/L/F/F/X/L"
```

## Тестирование новой стратегии

Создайте тестовые файлы для вашей стратегии:

1. Создайте `tests/myCustom_test.input` с входными данными
2. Запустите приложение для генерации ожидаемого вывода:
   ```bash
   go run main.go test_runner.go -strategy myCustom -input "ваши/события" -output temp.csv
   ```
3. Скопируйте содержимое `temp.csv` в `tests/myCustom_test.expected`
4. Запустите тесты:
   ```bash
   go run main.go test_runner.go -test -strategy myCustom
   ```

## Советы по созданию стратегий

1. **Используйте утилитарные функции**:
   - `calcBet(value, odd)` - расчет ставки
   - `roundUp(value)` - округление вверх

2. **Работайте с конфигурацией**:
   - `config.DefaultBetF` - базовая ставка
   - `config.RoundUp` - шаг округления

3. **Учитывайте хоккей**:
   - Проверяйте параметр `hockey` для специальной логики

4. **Обрабатывайте паттерны**:
   - Стратегии могут учитывать `previous.Pattern` для принятия решений

5. **Тестируйте thoroughly**:
   - Создавайте тесты для различных сценариев
   - Проверяйте крайние случаи (длинные серии, большие убытки)

## Доступные стратегии

На данный момент доступны следующие стратегии:

1. **xlWithSupport** - Стратегия "Ставка с поддержкой" с распределением убытков
2. **basic** - Базовая стратегия с фиксированными ставками

Вы можете использовать их как пример для создания собственных стратегий.