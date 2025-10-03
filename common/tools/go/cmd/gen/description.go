package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	LOG_DESCRIPTION_FILENAME = "log_desc.go"
	LOG_DESCRIPTION_REGEX    = `%s\s*=\s*"(.*?)"`
)

type LogDescriptionParser interface {
	Parse(desc ast.Expr) (string, error)
}

type LogDescriptionString struct{}

func (lds *LogDescriptionString) Parse(expr ast.Expr) (string, error) {
	// Remove quotes from a string, to prevent comments like:
	// "This is quoted comment"
	desc, ok := expr.(*ast.BasicLit)
	if !ok {
		return "", fmt.Errorf("cannot cast expr to *ast.BasicLit")
	}

	return strings.Trim(desc.Value, "\""), nil
}

type LogDescriptionConstExpr struct {
	// LogDescAbsDirPath is a directory that contains log_desc.go file
	LogDescAbsDirPath string
}

func (ldce *LogDescriptionConstExpr) Parse(expr ast.Expr) (string, error) {
	desc, ok := expr.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("cannot cast expr to *ast.Ident")
	}

	f := new(os.File)
	defer f.Close()
	var err error

	f, err = os.Open(fmt.Sprintf("%s/%s", filepath.Dir(ldce.LogDescAbsDirPath), LOG_DESCRIPTION_FILENAME))
	defer f.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read a local %s file: %v\n", LOG_DESCRIPTION_FILENAME, err)
	}

	scanner := bufio.NewScanner(f)

	// Read log_desc.go file line by line until EOF or any non-EOF error encounter
	for scanner.Scan() {
		line := scanner.Text()

		pattern := fmt.Sprintf(LOG_DESCRIPTION_REGEX, desc.Name)
		match := regexp.MustCompile(pattern).FindStringSubmatch(strings.TrimSpace(line))
		if len(match) == 0 {
			continue
		}

		return match[1], nil
	}

	// Any errors encountered during Scan(), except EOF, will be registered in Err()
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read a line from a local %s file: %v\n", LOG_DESCRIPTION_FILENAME, err)
	}

	return "", fmt.Errorf("no log description found in a local %s file for a %s log record\n", LOG_DESCRIPTION_FILENAME, desc.Name)
}
