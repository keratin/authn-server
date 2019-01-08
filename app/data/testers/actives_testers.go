package testers

import (
	"strconv"
	"testing"
	"time"

	"github.com/keratin/authn-server/app/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ActivesTesters = []func(*testing.T, data.Actives){
	testActivesTrack,
	testActivesActivesByDay,
	testActivesActivesByWeek,
	testActivesActivesByMonth,
}

func testActivesTrack(t *testing.T, actives data.Actives) {
	actives.Track(1)
	actives.Track(5)
	actives.Track(6)

	report, err := actives.ActivesByDay()
	require.NoError(t, err)
	if assert.Len(t, report, 1) {
		assert.Equal(t, []int{3}, mapVals(report))
	}
}

func testActivesActivesByDay(t *testing.T, actives data.Actives) {
	actives.Track(1)

	report, err := actives.ActivesByDay()
	require.NoError(t, err)
	if assert.Len(t, report, 1) {
		assert.Equal(t, map[string]int{time.Now().In(time.UTC).Format("2006-01-02"): 1}, report)
	}
}

func testActivesActivesByWeek(t *testing.T, actives data.Actives) {
	actives.Track(1)

	report, err := actives.ActivesByWeek()
	require.NoError(t, err)
	if assert.Len(t, report, 1) {
		y, w := time.Now().In(time.UTC).ISOWeek()
		label := strconv.Itoa(y) + "-W" + strconv.Itoa(w)
		assert.Equal(t, map[string]int{label: 1}, report)
	}
}

func testActivesActivesByMonth(t *testing.T, actives data.Actives) {
	actives.Track(1)

	report, err := actives.ActivesByMonth()
	require.NoError(t, err)
	if assert.Len(t, report, 1) {
		assert.Equal(t, map[string]int{time.Now().In(time.UTC).Format("2006-01"): 1}, report)
	}
}

func mapVals(m map[string]int) []int {
	vals := make([]int, 0, len(m))
	for _, x := range m {
		vals = append(vals, x)
	}
	return vals
}
