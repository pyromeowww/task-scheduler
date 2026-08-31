package api

import (
	"errors"
	"fmt"
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
		return nextWeek()
	case "m":
		return nextMonthDays()
	case "y":
		return nextYear()
	default:
		return "", errors.New("unknown format")
	}
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
