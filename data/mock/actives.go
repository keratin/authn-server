package mock

import (
	"strconv"
	"time"
)

type actives struct {
	byDay   map[string][]int
	byWeek  map[string][]int
	byMonth map[string][]int
}

func NewActives() *actives {
	return &actives{
		byDay:   make(map[string][]int, 0),
		byWeek:  make(map[string][]int, 0),
		byMonth: make(map[string][]int, 0),
	}
}

func (a *actives) Track(accountID int) error {
	t := time.Now().In(time.UTC)
	a.byDay = appendUniq(a.byDay, dayKey(t), accountID)
	a.byWeek = appendUniq(a.byWeek, weekKey(t), accountID)
	a.byMonth = appendUniq(a.byMonth, monthKey(t), accountID)

	return nil
}

func (a *actives) ActivesByDay() (map[string]int, error) {
	return countUniqs(a.byDay), nil
}

func (a *actives) ActivesByWeek() (map[string]int, error) {
	return countUniqs(a.byWeek), nil
}

func (a *actives) ActivesByMonth() (map[string]int, error) {
	return countUniqs(a.byMonth), nil
}

//-- UTIL

func countUniqs(data map[string][]int) map[string]int {
	counts := make(map[string]int, 0)

	for k, items := range data {
		counts[k] = len(items)
	}

	return counts
}

func appendUniq(data map[string][]int, key string, val int) map[string][]int {
	if data[key] != nil {
		for _, existing := range data[key] {
			if existing == val {
				return data
			}
		}
	} else {
		data[key] = make([]int, 0)
	}
	data[key] = append(data[key], val)
	return data
}

func dayKey(t time.Time) string {
	return t.Format("2006-01-02") // %Y-%m-%d
}

func weekKey(t time.Time) string {
	y, w := t.ISOWeek()
	return strconv.Itoa(y) + "-W" + strconv.Itoa(w) // %G-W%V
}

func monthKey(t time.Time) string {
	return t.Format("2006-01") // %Y-%m
}
