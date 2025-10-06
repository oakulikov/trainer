#!/bin/bash

# Создаем папку для результатов
mkdir -p results_ratio

echo "ТЕСТИРОВАНИЕ С КРИТЕРИЕМ: total/expected_total"
echo "================================================"
echo "Valid: ratio >= 0.5"
echo "Invalid: ratio < 0.5"
echo ""

# Ожидаемый total = количество событий * базовая прибыль (5000)
# У нас 337 событий
EXPECTED_TOTAL=$((337 * 5000))
echo "Ожидаемый total: $EXPECTED_TOTAL"
echo ""

# Запускаем 20 раз
valid_count=0
invalid_count=0
total_sum=0

for i in {1..20}; do
    echo "Запуск $i..."
    
    # Запускаем программу и сохраняем вывод
    output=$(go run main.go -output "results_ratio/run${i}.csv" 2>&1)
    
    # Извлекаем итоговый результат
    total=$(echo "$output" | grep "Итоговый результат:" | grep -oE '[-]?[0-9]+')
    
    # Вычисляем соотношение
    ratio=$(echo "scale=3; $total / $EXPECTED_TOTAL" | bc)
    
    # Извлекаем максимальные убытки для информации
    max_losses=$(echo "$output" | grep "МАКСИМАЛЬНЫЕ УБЫТКИ" -A 3 | tail -3)
    loss_f=$(echo "$max_losses" | grep "F:" | grep -oE '[0-9]+')
    loss_x=$(echo "$max_losses" | grep "X:" | grep -oE '[0-9]+')
    loss_l=$(echo "$max_losses" | grep "L:" | grep -oE '[0-9]+')
    
    # Определяем valid или invalid по соотношению
    if (( $(echo "$ratio >= 0.5" | bc -l) )); then
        mv "results_ratio/run${i}.csv" "results_ratio/valid-run${i}.csv"
        echo "  ✅ VALID - Ratio: $ratio | Total: $total | Убытки: F=$loss_f, X=$loss_x, L=$loss_l"
        ((valid_count++))
    else
        mv "results_ratio/run${i}.csv" "results_ratio/invalid-run${i}.csv"
        echo "  ❌ INVALID - Ratio: $ratio | Total: $total | Убытки: F=$loss_f, X=$loss_x, L=$loss_l"
        ((invalid_count++))
    fi
    
    total_sum=$((total_sum + total))
done

echo ""
echo "========================================="
echo "ИТОГОВАЯ СТАТИСТИКА"
echo "========================================="
echo "Valid запусков: $valid_count ($(echo "scale=1; $valid_count*100/20" | bc)%)"
echo "Invalid запусков: $invalid_count ($(echo "scale=1; $invalid_count*100/20" | bc)%)"
echo ""
echo "Средний total: $((total_sum / 20))"
echo "Ожидаемый total: $EXPECTED_TOTAL"
echo "Среднее соотношение: $(echo "scale=3; ($total_sum / 20) / $EXPECTED_TOTAL" | bc)"
echo ""

# Анализ по категориям
echo "ДЕТАЛЬНЫЙ АНАЛИЗ:"
echo "-----------------"

# Подсчет по диапазонам ratio
excellent=0  # >= 0.75
good=0       # 0.5 - 0.75
poor=0       # 0.25 - 0.5
bad=0        # < 0.25

# Используем сохраненные результаты из первого цикла
# Запускаем еще раз те же 20 запусков для подсчета категорий
for i in {1..20}; do
    # Запускаем программу для получения total
    output=$(go run main.go -output "/dev/null" 2>&1)
    total=$(echo "$output" | grep "Итоговый результат:" | grep -oE '[-]?[0-9]+')
    ratio=$(echo "scale=3; $total / $EXPECTED_TOTAL" | bc)
    
    if (( $(echo "$ratio >= 0.75" | bc -l) )); then
        ((excellent++))
    elif (( $(echo "$ratio >= 0.5" | bc -l) )); then
        ((good++))
    elif (( $(echo "$ratio >= 0.25" | bc -l) )); then
        ((poor++))
    else
        ((bad++))
    fi
done

echo "Отличные (ratio >= 0.75): $excellent"
echo "Хорошие (0.5 <= ratio < 0.75): $good"
echo "Плохие (0.25 <= ratio < 0.5): $poor"
echo "Критические (ratio < 0.25): $bad"