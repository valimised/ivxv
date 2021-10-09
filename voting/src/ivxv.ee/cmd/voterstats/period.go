package main

import (
	"strconv"
	"time"
)

// period is a labeled time period [from, to) in which to report statistics.
// Either time can also be the zero value in which case that end is unbound.
type period struct {
	label string
	from  time.Time
	to    time.Time
}

func (p period) contains(point time.Time) bool {
	return (p.from.IsZero() || !point.Before(p.from)) &&
		(p.to.IsZero() || point.Before(p.to))
}

// findPeriodLabels returns the labels of periods which contain point.
func findPeriodLabels(periods []period, point time.Time) []string {
	var labels []string
	for _, period := range periods {
		if period.contains(point) {
			labels = append(labels, period.label)
		}
	}
	return labels
}

// cumulativePeriods returns cumulative periods for days from start to stop.
//
// Each day represents the period from start up to the end of that day in
// location loc. So each period contains the cumulative statistics up to and
// including that day. The last period ends at stop instead.
//
// The days are labeled in reverse order: the last day (stop) gets the label
// "1", the second to last gets "2", etc.
func cumulativePeriods(start, stop time.Time, loc *time.Location) []period {
	// Always at least one period from start to stop.
	label := 1
	periods := []period{
		{label: strconv.Itoa(label), from: start, to: stop},
	}

	// Prepend periods from start to midnight for every day in between.
	midnight := previousMidnight(stop, loc)
	for midnight.After(start) {
		label++
		periods = append([]period{
			{label: strconv.Itoa(label), from: start, to: midnight},
		}, periods...)
		midnight = previousMidnight(midnight, loc)
	}

	return periods
}

func previousMidnight(t time.Time, loc *time.Location) time.Time {
	year, month, day := t.Add(-1).In(loc).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}
