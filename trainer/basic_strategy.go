package trainer

// BasicStrategy простая стратегия без поддержки
type BasicStrategy struct{}

func (s *BasicStrategy) Name() string {
	return "basic"
}

func (s *BasicStrategy) Description() string {
	return "Базовая стратегия с фиксированными ставками"
}

func (s *BasicStrategy) Calculate(current, previous *TrainerRecord, hockey bool) {
	baseAmount := config.DefaultBetF

	// Инициализация потерь
	lossF := baseAmount
	lossX := baseAmount
	lossL := baseAmount
	total := previous.Total
	uf := previous.UF
	ux := previous.UX
	ul := previous.UL

	// Простые фиксированные ставки
	betF := calcBet(lossF, current.OddF)
	betX := calcBet(lossX, current.OddX)
	betL := calcBet(lossL, current.OddL)

	// Обработка результата
	if current.Result == "F" {
		uf = 0
		ux++
		ul++
		lossF = 0
		lossX += betX
		lossL += betL
	} else if current.Result == "X" {
		uf++
		ux = 0
		ul++
		lossF += betF
		lossX = 0
		lossL += betL
	} else if current.Result == "L" {
		uf++
		ux++
		ul = 0
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
