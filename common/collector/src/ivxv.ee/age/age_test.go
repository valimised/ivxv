package age

import (
	"fmt"
	"testing"
	"time"

	"ivxv.ee/errors"
)

func TestCheck(t *testing.T) {
	c := &Checker{
		get: func(dob string) (time.Time, error) {
			return time.Parse("2006-01-02", dob)
		},
		loc: time.UTC,
	}

	tests := []struct {
		dob   string
		now   string
		limit int
		ok    bool
	}{
		{"1999-01-01", "2016-01-01", 16, true},
		{"2000-01-01", "2016-01-01", 16, true},
		{"2000-01-02", "2016-02-01", 16, true},
		{"2000-02-28", "2016-02-29", 16, true},
		{"2000-02-29", "2016-02-29", 16, true},
		{"2000-03-01", "2016-02-29", 16, false},
		{"2000-03-02", "2016-03-01", 16, false},
		{"2001-01-01", "2016-01-01", 16, false},

		{"1996-02-29", "2017-02-28", 21, false},
		{"1996-02-29", "2017-03-01", 21, true},

		{"1970-01-02", "1970-01-01", 21, false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s in %s", test.dob, test.now), func(t *testing.T) {
			c.now = func() time.Time {
				n, err := time.Parse("2006-01-02", test.now)
				if err != nil {
					t.Fatal("failed to parse now:", err)
				}
				return n
			}
			c.limit = test.limit

			err := c.Check(test.dob)
			if test.ok && err != nil {
				t.Error("unexpected error:", err)
			}
			if !test.ok && errors.CausedBy(err, new(TooYoungError)) == nil {
				t.Errorf("unexpected error: got %v, want TooYoungError", err)
			}
		})
	}
}

func TestEstPIC(t *testing.T) {
	tests := []struct {
		identity string
		year     int
		month    time.Month
		day      int
	}{
		{"10001010000", 1800, time.January, 1},
		{"25202290000", 1852, time.February, 29},
		{"37001010000", 1970, time.January, 1},
		{"49108200000", 1991, time.August, 20},
		{"51703150000", 2017, time.March, 15},
		{"69912310000", 2099, time.December, 31},
	}

	for _, test := range tests {
		t.Run(test.identity, func(t *testing.T) {
			got, err := estpic(test.identity)
			if err != nil {
				t.Fatal("failed to get voter DOB:", err)
			}
			exp := time.Date(test.year, test.month, test.day, 0, 0, 0, 0, time.UTC)
			if !exp.Equal(got) {
				t.Errorf("unexpected DOB: got %s, want %s", got, exp)
			}
		})
	}
}
