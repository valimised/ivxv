package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func mustRFC3339(t *testing.T, value string) time.Time {
	tm, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func assertPeriods(t *testing.T, periods []period, expected string) {
	var buf bytes.Buffer
	for _, period := range periods {
		fmt.Fprintf(&buf, "%s [%s, %s)\n",
			period.label,
			period.from.Format(time.RFC3339),
			period.to.Format(time.RFC3339))
	}

	if got := buf.String(); got != expected {
		t.Errorf("unexpected periods:\ngot:\n%swant:\n%s", got, expected)
	}
}

func TestCumulativePeriods(t *testing.T) {
	// Include DST transition in the period.
	start := mustRFC3339(t, "2021-03-27T09:00:00+02:00")
	stop := mustRFC3339(t, "2021-03-29T20:15:00+03:00")
	loc, err := time.LoadLocation("Europe/Tallinn")
	if err != nil {
		t.Fatal(err)
	}

	periods := cumulativePeriods(start, stop, loc)

	assertPeriods(t, periods, ""+
		"3 [2021-03-27T09:00:00+02:00, 2021-03-28T00:00:00+02:00)\n"+
		"2 [2021-03-27T09:00:00+02:00, 2021-03-29T00:00:00+03:00)\n"+
		"1 [2021-03-27T09:00:00+02:00, 2021-03-29T20:15:00+03:00)\n")
}
