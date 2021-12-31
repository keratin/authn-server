package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisPrefix = "actives:"

type actives struct {
	client  *redis.Client
	tz      *time.Location
	days    int
	dayTTL  time.Duration
	weeks   int
	weekTTL time.Duration
	months  int
}

func NewActives(client *redis.Client, tz *time.Location, days int, weeks int, months int) *actives {
	return &actives{
		client:  client,
		tz:      tz,
		days:    days,
		dayTTL:  time.Duration(days*24) * time.Hour,
		weeks:   weeks,
		weekTTL: time.Duration(weeks*24*7) * time.Hour,
		months:  months,
	}
}

func (a *actives) Track(accountID int) error {
	ctx := context.TODO()
	t := time.Now().In(a.tz)
	pipe := a.client.Pipeline()

	// increment daily
	dayKey := redisPrefix + dayKey(t)
	pipe.PFAdd(ctx, dayKey, accountID)
	pipe.Expire(ctx, dayKey, a.dayTTL)

	// increment weekly
	weekKey := redisPrefix + weekKey(t)
	pipe.PFAdd(ctx, weekKey, accountID)
	pipe.Expire(ctx, weekKey, a.weekTTL)

	// increment monthly
	monthKey := redisPrefix + monthKey(t)
	pipe.PFAdd(ctx, monthKey, accountID)

	_, err := pipe.Exec(ctx)
	return err
}

func (a *actives) ActivesByDay() (map[string]int, error) {
	now := time.Now().In(a.tz)

	days := make([]string, a.days)
	for i := range days {
		days[i] = dayKey(now.Add(time.Duration(i*-24) * time.Hour))
	}

	return a.report(days)
}

func (a *actives) ActivesByWeek() (map[string]int, error) {
	now := time.Now().In(a.tz)

	weeks := make([]string, a.weeks)
	for i := range weeks {
		weeks[i] = weekKey(now.AddDate(0, 0, -7*i))
	}

	return a.report(weeks)
}

func (a *actives) ActivesByMonth() (map[string]int, error) {
	now := time.Now().In(a.tz)

	months := make([]string, a.months)
	for i := range months {
		months[i] = monthKey(now.AddDate(0, -1*i, 1-now.Day()))
	}

	return a.report(months)
}

func (a *actives) report(keys []string) (map[string]int, error) {
	pipe := a.client.Pipeline()

	// construct requests
	metrics := make([]metric, len(keys))
	for i := range metrics {
		metrics[i] = newMetric(pipe, keys[i])
	}

	// to redis
	_, err := pipe.Exec(context.TODO())
	if err != nil {
		return nil, err
	}

	// extract values
	report := make(map[string]int, len(keys))
	for _, m := range trim(metrics) {
		report[m.label] = m.val()
	}

	return report, nil
}

//-- METRIC

type metric struct {
	label  string
	future *redis.IntCmd
}

func newMetric(pipe redis.Pipeliner, key string) metric {
	return metric{key, pipe.PFCount(context.TODO(), redisPrefix+key)}
}

func (m metric) val() int {
	return int(m.future.Val())
}

//-- UTIL

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

// takes an ordered list of metrics and slices out the oldest entries that are zeroes
func trim(metrics []metric) []metric {
	lastNonZeroIndex := -1
	for i, metric := range metrics {
		if metric.future.Val() > 0 {
			lastNonZeroIndex = i
		}
	}
	return metrics[:lastNonZeroIndex+1]
}
