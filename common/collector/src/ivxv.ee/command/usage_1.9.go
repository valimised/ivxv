// +build !go1.10

package command

import "strings"

func indentUsage(usage string) string {
	return strings.Replace(usage, "\n", "\n    \t", -1)
}
