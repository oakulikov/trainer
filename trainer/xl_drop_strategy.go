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

    deferLoss := map[string]bool{
        "F": false,
        "X": false,
        "L": false,
    }

    if flags.Debug {
        fmt.Printf("DEBUG: INITIALIZE: Event %d: lossF: %.0f, lossX: %.0f, lossL: %.0f\n", eventNumber, lossF, lossX, lossL)
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
            halfPart := roundUp(realLoss / 2)
            total -= halfPart
            realLoss = halfPart
        } else if pattern == "YELLOW" {
            // if flags.Hockey {
            //  halfPart := roundUp(realLoss / 2)
            //  total -= halfPart
            //  realLoss = halfPart
            // }
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

            if flags.Hockey {
                lossX += bigPart
                lossL += smallPart
            } else {
                lossX += smallPart
                lossL += bigPart
            }
        }
    }

    betF := calcBet(lossF, current.OddF)

    betX := calcBet(lossX, current.OddX)
    if ux >= 5 {
        betX = calcBet(baseAmount, current.OddX)
        deferLoss["X"] = true
    }

    betL := calcBet(lossL, current.OddL)
    if ul >= 6 {
        betL = calcBet(baseAmount, current.OddL)
        deferLoss["L"] = true
    }

    if flags.Debug {
        fmt.Printf("DEBUG: Event %d: deferLoss_X: %v, deferLoss_L: %v\n", eventNumber, deferLoss["X"], deferLoss["L"])
        fmt.Printf("DEBUG: Event %d: lossF: %.0f, lossX: %.0f, lossL: %.0f\n", eventNumber, lossF, lossX, lossL)
        fmt.Printf("DEBUG: Event %d: betF: %.0f, betX: %.0f, betL: %.0f\n", eventNumber, betF, betX, betL)
        fmt.Printf("DEBUG: Event %d: total BEFORE process %.0f\n", eventNumber, total)
        fmt.Printf("DEBUG: Event %d: RESULT %s\n", eventNumber, current.Result)
    }

    // Обработка результата
    if current.Result == "F" {
        // Серии
        uf = 0
        ux++
        ul++
        // Потери
        lossF = 0
        lossX += betX
        lossL += betL
        total += baseAmount
    }
    if current.Result == "X" {
        // Серии
        uf++
        ux = 0
        ul++
        // Потери
        lossF += betF
        lossL += betL
        if deferLoss["X"] {
            lossX -= baseAmount
        } else {
            lossX = 0
        }
        total += baseAmount
    }
    if current.Result == "L" {
        // Серии
        uf++
        ux++
        ul = 0
        // Потери
        lossF += betF
        lossX += betX
        if deferLoss["L"] {
            lossL -= baseAmount
        } else {
            lossL = 0
        }
        total += baseAmount
    }
    if current.Result == "N" && flags.Real {
        // Серии
        uf = -1
        ux = -1
        ul = -1
        // Потери
        lossF = -1
        lossX = -1
        lossL = -1
    } else if current.Result == "N" && !flags.Testing {
        panic("current.Result N w/o flags.Real")
    }

    if flags.Debug {
        fmt.Printf("DEBUG: Event %d: total FINAL %.0f\n", eventNumber, total)
        fmt.Printf("DEBUG: Event %d: END GAME: lossF: %.0f, lossX: %.0f, lossL: %.0f\n", eventNumber, lossF, lossX, lossL)
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
