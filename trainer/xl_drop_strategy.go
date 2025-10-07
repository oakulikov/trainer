package trainer

// XLDropStrategy реализует стратегию "Ставка с ограниченной поддержкой"
type XLDropStrategy struct{}

func (s *XLDropStrategy) Name() string {
	return "xlDrop"
}

func (s *XLDropStrategy) Description() string {
	return "Стратегия 'Ставка с ограниченной поддержкой' с пессимизацией страховки"
}

func (s *XLDropStrategy) Calculate(current, previous *TrainerRecord, hockey bool) {
	lossF := previous.LossF
	lossX := previous.LossX
	lossL := previous.LossL
	total := previous.Total
	uf := previous.UF
	ux := previous.UX
	ul := previous.UL
	pattern := previous.Pattern

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

	deferLoss := map[string]float64{
		"F": 0,
		"X": 0,
		"L": 0,
	}
	coverageRatio := map[string]float64{
		"X": 0,
		"L": 0,
	}

	if uf > 0 || ux > 0 || ul > 0 {
		realLoss := lossF + lossX + lossL - baseAmount*3
		lossF = baseAmount
		lossX = baseAmount
		lossL = baseAmount

		// GREEN: метрика > 50,000
		// YELLOW: 2+ метрики > 50,000 ИЛИ любая > 100,000
		// RED: 3+ метрики > 100,000 (катастрофа)
		if pattern == "RED" {
			total -= realLoss
			realLoss = 0
		} else if pattern == "YELLOW" {
			total -= realLoss
			realLoss = 0
		} else if pattern == "GREEN" {
			total -= realLoss
			realLoss = 0
		}
		if realLoss > 0 {
			ratio := 0.3
			smallPart := roundUp(ratio * realLoss)
			if ux < 5 {
				lossX += smallPart
				if ux > 0 {
					coverageRatio["X"] = 1 / ux
				}
			} else {
				deferLoss["X"] = smallPart
			}
			if ul < 6 {
				lossL += roundUp(realLoss - smallPart)
				if ul > 0 {
					coverageRatio["L"] = 1 / ul
				}
			} else {
				deferLoss["L"] = roundUp(realLoss - smallPart)
			}
		}
	}

	betX := calcBet(lossX, current.OddX)
	betL := calcBet(lossL, current.OddL)

	// Корректировка lossF в зависимости от покрытия
	coverX := roundUp(betX * coverageRatio["X"])
	coverL := roundUp(betL * coverageRatio["L"])
	lossF += coverX
	lossF += coverL

	betF := calcBet(lossF, current.OddF)

	// Обработка результата
	if current.Result == "F" {
		// Серии
		uf = 0
		ux++
		ul++
		// Потери
		lossF = 0
		lossX += betX - coverX
		lossL += betL - coverL
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
	if deferLoss[current.Result] > 0 {
		total -= deferLoss[current.Result]
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
