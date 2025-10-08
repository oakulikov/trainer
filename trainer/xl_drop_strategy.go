package trainer

import "fmt"

// XLDropStrategy реализует стратегию "Ставка с ограниченной поддержкой"
type XLDropStrategy struct{}

func (s *XLDropStrategy) Name() string {
	return "xlDrop"
}

func (s *XLDropStrategy) Description() string {
	return "Стратегия 'Ставка с ограниченной поддержкой' с пессимизацией страховки"
}

func (s *XLDropStrategy) Calculate(current, previous *TrainerRecord, flags Flags) {
	if flags.Debug {
		fmt.Printf("\n=== DEBUG: Calculate with strategy %s ===\n", s.Name())
	}

	eventNumber := current.EventNumber

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

		if flags.Debug {
			fmt.Printf("DEBUG: Event %d: lossF: %.0f, lossX: %.0f, lossL: %.0f\n", eventNumber, lossF, lossX, lossL)
			fmt.Printf("DEBUG: Event %d: realLoss BEFORE patterns %.0f\n", eventNumber, realLoss)
		}

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
			// total -= realLoss
			// realLoss = 0
			// partLoss := roundUp(realLoss / 2)
			// total -= partLoss
			// realLoss -= partLoss
		}

		if flags.Debug {
			fmt.Printf("DEBUG: Event %d: realLoss AFTER patterns %.0f\n", eventNumber, realLoss)
		}

		if realLoss > 0 {
			ratio := 0.3
			smallPart := roundUp(ratio * realLoss)
			bigPart := roundUp(realLoss - smallPart)

			if flags.Debug {
				fmt.Printf("DEBUG: Event %d: ratio: %.2f, smallPart: %.0f, bigPart: %.0f\n", eventNumber, ratio, smallPart, bigPart)
			}

			if ux < 5 {
				lossX += smallPart
			} else {
				deferLoss["X"] = smallPart
			}
			if ul < 6 {
				lossL += bigPart
			} else {
				deferLoss["L"] = bigPart
			}
		}
	}

	if flags.Debug {
		fmt.Printf("DEBUG: Event %d: coverageRatio_X: %.2f, deferLoss_X: %.0f\n", eventNumber, coverageRatio["X"], deferLoss["X"])
		fmt.Printf("DEBUG: Event %d: coverageRatio_L: %.2f, deferLoss_L: %.0f\n", eventNumber, coverageRatio["L"], deferLoss["L"])
	}

	betX := calcBet(lossX, current.OddX)
	betL := calcBet(lossL, current.OddL)

	// Корректировка lossF в зависимости от покрытия
	coverX := roundUp(betX * coverageRatio["X"])
	coverL := roundUp(betL * coverageRatio["L"])
	lossF += coverX
	lossF += coverL

	betF := calcBet(lossF, current.OddF)

	if flags.Debug {
		fmt.Printf("DEBUG: Event %d: lossF: %.0f, lossX: %.0f, lossL: %.0f\n", eventNumber, lossF, lossX, lossL)
		fmt.Printf("DEBUG: Event %d: betF: %.0f, betX: %.0f, betL: %.0f\n", eventNumber, betF, betX, betL)
		fmt.Printf("DEBUG: Event %d: coverX: %.0f, coverL: %.0f\n", eventNumber, coverX, coverL)
		fmt.Printf("DEBUG: Event %d: total BEFORE process %.0f\n", eventNumber, total)
	}

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
		if current.Result == "X" {
			lossX += deferLoss["X"]
		}
		if current.Result == "L" {
			lossL += deferLoss["L"]
		}
	}

	if flags.Debug {
		fmt.Printf("DEBUG: Event %d: lossX: %.0f, lossL: %.0f AFTER deferLoss\n", eventNumber, lossX, lossL)
	}

	total += baseAmount

	if flags.Debug {
		fmt.Printf("DEBUG: Event %d: total FINAL %.0f\n", eventNumber, total)
	}

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
