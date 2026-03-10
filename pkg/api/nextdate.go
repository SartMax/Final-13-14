package api

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HandleNextDate обрабатывает запросы к /api/nextdate
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	now := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	var nowTime time.Time
	var err error
	if now == "" {
		nowTime = time.Now()
	} else {
		nowTime, err = time.Parse(DateFormat, now) // Используем DateFormat из api.go
		if err != nil {
			http.Error(w, "Invalid now parameter", http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(next))
}

// nextDate (внутренняя функция)
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("empty repeat rule")
	}

	start, err := time.Parse(DateFormat, date) // Используем DateFormat
	if err != nil {
		return "", errors.New("invalid date format")
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("invalid repeat rule")
	}

	switch parts[0] {
	case "d":
		return handleDaily(now, start, parts)
	case "y":
		return handleYearly(now, start)
	case "w":
		return handleWeekly(now, start, parts)
	case "m":
		return handleMonthly(now, start, parts)
	default:
		return "", errors.New("unsupported repeat rule")
	}
}

func handleDaily(now time.Time, start time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", errors.New("missing days count")
	}

	days, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", errors.New("invalid days count")
	}

	if days < 1 || days > 400 {
		return "", errors.New("days count must be 1-400")
	}

	next := start

	if next.After(now) {
		next = next.AddDate(0, 0, days)
		return next.Format(DateFormat), nil
	}

	for !next.After(now) {
		next = next.AddDate(0, 0, days)
	}

	return next.Format(DateFormat), nil
}

func handleYearly(now time.Time, start time.Time) (string, error) {
	next := start

	if next.After(now) {
		next = next.AddDate(1, 0, 0)
		return next.Format(DateFormat), nil
	}

	for !next.After(now) {
		next = next.AddDate(1, 0, 0)
	}

	return next.Format(DateFormat), nil
}

func handleWeekly(now time.Time, start time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", errors.New("missing weekdays")
	}

	weekdays := strings.Split(parts[1], ",")
	days := make(map[time.Weekday]bool)

	for _, d := range weekdays {
		day, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil {
			return "", errors.New("invalid weekday")
		}
		if day < 1 || day > 7 {
			return "", errors.New("weekday must be 1-7")
		}
		var wd time.Weekday
		if day == 7 {
			wd = time.Sunday
		} else {
			wd = time.Weekday(day)
		}
		days[wd] = true
	}

	current := start

	for {
		if days[current.Weekday()] {
			if current.After(now) {
				return current.Format(DateFormat), nil
			}
		}
		current = current.AddDate(0, 0, 1)
	}
}

func handleMonthly(now time.Time, start time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", errors.New("missing days")
	}

	// Парсим дни месяца
	dayParts := strings.Split(parts[1], ",")
	days := make([]int, 0)
	for _, d := range dayParts {
		day, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil {
			return "", errors.New("invalid day")
		}
		if day < -2 || day == 0 || day > 31 {
			return "", errors.New("day must be -2,-1 or 1-31")
		}
		days = append(days, day)
	}

	// Сортируем дни: сначала положительные, потом отрицательные (-2, -1)
	sort.Slice(days, func(i, j int) bool {
		// Оба положительные
		if days[i] > 0 && days[j] > 0 {
			return days[i] < days[j]
		}
		// Оба отрицательные
		if days[i] < 0 && days[j] < 0 {
			return days[i] < days[j]
		}
		// Один положительный, один отрицательный - положительный первый
		return days[i] > 0
	})

	// Парсим месяцы
	months := make([]int, 0)
	if len(parts) >= 3 {
		monthParts := strings.Split(parts[2], ",")
		for _, m := range monthParts {
			month, err := strconv.Atoi(strings.TrimSpace(m))
			if err != nil {
				return "", errors.New("invalid month")
			}
			if month < 1 || month > 12 {
				return "", errors.New("month must be 1-12")
			}
			months = append(months, month)
		}
		sort.Ints(months)
	}

	// Начинаем с текущего месяца
	currentYear := start.Year()
	currentMonth := start.Month()

	for {
		// Проверяем, подходит ли месяц
		if len(months) > 0 {
			monthOk := false
			for _, m := range months {
				if int(currentMonth) == m {
					monthOk = true
					break
				}
			}
			if !monthOk {
				// Переходим к следующему месяцу
				if currentMonth == 12 {
					currentMonth = 1
					currentYear++
				} else {
					currentMonth++
				}
				continue
			}
		}

		// Для каждого дня в текущем месяце
		for _, day := range days {
			targetDay := day

			// Обработка отрицательных дней
			if day < 0 {
				lastDay := getLastDayOfMonth(currentYear, int(currentMonth))
				if day == -1 {
					targetDay = lastDay
				} else if day == -2 {
					targetDay = lastDay - 1
				}
			}

			// Проверяем, что день существует в этом месяце
			if targetDay < 1 || targetDay > getLastDayOfMonth(currentYear, int(currentMonth)) {
				continue
			}

			// Создаем дату
			targetDate := time.Date(currentYear, currentMonth, targetDay, 0, 0, 0, 0, time.UTC)

			// Если это первый месяц, проверяем что дата > start
			if currentYear == start.Year() && currentMonth == start.Month() {
				if targetDate.After(start) && targetDate.After(now) {
					return targetDate.Format(DateFormat), nil
				}
				continue
			}

			// Для следующих месяцев просто проверяем, что дата > now
			if targetDate.After(now) {
				return targetDate.Format(DateFormat), nil
			}
		}

		// Переходим к следующему месяцу
		if currentMonth == 12 {
			currentMonth = 1
			currentYear++
		} else {
			currentMonth++
		}
	}
}

// getLastDayOfMonth возвращает последний день месяца
func getLastDayOfMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}
