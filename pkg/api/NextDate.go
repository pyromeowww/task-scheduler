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

/*
now — текущая дата.
dstart — дата, когда задача была создана или выполнена в прошлый раз.
repeat — правило (инструкция, как двигать дату вперед).
Функция возвращает следующую дату в формате 20060102 и ошибку.
*/
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat rule cannot be empty")
	}

	// Парсим начальную дату
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid dstart format: %w", err)
	}

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

func NextDateHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	dateStr := req.URL.Query().Get("date")
	repeatStr := req.URL.Query().Get("repeat")

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(nextDate))
}

func afterNow(date time.Time, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	return time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC).After(time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC))
}

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

	for {
		date = date.AddDate(0, 0, interval)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

func nextWeek(date time.Time, now time.Time, fields []string) (string, error) {
	if len(fields) < 2 {
		return "", errors.New("invalid week format")
	}

	// Парсим разрешённые дни недели в map для быстрой проверки
	daysMap := map[int]bool{}
	for _, s := range strings.Split(fields[1], ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 1 || n > 7 {
			return "", errors.New("invalid weekday value")
		}
		daysMap[n] = true
	}
	for {
		date = date.AddDate(0, 0, 1)
		wd := int(date.Weekday())
		if wd == 0 {
			wd = 7
		}
		if daysMap[wd] && afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

func nextMonthDays(date time.Time, now time.Time, fields []string) (string, error) {
	if len(fields) < 2 || len(fields) > 3 {
		return "", errors.New("invalid month rule format")
	}

	// Парсим дни месяца в map
	daysMap := map[int]bool{}
	for _, s := range strings.Split(fields[1], ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n == 0 || n < -2 || n > 31 {
			return "", errors.New("invalid month day value")
		}
		daysMap[n] = true
	}

	// Парсим месяцы, если они указаны
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

	// Поиск ближащей даты
	for {
		date = date.AddDate(0, 0, 1)
		// Если месяцы были ограничены (len > 0), проверяем текущий месяц
		if len(monthMap) > 0 && !monthMap[int(date.Month())] {
			continue // Месяц не подходит, идем дальше
		}
		// Проверяем, подходит ли день месяца
		if isMatchDay(date, daysMap) && afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

// isMatchDay — вспомогательная функция для проверки дня месяца
func isMatchDay(d time.Time, daysMap map[int]bool) bool {
	day := d.Day() // Текущий день

	// Находим последний день текущего месяца
	lastDay := time.Date(d.Year(), d.Month()+1, 1, 0, 0, 0, 0, d.Location()).AddDate(0, 0, -1).Day()

	if daysMap[day] {
		return true
	}
	// Проверка на последний день месяца
	if daysMap[-1] && day == lastDay {
		return true
	}
	// Проверка на предпоследний день месяца
	if daysMap[-2] && day == lastDay-1 {
		return true
	}
	return false
}

func nextYear(date time.Time, now time.Time, fields []string) (string, error) {
	if len(fields) != 1 {
		return "", errors.New("invalid year format: 'y' does not accept arguments")
	}
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}
