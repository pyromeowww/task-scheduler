package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)


const DateFormat = "20060102"

// NextDate вычисляет следующую дату выполнения задачи по правилу повторения.
// now — текущая дата;
// dstart — дата создания или последнего выполнения задачи;
// repeat — правило повторения.
// Возвращает следующую дату в формате DateFormat и ошибку.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// Без правила повтора посчитать следующую дату невозможно.
	if repeat == "" {
		return "", errors.New("repeat rule cannot be empty")
	}

	// Преобразуем строку с начальной датой в объект time.Time.
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid dstart format: %w", err)
	}

	// Первая буква правила определяет тип повтора:
	// d — каждый день, w — по дням недели, m — по дням месяца, y — каждый год.
	repeatParts := strings.Split(repeat, " ")
	switch repeatParts[0] {
	case "d":
		return nextDay(date, now, repeatParts)
	case "w":
		return nextWeek(date, now, repeatParts)
	case "m":
		return nextMonthDays(date, now, repeatParts)
	case "y":
		return nextYear(date, now, repeatParts)
	default:
		return "", errors.New("unknown format")
	}
}

// NextDateHandler обрабатывает GET-запрос.
// Принимает query-параметры
func NextDateHandler(res http.ResponseWriter, req *http.Request) {
	// Разрешён только метод GET.
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Если параметр not не передан, берём текущее время.
	var now time.Time
	nowStr := req.URL.Query().Get("now")
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(res, "invalid now date", http.StatusBadRequest)
			return
		}
	}

	// Читаем начальную дату и правило повторения из параметров запроса.
	dateStr := req.URL.Query().Get("date")
	repeatStr := req.URL.Query().Get("repeat")

	// Вычисляем следующую дату и возвращаем её либо сообщение об ошибке.
	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(nextDate))
}

// AfterNow сравнивает даты по календарному дню (без времени).
// Возвращает true, если date идёт позже, чем now.
func AfterNow(date time.Time, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	return time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC).After(time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC))
}

// nextDay рассчитывает следующую дату для правила.
func nextDay(date time.Time, now time.Time, fields []string) (string, error) {
	if len(fields) < 2 {
		return "", errors.New("is no interval")
	}
	interval, err := strconv.Atoi(fields[1])
	if err != nil {
		return "", err
	}
	if interval < 1 || interval > 400 {
		return "", errors.New("interval is out of range")
	}

	// Прибавляем интервал, пока дата не станет позже сегодняшней.
	for {
		date = date.AddDate(0, 0, interval)
		if AfterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

// nextWeek рассчитывает следующую дату для правила — по дням недели.
// Дни задаются числами.
func nextWeek(date time.Time, now time.Time, fields []string) (string, error) {
	if len(fields) < 2 {
		return "", errors.New("invalid week format")
	}

	// Разбираем список разрешённых дней недели в map для быстрой проверки.
	daysMap := map[int]bool{}
	for _, s := range strings.Split(fields[1], ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 1 || n > 7 {
			return "", errors.New("invalid weekday value")
		}
		daysMap[n] = true
	}

	// Перебираем дни по одному, пока не попадём на разрешённый день после сегодня.
	for {
		date = date.AddDate(0, 0, 1)
		// time.Weekday() возвращает 0 для воскресенья, а нам нужно 7.
		wd := int(date.Weekday())
		if wd == 0 {
			wd = 7
		}
		if daysMap[wd] && AfterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

// nextMonthDays рассчитывает следующую дату для правила m.
// Дни могут быть положительными (номер дня) или отрицательными
// (-1 — последний день месяца, -2 — предпоследний).
func nextMonthDays(date time.Time, now time.Time, fields []string) (string, error) {
	if len(fields) < 2 || len(fields) > 3 {
		return "", errors.New("invalid month rule format")
	}

	// Разбираем список дней месяца в map.
	daysMap := map[int]bool{}
	for _, s := range strings.Split(fields[1], ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n == 0 || n < -2 || n > 31 {
			return "", errors.New("invalid month day value")
		}
		daysMap[n] = true
	}

	// Если третий параметр задан, разбираем список месяцев в map.
	monthMap := map[int]bool{}
	if len(fields) == 3 {
		for _, s := range strings.Split(fields[2], ",") {
			m, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || m < 1 || m > 12 {
				return "", errors.New("invalid month value")
			}
			monthMap[m] = true
		}
	}

	// Ищем ближайший подходящий день, перебирая даты по одному дню.
	for {
		date = date.AddDate(0, 0, 1)
		// Если месяцы ограничены списком, пропускаем неподходящие.
		if len(monthMap) > 0 && !monthMap[int(date.Month())] {
			continue // Месяц не подходит — идём дальше.
		}
		// Дата подходит, если день совпал и она позже сегодняшней.
		if isMatchDay(date, daysMap) && AfterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

// isMatchDay проверяет, входит ли день даты d в список разрешённых дней месяца.
// Дополнительно учитываются значения -1 (последний день) и -2 (предпоследний день).
func isMatchDay(d time.Time, daysMap map[int]bool) bool {
	day := d.Day()

	// Последний день месяца: берём первый день следующего месяца и вычитаем один день.
	lastDay := time.Date(d.Year(), d.Month()+1, 1, 0, 0, 0, 0, d.Location()).AddDate(0, 0, -1).Day()

	// Совпадение по точному номеру дня.
	if daysMap[day] {
		return true
	}
	// Правило "-1" — последний день месяца.
	if daysMap[-1] && day == lastDay {
		return true
	}
	// Правило "-2" — предпоследний день месяца.
	if daysMap[-2] && day == lastDay-1 {
		return true
	}
	return false
}

// nextYear рассчитывает следующую дату для правила "y" — раз в год.
func nextYear(date time.Time, now time.Time, fields []string) (string, error) {
	if len(fields) != 1 {
		return "", errors.New("invalid year format: 'y' does not accept arguments")
	}
	// Прибавляем по одному году, пока дата не станет позже сегодняшней.
	for {
		date = date.AddDate(1, 0, 0)
		if AfterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}
