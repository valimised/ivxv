package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// resolve resolves package import paths into packages. It emulates the "..."
// wilcard behavior of matchPackages in cmd/go/main.go.
func resolve(base string, paths []string) (pkgs []*build.Package, ok bool) {
	ok = true
	for _, p := range paths {
		// Check if local import and if so, make absolute so that
		// build.Import can determine the package import path.
		local := filepath.IsAbs(p) || build.IsLocalImport(filepath.ToSlash(p))
		path := p
		if local && !filepath.IsAbs(p) {
			path = filepath.Join(pwd, p)
		}
		// Search for packages matching the wildcards.
		if strings.Contains(p, "...") {
			// List of source directories to search.
			srcdirs := []string{""}
			if !local {
				srcdirs = build.Default.SrcDirs()
				path = filepath.FromSlash(p)
			}
			before := len(pkgs)
			for _, src := range srcdirs {
				expanded, expok := expand(path, src, base)
				pkgs = append(pkgs, expanded...)
				ok = ok && expok
			}
			if len(pkgs) == before {
				fmt.Fprintln(os.Stderr, "warning: no packages match", p)
			}
			continue
		}
		// Attempt to import non-wildcard package.
		imp, src := p, ""
		if local {
			imp, src = ".", path // Import package "." from abs(p).
		}
		pkg, err := build.Import(imp, src, 0)
		pkg.ImportPath = strings.TrimPrefix(imp, base)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: package %s: %v\n", p, err)
			ok = false
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	return
}

// expand expands "..." wildcards in the import path p into packages rooted at
// the given directory.
func expand(p, src, base string) (expanded []*build.Package, ok bool) {
	ok = true

	// Clean src and p, and append the non-wildcard prefix of p to src for
	// the walking root directory.
	if len(src) > 0 {
		src = filepath.Clean(src) + string(filepath.Separator)
	}
	p = filepath.Clean(p)
	dir := src + p[:strings.Index(p, "...")]

	// Compile a regular expression from the slash-separated path with
	// "..." wildcards.
	re := regexp.QuoteMeta(filepath.ToSlash(p))
	re = strings.ReplaceAll(re, `\.\.\.`, ".*")
	if strings.HasSuffix(re, "/.*") {
		// Special case: "foo" matches "foo/...".
		re = re[:len(re)-len("/.*")] + "(/.*)?"
	}
	match := regexp.MustCompile("^" + re + "$").MatchString

	// Walk the directory looking for paths that match.
	//
	//nolint:errcheck // Only returns nil.
	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		// Ignore faulty or non-directory files.
		if err != nil || !fi.IsDir() {
			return nil
		}

		// Ignore "testdata" and directories that start with a dot or underscore.
		name := fi.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata" {
			return filepath.SkipDir
		}

		// Convert the path to the format expected by match: remove the
		// src prefix and convert to slash-separated path.
		imp := filepath.ToSlash(path[len(src):])

		// Check if the path matches the pattern and attempt to import
		// the package, ignoring non-Go directories that match.
		if match(imp) {
			pkg, err := build.ImportDir(path, 0)

			pkg.ImportPath = strings.TrimPrefix(path, base)
			if err != nil {
				if _, nogo := err.(*build.NoGoError); nogo {
					return nil
				}
				fmt.Fprintf(os.Stderr, "error: matched package %s: %v\n", rel(path), err)
				ok = false
			}
			expanded = append(expanded, pkg)
		}
		return nil
	})
	return
}
