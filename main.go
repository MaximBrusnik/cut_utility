package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FieldRange представляет диапазон полей
type FieldRange struct {
	Start int
	End   int
}

// Config содержит конфигурацию программы
type Config struct {
	Fields     string
	Delimiter  string
	Separated  bool
	FieldRanges []FieldRange
}

// parseFieldRanges парсит строку с номерами полей и диапазонами
func parseFieldRanges(fields string) ([]FieldRange, error) {
	if fields == "" {
		return nil, fmt.Errorf("пустая строка полей")
	}

	var ranges []FieldRange
	parts := strings.Split(fields, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Обработка диапазона (например, "3-5")
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("неверный формат диапазона: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("неверный номер начала диапазона: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("неверный номер конца диапазона: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("начало диапазона больше конца: %s", part)
			}

			ranges = append(ranges, FieldRange{Start: start, End: end})
		} else {
			// Обработка отдельного номера поля
			fieldNum, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("неверный номер поля: %s", part)
			}

			if fieldNum < 1 {
				return nil, fmt.Errorf("номер поля должен быть положительным: %d", fieldNum)
			}

			ranges = append(ranges, FieldRange{Start: fieldNum, End: fieldNum})
		}
	}

	return ranges, nil
}

// getFields возвращает номера полей для вывода
func getFields(ranges []FieldRange) []int {
	var fields []int
	seen := make(map[int]bool)

	for _, r := range ranges {
		for i := r.Start; i <= r.End; i++ {
			if !seen[i] {
				fields = append(fields, i)
				seen[i] = true
			}
		}
	}

	return fields
}

// processLine обрабатывает одну строку и возвращает результат
func processLine(line string, config Config) (string, bool) {
	// Если указан флаг -s и строка не содержит разделитель, пропускаем её
	if config.Separated && !strings.Contains(line, config.Delimiter) {
		return "", false
	}

	// Разбиваем строку по разделителю
	parts := strings.Split(line, config.Delimiter)
	
	// Если нет полей для вывода, возвращаем всю строку
	if len(config.FieldRanges) == 0 {
		return line, true
	}

	// Получаем номера полей для вывода
	fieldsToShow := getFields(config.FieldRanges)
	
	var result []string
	for _, fieldNum := range fieldsToShow {
		// Нумерация полей начинается с 1, но индексы с 0
		index := fieldNum - 1
		if index >= 0 && index < len(parts) {
			result = append(result, parts[index])
		}
	}

	return strings.Join(result, config.Delimiter), true
}

func main() {
	var config Config

	// Парсим флаги командной строки
	flag.StringVar(&config.Fields, "f", "", "номера полей для вывода (например: 1,3-5)")
	flag.StringVar(&config.Delimiter, "d", "\t", "разделитель полей")
	flag.BoolVar(&config.Separated, "s", false, "только строки, содержащие разделитель")
	flag.Parse()

	// Парсим номера полей
	if config.Fields != "" {
		ranges, err := parseFieldRanges(config.Fields)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка парсинга полей: %v\n", err)
			os.Exit(1)
		}
		config.FieldRanges = ranges
	}

	// Создаем сканер для чтения STDIN
	scanner := bufio.NewScanner(os.Stdin)

	// Обрабатываем каждую строку
	for scanner.Scan() {
		line := scanner.Text()
		result, shouldOutput := processLine(line, config)
		
		if shouldOutput {
			fmt.Println(result)
		}
	}

	// Проверяем ошибки сканера
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения: %v\n", err)
		os.Exit(1)
	}
}
