package lib

import (
	"math"
	"time"
)

// epochIntervalTicker is a ticker that emits monotonic numbers indicating how many multiples of a
// configured interval have passed since the Unix Epoch. No matter when the ticker is started, it
// will emit the same values at the same time (within a second) as any other ticker that has been
// configured with the same interval.
//
// Except in the case of a laptop going to sleep. When the laptop wakes up, this ticker will emit
// an event shortly (within a minute) for the last missed interval number. This behavior exists
// to address a common case during development of apps that use AuthN.
type epochIntervalTicker struct {
	interval             time.Duration
	C                    chan int
	lastReportedInterval int
}

// EpochIntervalTick returns only the channel from an epochIntervalTicker
func EpochIntervalTick(interval time.Duration) <-chan int {
	ticker := &epochIntervalTicker{
		interval:             interval,
		C:                    make(chan int),
		lastReportedInterval: 0,
	}
	ticker.lastReportedInterval = ticker.currentInterval()

	go ticker.tick()

	return ticker.C
}

func (t *epochIntervalTicker) currentInterval() int {
	return int(time.Now().Unix() / int64(t.interval/time.Second))
}

func (t *epochIntervalTicker) sleep() {
	elapsed := time.Duration(
		time.Now().Unix()%int64(t.interval/time.Second),
	) * time.Second

	alarm := time.Duration(
		math.Min(30, float64((t.interval-elapsed)/time.Second)),
	) * time.Second

	time.Sleep(alarm)
}

func (t *epochIntervalTicker) tick() {
	for {
		t.sleep()
		current := t.currentInterval()
		if current > t.lastReportedInterval {
			t.lastReportedInterval = current
			t.C <- current
		}
	}
}
