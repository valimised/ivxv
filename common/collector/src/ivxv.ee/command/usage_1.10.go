// +build go1.10

package command

func indentUsage(usage string) string {
	// Starting from Go 1.10 the flag.PrintDefaults function automatically
	// indents all usage texts after a newline.
	return usage
}
