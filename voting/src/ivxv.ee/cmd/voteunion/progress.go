package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type progress struct {
	stack []string
}

func (p *progress) push(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	fmt.Print(s)
	p.stack = append(p.stack, s)
}

func (p *progress) pop() {
	p.delete(p.stack[len(p.stack)-1])
	p.stack = p.stack[:len(p.stack)-1]
}

func (p progress) delete(s string) {
	fmt.Print(strings.Repeat("\b \b", utf8.RuneCountInString(s)))
}

func (p progress) hide() {
	for _, s := range p.stack {
		p.delete(s)
	}
}

func (p progress) show() {
	for _, s := range p.stack {
		fmt.Print(s)
	}
}
