/*
Package age determines a voter's age and enforces checks on it.

In some systems, voters are added to the list of voters if they are not of age
yet, but will become during election. In these cases, we must explicitly check
that a voter is of age at the moment they wish to vote.

This requires knowing the date of birth of the voter, determining of which can
differ based on how this information is stored and represented. Package age
also acts as a registry for methods to get a voter's date of birth.
*/
package age // import "ivxv.ee/age"

import (
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Method identifies the method used to get the voter's date of birth.
type Method string

// Enumeration of date of birth methods.
const (
	EstPIC Method = "estpic" // Estonian personal identification code. Built-in.
)

// Getter is the type of functions that get a voter's date of birth given their
// identity. Only the year, month, and day of the returned time will be used.
type Getter func(voter string) (dob time.Time, err error)

var (
	reglock  sync.RWMutex
	registry = map[Method]Getter{
		EstPIC: estpic,
	}
)

// Register registers a date of birth method. It is intended to be called from
// init functions of packages that implement date of birth methods.
func Register(m Method, g Getter) {
	reglock.Lock()
	defer reglock.Unlock()
	registry[m] = g
}

// Conf is the age checker configuration.
type Conf struct {
	Method   Method // The method used to get a voters date of birth.
	TimeZone string // The IANA Time Zone name where to calculate the age. UTC by default.
	Limit    uint64 // The minimum age of a voter who is allowed to vote. Age check disabled if zero.
}

// Checker gets the voter's age using a configured method and checks it againt
// the limit.
type Checker struct {
	get   Getter
	loc   *time.Location
	limit int
	now   func() time.Time // Returns the current time, can be replaced for testing.
}

// New initializes a new voter age checker with the provided configuration.
func New(c *Conf) (checker *Checker, err error) {
	reglock.RLock()
	defer reglock.RUnlock()
	g, ok := registry[c.Method]
	if !ok {
		return nil, UnlinkedMethodError{Method: c.Method}
	}
	loc, err := time.LoadLocation(c.TimeZone)
	if err != nil {
		return nil, LocationError{Err: err}
	}
	return &Checker{g, loc, int(c.Limit), time.Now}, nil
}

// Check gets the voters's date of birth, calculates their age in the
// configured location, and ensures it is at least the configured lower limit.
func (c *Checker) Check(voter string) (err error) {
	if c.limit == 0 {
		return nil
	}

	dob, err := c.get(voter)
	if err != nil {
		return GetAgeError{Err: err}
	}
	now := c.now().In(c.loc)       // Get current time in configured zone.
	age := now.Year() - dob.Year() // Years since date of birth.
	if now.Month() < dob.Month() || now.Month() == dob.Month() && now.Day() < dob.Day() {
		age-- // If before the date, then subtract one year.
	}
	if age < c.limit {
		err = TooYoungError{Age: age, Limit: c.limit}
	}
	return
}

var estpicre = regexp.MustCompile(`^[1-6][0-9]{2}(0[1-9]|1[012])(0[1-9]|[12][0-9]|3[01])[0-9]{4}$`)

// estpic checks if the voter's identity is an Estonian personal identification
// code, which has the voter's date of birth encoded into it, and returns the
// voter's age.
func estpic(voter string) (dob time.Time, err error) {
	if !estpicre.MatchString(voter) {
		return dob, EstPICInvalidFormat{}
	}
	dobs := make([]byte, 0, 8)   // yyyymmdd.
	pre := 18 + (voter[0]-'1')/2 // '1' and '2' = 18xx, '3' and '4' = 19xx, ...
	dobs = strconv.AppendUint(dobs, uint64(pre), 10)[:8]
	copy(dobs[2:], voter[1:7]) // Copy yymmdd from pic.
	if dob, err = time.Parse("20060102", string(dobs)); err != nil {
		err = EstPICParseDOBError{Err: err}
	}
	return
}
