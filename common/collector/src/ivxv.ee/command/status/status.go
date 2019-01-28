/*
Package status implements displaying the status of a process.

If running on Linux and os.Stdout is a terminal then any updates to the status
line will be reflected on the same line. These lines are called "rewritable".
Otherwise any updates will print a new line.
*/
package status // import "ivxv.ee/command/status"

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"
)

// Line is the current status of the process displayed in a single line.
//
// All methods of a Line can also be called on nil without any effect. This
// simplifies optional quieting the status line.
type Line struct {
	tty    bool
	stack  []interface{}
	length int
	mu     sync.Mutex
}

// New initializes a new status line.
func New() *Line {
	return &Line{tty: isatty(os.Stdout.Fd())}
}

// Redraw redraws the current status line.
func (l *Line) Redraw() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.redraw()
}

func (l *Line) redraw() {
	l.setLine(fmt.Sprint(l.stack...))
}

func (l *Line) setLine(s string) {
	if !l.tty {
		fmt.Println(s)
		return
	}

	fmt.Print("\r", s)

	n := utf8.RuneCountInString(s)
	if tail := l.length - n; tail > 0 {
		fmt.Print(strings.Repeat(" ", tail), strings.Repeat("\b", tail))
	}

	l.length = n
}

// Static appends a static string to the status line.
func (l *Line) Static(s string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.push(updatable{s: s}) // Reuse updatable but with no way of updating.
}

// Update is the type of functions used to update an updatable string. Returns
// the previous value.
type Update func(updated string) string

// Updatable appends an updatable string to the status line. If redraw is true,
// then the status line will be redrawn after each call to Update. If l is nil,
// then Update is a noop.
func (l *Line) Updatable(initial string, redraw bool) Update {
	if l == nil {
		return func(string) string { return "" }
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	u := &updatable{l: l, s: initial, redraw: redraw}
	l.push(u)
	return u.update
}

type updatable struct {
	l      *Line
	redraw bool
	s      string
}

func (u *updatable) update(s string) string {
	u.l.mu.Lock()
	defer u.l.mu.Unlock()
	old := u.s
	u.s = s
	if u.redraw {
		u.l.redraw()
	}
	return old
}

func (u updatable) String() string { return u.s }

// Add is the type of function used to update Count and Percentage displays.
// Returns the new total value.
type Add func(count uint64) uint64

// Count appends a counter to the status line. If total is non-zero, then the
// count is displayed as "<current>/<total>". If redraw is true, then the
// status line will be redrawn after each call to Add. If l is nil, then Add is
// a noop, but the returned total will still be accurate.
func (l *Line) Count(total uint64, redraw bool) Add {
	if l == nil {
		return (&count{l: new(Line)}).add
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	p := &count{l: l, format: "%d", redraw: redraw}
	if total > 0 {
		tstr := fmt.Sprint(total)
		p.format = fmt.Sprintf("%%%dd/%s", len(tstr), tstr)
	}
	l.push(p)
	return p.add
}

type count struct {
	l       *Line
	redraw  bool
	format  string
	current uint64
}

func (c *count) add(count uint64) uint64 {
	c.l.mu.Lock()
	defer c.l.mu.Unlock()
	c.current += count
	if c.redraw {
		c.l.redraw()
	}
	return c.current
}

func (c count) String() string { return fmt.Sprintf(c.format, c.current) }

// Percent appends a percentage indicator to the status line. It calculates the
// percentage itself based on current and total values. If total is 0, then
// 100% is always displayed. If redraw is true, then the status line will be
// redrawn after each call to Add that changes the percentage value. If l is
// nil, then Add is a noop, but the returned total will still be accurate.
func (l *Line) Percent(total uint64, redraw bool) Add {
	if l == nil {
		return (&percent{l: new(Line)}).add
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	p := &percent{l: l, total: total, redraw: redraw}
	if p.total == 0 {
		p.percent = 100
	}
	p.step = p.total / 100
	p.recalc = p.step
	l.push(p)
	return p.add
}

type percent struct {
	l       *Line
	redraw  bool
	current uint64
	total   uint64
	percent uint64
	step    uint64 // Minimum count that must be added for percent to change.
	recalc  uint64 // Next value that percent must be recalculated on.
}

func (p *percent) add(count uint64) uint64 {
	p.l.mu.Lock()
	defer p.l.mu.Unlock()
	p.current += count
	if p.total > 0 && (p.step == 0 || p.current >= p.recalc) {
		oldpc := p.percent // To check if percent actually changed.
		p.percent = p.current * 100 / p.total
		p.recalc = p.percent*p.step + p.step
		if p.redraw && oldpc != p.percent {
			p.l.redraw()
		}
	}
	return p.current
}

func (p percent) String() string { return fmt.Sprintf("%3d%%", p.percent) }

func (l *Line) push(stringer fmt.Stringer) {
	l.stack = append(l.stack, stringer)
}

// Pop removes the last appended element from the status line.
func (l *Line) Pop() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.stack) > 0 {
		l.stack = l.stack[:len(l.stack)-1]
	}
}

// Hide hides a rewritable status line. This can be used to display a permanent
// line before redrawing the rewritable line.
//
// Hide does nothing if the line is not rewritable.
func (l *Line) Hide() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tty {
		l.setLine("")
	}
}

// Show unhides a rewritable status line.
//
// Show does nothing if the line is not rewritable.
func (l *Line) Show() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tty {
		l.redraw()
	}
}

// Keep keeps the last output of a rewritable line and resets the line. This
// can be used to keep the end result of a progress counter visible and start a
// new status line.
//
// If the line is not rewritable, then Keep only resets it.
func (l *Line) Keep() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tty {
		fmt.Println()
	}
	l.stack = nil
	l.length = 0
}
