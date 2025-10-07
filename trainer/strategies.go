package trainer

import (
	"fmt"
	"strings"
)

// Strategy интерфейс для различных стратегий ставок
type Strategy interface {
	Name() string
	Description() string
	Calculate(current, previous *TrainerRecord, hockey bool)
}

// Регистр доступных стратегий
var strategies = map[string]Strategy{}

// RegisterStrategy регистрирует новую стратегию
func RegisterStrategy(strategy Strategy) {
	strategies[strategy.Name()] = strategy
}

// GetStrategy возвращает стратегию по имени
func GetStrategy(name string) (Strategy, error) {
	strategy, exists := strategies[name]
	if !exists {
		availableStrategies := make([]string, 0, len(strategies))
		for k := range strategies {
			availableStrategies = append(availableStrategies, k)
		}
		return nil, fmt.Errorf("стратегия '%s' не найдена. Доступные стратегии: %s",
			name, strings.Join(availableStrategies, ", "))
	}
	return strategy, nil
}

func init() {
	RegisterStrategy(&XLDropStrategy{})
	RegisterStrategy(&XLWithSupportStrategy{})
}
