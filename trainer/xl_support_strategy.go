package trainer

// XLWithSupportStrategy реализует стратегию "Ставка с поддержкой"
type XLWithSupportStrategy struct{}

func (s *XLWithSupportStrategy) Name() string {
	return "xlWithSupport"
}

func (s *XLWithSupportStrategy) Description() string {
	return "Стратегия 'Ставка с поддержкой' с распределением убытков"
}

func (s *XLWithSupportStrategy) Calculate(current, previous *TrainerRecord, hockey bool) {
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

	fullCoverage := ""
	partialCoverage := ""

	if uf > 0 || ux > 0 || ul > 0 {
		realLoss := lossF + lossX + lossL - baseAmount*3
		lossF = baseAmount
		lossX = baseAmount
		lossL = baseAmount

		// 🔸 Желтый: метрика > 50,000
		// 🔴 Красный: 2+ метрики > 50,000 ИЛИ любая > 100,000
		// ⚫ Черный: 3+ метрики > 100,000 (катастрофа)
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
			lossX += smallPart
			lossL += roundUp(realLoss - smallPart)
			fullCoverage = "X"
			if lossL > baseAmount*PARTIAL_COVERAGE_MULT {
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
			lossF += betL - baseAmount*PARTIAL_COVERAGE_MULT
		}
	} else if fullCoverage == "L" {
		lossF += betL
		if partialCoverage == "X" {
			lossF += betX - baseAmount*PARTIAL_COVERAGE_MULT
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
			// X L были покрыты полностью, убытки не растут
		} else if fullCoverage == "X" {
			// X был покрыт полностью, убытки не растут
			// lossX остается прежним
			if partialCoverage == "L" {
				lossL += baseAmount * PARTIAL_COVERAGE_MULT
			} else {
				lossL += betL
			}
		} else if fullCoverage == "L" {
			// L был покрыт полностью, убытки не растут
			// lossL остается прежним
			if partialCoverage == "X" {
				lossX += baseAmount * PARTIAL_COVERAGE_MULT
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
